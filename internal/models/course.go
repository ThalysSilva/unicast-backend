package models

import "time"

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
