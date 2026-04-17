package enrollment

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

// Gerencia operações de banco para Enrollment
type sqlRepository struct {
	db    database.DB
	sqlDB *sql.DB
}

func newSQLRepository(db *sql.DB) Repository {
	newDb := database.NewSQLTx(db)
	return &sqlRepository{
		db: newDb.DB,
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

// Insere uma nova matrícula
func (r *sqlRepository) Create(ctx context.Context, disciplineID, studentID string) error {
	query := `
        INSERT INTO enrollments (discipline_id, student_id)
        VALUES ($1, $2)
    `
	_, err := r.db.ExecContext(ctx, query, disciplineID, studentID)
	return err
}

// Busca uma matrícula pelo ID
func (r *sqlRepository) FindByID(ctx context.Context, id string) (*Enrollment, error) {
	query := `
        SELECT id, discipline_id, student_id, self_registration_completed_at, self_registration_count, created_at, updated_at
        FROM enrollments
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	return scanEnrollment(row)
}

func (r *sqlRepository) FindByDisciplineAndStudent(ctx context.Context, disciplineID, studentID string) (*Enrollment, error) {
	query := `
        SELECT id, discipline_id, student_id, self_registration_completed_at, self_registration_count, created_at, updated_at
        FROM enrollments
        WHERE discipline_id = $1 AND student_id = $2
    `
	row := r.db.QueryRowContext(ctx, query, disciplineID, studentID)

	return scanEnrollment(row)
}

func scanEnrollment(scanner interface {
	Scan(dest ...any) error
}) (*Enrollment, error) {
	enrollment := &Enrollment{}
	var selfRegistrationCompletedAt sql.NullTime
	err := scanner.Scan(
		&enrollment.ID,
		&enrollment.DisciplineID,
		&enrollment.StudentID,
		&selfRegistrationCompletedAt,
		&enrollment.SelfRegistrationCount,
		&enrollment.CreatedAt,
		&enrollment.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if selfRegistrationCompletedAt.Valid {
		completedAt := time.Time(selfRegistrationCompletedAt.Time)
		enrollment.SelfRegistrationCompletedAt = &completedAt
	}
	return enrollment, nil
}

// Atualiza uma matrícula
func (r *sqlRepository) Update(ctx context.Context, id string, fields map[string]any) error {
	err := database.Update(ctx, r.db, "enrollments", id, fields)
	return err
}

// Remove uma matrícula pelo ID
func (r *sqlRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM enrollments WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *sqlRepository) DeleteByDisciplineID(ctx context.Context, disciplineID string) error {
	query := `DELETE FROM enrollments WHERE discipline_id = $1`
	_, err := r.db.ExecContext(ctx, query, disciplineID)
	return err
}
