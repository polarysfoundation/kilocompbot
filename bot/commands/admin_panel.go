package commands

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	admin                 = "admin"
	change_text           = "Change Text"
	change_video          = "Change Video"
	change_button_content = "Change Button Link"
	change_button_context = "Change Button Name"
	exit                  = "Exit"
	total_groups          = "Total Groups"
	send_announcement     = "Send Announcement"
	sending_announcement  = "Sending Announcement"
)

var (
	errEmptyParam           = errors.New("error: empty param")
	errAdminAlreadyExist    = errors.New("error: administrador con ese usuario ya existe")
	errAdminNotExist        = errors.New("error: administrador con ese usuario no existe")
	errCommandAlreadyActive = errors.New("error: commando ya esta activo")
	errCommandNotActive     = errors.New("error: commando ya esta desactivado")
	errAdminAlreadyLogged   = errors.New("error: administrador ya esta loggueado")
	errAdminNotLogged       = errors.New("error: administrador no esta loggueado")
)

const (
	onlyAdmins         = "Sorry only administrators can use that command, please if you have a question you can contact admins at t.me/KiloTonCoin"
	notGroups          = "Sorry, for this mode it is not allowed to be used in groups. "
	errAdminNotAllowed = "Sorry, no action allowed, any questions you can contact a bot administrator. "
	errAlreadyLogued   = "Sorry, but you're already on the admin panel."
	errDoubleSession   = "There is already an administrator with an active session, please try later. "
	errNoLoggued       = "I'm sorry, there's no active session. "

	addNewText          = "Send new text"
	addNewVideo         = "Send new video"
	addNewButtonContent = "Send new button content"
	addNewButtonContext = "Send new button context"
	addAnnouncement     = "Send new accouncement"

	cancelMarkup = "Cancel"
	cancel       = "cancel"

	textUpdated          = "promo text updated. "
	mediaUpdated         = "The ad video has already been updated."
	buttonContextUpdated = "The contents of the button have been updated. "
	buttonContentUpdated = "The button name have been updated. "
)

type Admins struct {
	AdminAllowed  map[string]struct{}
	CommandActive map[string]bool
	Signed        map[string]bool
	Total         int64
	mutex         sync.RWMutex
}

func InitAdmins() *Admins {
	return &Admins{
		AdminAllowed:  make(map[string]struct{}),
		CommandActive: make(map[string]bool),
		Signed:        make(map[string]bool),
	}
}

func (a *Admins) IsLogged(username string) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if username == "" {
		return false
	}

	if _, exist := a.AdminAllowed[username]; !exist {
		return false
	}

	if _, exist := a.Signed[username]; exist {
		return true
	}

	return false
}

func (a *Admins) SignIn(username string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if username == "" {
		return errEmptyParam
	}

	if _, exist := a.AdminAllowed[username]; !exist {
		return errAdminNotExist
	}

	if _, exist := a.Signed[username]; exist {
		return errAdminAlreadyLogged
	}

	a.Total = 1
	a.Signed[username] = true

	return nil
}

func (a *Admins) SignOut(username string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if username == "" {
		return errEmptyParam
	}

	if _, exist := a.AdminAllowed[username]; !exist {
		return errAdminNotExist
	}

	if _, exist := a.Signed[username]; !exist {
		return errAdminNotLogged
	}

	a.Total = 0
	delete(a.Signed, username)

	return nil
}

func (a *Admins) AdminExist(username string) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if username == "" {
		return false
	}

	if _, exist := a.AdminAllowed[username]; exist {
		return true
	}

	return false
}

func (a *Admins) AddAdmin(username string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if username == "" {
		return errEmptyParam
	}

	if _, exist := a.AdminAllowed[username]; exist {
		return errAdminAlreadyExist
	}

	a.AdminAllowed[username] = struct{}{}

	return nil
}

func (a *Admins) RemoveAdmin(username string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if username == "" {
		return errEmptyParam
	}

	if _, exist := a.AdminAllowed[username]; !exist {
		return errAdminNotExist
	}

	delete(a.AdminAllowed, username)

	return nil
}

func (a *Admins) ActiveCommand(command string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if command == "" {
		return errEmptyParam
	}

	if _, exist := a.CommandActive[command]; exist {
		return errCommandAlreadyActive
	}

	a.CommandActive[command] = true

	return nil
}

