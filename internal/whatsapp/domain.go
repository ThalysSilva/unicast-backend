package whatsapp

import (
	"database/sql"
	"time"
)

type Instance struct {
	ID         string    `json:"id"`
	Phone      string    `json:"phone" validate:"required"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
	UserID     string    `json:"-"`
	InstanceID string    `json:"instanceId"`
}

type Repository interface {
	Create(phone, userID, instanceID string) error
	FindByID(id string) (*Instance, error)
	FindByPhoneAndUserId(phone, userId string) (*Instance, error)
	Update(instance *Instance) error
	Delete(id string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return newNativeRepository(db)
}
