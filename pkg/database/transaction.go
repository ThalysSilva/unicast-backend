package database

import (
	"context"
	"database/sql"
	"fmt"
)

// DB abstrai métodos de *sql.DB e *sql.Tx.
type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
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
func MakeTransaction[T any](ctx context.Context, repos []Transactional, fn func() (T, error)) (T, error) {
	var zero T

	if len(repos) == 0 {
		return zero, fmt.Errorf("nenhum repositório fornecido")
	}

	tx, err := resolveTransactionBackend(repos)
	if err != nil {
		return zero, err
	}
	txHandle, err := tx.Begin(ctx)
	if err != nil {
		return zero, fmt.Errorf("falha ao iniciar transação: %w", err)
	}

	reposWithTx, err := withTransaction(repos, txHandle)
	if err != nil {
		_ = tx.Rollback(txHandle)
		return zero, err
	}
	repos = reposWithTx

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
	data, err := fn()
	if err != nil {
		if rollbackErr := tx.Rollback(txHandle); rollbackErr != nil {
			return zero, fmt.Errorf("falha ao reverter transação: %v; erro original: %w", rollbackErr, err)
		}
		return zero, err
	}

	if err := tx.Commit(txHandle); err != nil {
		return zero, fmt.Errorf("falha ao confirmar transação: %w", err)
	}

	return data, nil
}

func resolveTransactionBackend(repos []Transactional) (Tx, error) {
	var tx Tx

	for i, repo := range repos {
		backend := repo.TransactionBackend()
		switch db := backend.(type) {
		case *sql.DB:
			if tx == nil {
				tx = NewSQLTx(db)
			} else if _, ok := tx.(*SQLTx); !ok {
				return nil, fmt.Errorf("repositório %d usa backend incompatível (esperado *sql.DB)", i)
			}
		default:
			return nil, fmt.Errorf("repositório %d usa backend desconhecido: %T", i, backend)
		}
	}

	if tx == nil {
		return nil, fmt.Errorf("nenhum backend válido fornecido")
	}

	return tx, nil
}

func withTransaction(repos []Transactional, txHandle any) ([]Transactional, error) {
	updated := make([]Transactional, len(repos))

	for i, repo := range repos {
		newRepo := repo.WithTransaction(txHandle)
		if newRepo == nil {
			return nil, fmt.Errorf("WithTransaction retornou nil para repositório %d", i)
		}

		castedRepo, ok := newRepo.(Transactional)
		if !ok {
			return nil, fmt.Errorf("repositório %d não implementa Transactional após WithTransaction", i)
		}

		updated[i] = castedRepo
	}

	return updated, nil
}