func (a *Admins) DeactivateCommand(command string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if command == "" {
		return errEmptyParam
	}

	if _, exist := a.CommandActive[command]; !exist {
		return errCommandNotActive
	}

	delete(a.CommandActive, command)

	return nil
}

func (a *Admins) CommandStatus(command string) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if command == "" {
		return false
	}

	if _, exist := a.CommandActive[command]; exist {
		return true
	}

	return false
}

func (p *Commands) handleAdminCommands(update tgbotapi.Update) {
	/* 	chatIDStr := strconv.FormatInt(update.Message.Chat.ID, 10) */
	chatID := update.Message.Chat.ID

	chatConfig := tgbotapi.ChatConfig{ChatID: chatID}
	chat, err := p.BotAPI.GetChat(chatConfig)
	if err != nil {
		log.Println(err)
	}

	if update.Message.IsCommand() {
		userName := update.Message.From.UserName

		switch update.Message.Command() {
		case admin:

			log.Printf("mensaje desde el grupo %v", chatID)
			log.Printf("el grupo %v es un %v", chatID, chat.Type)

			if chat.IsGroup() && chat.IsSuperGroup() || chat.IsSuperGroup() || chat.IsGroup() {
				log.Printf("the current group %v, es un grupo o un supergrupo", chatID)
				p.send(chatID, notGroups)
				return
			}

			if !p.Admins.AdminExist(userName) {
				log.Printf("el usuario %s no es un administrador del bot", userName)
				p.send(chatID, errAdminNotAllowed)
				return
			}

			if p.Admins.IsLogged(userName) {
				log.Printf("el usuario %s ya esta logueado", userName)
				p.send(chatID, errAlreadyLogued)
				return
			}

			if p.Admins.Total == 1 {
				log.Print("Ya hay un usuario logueado")
				p.send(chatID, errDoubleSession)
				return
			}

			err := p.Admins.SignIn(userName)
			if err != nil {
				log.Printf("Error mientras se iniciaba sesion con el usuario %s", userName)
				return
			}

			p.sendOptions(chatID)
			return
		default:
			return
		}

	}
}

