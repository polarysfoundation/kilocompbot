package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Database struct {
	Client *sql.DB
}

// Inicializar database postgresql
func Init() (*Database, error) {

	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error cargando el archivo .env: %v", err)
	}

	host := os.Getenv("HOST_ADDRESS")
	if host == "" {
		return nil, fmt.Errorf("la variable HOST_ADDRESS no está definida en el archivo .env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		return nil, fmt.Errorf("la variable PORT no está definida en el archivo .env")
	}

	user := os.Getenv("DBUSER")
	if user == "" {
		return nil, fmt.Errorf("la variable USER no está definida en el archivo .env")
	}

	password := os.Getenv("DBPASSWORD")
	if password == "" {
		return nil, fmt.Errorf("la variable PASSWORD no está definida en el archivo .env")
	}

	dbname := os.Getenv("DBNAME")
	if dbname == "" {
		return nil, fmt.Errorf("la variable DBNAME no está definida en el archivo .env")
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("error abriendo la conexión a la base de datos: %v", err)
	}

	return &Database{
		Client: db,
	}, nil

}
