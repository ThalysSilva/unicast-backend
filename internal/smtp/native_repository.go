package smtp

import (
	"database/sql"
)

// Gerencia operações de banco para SmtpInstance
type nativeRepository = repository

// Cria uma nova instância do repositório
func newNativeRepository(db *sql.DB) Repository {
	return &nativeRepository{db: db}
}

// Insere uma nova instância SMTP
func (r *nativeRepository) Create(instance *SmtpInstance) error {
	query := `
        INSERT INTO smtp_instances (id, host, port, email, password, iv, user_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := r.db.Exec(query, instance.ID, instance.Host, instance.Port, instance.Email, instance.Password, instance.IV, instance.UserID)
	return err
}

// Busca uma instância SMTP pelo ID
func (r *nativeRepository) FindByID(id string) (*SmtpInstance, error) {
	query := `
        SELECT id, host, port, email, password, iv, created_at, updated_at, user_id
        FROM smtp_instances
        WHERE id = $1
    `
	row := r.db.QueryRow(query, id)

	instance := &SmtpInstance{}
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
func (r *nativeRepository) Update(instance *SmtpInstance) error {
	query := `
        UPDATE smtp_instances
        SET host = $2, port = $3, email = $4, password = $5, iv = $6,  user_id = $7
        WHERE id = $1
    `
	_, err := r.db.Exec(query, instance.ID, instance.Host, instance.Port, instance.Email, instance.Password, instance.IV, instance.UserID)
	return err
}

// Remove uma instância SMTP pelo ID
func (r *nativeRepository) Delete(id string) error {
	query := `DELETE FROM smtp_instances WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
