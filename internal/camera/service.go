package camera

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrInvalid   = errors.New("invalid camera")
	ErrDuplicate = errors.New("duplicate camera")
	ErrInUse     = errors.New("camera in use")
)

var codePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9._-]{1,49}$`)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service                     { return &Service{repository: repository} }
func (s *Service) Cameras(ctx context.Context) ([]Camera, error)     { return s.repository.Cameras(ctx) }
func (s *Service) Stores(ctx context.Context) ([]StoreOption, error) { return s.repository.Stores(ctx) }
func (s *Service) Areas(ctx context.Context) ([]AreaOption, error)   { return s.repository.Areas(ctx) }

func (s *Service) Create(ctx context.Context, input Input) error {
	input, err := s.normalize(ctx, input)
	if err != nil {
		return err
	}
	return databaseError(s.repository.Create(ctx, input))
}

func (s *Service) Update(ctx context.Context, id uint64, input Input) error {
	if _, err := s.repository.Camera(ctx, id); err != nil {
		return err
	}
	input, err := s.normalize(ctx, input)
	if err != nil {
		return err
	}
	return databaseError(s.repository.Update(ctx, id, input))
}

func (s *Service) Delete(ctx context.Context, id uint64) error {
	if _, err := s.repository.Camera(ctx, id); err != nil {
		return err
	}
	return databaseError(s.repository.Delete(ctx, id))
}

func (s *Service) normalize(ctx context.Context, input Input) (Input, error) {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.StoreID == 0 || input.AreaID == 0 || !codePattern.MatchString(input.Code) || input.Name == "" || utf8.RuneCountInString(input.Name) > 150 || utf8.RuneCountInString(input.Description) > 2000 || (input.Status != "active" && input.Status != "inactive") {
		return Input{}, ErrInvalid
	}
	exists, err := s.repository.AreaBelongsToStore(ctx, input.AreaID, input.StoreID)
	if err != nil {
		return Input{}, err
	}
	if !exists {
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
