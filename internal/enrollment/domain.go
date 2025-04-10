package enrollment

import (
	"database/sql"
	"time"
)

type Enrollment struct {
	ID        string    `json:"id"`
	CourseID  string    `json:"courseId"`
	StudentID string    `json:"studentId"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Repository interface {
	Create(enrollment *Enrollment) error
	FindByID(id string) (*Enrollment, error)
	Update(enrollment *Enrollment) error
	Delete(id string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return newNativeRepository(db)
}
