package commands

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/polarysfoundation/kilocompbot/bot/notificator"
	"github.com/polarysfoundation/kilocompbot/bot/promotions"
	"github.com/polarysfoundation/kilocompbot/core"
	"github.com/polarysfoundation/kilocompbot/getters"
	"github.com/polarysfoundation/kilocompbot/groups"
)

const (
	start        = "startkilo"
	addtoken     = "addtoken"
	removetoken  = "removetoken"
	startnewcomp = "startnewcomp"
	stopcomp     = "stopcomp"
	removebuyer  = "removebuyer"
	addemoji     = "addemoji"
	list         = "list"
)

const (
	invalidJetton           = "Invalid jetton, please check and try again"
	onlyAdmin               = "This command only can be used by admins"
	onlyGroups              = "Please first add me to a group and make me an administrator"
	groupAdded              = "Group successfully added"
	groupAlreadyExist       = "This group has already been added and started successfully, try another command. "
	addToken                = "Please send the jetton token address"
	initGroup               = "First initialise the bot with the start kilo command"
	tokenRemoved            = "Token removed successfully"
	errorUnexpected         = "Unexpected error, please try again."
	errorJettonAlreadyAdded = "You already have a jetton address, please delete it and add the new one."
	tokenAdded              = "New jetton address successfully added."
	errNothingToDelete      = "Nothing to delete"
	errCompAlreadyActive    = "This group already has an active competition, please wait for it to stop or stop manually with /stopcomp."
	errWithoutJetton        = "I'm sorry this group doesn't have a valid jetton address. "
	addTimestamp            = "How many hours should the contest last? Reply with 24, 48 or 72"
	errInvalidFormatHours   = "Invalid hours format for the competition, check and try again."
	competitionStarted      = "The competition has started, let the buys begin!\n\nOnly direct buys with TON will be included. If you sell you will be removed from the contest and your future buys won't count."
	errCompNotActive        = "Sorry, the group has no active competition. "
	addNewEmoji             = "Cool, send the new emoji. "
	emojiAdded              = "The emoji has been changed."
	competitionEnded        = "The competition is over."
	emptyList               = "The list of competitors is empty. "

	purchaseRemoved = "The purchase has been removed"

	actionCanceled = "action canceled"
)

type Commands struct {
	Groups *groups.Groups
	Temps  *groups.ActiveTemps
	Comps  *core.Competition
	Admins *Admins

	events *notificator.Groups

	promotions *promotions.Params

	BotAPI *tgbotapi.BotAPI
}

func InitCommands(groups *groups.Groups, temps *groups.ActiveTemps, comps *core.Competition, admins *Admins, bot *tgbotapi.BotAPI, promo *promotions.Params, events *notificator.Groups) *Commands {
	return &Commands{
		Groups:     groups,
		Temps:      temps,
		Comps:      comps,
		Admins:     admins,
		promotions: promo,
		events:     events,
		BotAPI:     bot,
	}
}

func (c *Commands) HandleGroup(updates <-chan tgbotapi.Update) {

	c.Admins.AddAdmin("uranusnfts")
	c.Admins.AddAdmin("Mrforrestgump84")
	c.Admins.AddAdmin("Tobster_LCL")

	/* 	purchases := core.InitPurchases()
	   	sales := core.InitSales()
	*/
	for update := range updates {

		if update.CallbackQuery != nil {
			callbackQuery := update.CallbackQuery
			buttonContext := callbackQuery.Data

			if buttonContext == cancelMarkup {
				err := c.Admins.DeactivateCommand(change_text)
				if err != nil {
					log.Print("error mientras se cerraba session y desactivaba el comando")
				}

				err = c.Admins.DeactivateCommand(change_video)
				if err != nil {
					log.Print("error mientras se cerraba session y desactivaba el comando")
				}

				err = c.Admins.DeactivateCommand(change_button_content)
				if err != nil {
					log.Print("error mientras se cerraba session y desactivaba el comando")
				}

				err = c.Admins.DeactivateCommand(change_button_context)
				if err != nil {
					log.Print("error mientras se cerraba session y desactivaba el comando")
				}

				c.send(update.CallbackQuery.Message.Chat.ID, actionCanceled)
			}
			continue
		}

		if update.Message != nil {
			c.handleAdminCommands(update)
			c.handleCommands(update)
			c.handleNoCommands(update)
			c.handleAdminsNoCommands(update)
		}
	}
}

