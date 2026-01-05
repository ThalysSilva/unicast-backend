package enrollment

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Enrollment struct {
	ID        string    `json:"id"`
	CourseID  string    `json:"courseId"`
	StudentID string    `json:"studentId"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Repository interface {
	database.Transactional
	Create(ctx context.Context, courseID, studentID string) error
	FindByID(ctx context.Context, cid string) (*Enrollment, error)
	FindByCourseAndStudent(ctx context.Context, courseID, studentID string) (*Enrollment, error)
	DeleteByCourseID(ctx context.Context, courseID string) error
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, cid string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
