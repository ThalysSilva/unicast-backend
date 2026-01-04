package invite

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ThalysSilva/unicast-backend/internal/course"
	"github.com/ThalysSilva/unicast-backend/internal/enrollment"
	"github.com/ThalysSilva/unicast-backend/internal/student"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/lib/pq"
)

type Service interface {
	Create(ctx context.Context, courseID, userID string, expiresAt *time.Time) (*Invite, error)
	SelfRegister(ctx context.Context, code, studentID, name, phone, email string) error
}

type inviteService struct {
	inviteRepository     Repository
	courseRepository     course.Repository
	enrollmentRepository enrollment.Repository
	studentRepository    student.Repository
}

var (
	ErrInviteNotFound     = customerror.Make("convite não encontrado", http.StatusNotFound, errors.New("ErrInviteNotFound"))
	ErrInviteInactive     = customerror.Make("convite inativo", http.StatusBadRequest, errors.New("ErrInviteInactive"))
	ErrInviteExpired      = customerror.Make("convite expirado", http.StatusBadRequest, errors.New("ErrInviteExpired"))
	ErrNotCourseOwner     = customerror.Make("você não tem permissão para este curso", http.StatusForbidden, errors.New("ErrNotCourseOwner"))
	ErrEnrollmentNotFound = customerror.Make("estudante não está vinculado à disciplina", http.StatusBadRequest, errors.New("ErrEnrollmentNotFound"))
	ErrStudentNotPending  = customerror.Make("estudante já cadastrou os dados", http.StatusConflict, errors.New("ErrStudentNotPending"))
	ErrStudentNotFound    = customerror.Make("estudante não encontrado", http.StatusNotFound, errors.New("ErrStudentNotFound"))
)

func NewService(
	inviteRepository Repository,
	courseRepository course.Repository,
	enrollmentRepository enrollment.Repository,
	studentRepository student.Repository,
) Service {
	return &inviteService{
		inviteRepository:     inviteRepository,
		courseRepository:     courseRepository,
		enrollmentRepository: enrollmentRepository,
		studentRepository:    studentRepository,
	}
}

func (s *inviteService) Create(ctx context.Context, courseID, userID string, expiresAt *time.Time) (*Invite, error) {
	courseWithOwner, err := s.courseRepository.FindByIDWithUserOwnerID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if courseWithOwner == nil || courseWithOwner.UserOwnerID != userID {
		return nil, ErrNotCourseOwner
	}

	// Gera código determinístico (hash curto) variando salt/tempo; evita colisão com retry em unique.
	for attempts := 0; attempts < 10; attempts++ {
		code := s.generateCode(courseID, attempts)

		err = s.inviteRepository.Create(ctx, courseID, code, expiresAt)
		if err != nil {
			if isUniqueViolation(err) {
				continue
			}
			return nil, err
		}

		return &Invite{
			CourseID:  courseID,
			Code:      code,
			ExpiresAt: expiresAt,
			Active:    true,
		}, nil
	}

	return nil, customerror.Trace("inviteService.Create", errors.New("falha ao gerar código único após múltiplas tentativas"))
}

func (s *inviteService) SelfRegister(ctx context.Context, code, studentID, name, phone, email string) error {
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

	enrollmentFound, err := s.enrollmentRepository.FindByCourseAndStudent(ctx, inviteFound.CourseID, studentID)
	if err != nil {
		return err
	}
	if enrollmentFound == nil {
		return ErrEnrollmentNotFound
	}

	studentFound, err := s.studentRepository.FindByStudentID(ctx, studentID)
	if err != nil {
		return err
	}
	if studentFound == nil {
		return ErrStudentNotFound
	}
	if studentFound.Status != student.StudentStatusPending {
		return ErrStudentNotPending
	}

	fields := map[string]any{
		"status": student.StudentStatusActive,
	}
	if name != "" {
		fields["name"] = name
	}
	if phone != "" {
		fields["phone"] = phone
	}
	if email != "" {
		fields["email"] = email
	}

	return s.studentRepository.Update(ctx, studentFound.ID, fields)
}

func (s *inviteService) generateCode(courseID string, attempt int) string {
	now := time.Now().UnixNano()
	payload := courseID + ":" + strconv.Itoa(attempt) + ":" + strconv.FormatInt(now, 10)
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
