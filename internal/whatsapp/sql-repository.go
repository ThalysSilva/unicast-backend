package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
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

func (r *sqlRepository) Create(ctx context.Context, phone, userID, instanceID string) error {
	query := `
		INSERT INTO whatsapp_instances (phone, user_id, instance_id)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query, phone, userID, instanceID)
	return err
}

func (r *sqlRepository) FindByID(ctx context.Context, id string) (*Instance, error) {
	query := `
		SELECT id, phone, created_at, updated_at, user_id, instance_id
		FROM whatsapp_instances
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	instance := &Instance{}
	err := row.Scan(&instance.ID, &instance.Phone, &instance.CreatedAt, &instance.UpdatedAt, &instance.UserID, &instance.InstanceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("falha ao buscar instância %s: %w", id, err)
	}
	return instance, nil
}

func (r *sqlRepository) FindByPhoneAndUserId(ctx context.Context, phone, userId string) (*Instance, error) {
	query := `
		SELECT id, phone, created_at, updated_at, user_id, instance_id
		FROM whatsapp_instances
		WHERE phone = $1 AND user_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, phone, userId)

	instance := &Instance{}
	err := row.Scan(&instance.ID, &instance.Phone, &instance.CreatedAt, &instance.UpdatedAt, &instance.UserID, &instance.InstanceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("falha ao buscar instância por telefone %s e usuário %s: %w", phone, userId, err)
	}
	return instance, nil
}

func (r *sqlRepository) FindAllByUserId(ctx context.Context, userId string) ([]*Instance, error) {
	query := `
		SELECT id, phone, created_at, updated_at, user_id, instance_id
		FROM whatsapp_instances
		WHERE user_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar instâncias: %w", err)
	}
	defer rows.Close()

	var instances []*Instance
	for rows.Next() {
		instance := &Instance{}
		err := rows.Scan(&instance.ID, &instance.Phone, &instance.CreatedAt, &instance.UpdatedAt, &instance.UserID, &instance.InstanceID)
		if err != nil {
			return nil, fmt.Errorf("falha ao escanear instância: %w", err)
		}
		instances = append(instances, instance)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar instâncias: %w", err)
	}
	return instances, nil
}

// Atualiza os campos de uma instância no banco de dados. Campos não fornecidos não serão atualizados.
// Campos: phone, user_id, instance_id
func (r *sqlRepository) Update(ctx context.Context, id string, fields map[string]any) error {
	err := database.Update(ctx, r.db, "whatsapp_instances", id, fields)
	if err != nil {
		return err
	}
	return nil
}

func (r *sqlRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM whatsapp_instances WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
