package course

import (
	"database/sql"
	"time"
)

type Course struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	Year        int       `json:"year" validate:"required"`
	Semester    int       `json:"semester" validate:"required"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
	ProgramID   string    `json:"-"`
}

type Repository interface {
	Create(course *Course) error
	FindByID(id string) (*Course, error)
	Update(course *Course) error
	Delete(id string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return newNativeRepository(db)
}
