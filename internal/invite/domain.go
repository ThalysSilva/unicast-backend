package invite

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Invite struct {
	ID        string     `json:"id"`
	CourseID  string     `json:"courseId"`
	Code      string     `json:"code"`
	ExpiresAt *time.Time `json:"expiresAt"`
	Active    bool       `json:"active"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
}

type Repository interface {
	database.Transactional
	Create(ctx context.Context, courseID, code string, expiresAt *time.Time) error
	FindByID(ctx context.Context, id string) (*Invite, error)
	FindByCode(ctx context.Context, code string) (*Invite, error)
	FindLatestByCourseID(ctx context.Context, courseID string) (*Invite, error)
	FindByCourseID(ctx context.Context, courseID string) ([]*Invite, error)
	Delete(ctx context.Context, id string) error
}

func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
