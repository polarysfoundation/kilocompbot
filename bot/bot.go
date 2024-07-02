package bot

import (
	"context"
	"database/sql"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/polarysfoundation/kilocompbot/bot/backups"
	"github.com/polarysfoundation/kilocompbot/bot/commands"
	"github.com/polarysfoundation/kilocompbot/bot/notificator"
	"github.com/polarysfoundation/kilocompbot/bot/promotions"
	"github.com/polarysfoundation/kilocompbot/core"
	"github.com/polarysfoundation/kilocompbot/groups"
)

type Bot struct {
	API     *tgbotapi.BotAPI
	DB      *sql.DB
	Context context.Context
	TONAPI  string
}

func InitBot(token string, db *sql.DB, ctx context.Context, tonAPI string) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		API:     botAPI,
		DB:      db,
		Context: ctx,
		TONAPI:  tonAPI,
	}, nil
}

func (b *Bot) Run() {
	log.Printf("Bot iniciado como %s", b.API.Self.UserName)

	// Iniciar instancias
	log.Print("Iniciando instancias necesarias")
	groupsMap := groups.InitGroups()
	temps := groups.InitTemp()
	comps := core.InitComp()
	promo := promotions.InitParams()
	events := notificator.InitEvents()

	backup := backups.InitBackup(b.DB, groupsMap, comps, promo, temps)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.API.GetUpdatesChan(u)
	if err != nil {
		log.Println(err)
	}

	event := notificator.Init(b.API, groupsMap, comps, events, b.TONAPI, promo, b.DB)

	backup.LoadData(event)

	admins := commands.InitAdmins()
	handler := commands.InitCommands(groupsMap, temps, comps, admins, b.API, promo, event)

	var wg sync.WaitGroup
	wg.Add(3)
	go func(updates <-chan tgbotapi.Update) {
		defer wg.Done()
		handler.HandleGroup(updates)
	}(updates)

	go func() {
		defer wg.Done()
		event.HandleUpdate()
	}()

	go func() {
		defer wg.Done()
		backup.HandleBackup()
	}()

	// Esperar a que todas las goroutines terminen
	wg.Wait()
}
