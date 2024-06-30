package main

import (
	"context"
	"log"

	"github.com/polarysfoundation/kilocompbot/bot"
	"github.com/polarysfoundation/kilocompbot/config"
	"github.com/polarysfoundation/kilocompbot/database"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("error iniciando config: %v", err)
	}

	db, err := database.Init()
	if err != nil {
		log.Fatalf("error iniciando database: %v", err)
	}

	client := db.Client

	err = client.Ping()
	if err != nil {
		log.Fatalf("No se pudo conectar a la base de datos: %v", err)
	}

	defer client.Close()

	bot, err := bot.InitBot(cfg.BotToken, client, ctx, cfg.TONCenterAPI)
	if err != nil {
		log.Fatalf("Error iniciando el bot: %v", err)
	}

	bot.Run()
}
