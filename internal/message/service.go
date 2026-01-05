package message

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ThalysSilva/unicast-backend/internal/auth"
	"github.com/ThalysSilva/unicast-backend/internal/config/env"
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
	logRepository      LogRepository
	jweSecret          []byte
	defaultCountryCode string
}

var (
	ErrSmtpNotFound     = customerror.Make("smtp não encontrado.", 404, errors.New("ErrSmtpNotFound"))
	ErrWhatsAppNotFound = customerror.Make("whatsapp não encontrado.", 404, errors.New("ErrWhatsAppNotFound"))
	ErrStudentsNotFound = customerror.Make("estudantes não encontrado.", 404, errors.New("ErrStudentsNotFound"))
	ErrPhoneMissing     = customerror.Make("estudante sem telefone configurado", 400, errors.New("ErrPhoneMissing"))
	ErrPhoneInvalid     = customerror.Make("telefone inválido para WhatsApp", 400, errors.New("ErrPhoneInvalid"))
)

func NewMessageService(whatsAppRepository whatsapp.Repository, smtpRepository smtp.Repository, userRepository user.Repository, studentRepository student.Repository, logRepository LogRepository, jweSecret []byte) Service {
	cfg, _ := env.Load()
	defaultCountry := "55"
	if cfg != nil && cfg.Defaults.CountryCode != "" {
		defaultCountry = cfg.Defaults.CountryCode
	}

	return &service{
		whatsAppRepository: whatsAppRepository,
		smtpRepository:     smtpRepository,
		userRepository:     userRepository,
		studentRepository:  studentRepository,
		logRepository:      logRepository,
		jweSecret:          jweSecret,
		defaultCountryCode: defaultCountry,
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

	waInstance, err := s.whatsAppRepository.FindByID(ctx, message.WhatsappId)
	if err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}
	if waInstance == nil {
		return nil, nil, customerror.Trace("Send", ErrWhatsAppNotFound)
	}

	decryptedJwe, err := auth.DecryptJWE[auth.JwePayload](message.Jwe, s.jweSecret)
	if err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}

	if err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}
	decryptedSmtpPassword, err := encryption.DecryptSmtpPassword([]byte(smtp.Password), []byte(decryptedJwe.SmtpKeyEncoded), []byte(smtp.IV))
	if err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}
	sender := mailer.NewEmailSender(mailer.SmtpAuthentication{
		Host:     smtp.Host,
		Port:     smtp.Port,
		Username: smtp.Email,
		Password: decryptedSmtpPassword,
	})

	attachments := []mailer.Attachment{}
	var attachmentNames []string
	if message.Attachments != nil {
		for _, attachment := range *message.Attachments {
			attachments = append(attachments, mailer.Attachment{
				FileName: attachment.FileName,
				Data:     attachment.Data,
			})
			attachmentNames = append(attachmentNames, attachment.FileName)
		}
	}
	attachmentNamesStr := ""
	if len(attachmentNames) > 0 {
		attachmentNamesStr = strings.Join(attachmentNames, ",")
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

	emailSendErr := sender.SendEmails(4, 4, 10, 5*time.Second)
	if emailSendErr != nil {
		failedStudents, e := extractEmailFailedStudents(emailSendErr, students)
		if e != nil {
			return nil, nil, customerror.Trace("Send", e)
		}
		if len(failedStudents) > 0 {
			emailFailedStudents = &failedStudents
		}
	}
	whatsappFailedStudents := &[]student.Student{}

	// Envio por WhatsApp via Evolution API
	var failedWhats []student.Student
	for _, stud := range students {
		if stud.Phone == nil || *stud.Phone == "" {
			failedWhats = append(failedWhats, *stud)
			continue
		}

		normalized, err := whatsapp.NormalizeNumber(*stud.Phone, s.defaultCountryCode)
		if err != nil {
			failedWhats = append(failedWhats, *stud)
			continue
		}

		if err := sendWhatsAppWithRetry(waInstance.InstanceID, normalized, message.Body, 3, 1*time.Second); err != nil {
			fmt.Printf("falha ao enviar whatsapp para %s: %v\n", *stud.Phone, err)
			failedWhats = append(failedWhats, *stud)
			continue
		}
	}

	if len(failedWhats) > 0 {
		whatsappFailedStudents = &failedWhats
	}

	// Persist logs
	emailFailedSet := make(map[string]string)
	if emailSendErr != nil && emailFailedStudents != nil {
		for _, s := range *emailFailedStudents {
			emailFailedSet[s.ID] = emailSendErr.Error()
		}
	}
	for _, stud := range students {
		errText, failed := emailFailedSet[stud.ID]
		if err := s.logRepository.Save(ctx, &Log{
			StudentID:       stud.ID,
			Channel:         ChannelEmail,
			Success:         !failed,
			ErrorText:       nullableString(errText, failed),
			Subject:         &message.Subject,
			Body:            &message.Body,
			SMTPID:          &message.SmtpId,
			AttachmentNames: nullableString(attachmentNamesStr, len(attachmentNames) > 0),
			AttachmentCount: len(attachmentNames),
		}); err != nil {
			fmt.Printf("falha ao salvar log email student %s: %v\n", stud.ID, err)
		}
	}

	whatsFailedSet := make(map[string]string)
	for _, s := range failedWhats {
		whatsFailedSet[s.ID] = "failed to send whatsapp"
	}
	for _, stud := range students {
		errText, failed := whatsFailedSet[stud.ID]
		if err := s.logRepository.Save(ctx, &Log{
			StudentID:          stud.ID,
			Channel:            ChannelWhatsApp,
			Success:            !failed,
			ErrorText:          nullableString(errText, failed),
			Subject:            &message.Subject,
			Body:               &message.Body,
			WhatsAppInstanceID: &message.WhatsappId,
			AttachmentNames:    nullableString(attachmentNamesStr, len(attachmentNames) > 0),
			AttachmentCount:    len(attachmentNames),
		}); err != nil {
			fmt.Printf("falha ao salvar log whatsapp student %s: %v\n", stud.ID, err)
		}
	}

	return emailFailedStudents, whatsappFailedStudents, nil
}

// sendWhatsAppWithRetry encapsula retentativa simples para envio de WhatsApp.
func sendWhatsAppWithRetry(instanceID, number, body string, attempts int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		lastErr = whatsapp.SendText(instanceID, number, body)
		if lastErr == nil {
			return nil
		}
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}
	return lastErr
}

func nullableString(val string, set bool) *string {
	if !set {
		return nil
	}
	return &val
}
