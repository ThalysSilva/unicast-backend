package message

import (
	"context"
	"errors"
	"time"

	"github.com/ThalysSilva/unicast-backend/internal/auth"
	"github.com/ThalysSilva/unicast-backend/internal/encryption"
	"github.com/ThalysSilva/unicast-backend/internal/smtp"
	"github.com/ThalysSilva/unicast-backend/internal/student"
	"github.com/ThalysSilva/unicast-backend/internal/user"
	"github.com/ThalysSilva/unicast-backend/internal/whatsapp"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/mailer"
)

type SendResponse struct {
	EmailsFailed   []student.Student `json:"emailsFailed"`
	WhatsappFailed []student.Student `json:"whatsappFailed"`
}

type Service interface {
	Send(ctx context.Context, message *Message) (emailsFails, whatsappFails *[]student.Student, err error)
}

type service struct {
	whatsAppRepository whatsapp.Repository
	smtpRepository     smtp.Repository
	userRepository     user.Repository
	studentRepository  student.Repository
	jweSecret          []byte
}

var (
	ErrSmtpNotFound     = customerror.Make("smtp não encontrado.", 404, errors.New("ErrSmtpNotFound"))
	ErrWhatsAppNotFound = customerror.Make("whatsapp não encontrado.", 404, errors.New("ErrWhatsAppNotFound"))
	ErrStudentsNotFound = customerror.Make("estudantes não encontrado.", 404, errors.New("ErrStudentsNotFound"))
)

func NewMessageService(whatsAppRepository whatsapp.Repository, smtpRepository smtp.Repository, userRepository user.Repository, studentRepository student.Repository, jweSecret []byte) Service {
	return &service{
		whatsAppRepository: whatsAppRepository,
		smtpRepository:     smtpRepository,
		userRepository:     userRepository,
		studentRepository:  studentRepository,
		jweSecret:          jweSecret,
	}
}

func extractEmailFailedStudents(err error, students []*student.Student) ([]student.Student, error) {
	var emailErr *mailer.EmailSentError
	if errors.As(err, &emailErr) {
		emailToStudent := make(map[string]*student.Student, len(students))
		for _, student := range students {
			if student.Email != nil {
				emailToStudent[*student.Email] = student
			}
		}

		var failedStudents []student.Student
		for _, failedEmail := range emailErr.To {
			if student, exists := emailToStudent[failedEmail]; exists {
				failedStudents = append(failedStudents, *student)
			}
		}
		return failedStudents, err
	}
	return nil, err
}

func (s *service) Send(ctx context.Context, message *Message) (emailsFails, whatsappFails *[]student.Student, err error) {
	students, err := s.studentRepository.FindByIDs(ctx, message.To)
	if err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}
	if len(students) == 0 {
		return nil, nil, customerror.Trace("Send", ErrStudentsNotFound)
	}

	smtp, err := s.smtpRepository.FindByID(ctx, message.SmtpId)
	if err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}
	if smtp == nil {
		return nil, nil, customerror.Trace("Send", ErrSmtpNotFound)
	}

	whatsapp, err := s.whatsAppRepository.FindByID(ctx, message.WhatsappId)
	if err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}
	if whatsapp == nil {
		return nil, nil, customerror.Trace("Send", ErrSmtpNotFound)
	}

	decryptedJwe, err := auth.DecryptJWE[auth.JwePayload](message.Jwe, s.jweSecret)
	if err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}

	decryptedSmtpPassword, err := encryption.DecryptSmtpPassword([]byte(smtp.Password), []byte(decryptedJwe.SmtpKey), []byte(smtp.IV))
	sender := mailer.NewEmailSender(mailer.SmtpAuthentication{
		Host:     smtp.Host,
		Port:     smtp.Port,
		Username: smtp.Email,
		Password: decryptedSmtpPassword,
	})
	if err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}

	attachments := []mailer.Attachment{}
	if message.Attachments != nil {
		for _, attachment := range *message.Attachments {
			attachments = append(attachments, mailer.Attachment{
				FileName: attachment.FileName,
				Data:     attachment.Data,
			})
		}
	}
	emailFailedStudents := &[]student.Student{}

	if err := sender.SetData(&mailer.MailerData{
		From:        message.From,
		To:          message.To,
		Subject:     message.Subject,
		Body:        message.Body,
		Attachments: &attachments,
		ContentType: mailer.TextPlain,
	}); err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}

	if err := sender.SendEmails(4, 4, 10, 5*time.Second); err != nil {
		failedStudents, err := extractEmailFailedStudents(err, students)
		if err != nil {
			return nil, nil, customerror.Trace("Send", err)
		}
		if len(failedStudents) > 0 {
			emailFailedStudents = &failedStudents
		}

	}
	whatsappFailedStudents := &[]student.Student{}

	// Ainda falta implementar a parte do whatsapp

	return emailFailedStudents, whatsappFailedStudents, nil
}