func (c *Commands) handleCommands(update tgbotapi.Update) {

	chatIDStr := strconv.FormatInt(update.Message.Chat.ID, 10)
	chatID := update.Message.Chat.ID

	exist := c.Groups.GroupExist(chatIDStr)

	chatConfig := tgbotapi.ChatConfig{ChatID: chatID}
	chat, err := c.BotAPI.GetChat(chatConfig)
	if err != nil {
		log.Println(err)
	}

	if update.Message.IsCommand() {
		userID := update.Message.From.ID

		param := update.Message.Text

		switch update.Message.Command() {
		case start:

			if !chat.IsGroup() && !chat.IsSuperGroup() {
				log.Printf("the current group %v, no es un grupo o un supergrupo", chatID)
				c.send(chatID, onlyGroups)
				return
			}

			if !c.isAdmin(userID, chatID) {
				log.Printf("el usuario %s, no es un administrador", update.Message.From.UserName)
				c.send(chatID, onlyAdmin)
				return
			}

			if !exist {
				err := c.Groups.AddGroup(chatIDStr)
				if err != nil {
					log.Printf("error agregando grupo %v", err)
				}

				err = c.Temps.AddTemp(chatIDStr)
				if err != nil {
					log.Printf("error agregando grupo %v", err)
				}

				log.Printf("grupo con ID: %s, agregado correctamente", chatIDStr)
				c.send(chatID, groupAdded)
				return
			} else {
				log.Printf("grupo con ID: %s, no puede agregarse por que ya existe", chatIDStr)
				c.send(chatID, groupAlreadyExist)
			}
		case addtoken:
			if !chat.IsGroup() && !chat.IsSuperGroup() {
				log.Printf("the current group %v, no es un grupo o un supergrupo", chatID)
				c.send(chatID, onlyGroups)
				return
			}

			if !c.isAdmin(userID, chatID) {
				log.Printf("el usuario %s, no es un administrador", update.Message.From.UserName)
				c.send(chatID, onlyAdmin)
				return
			}

			if exist {
				group, err := c.Groups.GetDataGroup(chatIDStr)
				if err != nil {
					log.Printf("no se pudo obtener los datos del grupo, %v", err)
					return
				}

				err = c.Temps.ChangeTemp(0, chatIDStr, true)
				if err != nil {
					log.Printf("no se pudo actualizar el temp, %v", err)
					return
				}

				if group.JettonAddress != "" {
					log.Printf("el grupo %s, ya tiene una direccion activa", chatIDStr)
					c.send(chatID, errorJettonAlreadyAdded)
					return
				}

				c.send(chatID, addToken)
				return
			} else {
				c.send(chatID, initGroup)
				return
			}
		case removetoken:
			if !chat.IsGroup() && !chat.IsSuperGroup() {
				log.Printf("the current group %v, no es un grupo o un supergrupo", chatID)
				c.send(chatID, onlyGroups)
				return
			}

			if !c.isAdmin(userID, chatID) {
				log.Printf("el usuario %s, no es un administrador", update.Message.From.UserName)
				c.send(chatID, onlyAdmin)
				return
			}

			if exist {
				group, err := c.Groups.GetDataGroup(chatIDStr)
				if err != nil {
					log.Printf("no se pudo obtener los datos del grupo, %v", err)
					return
				}

				if group.JettonAddress == "" {
					log.Printf("no existe una direccion que remover para el grupo %s", chatIDStr)
					c.send(chatID, errNothingToDelete)
					return
				}

				group.JettonAddress = ""

				c.send(chatID, tokenRemoved)
				return
			} else {
				c.send(chatID, initGroup)
				return
			}
		case startnewcomp:
			if !chat.IsGroup() && !chat.IsSuperGroup() {
				log.Printf("the current group %v, no es un grupo o un supergrupo", chatID)
				c.send(chatID, onlyGroups)
				return
			}

			if !c.isAdmin(userID, chatID) {
				log.Printf("el usuario %s, no es un administrador", update.Message.From.UserName)
				c.send(chatID, onlyAdmin)
				return
			}

			if exist {
				group, err := c.Groups.GetDataGroup(chatIDStr)
				if err != nil {
					log.Printf("no se pudo obtener los datos del grupo, %v", err)
					return
				}

				if group.CompActive {
					log.Printf("ya existe una competicion activa para el grupo %s", chatIDStr)
					c.send(chatID, errCompAlreadyActive)
					return
				}

				if group.JettonAddress == "" {
					log.Printf("el grupo %s, no tiene una direccion activa", chatIDStr)
					c.send(chatID, errWithoutJetton)
					return
				}

				err = c.Comps.RemoveCompActive(chatIDStr)
				if err != nil {
					log.Printf("no se pudo remover la comp para el grupo %s", chatIDStr)
				}

				err = c.Temps.ChangeTemp(2, chatIDStr, true)
				if err != nil {
					log.Printf("no se pudo actualizar el temp, %v", err)
					return
				}

				c.send(chatID, addTimestamp)
				return
			} else {
				c.send(chatID, initGroup)
				return
			}
		case stopcomp:
			if !chat.IsGroup() && !chat.IsSuperGroup() {
				log.Printf("the current group %v, no es un grupo o un supergrupo", chatID)
				c.send(chatID, onlyGroups)
				return
			}

			if !c.isAdmin(userID, chatID) {
				log.Printf("el usuario %s, no es un administrador", update.Message.From.UserName)
				c.send(chatID, onlyAdmin)
				return
			}

			if exist {
				group, err := c.Groups.GetDataGroup(chatIDStr)
				if err != nil {
					log.Printf("no se pudo obtener los datos del grupo, %v", err)
					return
				}

				if !group.CompActive {
					log.Printf("no existe una competicion activa para el grupo %s", chatIDStr)
					c.send(chatID, errCompNotActive)
					return
				}

				if group.JettonAddress == "" {
					log.Printf("el grupo %s, no tiene una direccion activa", chatIDStr)
					c.send(chatID, errWithoutJetton)
					return
				}

				err = c.events.RemoveTicker(chatIDStr)
				if err != nil {
					log.Printf("no se pudo remover el ticker para el grupo %s, %v", chatIDStr, err)
				}

				err = c.Comps.RemoveBlacklistActive(chatIDStr)
				if err != nil {
					log.Printf("no se pudo remover el blacklist para el grupo %s", chatIDStr)
				}

				err = c.Comps.RemoveTimestampActive(chatIDStr)
				if err != nil {
					log.Printf("no se pudo remover el timestamp para el grupo %s", chatIDStr)
				}

				err = c.Groups.UpdateCompStatus(chatIDStr, false)
				if err != nil {
					log.Printf("no se pudo remover el status para el grupo %s", chatIDStr)
				}

				c.send(chatID, competitionEnded)
				return
			} else {
				c.send(chatID, initGroup)
				return
			}
		case addemoji:
			if !chat.IsGroup() && !chat.IsSuperGroup() {
				log.Printf("the current group %v, no es un grupo o un supergrupo", chatID)
				c.send(chatID, onlyGroups)
				return
			}

			if !c.isAdmin(userID, chatID) {
				log.Printf("el usuario %s, no es un administrador", update.Message.From.UserName)
				c.send(chatID, onlyAdmin)
				return
			}

			if exist {
				err = c.Temps.ChangeTemp(1, chatIDStr, true)
				if err != nil {
					log.Printf("no se pudo actualizar el temp, %v", err)
					return
				}

				c.send(chatID, addNewEmoji)
				return
			} else {
				c.send(chatID, initGroup)
				return
			}
		case list:
			if !chat.IsGroup() && !chat.IsSuperGroup() {
				log.Printf("the current group %v, no es un grupo o un supergrupo", chatID)
				c.send(chatID, onlyGroups)
				return
			}

			if exist {

				order, err := c.Comps.GetComp(chatIDStr)
				if err != nil {
					log.Printf("no se pudo obtener las ordenes de compra para el grupo %s", chatIDStr)
				}

				compList := order.GetCompList()

				if len(compList) == 0 {
					c.send(chatID, emptyList)
					return
				}

				msg := listMessage(compList)
				c.send(chatID, msg)
				return
			} else {
				c.send(chatID, initGroup)
				return
			}
		case admin:
			return
		case "ca":
			return
		case removebuyer:
			if !chat.IsGroup() && !chat.IsSuperGroup() {
				log.Printf("the current group %v, no es un grupo o un supergrupo", chatID)
				c.send(chatID, onlyGroups)
				return
			}

			if !c.isAdmin(userID, chatID) {
				log.Printf("el usuario %s, no es un administrador", update.Message.From.UserName)
				c.send(chatID, onlyAdmin)
				return
			}

			if exist {
				parts := strings.SplitN(param, " ", 2)
				if len(parts) < 2 {
					c.send(chatID, "Please provide the buyer's address. Usage: /removebuyer <address>")
					return
				}

				buyerAddress := parts[1]

				purchase, err := c.Comps.GetComp(chatIDStr)
				if err != nil {
					log.Printf("no se pudo obtener la competencia para el grupo %s", chatIDStr)
					return
				}
				_, hash, err := purchase.GetPurchaseAndHashByBuyer(buyerAddress)
				if err != nil {
					log.Printf("no se pudo obtener la compra para el grupo %s", chatIDStr)
					return
				}

				err = purchase.RemovePurchase(hash)
				if err != nil {
					log.Printf("no se pudo remover la compra para el grupo %s", chatIDStr)
					return
				}

				c.send(chatID, purchaseRemoved)
				return
			} else {
				c.send(chatID, initGroup)
				return
			}
		default:
			c.defaultHandler(update)
			return
		}
	}
}

