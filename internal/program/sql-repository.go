package program

import (
	"context"
	"database/sql"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

// Gerencia operações de banco para Program
type sqlRepository struct {
	db    database.DB
	sqlDB *sql.DB
}

func newSQLRepository(db *sql.DB) Repository {
	newDb := database.NewSQLTx(db)
	return &sqlRepository{
		db: newDb.DB,
	}
}
func (r *sqlRepository) WithTransaction(tx any) any {
	return &sqlRepository{
		db:    database.NewSQLTx(nil).WithSQLTransaction(tx).DB,
		sqlDB: r.sqlDB,
	}
}

func (r *sqlRepository) TransactionBackend() any {
	return r.sqlDB
}

// Insere um novo programa
func (r *sqlRepository) Create(ctx context.Context, name, description, campusID string, active bool) error {
	query := `
        INSERT INTO programs (name, description, campus_id, active)
        VALUES ($1, $2, $3, $4)
    `
	_, err := r.db.ExecContext(ctx, query, name, description, campusID, active)
	return err
}

// FindByID busca um programa pelo ID
func (r *sqlRepository) FindByID(ctx context.Context, id string) (*Program, error) {
	query := `
        SELECT id, name, description, created_at, updated_at, campus_id, active
        FROM programs
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	program := &Program{}
	err := row.Scan(&program.ID, &program.Name, &program.Description, &program.CreatedAt, &program.UpdatedAt, &program.CampusID, &program.Active)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return program, nil
}

// Atualiza os campos de uma instância no banco de dados. Campos não fornecidos não serão atualizados.
// Campos: name, description, campus_id, active
func (r *sqlRepository) Update(ctx context.Context, id string, fields map[string]any) error {
	err := database.Update(ctx, r.db, "programs", id, fields)
	if err != nil {
		return err
	}
	return nil
}

// Delete remove um programa pelo ID
func (r *sqlRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM programs WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
