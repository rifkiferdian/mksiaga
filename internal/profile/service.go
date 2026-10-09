package profile

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidName     = errors.New("invalid profile name")
	ErrInvalidPassword = errors.New("invalid password")
	ErrWrongPassword   = errors.New("wrong current password")
)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }
func (s *Service) Profile(ctx context.Context, userID, userStoreID uint64) (Profile, error) {
	return s.repository.Profile(ctx, userID, userStoreID)
}
func (s *Service) UpdateName(ctx context.Context, userID uint64, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 150 {
		return "", ErrInvalidName
	}
	return name, s.repository.UpdateName(ctx, userID, name)
}
func (s *Service) UpdatePassword(ctx context.Context, userID uint64, currentPassword, newPassword, confirmation string) error {
	if len(currentPassword) == 0 || len(currentPassword) > 72 || len(newPassword) < 8 || len(newPassword) > 72 || newPassword != confirmation || currentPassword == newPassword {
		return ErrInvalidPassword
	}
	hash, err := s.repository.PasswordHash(ctx, userID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(currentPassword)) != nil {
		return ErrWrongPassword
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repository.UpdatePassword(ctx, userID, string(newHash))
}
