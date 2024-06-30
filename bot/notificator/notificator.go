package notificator

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strconv"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/polarysfoundation/kilocompbot/bot/promotions"
	"github.com/polarysfoundation/kilocompbot/core"
	"github.com/polarysfoundation/kilocompbot/database"
	"github.com/polarysfoundation/kilocompbot/groups"
	"github.com/polarysfoundation/kilocompbot/indexer"
)

var (
	errIDAlreadyExist    = errors.New("error: el grupo ya existe")
	errIDNotExist        = errors.New("error: el grupo no existe")
	errEventAlreadyExist = errors.New("error: instancia de evento ya existe")
	errEventNotExist     = errors.New("error: instancia de evento no existe")
)

const (
	compEnded = "The competition is over."
)

type Events struct {
	events map[string]*indexer.Events
	mutex  sync.RWMutex
}

func InitEvents() *Events {
	return &Events{
		events: make(map[string]*indexer.Events),
	}
}

func (e *Events) AddNewInstance(id string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	_, exist := e.events[id]
	if exist {
		return errEventAlreadyExist
	}

	newEvent := indexer.Init()

	e.events[id] = newEvent

	return nil
}

func (e *Events) RemoveInstance(id string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	_, exist := e.events[id]
	if !exist {
		return errEventNotExist
	}

	delete(e.events, id)

	return nil
}

func (e *Events) GetEvent(id string) *indexer.Events {
	return e.events[id]
}

type Groups struct {
	ID     []string
	Ticker map[string]*time.Ticker
	Groups *groups.Groups
	DB     *sql.DB
	BotAPI *tgbotapi.BotAPI

	events *Events
	comps  *core.Competition
	api    string

	promotions *promotions.Params

	mutex sync.RWMutex
}

func Init(bot *tgbotapi.BotAPI, groups *groups.Groups, comps *core.Competition, events *Events, api string, params *promotions.Params, db *sql.DB) *Groups {
	return &Groups{
		ID:         make([]string, 0),
		Ticker:     make(map[string]*time.Ticker),
		Groups:     groups,
		DB:         db,
		BotAPI:     bot,
		events:     events,
		comps:      comps,
		api:        api,
		promotions: params,
	}
}

func (g *Groups) AddNewTicker(id string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	log.Printf("new event added for: %s", id)

	for _, exist := range g.ID {
		if exist == id {
			return errIDAlreadyExist
		}
	}

	g.ID = append(g.ID, id)

	return nil
}

func (g *Groups) RemoveTicker(id string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	for i, exist := range g.ID {
		if exist == id {
			g.ID = append(g.ID[:i], g.ID[i+1:]...)
		}
	}
	return errIDNotExist
}

func (g *Groups) HandleUpdate() {
	// Usar un solo mutex en lugar de dos para simplificar
	var tickerMutex sync.Mutex

	go func() {
		for {
			time.Sleep(5 * time.Second)

			tickerMutex.Lock()

			// Iniciar tickers para nuevos grupos
			for _, chatID := range g.ID {
				log.Printf("event found for: %s", chatID)
				if g.Groups.CompStatus(chatID) {
					if _, exist := g.Ticker[chatID]; !exist {
						ticker := time.NewTicker(5 * time.Second)
						g.Ticker[chatID] = ticker
						log.Printf("Iniciando ticker para el grupo %s", chatID)

						go func(chatID string, ticker *time.Ticker) {
							defer func() {
								ticker.Stop()
								tickerMutex.Lock()
								delete(g.Ticker, chatID)
								tickerMutex.Unlock()
								log.Printf("Ticker para el grupo %s detenido", chatID)
							}()

							err := g.events.AddNewInstance(chatID)
							if err != nil {
								log.Print("error creando una instancia de evento")
								return
							}

							for range ticker.C {
								g.handleEvent(chatID)
							}
						}(chatID, ticker)
					}
				}
			}

			// Detener tickers para los grupos que ya no están en g.ID
			for chatID, ticker := range g.Ticker {
				found := false
				for _, group := range g.ID {
					if group == chatID {
						found = true
						break
					}
				}
				if !found {
					ticker.Stop()
					delete(g.Ticker, chatID)
					log.Printf("Ticker para el grupo %s detenido", chatID)
				}
			}
			tickerMutex.Unlock()
		}
	}()
}

