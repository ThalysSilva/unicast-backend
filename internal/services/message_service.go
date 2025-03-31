package services

import (
	"errors"
	"time"

	"github.com/ThalysSilva/unicast-backend/internal/interfaces"
	"github.com/ThalysSilva/unicast-backend/internal/models"
	"github.com/ThalysSilva/unicast-backend/internal/models/entities"
	"github.com/ThalysSilva/unicast-backend/pkg/auth"
	"github.com/ThalysSilva/unicast-backend/pkg/encryption"
	"github.com/ThalysSilva/unicast-backend/pkg/mailer"
)

type SendResponse struct {
	EmailsFailed   []entities.Student `json:"emailsFailed"`
	WhatsappFailed []entities.Student `json:"whatsappFailed"`
}

type MessageService interface {
	Send(message *models.Message) (emailsFails, whatsappFails *[]entities.Student, err error)
}

type messageService struct {
	whatsAppRepository interfaces.WhatsAppRepository
	smtpRepository     interfaces.SmtpRepository
	userRepository     interfaces.UserRepository
	studentRepository  interfaces.StudentRepository
	jweSecret          []byte
}

var (
	ErrSmtpNotFound     = makeError("smtp não encontrado.", 404)
	ErrWhatsAppNotFound = makeError("whatsapp não encontrado.", 404)
	ErrStudentsNotFound = makeError("estudantes não encontrado.", 404)
)

func NewMessageService(whatsAppRepository interfaces.WhatsAppRepository, smtpRepository interfaces.SmtpRepository, userRepository interfaces.UserRepository, studentRepository interfaces.StudentRepository, jweSecret []byte) MessageService {
	return &messageService{
		whatsAppRepository: whatsAppRepository,
		smtpRepository:     smtpRepository,
		userRepository:     userRepository,
		studentRepository:  studentRepository,
		jweSecret:          jweSecret,
	}
}

func extractEmailFailedStudents(err error, students []*entities.Student) ([]entities.Student, error) {
	var emailErr *mailer.EmailSentError
	if errors.As(err, &emailErr) {
		emailToStudent := make(map[string]*entities.Student, len(students))
		for _, student := range students {
			if student.Email != nil {
				emailToStudent[*student.Email] = student
			}
		}

		var failedStudents []entities.Student
		for _, failedEmail := range emailErr.To {
			if student, exists := emailToStudent[failedEmail]; exists {
				failedStudents = append(failedStudents, *student)
			}
		}
		return failedStudents, err
	}
	return nil, err
}

func (s *messageService) Send(message *models.Message) (emailsFails, whatsappFails *[]entities.Student, err error) {
	students, err := s.studentRepository.FindByIDs(message.To)
	if err != nil {
		return nil, nil, trace("Send", err)
	}
	if len(students) == 0 {
		return nil, nil, trace("Send", ErrStudentsNotFound)
	}

	smtp, err := s.smtpRepository.FindByID(message.SmtpId)
	if err != nil {
		return nil, nil, trace("Send", err)
	}
	if smtp == nil {
		return nil, nil, trace("Send", ErrSmtpNotFound)
	}

	whatsapp, err := s.whatsAppRepository.FindByID(message.WhatsappId)
	if err != nil {
		return nil, nil, trace("Send", err)
	}
	if whatsapp == nil {
		return nil, nil, trace("Send", ErrSmtpNotFound)
	}

	decryptedJwe, err := auth.DecryptJWE[JwePayload](message.Jwe, s.jweSecret)
	if err != nil {
		return nil, nil, trace("Send", err)
	}

	decryptedSmtpPassword, err := encryption.DecryptSmtpPassword([]byte(smtp.Password), []byte(decryptedJwe.SmtpKey), []byte(smtp.IV))
	sender := mailer.NewEmailSender(mailer.SmtpAuthentication{
		Host:     smtp.Host,
		Port:     smtp.Port,
		Username: smtp.Email,
		Password: decryptedSmtpPassword,
	})
	if err != nil {
		return nil, nil, trace("Send", err)
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
	emailFailedStudents := &[]entities.Student{}

	if err := sender.SetData(&mailer.MailerData{
		From:        message.From,
		To:          message.To,
		Subject:     message.Subject,
		Body:        message.Body,
		Attachments: &attachments,
		ContentType: mailer.TextPlain,
	}); err != nil {
		return nil, nil, trace("Send", err)
	}

	if err := sender.SendEmails(4, 4, 10, 5*time.Second); err != nil {
		failedStudents, err := extractEmailFailedStudents(err, students)
		if err != nil {
			return nil, nil, trace("Send", err)
		}
		if len(failedStudents) > 0 {
			emailFailedStudents = &failedStudents
		}

	}
	whatsappFailedStudents := &[]entities.Student{}

	// Ainda falta implementar a parte do whatsapp

	return emailFailedStudents, whatsappFailedStudents, nil
}
