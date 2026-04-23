package invite

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ThalysSilva/unicast-backend/internal/discipline"
	"github.com/ThalysSilva/unicast-backend/internal/enrollment"
	"github.com/ThalysSilva/unicast-backend/internal/student"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/lib/pq"
)

type Service interface {
	Create(ctx context.Context, disciplineID, userID string, expiresAt *time.Time) (*Invite, error)
	GetCurrent(ctx context.Context, disciplineID, userID string) (*Invite, error)
	ListByDiscipline(ctx context.Context, disciplineID, userID string) ([]*Invite, error)
	Delete(ctx context.Context, inviteID, userID string) error
	SelfRegister(ctx context.Context, code, studentID, name, phone string, noPhone bool, email string, consent bool) error
}

type inviteService struct {
	inviteRepository     Repository
	disciplineRepository discipline.Repository
	enrollmentRepository enrollment.Repository
	studentRepository    student.Repository
}

var (
	ErrInviteNotFound                 = customerror.Make("convite não encontrado", http.StatusNotFound, errors.New("ErrInviteNotFound"))
	ErrInviteInactive                 = customerror.Make("convite inativo", http.StatusBadRequest, errors.New("ErrInviteInactive"))
	ErrInviteExpired                  = customerror.Make("convite expirado", http.StatusBadRequest, errors.New("ErrInviteExpired"))
	ErrNotDisciplineOwner             = customerror.Make("você não tem permissão para esta disciplina", http.StatusForbidden, errors.New("ErrNotDisciplineOwner"))
	ErrEnrollmentNotFound             = customerror.Make("estudante não está vinculado à disciplina", http.StatusBadRequest, errors.New("ErrEnrollmentNotFound"))
	ErrEnrollmentRegistrationComplete = customerror.Make("cadastro desta matrícula já foi concluído para esta disciplina", http.StatusConflict, errors.New("ErrEnrollmentRegistrationComplete"))
	ErrStudentNotFound                = customerror.Make("estudante não encontrado", http.StatusNotFound, errors.New("ErrStudentNotFound"))
	ErrConsentRequired                = customerror.Make("é necessário aceitar o recebimento automatizado de notificações", http.StatusBadRequest, errors.New("ErrConsentRequired"))
	ErrContactRequired                = customerror.Make("preencha nome, email e telefone, ou informe que não possui telefone, para concluir o cadastro", http.StatusBadRequest, errors.New("ErrContactRequired"))
)

func NewService(
	inviteRepository Repository,
	disciplineRepository discipline.Repository,
	enrollmentRepository enrollment.Repository,
	studentRepository student.Repository,
) Service {
	return &inviteService{
		inviteRepository:     inviteRepository,
		disciplineRepository: disciplineRepository,
		enrollmentRepository: enrollmentRepository,
		studentRepository:    studentRepository,
	}
}

func (s *inviteService) Create(ctx context.Context, disciplineID, userID string, expiresAt *time.Time) (*Invite, error) {
	disciplineWithOwner, err := s.disciplineRepository.FindByIDWithUserOwnerID(ctx, disciplineID)
	if err != nil {
		return nil, err
	}
	if disciplineWithOwner == nil || disciplineWithOwner.UserOwnerID != userID {
		return nil, ErrNotDisciplineOwner
	}

	// Gera código determinístico (hash curto) variando salt/tempo; evita colisão com retry em unique.
	for attempts := 0; attempts < 10; attempts++ {
		code := s.generateCode(disciplineID, attempts)

		err = s.inviteRepository.Create(ctx, disciplineID, code, expiresAt)
		if err != nil {
			if isUniqueViolation(err) {
				continue
			}
			return nil, err
		}

		return &Invite{
			DisciplineID: disciplineID,
			Code:         code,
			ExpiresAt:    expiresAt,
			Active:       true,
		}, nil
	}

	return nil, customerror.Trace("inviteService.Create", errors.New("falha ao gerar código único após múltiplas tentativas"))
}

func (s *inviteService) GetCurrent(ctx context.Context, disciplineID, userID string) (*Invite, error) {
	if err := s.ensureDisciplineOwner(ctx, disciplineID, userID); err != nil {
		return nil, err
	}

	return s.inviteRepository.FindLatestByDisciplineID(ctx, disciplineID)
}

