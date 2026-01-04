package whatsapp

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Instance struct {
	ID         string    `json:"id"`
	Phone      string    `json:"phone" validate:"required"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
	UserID     string    `json:"-"`
	InstanceID string    `json:"instanceId"`
}

// SendText envia uma mensagem de texto via Evolution API usando a instância informada.
func SendText(instanceID, number, text string) error {
	return sendEvolutionText(instanceID, number, text)
}

// Repository define operações para instâncias WhatsApp.
type Repository interface {
	database.Transactional
	Create(ctx context.Context, phone, userID, instanceID string) error
	FindByID(ctx context.Context, id string) (*Instance, error)
	FindByPhoneAndUserId(ctx context.Context, phone, userId string) (*Instance, error)
	FindAllByUserId(ctx context.Context, userId string) ([]*Instance, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, id string) error
}

func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
