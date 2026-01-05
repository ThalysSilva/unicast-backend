package whatsapp

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Instance struct {
	ID           string    `json:"id"`
	Phone        string    `json:"phone" validate:"required"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
	UserID       string    `json:"-"`
	InstanceName string    `json:"instanceName"`
}

// SendText envia uma mensagem de texto via Evolution API usando a instância informada.
func SendText(instanceName, number, text string) error {
	return sendEvolutionText(instanceName, number, text)
}

// SendMedia envia um attachment via Evolution API (media pode ser URL ou base64).
func SendMedia(instanceName, number string, fileName string, data []byte, caption string) error {
	mime := http.DetectContentType(data)
	mediaType := inferMediaType(mime)
	encoded := base64.StdEncoding.EncodeToString(data)

	return sendEvolutionMedia(instanceName, sendMediaPayload{
		Number:    number,
		Media:     encoded,
		MediaType: mediaType,
		MimeType:  mime,
		FileName:  fileName,
		Caption:   caption,
	})
}

// SendMediaURL envia um attachment hospedado por URL via Evolution API.
func SendMediaURL(instanceName, number string, mediaURL string, fileName string, caption string) error {
	return sendEvolutionMedia(instanceName, sendMediaPayload{
		Number:    number,
		Media:     mediaURL,
		MediaType: inferMediaType(""),
		FileName:  fileName,
		Caption:   caption,
	})
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

func inferMediaType(mime string) string {
	if strings.HasPrefix(mime, "image/") {
		return "image"
	}
	if strings.HasPrefix(mime, "video/") {
		return "video"
	}
	if strings.HasPrefix(mime, "audio/") {
		return "audio"
	}
	if mime == "" {
		return "document"
	}
	return "document"
}
