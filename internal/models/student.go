package models

import "time"

type StudentStatus string

// StudentStatus enum
const (
	StudentStatusActive    StudentStatus = "ACTIVE"
	StudentStatusCanceled  StudentStatus = "CANCELED"
	StudentStatusGraduated StudentStatus = "GRADUATED"
	StudentStatusLocked    StudentStatus = "LOCKED"
)

type Student struct {
	ID         string        `json:"id"`
	StudentID  string        `json:"studentId"`
	Name       *string       `json:"name"`
	Phone      *string       `json:"phone"`
	Email      *string       `json:"email" validate:"email"`
	Annotation *string       `json:"annotation"`
	CreatedAt  time.Time     `json:"-"`
	UpdatedAt  time.Time     `json:"-"`
	Status     StudentStatus `json:"status"`
}
