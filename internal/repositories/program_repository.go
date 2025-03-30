package repositories

import (
	"database/sql"
	"github.com/ThalysSilva/unicast-backend/internal/models/entities"
)

// Gerencia operações de banco para Program
type programInstanceRepository struct {
	db *sql.DB
}

// Cria uma nova instância do repositório
func NewProgramRepository(db *sql.DB) ProgramRepository {
	return &programInstanceRepository{db: db}
}

// Insere um novo programa
func (r *programInstanceRepository) Create(program *entities.Program) error {
	query := `
        INSERT INTO programs (id, name, description, campus_id, active)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.Exec(query, program.ID, program.Name, program.Description, program.CampusID, program.Active)
	return err
}

// FindByID busca um programa pelo ID
func (r *programInstanceRepository) FindByID(id string) (*entities.Program, error) {
	query := `
        SELECT id, name, description, created_at, updated_at, campus_id, active
        FROM programs
        WHERE id = $1
    `
	row := r.db.QueryRow(query, id)

	program := &entities.Program{}
	err := row.Scan(&program.ID, &program.Name, &program.Description, &program.CreatedAt, &program.UpdatedAt, &program.CampusID, &program.Active)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return program, nil
}

// Update atualiza um programa
func (r *programInstanceRepository) Update(program *entities.Program) error {
	query := `
        UPDATE programs
        SET name = $2, description = $3, campus_id = $4, active = $5
        WHERE id = $1
    `
	_, err := r.db.Exec(query, program.ID, program.Name, program.Description, program.CampusID, program.Active)
	return err
}

// Delete remove um programa pelo ID
func (r *programInstanceRepository) Delete(id string) error {
	query := `DELETE FROM programs WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
