package user

import "context"

type Service interface {
	Create(ctx context.Context, name, email, password string) (string, error)
}

type userService struct {
	userRepository Repository
}

func NewService(userRepository Repository) Service {
	return &userService{
		userRepository: userRepository,
	}
}

func (s *userService) Create(ctx context.Context, name, email, password string) (string, error) {
	user := &User{
		Name:     name,
		Email:    email,
		Password: password,
	}

	userId, err := s.userRepository.Create(ctx, user)
	if err != nil {
		return "", err
	}

	return userId, nil
}