func (g *Groups) handleEvent(chatID string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	chatIDInt, _ := strconv.Atoi(chatID)

	pools := make([]string, 0)
	group := g.Groups.ActiveGroups[chatID]

	if group.Dedust != "" {
		pools = append(pools, group.Dedust)
	}

	if group.StonFi != "" {
		pools = append(pools, group.StonFi)
	}

	for _, pool := range pools {
		event := g.events.GetEvent(chatID)
		tx, err := event.GetLastEvent(pool, g.api)
		if err != nil {
			log.Printf("no se pudo obtener los eventos para la direccion %s", pool)
			log.Printf("error: %v", err)
			continue
		}

		if tx.SellOrder {
			if !g.comps.BlackListExist(chatID) {
				err = g.comps.NewBlacklist(chatID)
				if err != nil {
					log.Printf("error creando una blacklist para el grupo %s", chatID)
					return
				}
			}

			blacklist, err := g.comps.GetBlacklist(chatID)
			if err != nil {
				log.Printf("error obteniendo la blacklist del grupo %s", chatID)
				return
			}

			err = blacklist.AddSale(tx)
			if err != nil {
				log.Printf("error mientras se creaba una nueva venta")
				return
			}

			if g.comps.CompExist(chatID) {
				buy, err := g.comps.GetComp(chatID)
				if err != nil {
					log.Printf("no se pudo obtener la blacklist del grupo %s", chatID)
					return
				}

				hash, err := blacklist.GetSaleHash(tx)
				if err != nil {
					log.Print("error obteniendo el hash de la venta")
					return
				}

				sale, err := blacklist.GetSale(hash)
				if err != nil {
					log.Printf("error obteniendo la venta con el hash %s ", hash)
					return
				}

				for buyHash, purchase := range buy.Purchase {
					if sale.Seller == purchase.Buyer {
						err := buy.RemovePurchase(buyHash)
						if err != nil {
							log.Printf("no se pudo remover la compra con hash: %s", hash)
							return
						}
						break
					}
				}
			}
		}

		if tx.BuyOrder {
			if !g.comps.CompExist(chatID) {
				err = g.comps.NewComp(chatID)
				if err != nil {
					log.Printf("error creando una comp para el grupo %s", chatID)
					return
				}
			}

			buy, err := g.comps.GetComp(chatID)
			if err != nil {
				log.Printf("error obteniendo la comp del grupo %s", chatID)
				return
			}

			order, err := buy.AddPurchase(tx)
			if err != nil {
				log.Printf("error mientras se creaba una nueva compra")
				return
			}

			if g.comps.BlackListExist(chatID) {
				blacklist, err := g.comps.GetBlacklist(chatID)
				if err != nil {
					log.Printf("no se pudo obtener la blacklist del grupo %s", chatID)
					return
				}

				hash, err := buy.GetPurchaseHash(tx)
				if err != nil {
					log.Printf("no se pudo obtener el hash de la compra de la wallet %s", tx.Wallet)
					return
				}

				for _, sale := range blacklist.Sale {
					if order.Buyer == sale.Seller {
						err := buy.RemovePurchase(hash)
						if err != nil {
							log.Printf("no se pudo remover la compra con hash: %s", hash)
							return
						}
						break
					}
				}
			}

			compList := buy.GetCompList()

			msg := g.generateMessage(order, chatID, compList)
			keyboardMarkup := keyboardMarkup(g.promotions.ButtonName, g.promotions.ButtonLink)
			g.newNotification(msg, int64(chatIDInt), g.promotions.Media, keyboardMarkup)
		}

		if g.comps.IsEnded(chatID) {
			if _, exist := g.Ticker[chatID]; !exist {
				log.Printf("el ticker para el grupo %s, no existe", chatID)
				return
			}

			_, err := database.RemoveCompData(g.DB, chatID)
			if err != nil {
				log.Printf("no se pudo remover la competencia para el grupo %s", chatID)
				return
			}

			_, err = database.RemoveEndTimeData(g.DB, chatID)
			if err != nil {
				log.Printf("no se pudo remover el tiempo para el grupo %s", chatID)
				return
			}

			_, err = database.RemoveSaleData(g.DB, chatID)
			if err != nil {
				log.Printf("no se pudo remover la venta para el grupo %s", chatID)
				return
			}

			delete(g.Ticker, chatID)

			err = g.events.RemoveInstance(chatID)
			if err != nil {
				log.Printf("no se pudo eliminar los eventos para el grupo %s", chatID)
				return
			}

			err = g.comps.RemoveTimestampActive(chatID)
			if err != nil {
				log.Printf("no se pudo eliminar el timestamp para el grupo %s", chatID)
				return
			}

			err = g.Groups.UpdateCompStatus(chatID, false)
			if err != nil {
				log.Printf("no se pudo eliminar el timestamp para el grupo %s", chatID)
				return
			}

			g.send(int64(chatIDInt), compEnded)
		}
	}

}

func (g *Groups) newNotification(text string, chatID int64, media string, markup *tgbotapi.InlineKeyboardMarkup) {
	video := media
	msg := tgbotapi.NewVideoUpload(chatID, video)
	msg.Caption = text
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = markup
	_, err := g.BotAPI.Send(msg)
	if err != nil {
		log.Println("Error al enviar mensaje:", err)
	}
}