func (p *Commands) handleAdminsNoCommands(update tgbotapi.Update) {
	/* 	chatIDStr := strconv.FormatInt(update.Message.Chat.ID, 10) */
	chatID := update.Message.Chat.ID

	chatConfig := tgbotapi.ChatConfig{ChatID: chatID}
	chat, err := p.BotAPI.GetChat(chatConfig)
	if err != nil {
		log.Println(err)
	}

	if update.Message.From.ID == p.BotAPI.Self.ID {
		return
	}

	if !update.Message.IsCommand() && chat.IsPrivate() {
		userName := update.Message.From.UserName

		param := update.Message.Text

		if p.Admins.Total == 1 && p.Admins.IsLogged(userName) {
			if chat.IsGroup() && chat.IsSuperGroup() {
				log.Printf("the current group %v, es un grupo o un supergrupo", chatID)
				p.send(chatID, notGroups)
				return
			}

			if !p.Admins.AdminExist(userName) {
				log.Printf("el usuario %s no es un administrador del bot", userName)
				p.send(chatID, errAdminNotAllowed)
				return
			}

			if p.Admins.CommandStatus(change_text) {
				err := p.promotions.UpdateAdName(param)
				if err != nil {
					log.Printf("no se pudo actualizar el texto del usuario %s", userName)
					return
				}

				err = p.Admins.DeactivateCommand(change_text)
				if err != nil {
					log.Print("error mientras se cerraba session y desactivaba el comando")
					return
				}

				p.send(chatID, textUpdated)
				return
			}

			if p.Admins.CommandStatus(change_video) {
				if update.Message.Video != nil {
					fileID := update.Message.Video.FileID
					fileName := fmt.Sprintf("%s.mp4", fileID)
					p.saveVideo(fileID, fileName)
				}

				err = p.Admins.DeactivateCommand(change_video)
				if err != nil {
					log.Print("error mientras se cerraba session y desactivaba el comando")
					return
				}

				p.send(chatID, mediaUpdated)
				return
			}

			if p.Admins.CommandStatus(change_button_context) {
				err := p.promotions.UpdateButtonName(param)
				if err != nil {
					log.Printf("no se pudo actualizar el texto del usuario %s", userName)
					return
				}

				err = p.Admins.DeactivateCommand(change_button_context)
				if err != nil {
					log.Print("error mientras se cerraba session y desactivaba el comando")
					return
				}

				p.send(chatID, buttonContextUpdated)
				return
			}

			if p.Admins.CommandStatus(change_button_content) {
				err := p.promotions.UpdateButtonLink(param)
				if err != nil {
					log.Printf("no se pudo actualizar el texto del usuario %s", userName)
					return
				}

				err = p.Admins.DeactivateCommand(change_button_content)
				if err != nil {
					log.Print("error mientras se cerraba session y desactivaba el comando")
					return
				}

				p.send(chatID, buttonContentUpdated)
				return
			}

			if p.Admins.CommandStatus(send_announcement) {
				msg := param
				for i := range p.Groups.ActiveGroups {
					chatIDInt, _ := strconv.Atoi(i)
					p.send(int64(chatIDInt), msg)
					break
				}

				err = p.Admins.DeactivateCommand(send_announcement)
				if err != nil {
					log.Print("error mientras se cerraba session y desactivaba el comando")
					return
				}

				p.send(chatID, sending_announcement)
				return
			}

			switch param {
			case change_text:
				err := p.Admins.ActiveCommand(change_text)
				if err != nil {
					log.Printf("error mientras se activaba el comando %s", change_text)
					return
				}

				markup := p.keyboardMarkup(cancel, cancelMarkup)
				p.sendReplyWithMarkup(chatID, addNewText, markup)
				return
			case change_video:
				err := p.Admins.ActiveCommand(change_video)
				if err != nil {
					log.Printf("error mientras se activaba el comando %s", change_video)
					return
				}

				markup := p.keyboardMarkup(cancel, cancelMarkup)
				p.sendReplyWithMarkup(chatID, addNewVideo, markup)
				return
			case change_button_content:
				err := p.Admins.ActiveCommand(change_button_content)
				if err != nil {
					log.Printf("error mientras se activaba el comando %s", change_button_content)
					return
				}

				markup := p.keyboardMarkup(cancel, cancelMarkup)
				p.sendReplyWithMarkup(chatID, addNewButtonContent, markup)
				return
			case change_button_context:
				err := p.Admins.ActiveCommand(change_button_context)
				if err != nil {
					log.Printf("error mientras se activaba el comando %s", change_button_context)
					return
				}

				markup := p.keyboardMarkup(cancel, cancelMarkup)
				p.sendReplyWithMarkup(chatID, addNewButtonContext, markup)
				return
			case exit:

				status := p.Admins.CommandStatus(change_text)
				if status {
					err = p.Admins.DeactivateCommand(change_text)
					if err != nil {
						log.Print("error mientras se cerraba session y desactivaba el comando")
						return
					}
				}

				status = p.Admins.CommandStatus(change_video)
				if status {
					err = p.Admins.DeactivateCommand(change_video)
					if err != nil {
						log.Print("error mientras se cerraba session y desactivaba el comando")
						return
					}

				}

				status = p.Admins.CommandStatus(change_button_content)
				if status {
					err = p.Admins.DeactivateCommand(change_button_content)
					if err != nil {
						log.Print("error mientras se cerraba session y desactivaba el comando")
						return
					}
				}

				status = p.Admins.CommandStatus(change_button_context)
				if status {
					err = p.Admins.DeactivateCommand(change_button_context)
					if err != nil {
						log.Print("error mientras se cerraba session y desactivaba el comando")
						return
					}
				}

				status = p.Admins.CommandStatus(send_announcement)
				if status {
					err = p.Admins.DeactivateCommand(send_announcement)
					if err != nil {
						log.Print("error mientras se cerraba session y desactivaba el comando")
						return
					}
				}

				err := p.Admins.SignOut(userName)
				if err != nil {
					log.Print("error mientras se cerraba session")
					return
				}

				p.hideKeyboard(chatID)

				return
			case total_groups:
				totalGroups := make([]string, 0)
				for i := range p.Groups.ActiveGroups {
					totalGroups = append(totalGroups, i)
					break
				}

				msg := fmt.Sprintf("Total groups: %d", len(totalGroups))
				p.send(chatID, msg)
				return
			case send_announcement:
				err := p.Admins.ActiveCommand(send_announcement)
				if err != nil {
					log.Printf("error mientras se activaba el comando %s", change_button_context)
					return
				}

				markup := p.keyboardMarkup(cancel, cancelMarkup)
				p.sendReplyWithMarkup(chatID, addAnnouncement, markup)
				return
			default:
				return
			}

		} else {
			return
		}

	}
}

