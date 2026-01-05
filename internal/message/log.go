package message

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Log struct {
	ID                 string
	StudentID          string
	Channel            Channel
	Success            bool
	ErrorText          *string
	Subject            *string
	Body               *string
	SMTPID             *string
	WhatsAppInstanceID *string
	AttachmentNames    *string
	AttachmentCount    int
	CreatedAt          time.Time
}

type Channel string

const (
	ChannelEmail    Channel = "EMAIL"
	ChannelWhatsApp Channel = "WHATSAPP"
)

type LogRepository interface {
	database.Transactional
	Save(ctx context.Context, log *Log) error
}

func NewLogRepository(db *sql.DB) LogRepository {
	return newLogRepository(db)
}
