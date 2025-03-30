package repositories

import (
	"database/sql"
	"github.com/ThalysSilva/unicast-backend/internal/models/entities"
)

// Gerencia operações de banco para SmtpInstance
type smtpInstanceRepository struct {
	db *sql.DB
}

// Cria uma nova instância do repositório
func NewSmtpInstanceRepository(db *sql.DB) SmtpRepository {
	return &smtpInstanceRepository{db: db}
}

// Insere uma nova instância SMTP
func (r *smtpInstanceRepository) Create(instance *entities.SmtpInstance) error {
	query := `
        INSERT INTO smtp_instances (id, host, port, email, password, iv, user_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := r.db.Exec(query, instance.ID, instance.Host, instance.Port, instance.Email, instance.Password, instance.IV, instance.UserID)
	return err
}

// Busca uma instância SMTP pelo ID
func (r *smtpInstanceRepository) FindByID(id string) (*entities.SmtpInstance, error) {
	query := `
        SELECT id, host, port, email, password, iv, created_at, updated_at, user_id
        FROM smtp_instances
        WHERE id = $1
    `
	row := r.db.QueryRow(query, id)

	instance := &entities.SmtpInstance{}
	err := row.Scan(&instance.ID, &instance.Host, &instance.Port, &instance.Email, &instance.Password, &instance.IV, &instance.CreatedAt, &instance.UpdatedAt, &instance.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return instance, nil
}

// Atualiza uma instância SMTP
func (r *smtpInstanceRepository) Update(instance *entities.SmtpInstance) error {
	query := `
        UPDATE smtp_instances
        SET host = $2, port = $3, email = $4, password = $5, iv = $6,  user_id = $7
        WHERE id = $1
    `
	_, err := r.db.Exec(query, instance.ID, instance.Host, instance.Port, instance.Email, instance.Password, instance.IV, instance.UserID)
	return err
}

// Remove uma instância SMTP pelo ID
func (r *smtpInstanceRepository) Delete(id string) error {
	query := `DELETE FROM smtp_instances WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
