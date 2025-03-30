package services

import (
	"errors"
	"time"
	"unicast-api/internal/models"
	"unicast-api/internal/models/entities"
	"unicast-api/internal/repositories"
	"unicast-api/pkg/mailer"
)

type SendResponse struct {
	EmailsFailed   []entities.Student `json:"emailsFailed"`
	WhatsappFailed []entities.Student `json:"whatsappFailed"`
}

type MessageService interface {
	Send(message *models.Message) (emailsFails, whatsappFails []entities.Student, err error)
}

type messageService struct {
	whatsAppRepository repositories.WhatsAppRepository
	smtpRepository     repositories.SmtpRepository
	userRepository     repositories.UserRepository
	studentRepository  repositories.StudentRepository
}

var (
	ErrSmtpNotFound     = makeError("smtp não encontrado.", 404)
	ErrWhatsAppNotFound = makeError("whatsapp não encontrado.", 404)
	ErrStudentsNotFound = makeError("estudantes não encontrado.", 404)
)

func NewMessageService(whatsAppRepository repositories.WhatsAppRepository, smtpRepository repositories.SmtpRepository, userRepository repositories.UserRepository, studentRepository repositories.StudentRepository) MessageService {
	return &messageService{
		whatsAppRepository: whatsAppRepository,
		smtpRepository:     smtpRepository,
		userRepository:     userRepository,
		studentRepository:  studentRepository,
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

func (s *messageService) Send(message *models.Message) (emailsFails, whatsappFails []entities.Student, err error) {
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
	sender := mailer.NewEmailSender(mailer.SmtpAuthentication{
		Host:     smtp.Host,
		Port:     smtp.Port,
		Username: smtp.Email,
		Password: smtp.Password,
	})
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

	return *emailFailedStudents, *whatsappFailedStudents, nil
}
