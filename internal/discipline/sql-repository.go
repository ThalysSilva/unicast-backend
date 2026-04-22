package discipline

import (
	"context"
	"database/sql"

	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

// Gerencia operações de banco para disciplinas
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

// Insere uma nova disciplina
func (r *sqlRepository) Create(ctx context.Context, name, description, programID string, year, semester int) error {
	query := `
        INSERT INTO disciplines (name, description, year, semester, program_id)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.ExecContext(ctx, query, name, description, year, semester, programID)
	if err != nil {
		return customerror.Trace("disciplineRepository: create", err)
	}
	return nil
}

// Busca uma disciplina pelo ID
func (r *sqlRepository) FindByID(ctx context.Context, id string) (*Discipline, error) {
	query := `
        SELECT id, name, description, year, semester, program_id, created_at, updated_at
        FROM disciplines
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	discipline := &Discipline{}
	err := row.Scan(&discipline.ID, &discipline.Name, &discipline.Description, &discipline.Year, &discipline.Semester, &discipline.ProgramID, &discipline.CreatedAt, &discipline.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, customerror.Trace("disciplineRepository: findByID", err)
	}
	return discipline, nil
}

func (r *sqlRepository) FindByIDWithUserOwnerID(ctx context.Context, id string) (*DisciplineWithOwnerID, error) {
	query := `
		SELECT disciplines.id, disciplines.name, disciplines.description, disciplines.year, disciplines.semester, disciplines.program_id, disciplines.created_at, disciplines.updated_at, ca.user_owner_id
		FROM disciplines
		JOIN programs p ON p.id = disciplines.program_id
		JOIN campuses ca ON ca.id = p.campus_id
		WHERE disciplines.id = $1
		
	`
	row := r.db.QueryRowContext(ctx, query, id)

	discipline := &DisciplineWithOwnerID{}
	err := row.Scan(&discipline.ID, &discipline.Name, &discipline.Description, &discipline.Year, &discipline.Semester, &discipline.ProgramID, &discipline.CreatedAt, &discipline.UpdatedAt, &discipline.UserOwnerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, customerror.Trace("disciplineRepository: findByID", err)
	}
	return discipline, nil
}

func (r *sqlRepository) FindByProgramID(ctx context.Context, programID string) ([]*Discipline, error) {
	query := `
		SELECT id, name, description, year, semester, program_id, created_at, updated_at
		FROM disciplines
		WHERE program_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, programID)
	if err != nil {
		return nil, customerror.Trace("disciplineRepository: findByProgramID", err)
	}
	defer rows.Close()

	var disciplines []*Discipline
	for rows.Next() {
		discipline := &Discipline{}
		err := rows.Scan(&discipline.ID, &discipline.Name, &discipline.Description, &discipline.Year, &discipline.Semester, &discipline.ProgramID, &discipline.CreatedAt, &discipline.UpdatedAt)
		if err != nil {
			return nil, customerror.Trace("disciplineRepository: findByProgramID", err)
		}
		disciplines = append(disciplines, discipline)
	}
	return disciplines, nil
}

func (r *sqlRepository) FindByUserOwnerID(ctx context.Context, userOwnerID string) ([]*Discipline, error) {
	query := `
		SELECT c.id, c.name, c.description, c.year, c.semester, c.program_id, c.created_at, c.updated_at
		FROM disciplines c
		JOIN programs p ON p.id = c.program_id
		JOIN campuses ca ON ca.id = p.campus_id
		WHERE ca.user_owner_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, userOwnerID)
	if err != nil {
		return nil, customerror.Trace("disciplineRepository: findByUserOwnerID", err)
	}
	defer rows.Close()

	var disciplines []*Discipline
	for rows.Next() {
		discipline := &Discipline{}
		err := rows.Scan(&discipline.ID, &discipline.Name, &discipline.Description, &discipline.Year, &discipline.Semester, &discipline.ProgramID, &discipline.CreatedAt, &discipline.UpdatedAt)
		if err != nil {
			return nil, customerror.Trace("disciplineRepository: findByUserOwnerID", err)
		}
		disciplines = append(disciplines, discipline)
	}
	return disciplines, nil
}

func (r *sqlRepository) FindByNameAndProgramID(ctx context.Context, name, programID string) (*Discipline, error) {
	query := `
		SELECT id, name, description, year, semester, program_id, created_at, updated_at
		FROM disciplines
		WHERE name = $1 AND program_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, name, programID)

	discipline := &Discipline{}
	err := row.Scan(&discipline.ID, &discipline.Name, &discipline.Description, &discipline.Year, &discipline.Semester, &discipline.ProgramID, &discipline.CreatedAt, &discipline.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, customerror.Trace("disciplineRepository: findByNameAndProgramID", err)
	}
	return discipline, nil
}

// Atualiza uma disciplina
func (r *sqlRepository) Update(ctx context.Context, id string, fields map[string]any) error {

	err := database.Update(ctx, r.db, "disciplines", id, fields)
	return err
}

// Remove uma disciplina pelo ID
func (r *sqlRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM disciplines WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
