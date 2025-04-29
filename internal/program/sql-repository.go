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

func (r *sqlRepository) FindByIDWithUserOwnerID(ctx context.Context, id string) (*ProgramWithUserOwnerID, error) {
	query := `
				SELECT p.id, p.name, p.description, p.created_at, p.updated_at, p.campus_id, p.active, u.id
				FROM programs p
				JOIN campus ca ON p.campus_id = c.id
				JOIN users u ON ca.user_owner_id = u.id
				WHERE p.id = $1
		`

	row := r.db.QueryRowContext(ctx, query, id)

	program := &ProgramWithUserOwnerID{}
	err := row.Scan(&program.ID, &program.Name, &program.Description, &program.CreatedAt, &program.UpdatedAt, &program.CampusID, &program.Active, &program.UserOwnerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return program, nil
}

func (r *sqlRepository) FindByCampusID(ctx context.Context, campusID string) ([]*Program, error) {
	query := `
				SELECT id, name, description, created_at, updated_at, campus_id, active
				FROM programs
				WHERE campus_id = $1
		`

	rows, err := r.db.QueryContext(ctx, query, campusID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var programs []*Program
	for rows.Next() {
		program := &Program{}
		err := rows.Scan(&program.ID, &program.Name, &program.Description, &program.CreatedAt, &program.UpdatedAt, &program.CampusID, &program.Active)
		if err != nil {
			return nil, err
		}
		programs = append(programs, program)
	}
	return programs, nil
}

func (r *sqlRepository) FindByNameAndCampusID(ctx context.Context, name, campusID string) (*Program, error) {
	query := `
				SELECT id, name, description, created_at, updated_at, campus_id, active
				FROM programs
				WHERE name = $1 AND campus_id = $2
		`
	row := r.db.QueryRowContext(ctx, query, name, campusID)

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
