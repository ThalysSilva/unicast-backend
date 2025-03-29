package entities

import "time"

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name" validate:"required"`
	Email        string    `json:"email" validate:"required,email"`
	Password     string    `json:"-"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
	Salt         []byte    `json:"-"`
	RefreshToken *string   `json:"-"`
}
