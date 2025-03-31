package repositories

import (
	"database/sql"

	"github.com/ThalysSilva/unicast-backend/internal/interfaces"
	"github.com/ThalysSilva/unicast-backend/internal/repositories/native"
)

type Options struct {
	Db *sql.DB
}

type Container struct {
	User       interfaces.UserRepository
	Course     interfaces.CourseRepository
	Enrollment interfaces.EnrollmentRepository
	SmtpInstance interfaces.SmtpRepository
	WhatsAppInstance interfaces.WhatsAppRepository
	Campus     interfaces.CampusRepository
	Program    interfaces.ProgramRepository
	Student    interfaces.StudentRepository
}

func New(options Options) *Container {
	return &Container{
		User:       native.NewUserRepository(options.Db),
		Course:     native.NewCourseRepository(options.Db),
		Enrollment: native.NewEnrollmentRepository(options.Db),
		SmtpInstance: native.NewSmtpInstanceRepository(options.Db),
		WhatsAppInstance: native.NewWhatsAppInstanceRepository(options.Db),
		Campus:     native.NewCampusRepository(options.Db),
		Program:    native.NewProgramRepository(options.Db),
		Student:    native.NewStudentRepository(options.Db),
		
	}
}