func (c *Commands) handleNoCommands(update tgbotapi.Update) {
	chatIDStr := strconv.FormatInt(update.Message.Chat.ID, 10)
	chatID := update.Message.Chat.ID

	exist := c.Groups.GroupExist(chatIDStr)

	chatConfig := tgbotapi.ChatConfig{ChatID: chatID}
	chat, err := c.BotAPI.GetChat(chatConfig)
	if err != nil {
		log.Println(err)
	}

	if update.Message.From.ID == c.BotAPI.Self.ID {
		return
	}

	if !update.Message.IsCommand() && !chat.IsPrivate() {
		userID := update.Message.From.ID

		temps, err := c.Temps.GetActiveTemps(chatIDStr)
		if err != nil {
			log.Printf("error cargando los temps del grupo %s: %s", chatIDStr, err)
			return
		}

		if temps.AwaitingToken {
			param := update.Message.Text

			if !chat.IsGroup() && !chat.IsSuperGroup() {
				log.Printf("the current group %v, no es un grupo o un supergrupo", chatID)
				c.send(chatID, onlyGroups)
				return
			}

			if !c.isAdmin(userID, chatID) {
				log.Printf("el usuario %s, no es un administrador", update.Message.From.UserName)
				c.send(chatID, onlyAdmin)
				return
			}

			if exist {
				isAddress, err := getters.IsAddress(param)
				if err != nil {
					log.Printf("no se pudo comprobar si el parametro era una direccion")
					c.send(chatID, errorUnexpected)
					return
				}

				if isAddress {

					group, err := c.Groups.GetDataGroup(chatIDStr)
					if err != nil {
						log.Printf("no se pudo obtener los datos del grupo")
						return
					}

					if group.JettonAddress != "" {
						log.Printf("el grupo %s, ya tiene una direccion activa", chatIDStr)
						c.send(chatID, errorJettonAlreadyAdded)
						return
					}

					group.JettonAddress = param

					err = c.Groups.AddPools(chatIDStr)
					if err != nil {
						log.Printf("no se pudo obtener los pools, %v", err)
						return
					}

					err = c.Temps.ChangeTemp(0, chatIDStr, false)
					if err != nil {
						log.Printf("no se pudo actualizar el temp, %v", err)
						return
					}

					c.send(chatID, tokenAdded)
					return
				} else {
					log.Printf("direccion de jetton invalida para el grupo %s", chatIDStr)
					c.send(chatID, invalidJetton)
					return
				}
			} else {
				c.send(chatID, initGroup)
				return
			}

		}

		if temps.AwaitingEmoji {
			param := update.Message.Text

			if !chat.IsGroup() && !chat.IsSuperGroup() {
				log.Printf("the current group %v, no es un grupo o un supergrupo", chatID)
				c.send(chatID, onlyGroups)
				return
			}

			if !c.isAdmin(userID, chatID) {
				log.Printf("el usuario %s, no es un administrador", update.Message.From.UserName)
				c.send(chatID, onlyAdmin)
				return
			}

			if exist {
				group, err := c.Groups.GetDataGroup(chatIDStr)
				if err != nil {
					log.Printf("no se pudo obtener los datos del grupo")
					return
				}

				group.Emoji = param

				err = c.Temps.ChangeTemp(1, chatIDStr, false)
				if err != nil {
					log.Printf("no se pudo actualizar el temp, %v", err)
					return
				}

				c.send(chatID, emojiAdded)
				return
			} else {
				c.send(chatID, initGroup)
				return
			}

		}

		if temps.AwaitingTimestamp {
			param := update.Message.Text

			if !chat.IsGroup() && !chat.IsSuperGroup() {
				log.Printf("the current group %v, no es un grupo o un supergrupo", chatID)
				c.send(chatID, onlyGroups)
				return
			}

			if !c.isAdmin(userID, chatID) {
				log.Printf("el usuario %s, no es un administrador", update.Message.From.UserName)
				c.send(chatID, onlyAdmin)
				return
			}

			if exist {
				var timestamp int64
				switch param {
				case "24", "48", "72", "5":
					switch param {
					case "24":
						timestamp = time.Now().Unix() + int64(24*60*60)
					case "48":
						timestamp = time.Now().Unix() + int64(24*60*60*2)
					case "72":
						timestamp = time.Now().Unix() + int64(24*60*60*3)
					case "5":
						timestamp = time.Now().Unix() + int64(5*60)
					}
				default:
					log.Printf("parametro invalido para comenzar la competencia")
					c.send(chatID, errInvalidFormatHours)
					return
				}

				err := c.Comps.NewTimestamp(chatIDStr, timestamp)
				if err != nil {
					log.Printf("no se pudo crear el nuevo timestamp para el grupo %s", chatIDStr)
					return
				}

				err = c.events.AddNewTicker(chatIDStr)
				if err != nil {
					log.Printf("no se pudo agregar el ticker para el grupo %s", chatIDStr)
					return
				}

				c.Groups.ActiveGroups[chatIDStr].CompActive = true

				err = c.Temps.ChangeTemp(2, chatIDStr, false)
				if err != nil {
					log.Printf("no se pudo actualizar el temp, %v", err)
					return
				}

				c.send(chatID, competitionStarted)
				return
			} else {
				c.send(chatID, initGroup)
				return
			}
		}
	}
}

