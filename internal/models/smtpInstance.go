package models

import "time"

type SmtpInstance struct {
	ID        string    `json:"id"`
	Host      string    `json:"host" validate:"required"`
	Port      int       `json:"port" validate:"required"`
	Email     string    `json:"email" validate:"required"`
	Password  string    `json:"password" validate:"required"`
	IV        string    `json:"iv" validate:"required"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	UserID    string    `json:"-"`
}
