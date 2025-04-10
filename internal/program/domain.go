package program

import (
	"database/sql"
	"time"
)

type Program struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
	CampusID    string    `json:"-"`
	Active      bool      `json:"active"`
}

type Repository interface {
	Create(program *Program) error
	FindByID(id string) (*Program, error)
	Update(program *Program) error
	Delete(id string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return newNativeRepository(db)
}
