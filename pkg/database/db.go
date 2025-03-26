package database

import (
	"database/sql"
	"fmt"
	"os"
	"unicast-api/pkg/utils"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() error {
	trace := utils.TraceError("InitDB")
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	fmt.Println("string de conexão: ", connStr)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		errorMounted := fmt.Errorf("falha ao conectar ao banco: %w", err)
		return trace(errorMounted)
	}
	if err = DB.Ping(); err != nil {
		errorMounted := fmt.Errorf("falha ao efetuar ping no banco: %w", err)
		return trace(errorMounted)
	}
	fmt.Println("Banco de dados conectado!")
	return nil
}
