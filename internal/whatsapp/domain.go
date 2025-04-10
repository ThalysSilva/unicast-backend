package whatsapp

import (
	"database/sql"
	"time"
)

type WhatsAppInstance struct {
	ID         string    `json:"id"`
	Phone      string    `json:"phone" validate:"required"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
	UserID     string    `json:"-"`
	InstanceID string    `json:"instanceId"`
}

type Repository interface {
	Create(instance *WhatsAppInstance) error
	FindByID(id string) (*WhatsAppInstance, error)
	Update(instance *WhatsAppInstance) error
	Delete(id string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return newNativeRepository(db)
}
