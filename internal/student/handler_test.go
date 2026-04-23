package student

import (
	"strings"
	"testing"
)

type testMultipartFile struct {
	*strings.Reader
}

func (f testMultipartFile) Close() error {
	return nil
}

func TestParseImportCSVAcceptsMissingTrailingFields(t *testing.T) {
	file := testMultipartFile{strings.NewReader("studentId,name,phone,email,status\n2026996,Isabela Fernandes,5500000000010\n")}

	records, err := parseImportCSV(file)
	if err != nil {
		t.Fatalf("parseImportCSV() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("parseImportCSV() len = %d, want 1", len(records))
	}

	record := records[0]
	if record.StudentID != "2026996" {
		t.Fatalf("StudentID = %q, want %q", record.StudentID, "2026996")
	}
	if record.Name == nil || *record.Name != "Isabela Fernandes" {
		t.Fatalf("Name = %v, want Isabela Fernandes", record.Name)
	}
	if record.Phone == nil || *record.Phone != "5500000000010" {
		t.Fatalf("Phone = %v, want 5500000000010", record.Phone)
	}
	if record.Email != nil {
		t.Fatalf("Email = %v, want nil", record.Email)
	}
	if record.Status != StudentStatusPending {
		t.Fatalf("Status = %q, want %q", record.Status, StudentStatusPending)
	}
	if record.StatusProvided {
		t.Fatalf("StatusProvided = true, want false")
	}
}
