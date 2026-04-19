package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

const (
	defaultEnvFile  = ".env"
	defaultSeedFile = "scripts/demo-seed.sql"
)

func main() {
	envPath := flag.String("env", defaultEnvFile, "arquivo de ambiente")
	seedPath := flag.String("file", defaultSeedFile, "arquivo SQL da seed")
	flag.Parse()

	if err := loadEnvFile(*envPath); err != nil {
		exitWithError(err)
	}

	db, err := openDatabase()
	if err != nil {
		exitWithError(err)
	}
	defer db.Close()

	seed, err := os.ReadFile(filepath.Clean(*seedPath))
	if err != nil {
		exitWithError(fmt.Errorf("falha ao ler seed %s: %w", *seedPath, err))
	}

	if _, err := db.Exec(string(seed)); err != nil {
		exitWithError(fmt.Errorf("falha ao executar seed: %w", err))
	}

	fmt.Println("Seed de demonstração aplicada com sucesso.")
	fmt.Println("Login: demo@unicast.local")
	fmt.Println("Senha: Unicast@2026")
}

func openDatabase() (*sql.DB, error) {
	connStr := os.Getenv("POSTGRES_DATABASE_URL")
	if strings.TrimSpace(connStr) == "" {
		connStr = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_PORT"),
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_DB"),
		)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir conexão com Postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("falha ao conectar no Postgres: %w", err)
	}
	return db, nil
}

func loadEnvFile(path string) error {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("falha ao abrir %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("%s:%d sem '='", path, lineNumber)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			return fmt.Errorf("%s:%d com chave vazia", path, lineNumber)
		}

		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("falha ao definir %s: %w", key, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("falha ao ler %s: %w", path, err)
	}
	return nil
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
	os.Exit(1)
}