func (s *inviteService) ListByDiscipline(ctx context.Context, disciplineID, userID string) ([]*Invite, error) {
	if err := s.ensureDisciplineOwner(ctx, disciplineID, userID); err != nil {
		return nil, err
	}

	return s.inviteRepository.FindByDisciplineID(ctx, disciplineID)
}

func (s *inviteService) Delete(ctx context.Context, inviteID, userID string) error {
	inviteFound, err := s.inviteRepository.FindByID(ctx, inviteID)
	if err != nil {
		return err
	}
	if inviteFound == nil {
		return ErrInviteNotFound
	}
	if err := s.ensureDisciplineOwner(ctx, inviteFound.DisciplineID, userID); err != nil {
		return err
	}

	return s.inviteRepository.Delete(ctx, inviteID)
}

func (s *inviteService) ensureDisciplineOwner(ctx context.Context, disciplineID, userID string) error {
	disciplineWithOwner, err := s.disciplineRepository.FindByIDWithUserOwnerID(ctx, disciplineID)
	if err != nil {
		return err
	}
	if disciplineWithOwner == nil || disciplineWithOwner.UserOwnerID != userID {
		return ErrNotDisciplineOwner
	}

	return nil
}

func (s *inviteService) SelfRegister(ctx context.Context, code, studentID, name, phone string, noPhone bool, email string, consent bool) error {
	inviteFound, err := s.inviteRepository.FindByCode(ctx, code)
	if err != nil {
		return err
	}
	if inviteFound == nil {
		return ErrInviteNotFound
	}
	if !inviteFound.Active {
		return ErrInviteInactive
	}
	if inviteFound.ExpiresAt != nil && inviteFound.ExpiresAt.Before(time.Now()) {
		return ErrInviteExpired
	}

	disciplineFound, err := s.disciplineRepository.FindByIDWithUserOwnerID(ctx, inviteFound.DisciplineID)
	if err != nil {
		return err
	}
	if disciplineFound == nil {
		return ErrInviteNotFound
	}

	studentFound, err := s.studentRepository.FindByStudentID(ctx, studentID, disciplineFound.UserOwnerID)
	if err != nil {
		return err
	}
	if studentFound == nil {
		return ErrStudentNotFound
	}

	enrollmentFound, err := s.enrollmentRepository.FindByDisciplineAndStudent(ctx, inviteFound.DisciplineID, studentFound.ID)
	if err != nil {
		return err
	}
	if enrollmentFound == nil {
		return ErrEnrollmentNotFound
	}

	if enrollmentFound.SelfRegistrationCompletedAt != nil {
		return ErrEnrollmentRegistrationComplete
	}
	if !consent {
		return ErrConsentRequired
	}

	name = strings.TrimSpace(name)
	phone = strings.TrimSpace(phone)
	email = strings.TrimSpace(email)
	if name == "" || email == "" || (!noPhone && phone == "") {
		return ErrContactRequired
	}
	if noPhone {
		phone = ""
	}

	fields := map[string]any{
		"status":   student.StudentStatusActive,
		"consent":  true,
		"name":     name,
		"phone":    nullableValue(phone),
		"no_phone": noPhone,
		"email":    email,
	}

	if err := s.studentRepository.Update(ctx, studentFound.ID, fields); err != nil {
		return err
	}

	now := time.Now()
	return s.enrollmentRepository.Update(ctx, enrollmentFound.ID, map[string]any{
		"self_registration_completed_at": now,
		"self_registration_count":        enrollmentFound.SelfRegistrationCount + 1,
	})
}

func nullableValue(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func (s *inviteService) generateCode(disciplineID string, attempt int) string {
	now := time.Now().UnixNano()
	payload := disciplineID + ":" + strconv.Itoa(attempt) + ":" + strconv.FormatInt(now, 10)
	sum := sha256.Sum256([]byte(payload))

	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(sum[:])
	// Evita caracteres ambíguos substituindo I/L/O/0 por letras mais claras.
	encoded = sanitizeBase32(encoded)

	if len(encoded) < 7 {
		return encoded
	}
	return encoded[:7]
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

func sanitizeBase32(s string) string {
	// Remove caracteres potencialmente ambíguos.
	replacer := map[rune]rune{
		'I': 'X',
		'L': 'Y',
		'O': 'Z',
		'0': '2',
	}
	out := make([]rune, len(s))
	for i, r := range s {
		if repl, ok := replacer[r]; ok {
			out[i] = repl
		} else {
			out[i] = r
		}
	}
	return string(out)
}
