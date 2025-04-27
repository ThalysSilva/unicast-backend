package campus

import (
	"context"
	"database/sql"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

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

// Insere um novo campus
func (r *sqlRepository) Create(ctx context.Context, name, description, userOwnerID string) error {
	query := `
        INSERT INTO campuses (name, description, user_owner_id)
        VALUES ($1, $2, $3)
    `
	_, err := r.db.ExecContext(ctx, query, name, description, userOwnerID)
	return err
}

// Busca um campus pelo ID
func (r *sqlRepository) FindByID(ctx context.Context, id string) (*Campus, error) {
	query := `
        SELECT id, name, description, created_at, updated_at, user_owner_id
        FROM campuses
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	campus := &Campus{}
	err := row.Scan(&campus.ID, &campus.Name, &campus.Description, &campus.CreatedAt, &campus.UpdatedAt, &campus.UserOwnerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return campus, nil
}

func (r *sqlRepository) FindByUserOwnerId(ctx context.Context, userOwnerID string) ([]*Campus, error) {
	query := `
				SELECT id, name, description, created_at, updated_at, user_owner_id
				FROM campuses
				WHERE user_owner_id = $1
		`
	rows, err := r.db.QueryContext(ctx, query, userOwnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campuses []*Campus
	for rows.Next() {
		campus := &Campus{}
		err := rows.Scan(&campus.ID, &campus.Name, &campus.Description, &campus.CreatedAt, &campus.UpdatedAt, &campus.UserOwnerID)
		if err != nil {
			return nil, err
		}
		campuses = append(campuses, campus)
	}
	return campuses, nil
}

func (r *sqlRepository) Update(ctx context.Context, id string, fields map[string]any) error {
	err := database.Update(ctx, r.db, "campuses", id, fields)
	return err
}

// Remove um campus pelo ID
func (r *sqlRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM campuses WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
