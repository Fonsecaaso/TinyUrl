package service

import (
	"context"
	"errors"

	"github.com/fonsecaaso/TinyUrl/go-server/internal/model"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/repository"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/token"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, username, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type authService struct {
	userRepo repository.UserRepository
	logger   *zap.Logger
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
		logger:   zap.L().With(zap.String("component", "AuthService")),
	}
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
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, ErrEmailAlreadyExists
	}
	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	tokenString, err := token.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