func (g *Groups) generateMessage(tx *core.Purchase, id string, compList []*core.Purchase) string {
	header := fmt.Sprintf("🚨*%s New Buy*🚨\n\n", tx.JettonName)

	emojis := g.calcularCantidadEmoji(int(tx.Ton.Int64()), id)

	buyerIndex := 0

	for i, comp := range compList {
		if comp.Buyer == tx.Buyer {
			buyerIndex = i
		}
	}

	spent := fmt.Sprintf("\n\n💰Spent: %d *TON*\n", tx.Ton)
	got := fmt.Sprintf("🧳Bought: %s *%s*\n", formatWithCommas(tx.Token.Int64()), tx.JettonSymbol)
	compspot := fmt.Sprintf("📊Competition Spot: %d\n", buyerIndex)
	walletEnc := url.QueryEscape(tx.Buyer)
	lenWallet := len(tx.Buyer)
	wallet := fmt.Sprintf("💎Wallet: [%s...%s](https://tonviewer.com/%s/)\n", tx.Buyer[:6], tx.Buyer[lenWallet-6:], walletEnc)

	leadingBoard := "\n*Leading Buys:*\n"

	var place1, place2, place3 string
	if len(compList) > 0 {
		place1 = fmt.Sprintf("🥇%d *TON*  -  [%s...%s](https://tonviewer.com/%s/)\n", compList[0].Ton, compList[0].Buyer[:6], compList[0].Buyer[lenWallet-6:], url.QueryEscape(compList[0].Buyer))
	}
	if len(compList) > 1 {
		place2 = fmt.Sprintf("🥈%d *TON*  -  [%s...%s](https://tonviewer.com/%s/)\n", compList[1].Ton, compList[1].Buyer[:6], compList[1].Buyer[lenWallet-6:], url.QueryEscape(compList[1].Buyer))
	}
	if len(compList) > 2 {
		place3 = fmt.Sprintf("🥉%d *TON*  -  [%s...%s](https://tonviewer.com/%s/)\n", compList[2].Ton, compList[2].Buyer[:6], compList[2].Buyer[lenWallet-6:], url.QueryEscape(compList[2].Buyer))
	}

	endIn := timeUntilEnd(g.comps.Timestamp[id])

	foot := fmt.Sprintf("Buy competition end at %s", endIn)

	concatened := header + emojis + spent + got + compspot + wallet + leadingBoard + place1 + place2 + place3 + foot + "\n" + "\n" + g.promotions.AdName + "\n" + "\n"

	return concatened
}

func (g *Groups) calcularCantidadEmoji(amount int, id string) string {
	var cantidadBase int

	if amount < 10 {
		cantidadBase = rand.Intn(11) + 5 // Generar número aleatorio entre 5 y 15
	} else if amount < 50 {
		cantidadBase = rand.Intn(36) + 15 // Generar número aleatorio entre 15 y 50
	} else if amount < 100 {
		cantidadBase = rand.Intn(50) + 20 // Generar número aleatorio entre 20 y 70
	} else {
		cantidadBase = rand.Intn(600) + 25 // Generar número aleatorio entre 25 y 625
	}

	defaultEmoji := "🦾"

	var emojis string

	group, err := g.Groups.GetDataGroup(id)
	if err != nil {
		log.Printf("error obteniendo los datos del grupo %s", id)
		return ""
	}

	if group.Emoji != "" {
		emojis = repeatEmoji(group.Emoji, cantidadBase)
	} else {
		emojis = repeatEmoji(defaultEmoji, cantidadBase)
	}

	return emojis
}

func repeatEmoji(emoji string, times int) string {
	var result string
	for i := 0; i < times; i++ {
		result += emoji
	}
	return result
}

func formatWithCommas(num int64) string {
	// Convertir el int64 a una cadena sin comas
	str := strconv.FormatInt(num, 10)

	// Calcular la longitud de la cadena
	n := len(str)

	// Construir la cadena con comas
	var formatted string
	for i, char := range str {
		// Insertar una coma si la posición es divisible por 3 y no es el primer dígito
		if i > 0 && (n-i)%3 == 0 {
			formatted += ","
		}
		formatted += string(char)
	}

	return formatted
}

func timeUntilEnd(timestamp int64) string {
	endTime := time.Unix(timestamp, 0)
	duration := time.Until(endTime)

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	return fmt.Sprintf("*%d hours: %d minutes: %d sec*", hours, minutes, seconds)
}

func keyboardMarkup(buttonContext string, buttonContent string) *tgbotapi.InlineKeyboardMarkup {
	urlButton := tgbotapi.NewInlineKeyboardButtonURL(buttonContext, buttonContent)
	row := []tgbotapi.InlineKeyboardButton{urlButton}
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(row)
	return &inlineKeyboard
}

func (p *Groups) send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := p.BotAPI.Send(msg)
	if err != nil {
		log.Println("Error al enviar mensaje:", err)
	}
}
