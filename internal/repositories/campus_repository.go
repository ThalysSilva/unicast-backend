package repositories

import (
	"database/sql"
	"unicast-api/internal/models/entities"
)



// Gerencia operações de banco para Campus
type campusInstanceRepository struct {
    db *sql.DB
}

// Cria uma nova instância do repositório
func NewCampusRepository(db *sql.DB) CampusRepository {
    return &campusInstanceRepository{db: db}
}

// Insere um novo campus
func (r *campusInstanceRepository) Create(campus *entities.Campus) error {
    query := `
        INSERT INTO campuses (id, name, description,user_owner_id)
        VALUES ($1, $2, $3, $4)
    `
    _, err := r.db.Exec(query, campus.ID, campus.Name, campus.Description, campus.UserOwnerID)
    return err
}

// Busca um campus pelo ID
func (r *campusInstanceRepository) FindByID(id string) (*entities.Campus, error) {
    query := `
        SELECT id, name, description, created_at, updated_at, user_owner_id
        FROM campuses
        WHERE id = $1
    `
    row := r.db.QueryRow(query, id)

    campus := &entities.Campus{}
    err := row.Scan(&campus.ID, &campus.Name, &campus.Description, &campus.CreatedAt, &campus.UpdatedAt, &campus.UserOwnerID)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return campus, nil
}

// Atualiza um campus
func (r *campusInstanceRepository) Update(campus *entities.Campus) error {
    query := `
        UPDATE campuses
        SET name = $2, description = $3, user_owner_id = $4
        WHERE id = $1
    `
    _, err := r.db.Exec(query, campus.ID, campus.Name, campus.Description, campus.UserOwnerID)
    return err
}

// Remove um campus pelo ID
func (r *campusInstanceRepository) Delete(id string) error {
    query := `DELETE FROM campuses WHERE id = $1`
    _, err := r.db.Exec(query, id)
    return err
}