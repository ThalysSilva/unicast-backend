package repositories

import (
	"database/sql"
	"unicast-api/internal/models/entities"
)

// whatsAppInstanceRepository gerencia operações de banco para WhatsAppInstance
type whatsAppInstanceRepository struct {
	db *sql.DB
}

// Cria uma nova instância do repositório
func NewWhatsAppInstanceRepository(db *sql.DB) WhatsAppRepository {
	return &whatsAppInstanceRepository{db: db}
}

// Insere uma nova instância de WhatsApp
func (r *whatsAppInstanceRepository) Create(instance *entities.WhatsAppInstance) error {
	query := `
        INSERT INTO whatsapp_instances (id, phone, user_id, instance_id)
        VALUES ($1, $2, $3, $4)
    `
	_, err := r.db.Exec(query, instance.ID, instance.Phone, instance.UserID, instance.InstanceID)
	return err
}

// FindByID busca uma instância de WhatsApp pelo ID
func (r *whatsAppInstanceRepository) FindByID(id string) (*entities.WhatsAppInstance, error) {
	query := `
        SELECT id, phone, created_at, updated_at, user_id, instance_id
        FROM whatsapp_instances
        WHERE id = $1
    `
	row := r.db.QueryRow(query, id)

	instance := &entities.WhatsAppInstance{}
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
func (r *whatsAppInstanceRepository) Update(instance *entities.WhatsAppInstance) error {
	query := `
        UPDATE whatsapp_instances
        SET phone = $2, user_id = $3, instance_id = $4
        WHERE id = $1
    `
	_, err := r.db.Exec(query, instance.ID, instance.Phone, instance.UserID, instance.InstanceID)
	return err
}

// Remove uma instância de WhatsApp pelo ID
func (r *whatsAppInstanceRepository) Delete(id string) error {
	query := `DELETE FROM whatsapp_instances WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
