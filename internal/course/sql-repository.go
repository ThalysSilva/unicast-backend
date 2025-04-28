package course

import (
	"context"
	"database/sql"

	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

// Gerencia operações de banco para course
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

// Insere um novo course
func (r *sqlRepository) Create(ctx context.Context, name, description, programID string, year, semester int) error {
	query := `
        INSERT INTO courses (name, description, year, semester, program_id)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.ExecContext(ctx, query, name, description, year, semester, programID)
	if err != nil {
		return customerror.Trace("courseRepository: create", err)
	}
	return nil
}

// Busca um course pelo ID
func (r *sqlRepository) FindByID(ctx context.Context, id string) (*Course, error) {
	query := `
        SELECT id, name, description, year, semester, program_id, created_at, updated_at
        FROM courses
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	course := &Course{}
	err := row.Scan(&course.ID, &course.Name, &course.Description, &course.Year, &course.Semester, &course.ProgramID, &course.CreatedAt, &course.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, customerror.Trace("courseRepository: findByID", err)
	}
	return course, nil
}

func (r *sqlRepository) FindByIDWithUserOwnerID(ctx context.Context, id string) (*CourseWithOwnerID, error) {
	query := `
		SELECT id, name, description, year, semester, program_id, created_at, updated_at, ca.user_owner_id
		FROM courses
		JOIN programs p ON p.id = courses.program_id
		JOIN campuses ca ON ca.id = p.campus_id
		WHERE courses.id = $1
		
	`
	row := r.db.QueryRowContext(ctx, query, id)

	course := &CourseWithOwnerID{}
	err := row.Scan(&course.ID, &course.Name, &course.Description, &course.Year, &course.Semester, &course.ProgramID, &course.CreatedAt, &course.UpdatedAt, &course.UserOwnerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, customerror.Trace("courseRepository: findByID", err)
	}
	return course, nil
}

func (r *sqlRepository) FindByProgramID(ctx context.Context, programID string) ([]*Course, error) {
	query := `
		SELECT id, name, description, year, semester, program_id, created_at, updated_at
		FROM courses
		WHERE program_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, programID)
	if err != nil {
		return nil, customerror.Trace("courseRepository: findByProgramID", err)
	}
	defer rows.Close()

	var courses []*Course
	for rows.Next() {
		course := &Course{}
		err := rows.Scan(&course.ID, &course.Name, &course.Description, &course.Year, &course.Semester, &course.ProgramID, &course.CreatedAt, &course.UpdatedAt)
		if err != nil {
			return nil, customerror.Trace("courseRepository: findByProgramID", err)
		}
		courses = append(courses, course)
	}
	return courses, nil
}

func (r *sqlRepository) FindByNameAndProgramID(ctx context.Context, name, programID string) (*Course, error) {
	query := `
		SELECT id, name, description, year, semester, program_id, created_at, updated_at
		FROM courses
		WHERE name = $1 AND program_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, name, programID)

	course := &Course{}
	err := row.Scan(&course.ID, &course.Name, &course.Description, &course.Year, &course.Semester, &course.ProgramID, &course.CreatedAt, &course.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, customerror.Trace("courseRepository: findByNameAndProgramID", err)
	}
	return course, nil
}

// Atualiza um course
func (r *sqlRepository) Update(ctx context.Context, id string, fields map[string]any) error {

	err := database.Update(ctx, r.db, "courses", id, fields)
	return err
}

// Remove um course pelo ID
func (r *sqlRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM courses WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
