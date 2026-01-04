package whatsapp

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Instance struct {
	ID         string    `json:"id"`
	Phone      string    `json:"phone" validate:"required"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
	UserID     string    `json:"-"`
	InstanceID string    `json:"instanceId"`
}

// SendText envia uma mensagem de texto via Evolution API usando a instância informada.
func SendText(instanceID, number, text string) error {
	return sendEvolutionText(instanceID, number, text)
}

// NormalizeNumber sanitiza e tenta converter para um formato próximo de E.164 usando um DDI padrão.
// Se o número for muito curto, retorna erro.
func NormalizeNumber(raw, defaultCountryCode string) (string, error) {
	digits := make([]rune, 0, len(raw))
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
		}
	}

	if len(digits) < 10 {
		return "", errors.New("telefone muito curto")
	}

	num := string(digits)
	// Se já começa com o DDI informado, só prefixa o '+'
	if strings.HasPrefix(num, defaultCountryCode) {
		return "+" + num, nil
	}

	// Caso contrário, prefixa o DDI e retorna.
	return "+" + defaultCountryCode + num, nil
}

// Repository define operações para instâncias WhatsApp.
type Repository interface {
	database.Transactional
	Create(ctx context.Context, phone, userID, instanceID string) error
	FindByID(ctx context.Context, id string) (*Instance, error)
	FindByPhoneAndUserId(ctx context.Context, phone, userId string) (*Instance, error)
	FindAllByUserId(ctx context.Context, userId string) ([]*Instance, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, id string) error
}

func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
