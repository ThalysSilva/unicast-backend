package invite

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

// Gerencia operações de banco para Invite
type sqlRepository struct {
	db    database.DB
	sqlDB *sql.DB
}

func newSQLRepository(db *sql.DB) Repository {
	newDb := database.NewSQLTx(db)
	return &sqlRepository{
		db:    newDb.DB,
		sqlDB: db,
	}
}

func (r *sqlRepository) WithTransaction(tx any) any {
	return &sqlRepository{
		db:    database.NewSQLTx(nil).WithSQLTransaction(tx).DB,
		sqlDB: r.sqlDB,
	}
}

func (r *sqlRepository) TransactionBackend() any {
	return r.sqlDB
}

func (r *sqlRepository) Create(ctx context.Context, courseID, code string, expiresAt *time.Time) error {
	query := `
        INSERT INTO invites (course_id, code, expires_at)
        VALUES ($1, $2, $3)
    `
	_, err := r.db.ExecContext(ctx, query, courseID, code, expiresAt)
	if err != nil {
		return customerror.Trace("inviteRepository: create", err)
	}
	return nil
}

func (r *sqlRepository) FindByID(ctx context.Context, id string) (*Invite, error) {
	query := `
        SELECT id, course_id, code, expires_at, active, created_at, updated_at
        FROM invites
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	return scanInvite(row, "inviteRepository: findByID")
}

func (r *sqlRepository) FindByCode(ctx context.Context, code string) (*Invite, error) {
	query := `
        SELECT id, course_id, code, expires_at, active, created_at, updated_at
        FROM invites
        WHERE UPPER(code) = UPPER($1)
    `
	row := r.db.QueryRowContext(ctx, query, strings.TrimSpace(code))

	return scanInvite(row, "inviteRepository: findByCode")
}

func (r *sqlRepository) FindLatestByCourseID(ctx context.Context, courseID string) (*Invite, error) {
	query := `
        SELECT id, course_id, code, expires_at, active, created_at, updated_at
        FROM invites
        WHERE course_id = $1
        ORDER BY created_at DESC
        LIMIT 1
    `
	row := r.db.QueryRowContext(ctx, query, courseID)

	return scanInvite(row, "inviteRepository: findLatestByCourseID")
}

func scanInvite(scanner interface {
	Scan(dest ...any) error
}, trace string) (*Invite, error) {
	invite := &Invite{}
	var expiresAt sql.NullTime
	err := scanner.Scan(&invite.ID, &invite.CourseID, &invite.Code, &expiresAt, &invite.Active, &invite.CreatedAt, &invite.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, customerror.Trace(trace, err)
	}
	if expiresAt.Valid {
		invite.ExpiresAt = &expiresAt.Time
	}
	return invite, nil
}

func (r *sqlRepository) FindByCourseID(ctx context.Context, courseID string) ([]*Invite, error) {
	query := `
        SELECT id, course_id, code, expires_at, active, created_at, updated_at
        FROM invites
        WHERE course_id = $1
        ORDER BY created_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query, courseID)
	if err != nil {
		return nil, customerror.Trace("inviteRepository: findByCourseID", err)
	}
	defer rows.Close()

	invites := make([]*Invite, 0)
	for rows.Next() {
		invite := &Invite{}
		var expiresAt sql.NullTime
		err := rows.Scan(&invite.ID, &invite.CourseID, &invite.Code, &expiresAt, &invite.Active, &invite.CreatedAt, &invite.UpdatedAt)
		if err != nil {
			return nil, customerror.Trace("inviteRepository: findByCourseID", err)
		}
		if expiresAt.Valid {
			invite.ExpiresAt = &expiresAt.Time
		}
		invites = append(invites, invite)
	}
	return invites, nil
}

func (r *sqlRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM invites WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return customerror.Trace("inviteRepository: delete", err)
	}
	return nil
}
