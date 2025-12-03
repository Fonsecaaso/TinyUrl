package service

import (
	"context"
	"errors"

	"github.com/fonsecaaso/TinyUrl/go-server/internal/model"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/repository"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/token"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, username, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Register(ctx context.Context, username, email, password string) (*model.User, error) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
	}

	err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	tokenString, err := token.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
