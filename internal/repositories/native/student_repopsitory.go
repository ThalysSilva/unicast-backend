package native

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ThalysSilva/unicast-backend/internal/interfaces"
	"github.com/ThalysSilva/unicast-backend/internal/models/entities"
)

// Gerencia operações de banco para Student
type studentInstanceRepository struct {
	db *sql.DB
}

// Cia uma nova instância do repositório
func NewStudentRepository(db *sql.DB) interfaces.StudentRepository {
	return &studentInstanceRepository{db: db}
}

// Insere um novo estudante
func (r *studentInstanceRepository) Create(student *entities.Student) error {
	query := `
        INSERT INTO students (id, student_id, name, phone, email, annotation, status)
        VALUES ($1, $2, $3, $4, $5, $6, $9)
    `
	_, err := r.db.Exec(query, student.ID, student.StudentID, student.Name, student.Phone, student.Email, student.Annotation, student.Status)
	return err
}

// Busca um estudante pelo ID
func (r *studentInstanceRepository) FindByID(id string) (*entities.Student, error) {
	query := `
        SELECT id, student_id, name, phone, email, annotation, created_at, updated_at, status
        FROM students
        WHERE id = $1
    `
	row := r.db.QueryRow(query, id)

	student := &entities.Student{}
	var name, phone, email, annotation sql.NullString
	err := row.Scan(&student.ID, &student.StudentID, &name, &phone, &email, &annotation, &student.CreatedAt, &student.UpdatedAt, &student.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if name.Valid {
		student.Name = &name.String
	}
	if phone.Valid {
		student.Phone = &phone.String
	}
	if email.Valid {
		student.Email = &email.String
	}
	if annotation.Valid {
		student.Annotation = &annotation.String
	}
	return student, nil
}

// Busca estudantes por IDs
// Se a lista estiver vazia, retorna nil
func (r *studentInstanceRepository) FindByIDs(studentIds []string) ([]*entities.Student, error) {
	if len(studentIds) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(studentIds))
	args := make([]interface{}, len(studentIds))
	for i, id := range studentIds {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
			SELECT id, student_id, name, phone, email, annotation, created_at, updated_at, status
			FROM students
			WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := make([]*entities.Student, 0, len(studentIds))
	for rows.Next() {
		student := &entities.Student{}
		var name, phone, email, annotation sql.NullString
		err := rows.Scan(&student.ID, &student.StudentID, &name, &phone, &email, &annotation, &student.CreatedAt, &student.UpdatedAt, &student.Status)
		if err != nil {
			return nil, err
		}
		if name.Valid {
			student.Name = &name.String
		}
		if phone.Valid {
			student.Phone = &phone.String
		}
		if email.Valid {
			student.Email = &email.String
		}
		if annotation.Valid {
			student.Annotation = &annotation.String
		}
		students = append(students, student)
	}
	return students, nil
}

// Atualiza um estudante
func (r *studentInstanceRepository) Update(student *entities.Student) error {
	query := `
        UPDATE students
        SET student_id = $2, name = $3, phone = $4, email = $5, annotation = $6, status = $7
        WHERE id = $1
    `
	_, err := r.db.Exec(query, student.ID, student.StudentID, student.Name, student.Phone, student.Email, student.Annotation, student.Status)
	return err
}

// Remove um estudante pelo ID
func (r *studentInstanceRepository) Delete(id string) error {
	query := `DELETE FROM students WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
