package user

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/21v1u5/api_rest_jwt/internal/validator"
)

type Store interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*User, error) {
	return s.store.GetByID(ctx, id)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (in *RegisterInput) normalize() {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
}

func (in RegisterInput) validate() error {
	v := validator.New()

	v.Check(validator.NotBlank(in.Name), "name", "must not be blank")
	v.Check(validator.MaxChars(in.Name, 100), "name", "must not exceed 100 characters")

	v.Check(validator.NotBlank(in.Email), "email", "must not be blank")
	v.Check(validator.Matches(in.Email, validator.EmailRX), "email", "must be a valid email")

	v.Check(validator.MinChars(in.Password, 8), "password", "must be at least 8 characters")
	v.Check(len(in.Password) <= 72, "password", "must not exceed 72 bytes")

	return v.Err()
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*User, error) {
	in.normalize()

	if err := in.validate(); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &User{
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: string(hash),
	}

	if err := s.store.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}
