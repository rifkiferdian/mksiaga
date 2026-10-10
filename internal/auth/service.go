package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Service struct {
	repository *Repository
}

func (s *Service) CreateSession(ctx context.Context, id string, user User, ipAddress, userAgent string, expiresAt time.Time) error {
	return s.repository.CreateSession(ctx, id, user.ID, user.UserStoreID, ipAddress, userAgent, expiresAt)
}
func (s *Service) RevokeSession(ctx context.Context, id string) error {
	return s.repository.RevokeSession(ctx, id)
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Authenticate(ctx context.Context, login, password string) (User, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" || len(login) > 191 || len(password) > 72 {
		return User{}, ErrInvalidCredentials
	}

	user, err := s.repository.FindActiveByLogin(ctx, login)
	if errors.Is(err, ErrUserNotFound) {
		return User{}, ErrInvalidCredentials
	}
	if err != nil {
		return User{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return User{}, ErrInvalidCredentials
	}
	if err := s.repository.RecordLogin(ctx, user.ID); err != nil {
		return User{}, fmt.Errorf("finish authentication: %w", err)
	}

	return user, nil
}
