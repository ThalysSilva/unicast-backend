package repositories

import "unicast-api/internal/models/entities"

type CampusRepository interface {
	Create(program *entities.Campus) error
	FindByID(id string) (*entities.Campus, error)
	Update(program *entities.Campus) error
	Delete(id string) error
}

type ProgramRepository interface {
	Create(program *entities.Program) error
	FindByID(id string) (*entities.Program, error)
	Update(program *entities.Program) error
	Delete(id string) error
}

type SmtpRepository interface {
	Create(instance *entities.SmtpInstance) error
	FindByID(id string) (*entities.SmtpInstance, error)
	Update(instance *entities.SmtpInstance) error
	Delete(id string) error
}

type UserRepository interface {
	Create(user *entities.User) (userId string, err error)
	FindByEmail(email string) (*entities.User, error)
	SaveRefreshToken(userId string, refreshToken string) error
	Logout(userId string) error
	FindByID(id string) (*entities.User, error)
}

type WhatsAppRepository interface {
	Create(instance *entities.WhatsAppInstance) error
	FindByID(id string) (*entities.WhatsAppInstance, error)
	Update(instance *entities.WhatsAppInstance) error
	Delete(id string) error
}

type StudentRepository interface {
	Create(student *entities.Student) error
	FindByID(id string) (*entities.Student, error)
	Update(student *entities.Student) error
	Delete(id string) error
}

type EnrollmentRepository interface {
	Create(enrollment *entities.Enrollment) error
	FindByID(id string) (*entities.Enrollment, error)
	Update(enrollment *entities.Enrollment) error
	Delete(id string) error
}




