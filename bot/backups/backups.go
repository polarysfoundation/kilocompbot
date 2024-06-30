package backups

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/polarysfoundation/kilocompbot/bot/notificator"
	"github.com/polarysfoundation/kilocompbot/bot/promotions"
	"github.com/polarysfoundation/kilocompbot/core"
	"github.com/polarysfoundation/kilocompbot/database"
	"github.com/polarysfoundation/kilocompbot/groups"
)

type Backup struct {
	Group      *groups.Groups
	Comps      *core.Competition
	Promo      *promotions.Params
	DB         *sql.DB
	Lastupdate int64
}

func InitBackup(db *sql.DB, groups *groups.Groups, comps *core.Competition, promo *promotions.Params) *Backup {
	lastUpdate := time.Now().Unix()

	return &Backup{
		Group:      groups,
		Comps:      comps,
		Promo:      promo,
		DB:         db,
		Lastupdate: lastUpdate,
	}
}

func (b *Backup) HandleBackup() {
	ticker := time.NewTicker(20 * time.Minute)

	go func() {
		for {
			select {
			case <-ticker.C:
				b.HandleSync()
			case sig := <-signalChan():
				log.Printf("Received signal: %v", sig)
				b.HandleSync()
				os.Exit(0)
			}
		}
	}()
}

func (b *Backup) HandleSync() {
	log.Print("creating backup...")
	b.storeGroups()
	b.storePromo()
	b.storePurchase()
	b.storeSale()
}

func (b *Backup) LoadData(tickers *notificator.Groups) {
	log.Print("creating stored instances...")
	b.loadGroups(tickers)
	b.loadPurchase()
	b.loadSales()
	b.loadPromo()
}

func (b *Backup) loadPromo() {
	promo, err := database.GetPromos(b.DB, "promo")
	if err != nil {
		log.Printf("no se pudo obtener la promo: %v", err)
		return
	}

	b.Promo.UpdateAdName(promo.AdName)
	b.Promo.UpdateButtonLink(promo.ButtonLink)
	b.Promo.UpdateButtonName(promo.ButtonName)
	b.Promo.Media = promo.Media
	b.loadTimestamp()
}

func (b *Backup) loadTimestamp() {
	for _, group := range b.Group.ActiveGroups {
		if group.CompActive {
			timestamp, err := database.GetEndTime(b.DB, group.ID)
			if err != nil {
				log.Printf("no se pudo obtener la fecha de culminacion para el grupo %s, por el siguiente error, %v", group.ID, err)
				continue
			}
			err = b.Comps.NewTimestamp(group.ID, timestamp)
			if err != nil {
				log.Printf("hubo error mientras se añadia la fecha de culminacion para el grupo %s", group.ID)
				continue
			}
		}
	}
}

func (b *Backup) loadGroups(tickers *notificator.Groups) {
	groups, err := database.GetGroups(b.DB)
	if err != nil {
		log.Printf("Error obteniendo los grupos: %v", err)
		return
	}

	for _, group := range groups {
		b.Group.ActiveGroups[group.ID] = group
		if group.CompActive {
			err := tickers.AddNewTicker(group.ID)
			if err != nil {
				log.Printf("error agregando tickers %v", err)
				continue
			}
		}
	}

	if len(groups) == 0 {
		log.Println("No se encontraron grupos activos en la base de datos")
	} else {
		log.Printf("Se cargaron %d grupos activos", len(groups))
	}
}

func (b *Backup) loadPurchase() {
	for id := range b.Group.ActiveGroups {
		purchases, err := database.GetPurchase(b.DB, id)
		if err != nil {
			log.Printf("error obteniendo las compras para el grupo %s: %v", id, err)
			continue // Salta a la siguiente iteración en lugar de terminar la función
		}

		if len(purchases) > 0 {
			for _, purchase := range purchases {
				if !b.Comps.CompExist(id) {
					err := b.Comps.NewComp(id)
					if err != nil {
						log.Printf("no se pudo crear las nuevas comps, error: %v", err)
						continue
					}
				}
				err = b.Comps.Comps[id].StorePurchase(purchase)
				if err != nil {
					log.Printf("error guardando compra de la base de datos para el grupo %s: %v", id, err)
					// Puedes decidir si quieres continuar con la siguiente compra o saltar al siguiente grupo
					continue
				}
				// Elimina este break si quieres procesar todas las compras
				// break
			}
		}
	}
}

func (b *Backup) loadSales() {
	for id := range b.Group.ActiveGroups {
		sales, err := database.GetSale(b.DB, id)
		if err != nil {
			log.Printf("error obteniendo las compras para el grupo %s: %v", id, err)
			continue // Salta a la siguiente iteración en lugar de terminar la función
		}

		if len(sales) > 0 {
			for _, sale := range sales {
				err := b.Comps.BlackList[id].StoreSale(sale)
				if err != nil {
					log.Printf("error guardando venta de la base de datos para el grupo %s: %v", id, err)
					// Puedes decidir si quieres continuar con la siguiente compra o saltar al siguiente grupo
					continue
				}
				// Elimina este break si quieres procesar todas las compras
				// break
			}
		}
	}
}

func (b *Backup) storeGroups() {
	for _, group := range b.Group.ActiveGroups {
		err := database.WriteGroups(b.DB, group.ID, group.CompActive, group.JettonAddress, group.Dedust, group.StonFi, group.Emoji)
		if err != nil {
			log.Printf("error guardando grupo %s", group.ID)
			log.Println("error:", err)
			return
		}
		break
	}
}

func (b *Backup) storePurchase() {
	for id, comp := range b.Comps.Comps {
		for _, purchase := range comp.Purchase {
			err := database.WritePurchases(b.DB, id, purchase.JettonAddress, purchase.JettonName, purchase.JettonSymbol, purchase.JettonDecimals, purchase.Buyer, purchase.Ton, purchase.Token)
			if err != nil {
				log.Printf("error guardando compra para el grupo %s", id)
				log.Println("error:", err)
				return
			}
			break
		}
	}
}

func (b *Backup) storeSale() {
	for id, comp := range b.Comps.BlackList {
		for _, sale := range comp.Sale {
			err := database.WritePurchases(b.DB, id, sale.JettonAddress, sale.JettonName, sale.JettonSymbol, sale.JettonDecimals, sale.Seller, sale.Ton, sale.Token)
			if err != nil {
				log.Printf("error guardando venta para el grupo %s", id)
				log.Println("error:", err)
				return
			}
			break
		}
	}
}

func (b *Backup) storePromo() {
	err := database.WritePromo(b.DB, "promo", b.Promo.AdName, b.Promo.ButtonName, b.Promo.ButtonLink, b.Promo.Media)
	if err != nil {
		log.Print("error guardando la promo")
		log.Println("error:", err)
		return
	}
}

func signalChan() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}
