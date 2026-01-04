package enrollment

import (
	"context"
	"database/sql"

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
func (r *sqlRepository) Create(ctx context.Context, courseID, studentID string) error {
	query := `
        INSERT INTO enrollments (course_id, student_id)
        VALUES ($1, $2)
    `
	_, err := r.db.ExecContext(ctx, query, courseID, studentID)
	return err
}

// Busca uma matrícula pelo ID
func (r *sqlRepository) FindByID(ctx context.Context, id string) (*Enrollment, error) {
	query := `
        SELECT id, course_id, student_id, created_at, updated_at
        FROM enrollments
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	enrollment := &Enrollment{}
	err := row.Scan(&enrollment.ID, &enrollment.CourseID, &enrollment.StudentID, &enrollment.CreatedAt, &enrollment.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return enrollment, nil
}

func (r *sqlRepository) FindByCourseAndStudent(ctx context.Context, courseID, studentID string) (*Enrollment, error) {
	query := `
        SELECT id, course_id, student_id, created_at, updated_at
        FROM enrollments
        WHERE course_id = $1 AND student_id = $2
    `
	row := r.db.QueryRowContext(ctx, query, courseID, studentID)

	enrollment := &Enrollment{}
	err := row.Scan(&enrollment.ID, &enrollment.CourseID, &enrollment.StudentID, &enrollment.CreatedAt, &enrollment.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
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
