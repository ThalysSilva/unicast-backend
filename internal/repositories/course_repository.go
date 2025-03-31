package repositories

import (
	"database/sql"
	"github.com/ThalysSilva/unicast-backend/internal/models/entities"
)

// Gerencia operações de banco para course
type courseInstanceRepository struct {
	db *sql.DB
}

// Cria uma nova instância do repositório
func NewCourseRepository(db *sql.DB) CourseRepository {
	return &courseInstanceRepository{db: db}
}

// Insere um novo course
func (r *courseInstanceRepository) Create(course *entities.Course) error {
	query := `
        INSERT INTO coursees (id, name, description, year, semester, program_id)
        VALUES ($1, $2, $3, $4)
    `
	_, err := r.db.Exec(query, course.ID, course.Name, course.Description, course.Year, course.Semester, course.ProgramID)
	if err != nil {
		return Trace("courseRepository: create", err)
	}
	return nil
}

// Busca um course pelo ID
func (r *courseInstanceRepository) FindByID(id string) (*entities.Course, error) {
	query := `
        SELECT id, name, description, year, semester, program_id, created_at, updated_at
        FROM coursees
        WHERE id = $1
    `
	row := r.db.QueryRow(query, id)

	course := &entities.Course{}
	err := row.Scan(&course.ID, &course.Name, &course.Description, &course.Year, &course.Semester, &course.ProgramID, &course.CreatedAt, &course.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, Trace("courseRepository: findByID", err)
	}
	return course, nil
}

// Atualiza um course
func (r *courseInstanceRepository) Update(course *entities.Course) error {
	query := `
        UPDATE coursees
        SET name = $2, description = $3, year = $4, semester = $5, program_id = $6
        WHERE id = $1
    `
	_, err := r.db.Exec(query, course.ID, course.Name, course.Description, course.Year, course.Semester, course.ProgramID)
	if err != nil {
		return Trace("courseRepository: update", err)
	}
	return err
}

// Remove um course pelo ID
func (r *courseInstanceRepository) Delete(id string) error {
	query := `DELETE FROM courses WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
