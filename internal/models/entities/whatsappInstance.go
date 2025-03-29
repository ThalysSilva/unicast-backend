package entities

import "time"

type WhatsAppInstance struct {
	ID         string    `json:"id"`
	Phone      string    `json:"phone" validate:"required"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
	UserID     string    `json:"-"`
	InstanceID string    `json:"instanceId"`
}
