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
func (r *sqlRepository) Create(ctx context.Context, userOwnerID, studentID string, name, phone, email, annotation *string, noPhone bool, status StudentStatus) error {
	query := `
        INSERT INTO students (student_id, name, phone, no_phone, email, annotation, status, consent, user_owner_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
	_, err := r.db.ExecContext(ctx, query, studentID, name, phone, noPhone, email, annotation, status, false, userOwnerID)
	return err
}

// Busca um estudante pelo ID
func (r *sqlRepository) FindByID(ctx context.Context, id, userOwnerID string) (*Student, error) {
	query := `
        SELECT id, student_id, name, phone, no_phone, email, annotation, consent,
               COALESCE((
                 SELECT MAX(created_at) FROM message_logs ml
                 WHERE ml.student_id = students.id AND ml.channel = 'EMAIL' AND ml.success = false
               ), '-infinity'::timestamptz) >
               COALESCE((
                 SELECT MAX(created_at) FROM message_logs ml
                 WHERE ml.student_id = students.id AND ml.channel = 'EMAIL' AND ml.success = true
               ), '-infinity'::timestamptz) AS email_delivery_issue,
               COALESCE((
                 SELECT MAX(created_at) FROM message_logs ml
                 WHERE ml.student_id = students.id AND ml.channel = 'WHATSAPP' AND ml.success = false
               ), '-infinity'::timestamptz) >
               COALESCE((
                 SELECT MAX(created_at) FROM message_logs ml
                 WHERE ml.student_id = students.id AND ml.channel = 'WHATSAPP' AND ml.success = true
               ), '-infinity'::timestamptz) AS whatsapp_delivery_issue,
               created_at, updated_at, status, user_owner_id
        FROM students
        WHERE id = $1 AND user_owner_id = $2
    `
	row := r.db.QueryRowContext(ctx, query, id, userOwnerID)

	student, err := scanStudent(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return student, nil
}

func (r *sqlRepository) FindByStudentID(ctx context.Context, studentID, userOwnerID string) (*Student, error) {
	query := `
        SELECT id, student_id, name, phone, no_phone, email, annotation, consent,
               COALESCE((
                 SELECT MAX(created_at) FROM message_logs ml
                 WHERE ml.student_id = students.id AND ml.channel = 'EMAIL' AND ml.success = false
               ), '-infinity'::timestamptz) >
               COALESCE((
                 SELECT MAX(created_at) FROM message_logs ml
                 WHERE ml.student_id = students.id AND ml.channel = 'EMAIL' AND ml.success = true
               ), '-infinity'::timestamptz) AS email_delivery_issue,
               COALESCE((
                 SELECT MAX(created_at) FROM message_logs ml
                 WHERE ml.student_id = students.id AND ml.channel = 'WHATSAPP' AND ml.success = false
               ), '-infinity'::timestamptz) >
               COALESCE((
                 SELECT MAX(created_at) FROM message_logs ml
                 WHERE ml.student_id = students.id AND ml.channel = 'WHATSAPP' AND ml.success = true
               ), '-infinity'::timestamptz) AS whatsapp_delivery_issue,
               created_at, updated_at, status, user_owner_id
        FROM students
        WHERE student_id = $1 AND user_owner_id = $2
    `
	row := r.db.QueryRowContext(ctx, query, studentID, userOwnerID)

	student, err := scanStudent(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return student, nil
}

func (r *sqlRepository) FindByFilters(ctx context.Context, filters map[string]string) ([]*Student, error) {
	query, args := buildFilteredStudentsQuery(filters)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := make([]*Student, 0)
	for rows.Next() {
		student, err := scanStudent(rows)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}
	return students, nil
}

func buildFilteredStudentsQuery(filters map[string]string) (string, []any) {
	query := `
		SELECT DISTINCT s.id, s.student_id, s.name, s.phone, s.no_phone, s.email, s.annotation, s.consent,
		       COALESCE((
		         SELECT MAX(created_at) FROM message_logs ml
		         WHERE ml.student_id = s.id AND ml.channel = 'EMAIL' AND ml.success = false
		       ), '-infinity'::timestamptz) >
		       COALESCE((
		         SELECT MAX(created_at) FROM message_logs ml
		         WHERE ml.student_id = s.id AND ml.channel = 'EMAIL' AND ml.success = true
		       ), '-infinity'::timestamptz) AS email_delivery_issue,
		       COALESCE((
		         SELECT MAX(created_at) FROM message_logs ml
		         WHERE ml.student_id = s.id AND ml.channel = 'WHATSAPP' AND ml.success = false
		       ), '-infinity'::timestamptz) >
		       COALESCE((
		         SELECT MAX(created_at) FROM message_logs ml
		         WHERE ml.student_id = s.id AND ml.channel = 'WHATSAPP' AND ml.success = true
		       ), '-infinity'::timestamptz) AS whatsapp_delivery_issue,
		       s.created_at, s.updated_at, s.status, s.user_owner_id
		FROM students s
	`

	needsAcademicJoin := filters["discipline"] != "" ||
		filters["program"] != "" ||
		filters["campus"] != ""

	if needsAcademicJoin {
		query += `
			JOIN enrollments e ON e.student_id = s.id
			JOIN disciplines d ON d.id = e.discipline_id
			JOIN programs p ON p.id = d.program_id
			JOIN campuses ca ON ca.id = p.campus_id
		`
	}

	whereClause, args := buildWhereClause(filters)
	return query + whereClause, args
}

// Busca estudantes por IDs
// Se a lista estiver vazia, retorna nil
func (r *sqlRepository) FindByIDs(ctx context.Context, userOwnerID string, studentIds []string) ([]*Student, error) {
	if len(studentIds) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(studentIds))
	args := make([]interface{}, 0, len(studentIds)+1)
	for i, id := range studentIds {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}
	args = append([]interface{}{userOwnerID}, args...)

	query := fmt.Sprintf(`
			SELECT id, student_id, name, phone, no_phone, email, annotation, consent,
			       COALESCE((
			         SELECT MAX(created_at) FROM message_logs ml
			         WHERE ml.student_id = students.id AND ml.channel = 'EMAIL' AND ml.success = false
			       ), '-infinity'::timestamptz) >
			       COALESCE((
			         SELECT MAX(created_at) FROM message_logs ml
			         WHERE ml.student_id = students.id AND ml.channel = 'EMAIL' AND ml.success = true
			       ), '-infinity'::timestamptz) AS email_delivery_issue,
			       COALESCE((
			         SELECT MAX(created_at) FROM message_logs ml
			         WHERE ml.student_id = students.id AND ml.channel = 'WHATSAPP' AND ml.success = false
			       ), '-infinity'::timestamptz) >
			       COALESCE((
			         SELECT MAX(created_at) FROM message_logs ml
			         WHERE ml.student_id = students.id AND ml.channel = 'WHATSAPP' AND ml.success = true
			       ), '-infinity'::timestamptz) AS whatsapp_delivery_issue,
			       created_at, updated_at, status, user_owner_id
			FROM students
			WHERE user_owner_id = $1 AND id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := make([]*Student, 0, len(studentIds))
	for rows.Next() {
		student, err := scanStudent(rows)
		if err != nil {
			return nil, err
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

func buildWhereClause(filters map[string]string) (string, []any) {
	if len(filters) == 0 {
		return "", nil
	}

	parts := make([]string, 0, len(filters))
	args := make([]any, 0, len(filters))

	i := 1
	for key, value := range filters {
		column, ok := studentFilterColumns[key]
		if !ok || value == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s = $%d", column, i))
		args = append(args, value)
		i++
	}

	if len(parts) == 0 {
		return "", nil
	}

	return " WHERE " + strings.Join(parts, " AND "), args
}

var studentFilterColumns = map[string]string{
	"discipline": "d.id",
	"program":    "p.id",
	"campus":     "ca.id",
	"user":       "s.user_owner_id",
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanStudent(scanner rowScanner) (*Student, error) {
	student := &Student{}
	var name, phone, email, annotation sql.NullString

	err := scanner.Scan(
		&student.ID,
		&student.StudentID,
		&name,
		&phone,
		&student.NoPhone,
		&email,
		&annotation,
		&student.Consent,
		&student.EmailDeliveryIssue,
		&student.WhatsAppDeliveryIssue,
		&student.CreatedAt,
		&student.UpdatedAt,
		&student.Status,
		&student.UserOwnerID,
	)
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

	return student, nil
}
