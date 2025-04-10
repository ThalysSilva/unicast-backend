package whatsapp

import (
	"database/sql"
	"fmt"
)

// nativeRepository gerencia operações de banco para WhatsAppInstance
type nativeRepository = repository

// Cria uma nova instância do repositório
func newNativeRepository(db *sql.DB) Repository {
	return &nativeRepository{db: db}
}

// Insere uma nova instância de WhatsApp
func (r *nativeRepository) Create(phone, userID, instanceID string) error {
	query := `
        INSERT INTO whatsapp_instances (phone, user_id, instance_id)
        VALUES ($1, $2, $3)
    `
	_, err := r.db.Exec(query, phone, userID, instanceID)
	return err
}

// Busca uma instância de WhatsApp pelo ID
func (r *nativeRepository) FindByID(id string) (*Instance, error) {
	query := `
        SELECT id, phone, created_at, updated_at, user_id, instance_id
        FROM whatsapp_instances
        WHERE id = $1
    `
	row := r.db.QueryRow(query, id)

	instance := &Instance{}
	err := row.Scan(&instance.ID, &instance.Phone, &instance.CreatedAt, &instance.UpdatedAt, &instance.UserID, &instance.InstanceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return instance, nil
}

func (r *nativeRepository) FindByPhoneAndUserId(phone, userId string) (*Instance, error) {
	query := `
				SELECT id, phone, created_at, updated_at, user_id, instance_id
				FROM whatsapp_instances
				WHERE phone = $1 AND user_id = $2
		`
	row := r.db.QueryRow(query, phone, userId)
	fmt.Println("!!! passando aqui")

	instance := &Instance{}
	err := row.Scan(&instance.ID, &instance.Phone, &instance.CreatedAt, &instance.UpdatedAt, &instance.UserID, &instance.InstanceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return instance, nil
}

// Atualiza uma instância de WhatsApp
func (r *nativeRepository) Update(instance *Instance) error {
	query := `
        UPDATE whatsapp_instances
        SET phone = $2, user_id = $3, instance_id = $4
        WHERE id = $1
    `
	_, err := r.db.Exec(query, instance.ID, instance.Phone, instance.UserID, instance.InstanceID)
	return err
}

// Remove uma instância de WhatsApp pelo ID
func (r *nativeRepository) Delete(id string) error {
	query := `DELETE FROM whatsapp_instances WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
