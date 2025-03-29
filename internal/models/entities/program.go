package entities

import "time"

type Program struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
	CampusID    string    `json:"-"`
	Active      bool      `json:"active"`
}
