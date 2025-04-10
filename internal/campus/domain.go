package campus

import (
	"database/sql"
	"time"
)

type Campus struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
	UserOwnerID string    `json:"-"`
}

type repository struct {
	db *sql.DB
}

type Repository interface {
	Create(program *Campus) error
	FindByID(id string) (*Campus, error)
	Update(program *Campus) error
	Delete(id string) error
}

func NewRepository(db *sql.DB) Repository {
	return newNativeRepository(db)
}
