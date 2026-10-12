package checklist

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrInvalid   = errors.New("invalid checklist")
	ErrDuplicate = errors.New("duplicate checklist")
	ErrInUse     = errors.New("checklist in use")
)

var codePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9._-]{1,49}$`)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }
func (s *Service) Checklists(ctx context.Context) ([]Checklist, error) {
	return s.repository.Checklists(ctx)
}

func (s *Service) Create(ctx context.Context, input Input) error {
	input, err := normalize(input)
	if err != nil {
		return err
	}
	return databaseError(s.repository.Create(ctx, input))
}

func (s *Service) Update(ctx context.Context, id uint64, input Input) error {
	if _, err := s.repository.Checklist(ctx, id); err != nil {
		return err
	}
	input, err := normalize(input)
	if err != nil {
		return err
	}
	return databaseError(s.repository.Update(ctx, id, input))
}

func (s *Service) Delete(ctx context.Context, id uint64) error {
	if _, err := s.repository.Checklist(ctx, id); err != nil {
		return err
	}
	return databaseError(s.repository.Delete(ctx, id))
}

func normalize(input Input) (Input, error) {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)
	if !codePattern.MatchString(input.Code) || (input.Phase != "start" && input.Phase != "end") || input.Name == "" || utf8.RuneCountInString(input.Name) > 255 || (input.Status != "active" && input.Status != "inactive") {
		return Input{}, ErrInvalid
	}
	return input, nil
}

func databaseError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		if mysqlErr.Number == 1062 {
			return ErrDuplicate
		}
		if mysqlErr.Number == 1451 {
			return ErrInUse
		}
	}
	return err
}
