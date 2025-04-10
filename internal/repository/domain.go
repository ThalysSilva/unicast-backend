package repository

import (
	"database/sql"

	"github.com/ThalysSilva/unicast-backend/internal/campus"
	"github.com/ThalysSilva/unicast-backend/internal/course"
	"github.com/ThalysSilva/unicast-backend/internal/enrollment"
	"github.com/ThalysSilva/unicast-backend/internal/program"
	"github.com/ThalysSilva/unicast-backend/internal/smtp"
	"github.com/ThalysSilva/unicast-backend/internal/student"
	"github.com/ThalysSilva/unicast-backend/internal/user"
	"github.com/ThalysSilva/unicast-backend/internal/whatsapp"
)

type Repositories struct {
	User             user.Repository
	Course           course.Repository
	Enrollment       enrollment.Repository
	SmtpInstance     smtp.Repository
	WhatsAppInstance whatsapp.Repository
	Campus           campus.Repository
	Program          program.Repository
	Student          student.Repository
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		User:             user.NewRepository(db),
		Course:           course.NewRepository(db),
		Enrollment:       enrollment.NewRepository(db),
		SmtpInstance:     smtp.NewRepository(db),
		WhatsAppInstance: whatsapp.NewRepository(db),
		Campus:           campus.NewRepository(db),
		Program:          program.NewRepository(db),
		Student:          student.NewRepository(db),
	}
}
