package smtp

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type UpdateInstanceInput struct {
	Host     *string
	Port     *int
	Email    *string
	Password *string
	IV       *string
	UserID   *int
}

type sqlRepository struct {
	db    database.DB
	sqlDB *sql.DB
}

// Cria uma nova instância do repositório
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

// Insere uma nova instância SMTP
func (r *sqlRepository) Create(ctx context.Context, userID, email, password, host string, port int, iv []byte) error {

	query := `
        INSERT INTO smtp_instances (host, port, email, password, iv, user_id)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := r.db.ExecContext(ctx, query, host, port, email, password, iv, userID)
	return err
}

// Busca uma instância SMTP pelo ID
func (r *sqlRepository) FindByID(ctx context.Context, id string) (*Instance, error) {
	query := `
        SELECT id, host, port, email, password, iv, created_at, updated_at, user_id
        FROM smtp_instances
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	instance := &Instance{}
	err := row.Scan(&instance.ID, &instance.Host, &instance.Port, &instance.Email, &instance.Password, &instance.IV, &instance.CreatedAt, &instance.UpdatedAt, &instance.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return instance, nil
}

func (r *sqlRepository) GetInstances(ctx context.Context, userID string) ([]*Instance, error) {

	query := `
				SELECT id, host, port, email, password, iv, created_at, updated_at, user_id
				FROM smtp_instances
				WHERE user_id = $1
		`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var instances []*Instance
	for rows.Next() {
		instance := &Instance{}
		err := rows.Scan(&instance.ID, &instance.Host, &instance.Port, &instance.Email, &instance.Password, &instance.IV, &instance.CreatedAt, &instance.UpdatedAt, &instance.UserID)
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}
	return instances, nil
}


// Atualiza uma instância SMTP
func (r *sqlRepository) Update(ctx context.Context, id int, fields map[string]interface{}) error {

	if len(fields) == 0 {
		return fmt.Errorf("nenhum campo fornecido para atualização")
	}

	setters := make([]string, 0, len(fields))
	args := make([]interface{}, 0, len(fields)+1)
	args = append(args, id) // $1 é o id

	i := 2 // Começa em $2, pois $1 é o id
	for field, value := range fields {
		setters = append(setters, fmt.Sprintf("%s = $%d", field, i))
		args = append(args, value)
		i++
	}

	query := fmt.Sprintf("UPDATE smtp_instances SET %s WHERE id = $1", strings.Join(setters, ", "))
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("falha ao atualizar instância %d: %v", id, err)
	}

	return nil
}

// Remove uma instância SMTP pelo ID
func (r *sqlRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM smtp_instances WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
