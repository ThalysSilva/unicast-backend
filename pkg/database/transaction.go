package database

import (
	"context"
	"database/sql"
	"fmt"

)

// DB abstrai métodos de *sql.DB e *sql.Tx.
type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Transactional define repositórios que suportam transações e fornecem backend.
type Transactional interface {
	WithTransaction(tx any) any
	TransactionBackend() any
}

// Tx define a interface para transações.
type Tx interface {
	Begin(ctx context.Context) (any, error)
	Commit(tx any) error
	Rollback(tx any) error
}

// SQLTx gerencia contexto para database/sql.
type SQLTx struct {
	DB DB // Pode ser *sql.DB ou *sql.Tx
}

// NewSQLTx cria um SQLTx.
func NewSQLTx(db *sql.DB) *SQLTx {
	return &SQLTx{
		DB: db,
	}
}

// Begin inicia uma transação SQL.
func (t *SQLTx) Begin(ctx context.Context) (any, error) {
	db, ok := t.DB.(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("banco de dados não é *sql.DB")
	}
	return db.BeginTx(ctx, nil)
}

// Commit confirma uma transação SQL.
func (t *SQLTx) Commit(tx any) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return fmt.Errorf("tipo inválido para transação SQL")
	}
	return sqlTx.Commit()
}

// Rollback reverte uma transação SQL.
func (t *SQLTx) Rollback(tx any) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return fmt.Errorf("tipo inválido para transação SQL")
	}
	return sqlTx.Rollback()
}

// WithSQLTransaction configura a transação SQL.
func (t *SQLTx) WithSQLTransaction(tx any) *SQLTx {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return t
	}
	return &SQLTx{
		DB: sqlTx,
	}
}

// MakeTransaction executa uma transação inspirada no GORM.
func MakeTransaction(ctx context.Context, repos []Transactional, fn func() error) error {
	if len(repos) == 0 {
		return fmt.Errorf("nenhum repositório fornecido")
	}

	// Coleta o backend do primeiro repositório
	var tx Tx
	for _, repo := range repos {
		backend := repo.TransactionBackend()
		switch db := backend.(type) {
		case *sql.DB:
			if tx == nil {
				tx = NewSQLTx(db)
			} else if _, ok := tx.(*SQLTx); !ok {
				return fmt.Errorf("repositórios com backends mistos (SQL e GORM)")
			}
		/* case *gorm.DB:
		if tx == nil {
			tx = NewGORMTx(db)
		} else if _, ok := tx.(*GORMTx); !ok {
			return fmt.Errorf("repositórios com backends mistos (SQL e GORM)")
		} */
		default:
			return fmt.Errorf("backend desconhecido: %T", backend)
		}
	}

	txHandle, err := tx.Begin(ctx)
	if err != nil {
		return fmt.Errorf("falha ao iniciar transação: %w", err)
	}

	// Configura repositórios com a transação
	for i, repo := range repos {
		newRepo := repo.WithTransaction(txHandle)
		if newRepo == nil {
			_ = tx.Rollback(txHandle)
			return fmt.Errorf("WithTransaction retornou nil para repositório %d", i)
		}
		repos[i] = newRepo.(Transactional)
	}

	// Garante rollback em caso de pânico
	defer func() {
		if r := recover(); r != nil {
			if err := tx.Rollback(txHandle); err != nil {
				fmt.Printf("falha ao reverter transação: %v\n", err)
			}
			panic(r)
		}
	}()

	// Executa o callback
	if err := fn(); err != nil {
		if rollbackErr := tx.Rollback(txHandle); rollbackErr != nil {
			return fmt.Errorf("falha ao reverter transação: %v; erro original: %w", rollbackErr, err)
		}
		return err
	}

	return tx.Commit(txHandle)
}
