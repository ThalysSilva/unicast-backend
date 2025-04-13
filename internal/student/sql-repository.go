package student

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

// Gerencia operações de banco para Student
type sqlRepository struct {
	db    database.DB
	sqlDB *sql.DB
}

// Cia uma nova instância do repositório
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

// Insere um novo estudante
func (r *sqlRepository) Create(ctx context.Context, studentID string, name, phone, email, annotation *string, status StudentStatus) error {
	query := `
        INSERT INTO students (student_id, name, phone, email, annotation, status)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := r.db.ExecContext(ctx, query, studentID, name, phone, email, annotation, status)
	return err
}

// Busca um estudante pelo ID
func (r *sqlRepository) FindByID(ctx context.Context, id string) (*Student, error) {
	query := `
        SELECT id, student_id, name, phone, email, annotation, created_at, updated_at, status
        FROM students
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	student := &Student{}
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
func (r *sqlRepository) FindByIDs(ctx context.Context, studentIds []string) ([]*Student, error) {
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

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := make([]*Student, 0, len(studentIds))
	for rows.Next() {
		student := &Student{}
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
func (r *sqlRepository) Update(ctx context.Context, id string, fields map[string]any) error {
	err := database.Update(ctx, r.db, "students", id, fields)
	return err
}

// Remove um estudante pelo ID
func (r *sqlRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM students WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
