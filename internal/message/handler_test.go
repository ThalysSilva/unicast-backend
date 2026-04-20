package message

import (
	"testing"

	"github.com/ThalysSilva/unicast-backend/internal/student"
)

func TestFailedRecipientsOmitsContactData(t *testing.T) {
	name := "Aluno Teste"
	phone := "5511999999999"
	email := "aluno@example.com"
	annotation := "observação interna"

	recipients := failedRecipients([]student.Student{
		{
			ID:         "student-uuid",
			StudentID:  "2026996",
			Name:       &name,
			Phone:      &phone,
			Email:      &email,
			Annotation: &annotation,
			Consent:    true,
			Status:     student.StudentStatusActive,
		},
	})

	if len(recipients) != 1 {
		t.Fatalf("len = %d, want 1", len(recipients))
	}
	if recipients[0] != (FailedRecipient{ID: "student-uuid", StudentID: "2026996"}) {
		t.Fatalf("recipient = %+v", recipients[0])
	}
}
