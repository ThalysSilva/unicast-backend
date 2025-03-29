package entities

import "time"

type Enrollment struct {
	ID        string    `json:"id"`
	CourseID  string    `json:"courseId"`
	StudentID string    `json:"studentId"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