func (p *Commands) send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := p.BotAPI.Send(msg)
	if err != nil {
		log.Println("Error al enviar mensaje:", err)
	}
}

func (p *Commands) sendReplyWithMarkup(chatID int64, text string, markup *tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if markup != nil {
		msg.ReplyMarkup = markup
	}
	_, err := p.BotAPI.Send(msg)
	if err != nil {
		log.Println("Error al enviar mensaje:", err)
	}
}

func (a *Commands) keyboardMarkup(buttonContext string, buttonContent string) *tgbotapi.InlineKeyboardMarkup {
	button := tgbotapi.NewInlineKeyboardButtonData(buttonContext, buttonContent)
	row := []tgbotapi.InlineKeyboardButton{button}
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(row)
	return &inlineKeyboard
}

func (p *Commands) sendOptions(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Select an option below")

	// Crear los botones del teclado personalizado
	button1 := tgbotapi.NewKeyboardButton(change_video)
	button2 := tgbotapi.NewKeyboardButton(change_text)
	button3 := tgbotapi.NewKeyboardButton(change_button_context)
	button4 := tgbotapi.NewKeyboardButton(change_button_content)
	button5 := tgbotapi.NewKeyboardButton(total_groups)
	button6 := tgbotapi.NewKeyboardButton(send_announcement)
	button7 := tgbotapi.NewKeyboardButton(exit)

	// Crear las filas de botones
	row1 := tgbotapi.NewKeyboardButtonRow(button1, button2)
	row2 := tgbotapi.NewKeyboardButtonRow(button3, button4)
	row3 := tgbotapi.NewKeyboardButtonRow(button5, button6)
	row4 := tgbotapi.NewKeyboardButtonRow(button7)

	// Crear el teclado personalizado con las filas de botones
	replyKeyboard := tgbotapi.NewReplyKeyboard(row1, row2, row3, row4)
	msg.ReplyMarkup = replyKeyboard

	p.BotAPI.Send(msg)
}

func (p *Commands) hideKeyboard(chatID int64) {
	// Crear una estructura ReplyKeyboardRemove
	removeKeyboard := tgbotapi.NewRemoveKeyboard(true) // `true` para eliminar el teclado para todos los usuarios de este chat

	// Crear un mensaje con la estructura ReplyKeyboardRemove
	msg := tgbotapi.NewMessage(chatID, "leaving the administration panel")
	msg.ReplyMarkup = removeKeyboard

	// Enviar el mensaje
	_, err := p.BotAPI.Send(msg)
	if err != nil {
		log.Println("Error al enviar el mensaje:", err)
	}
}

func (c *Commands) saveVideo(fileID string, fileName string) {
	file, err := c.BotAPI.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		log.Println("Error al obtener el archivo:", err)
		return
	}

	filePath := fmt.Sprintf("assets/media/%s", fileName)

	err = c.promotions.UpdateMedia(fileName)
	if err != nil {
		log.Printf("no se pudo actualizar el video por el siguiente error %v", err)
		return
	}

	// URL completa para descargar el archivo
	fileURL := file.Link(c.BotAPI.Token)

	// Descargar el archivo
	response, err := http.Get(fileURL)
	if err != nil {
		log.Println("Error al descargar el archivo:", err)
		return
	}
	defer response.Body.Close()

	// Crear el archivo localmente en la ruta especificada
	out, err := os.Create(filePath)
	if err != nil {
		log.Println("Error al crear el archivo local:", err)
		return
	}
	defer out.Close()

	// Escribir el contenido descargado en el archivo local
	_, err = io.Copy(out, response.Body)
	if err != nil {
		log.Println("Error al guardar el archivo local:", err)
	}

	log.Println("Video guardado como:", filePath)
}
