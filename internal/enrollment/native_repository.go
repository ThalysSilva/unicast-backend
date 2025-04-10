package enrollment

import (
	"database/sql"
)

// Gerencia operações de banco para Enrollment
type nativeRepository = repository

// Cria uma nova instância do repositório
func newNativeRepository(db *sql.DB) Repository {
	return &nativeRepository{db: db}
}

// Insere uma nova matrícula
func (r *nativeRepository) Create(enrollment *Enrollment) error {
	query := `
        INSERT INTO enrollments (id, course_id, student_id)
        VALUES ($1, $2, $3)
    `
	_, err := r.db.Exec(query, enrollment.ID, enrollment.CourseID, enrollment.StudentID)
	return err
}

// Busca uma matrícula pelo ID
func (r *nativeRepository) FindByID(id string) (*Enrollment, error) {
	query := `
        SELECT id, course_id, student_id, created_at, updated_at
        FROM enrollments
        WHERE id = $1
    `
	row := r.db.QueryRow(query, id)

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
func (r *nativeRepository) Update(enrollment *Enrollment) error {
	query := `
        UPDATE enrollments
        SET course_id = $2, student_id = $3
        WHERE id = $1
    `
	_, err := r.db.Exec(query, enrollment.ID, enrollment.CourseID, enrollment.StudentID)
	return err
}

// Remove uma matrícula pelo ID
func (r *nativeRepository) Delete(id string) error {
	query := `DELETE FROM enrollments WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
