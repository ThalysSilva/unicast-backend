package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	_ "github.com/lib/pq"
)




func InitDB() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"))

	fmt.Println("string de conexão: ", connStr)

	var err error
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		errorMounted := fmt.Errorf("falha ao conectar ao banco: %w", err)
		return nil, customerror.Trace("InitDb", errorMounted)
	}
	if err = db.Ping(); err != nil {
		errorMounted := fmt.Errorf("falha ao efetuar ping no banco: %w", err)
		return nil, customerror.Trace("InitDb", errorMounted)
	}
	fmt.Println("Banco de dados conectado!")
	return db, nil
}

