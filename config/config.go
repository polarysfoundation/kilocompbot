package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

var (
	errorbotTokenNotExist = errors.New("error: el token del bot no existe")
	errorEnvFileNotExist  = errors.New("error: el archivo env no existe, o hubo error al cargarlo")
)

type Config struct {
	BotToken     string
	TONCenterAPI string
}

func Init() (*Config, error) {

	if err := godotenv.Load(); err != nil {
		return nil, errorEnvFileNotExist
	}
	// Acceder a la variable de entorno TELEGRAM_TOKEN
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		return nil, errorbotTokenNotExist
	}

	tonAPI := os.Getenv("TON_API")
	if telegramToken == "" {
		return nil, errorbotTokenNotExist
	}

	return &Config{
		BotToken:     telegramToken,
		TONCenterAPI: tonAPI,
	}, nil
}
