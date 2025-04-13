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

func NewRepositories(dbSQL *sql.DB) *Repositories {
	return &Repositories{
		User:             user.NewRepository(dbSQL),
		Course:           course.NewRepository(dbSQL),
		Enrollment:       enrollment.NewRepository(dbSQL),
		SmtpInstance:     smtp.NewRepository(dbSQL),
		WhatsAppInstance: whatsapp.NewRepository(dbSQL),
		Campus:           campus.NewRepository(dbSQL),
		Program:          program.NewRepository(dbSQL),
		Student:          student.NewRepository(dbSQL),
	}
}