func (c *Commands) isAdmin(id int, chatID int64) bool {
	// Verificar si el bot es un administrador del grupo
	memberConfig := tgbotapi.ChatConfigWithUser{
		ChatID: chatID,
		UserID: id,
	}
	member, err := c.BotAPI.GetChatMember(memberConfig)
	if err != nil {
		log.Println(err)
	}

	return member.IsAdministrator() || member.IsCreator()
}

func (b *Commands) defaultHandler(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I'm sorry, I didn't get your message. please contact with admins for more info")
	_, err := b.BotAPI.Send(msg)
	if err != nil {
		log.Println("Error al enviar mensaje:", err)
	}
}

// ListMessage function to generate the message list
func listMessage(buyers []*core.Purchase) string {
	header := "👑 *Top Buyers*:\n\n"

	for i := 0; i < 10; i++ {
		if i < len(buyers) && buyers[i] != nil {
			tx := buyers[i]
			tonSpent := tx.Ton
			wallet := tx.Buyer

			var walletShortened, walletLink string
			if wallet != "" {
				walletShortened = fmt.Sprintf("%s...%s", wallet[:6], wallet[len(wallet)-6:])
				walletLink = fmt.Sprintf("[%s](https://tonviewer.com/%s)", walletShortened, url.PathEscape(wallet))
			} else {
				walletShortened = "not set"
				walletLink = "not set"
			}

			header += fmt.Sprintf("%d.) %s *TON* - %s\n", i+1, tonSpent, walletLink)
		} else {
			header += fmt.Sprintf("%d.) not set *TON* - not set\n", i+1)
		}
	}

	return header
}
