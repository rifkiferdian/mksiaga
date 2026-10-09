package store

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrInvalid   = errors.New("invalid store")
	ErrDuplicate = errors.New("duplicate store")
	ErrProtected = errors.New("protected store")
	ErrInUse     = errors.New("store in use")
)
var codePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9._-]{1,49}$`)
var timezonePattern = regexp.MustCompile(`^[A-Za-z_]+(?:/[A-Za-z0-9_+.-]+)+$`)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service               { return &Service{repository: repository} }
func (s *Service) Stores(ctx context.Context) ([]Store, error) { return s.repository.Stores(ctx) }
func (s *Service) Create(ctx context.Context, input Input) error {
	input, err := normalize(input)
	if err != nil {
		return err
	}
	return duplicate(s.repository.Create(ctx, input))
}
func (s *Service) Update(ctx context.Context, id uint64, input Input) error {
	if _, err := s.repository.Store(ctx, id); err != nil {
		return err
	}
	input, err := normalize(input)
	if err != nil {
		return err
	}
	return duplicate(s.repository.Update(ctx, id, input))
}
func (s *Service) Delete(ctx context.Context, id, currentStoreID uint64) error {
	if id == currentStoreID {
		return ErrProtected
	}
	stores, err := s.repository.Stores(ctx)
	if err != nil {
		return err
	}
	for _, item := range stores {
		if item.ID == id && item.UserCount > 0 {
			return ErrInUse
		}
	}
	if _, err := s.repository.Store(ctx, id); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}
func normalize(input Input) (Input, error) {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)
	input.Address = strings.TrimSpace(input.Address)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Timezone = strings.TrimSpace(input.Timezone)
	if input.Status != "active" && input.Status != "inactive" {
		return Input{}, ErrInvalid
	}
	if !codePattern.MatchString(input.Code) || input.Name == "" || utf8.RuneCountInString(input.Name) > 150 || len(input.Address) > 2000 || len(input.Phone) > 30 || len(input.Timezone) > 64 || !timezonePattern.MatchString(input.Timezone) {
		return Input{}, ErrInvalid
	}
	return input, nil
}
func duplicate(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrDuplicate
	}
	return err
}
