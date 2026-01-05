package message

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type logRepository struct {
	db    database.DB
	sqlDB *sql.DB
}

func newLogRepository(db *sql.DB) LogRepository {
	newDb := database.NewSQLTx(db)
	return &logRepository{
		db:    newDb.DB,
		sqlDB: db,
	}
}

func (r *logRepository) WithTransaction(tx any) any {
	return &logRepository{
		db:    database.NewSQLTx(nil).WithSQLTransaction(tx).DB,
		sqlDB: r.sqlDB,
	}
}

func (r *logRepository) TransactionBackend() any {
	return r.sqlDB
}

func (r *logRepository) Save(ctx context.Context, log *Log) error {
	query := `
		INSERT INTO message_logs (student_id, channel, success, error_text, subject, body, smtp_id, whatsapp_instance_id, attachment_names, attachment_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		log.StudentID,
		string(log.Channel),
		log.Success,
		log.ErrorText,
		log.Subject,
		log.Body,
		log.SMTPID,
		log.WhatsAppInstanceID,
		log.AttachmentNames,
		log.AttachmentCount,
	)
	if err != nil {
		return fmt.Errorf("falha ao salvar log de mensagem: %w", err)
	}
	return nil
}
