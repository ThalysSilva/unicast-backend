package course

import (
	"context"
	"database/sql"

	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

// Gerencia operações de banco para course
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

// Insere um novo course
func (r *sqlRepository) Create(ctx context.Context, name, description, programID string, year, semester int) error {
	query := `
        INSERT INTO coursees (name, description, year, semester, program_id)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.ExecContext(ctx, query, name, description, year, semester, programID)
	if err != nil {
		return customerror.Trace("courseRepository: create", err)
	}
	return nil
}

// Busca um course pelo ID
func (r *sqlRepository) FindByID(ctx context.Context, id string) (*Course, error) {
	query := `
        SELECT id, name, description, year, semester, program_id, created_at, updated_at
        FROM coursees
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	course := &Course{}
	err := row.Scan(&course.ID, &course.Name, &course.Description, &course.Year, &course.Semester, &course.ProgramID, &course.CreatedAt, &course.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, customerror.Trace("courseRepository: findByID", err)
	}
	return course, nil
}

// Atualiza um course
func (r *sqlRepository) Update(ctx context.Context, id string, fields map[string]any) error {
	err := database.Update(ctx, r.db, "courses", id, fields)
	return err
}

// Remove um course pelo ID
func (r *sqlRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM courses WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
