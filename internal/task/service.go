package task

import (
	"context"
	"strings"

	"github.com/21v1u5/api_rest_jwt/internal/validator"
)

type Store interface {
	Create(ctx context.Context, userID int64, title, description string) (*Task, error)
	GetByID(ctx context.Context, userID, id int64) (*Task, error)
	List(ctx context.Context, userID int64, f ListFilter) ([]Task, int, error)
	Update(ctx context.Context, userID, id int64, in UpdateInput) (*Task, error)
	Delete(ctx context.Context, userID, id int64) error
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

type CreateInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (in *CreateInput) normalize() {
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
}

func (in CreateInput) validate() error {
	v := validator.New()
	v.Check(validator.NotBlank(in.Title), "title", "must not be blank")
	v.Check(validator.MaxChars(in.Title, 200), "title", "must not exceed 200 characters")
	v.Check(validator.MaxChars(in.Description, 2000), "description", "must not exceed 2000 characters")
	return v.Err()
}

type UpdateInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Done        *bool   `json:"done"`
}

func (in *UpdateInput) normalize() {
	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		in.Title = &t
	}
	if in.Description != nil {
		d := strings.TrimSpace(*in.Description)
		in.Description = &d
	}
}

func (in UpdateInput) validate() error {
	v := validator.New()
	v.Check(in.Title != nil || in.Description != nil || in.Done != nil,
		"body", "must contain at least one field")

	if in.Title != nil {
		v.Check(validator.NotBlank(*in.Title), "title", "must not be blank")
		v.Check(validator.MaxChars(*in.Title, 200), "title", "must not exceed 200 characters")
	}
	if in.Description != nil {
		v.Check(validator.MaxChars(*in.Description, 2000), "description", "must not exceed 2000 characters")
	}
	return v.Err()
}

func (f ListFilter) validate() error {
	v := validator.New()
	v.Check(f.Page >= 1, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000, "page", "must not exceed 10000")
	v.Check(f.PageSize >= 1 && f.PageSize <= 100, "page_size", "must be between 1 and 100")
	return v.Err()
}

func (s *Service) Create(ctx context.Context, userID int64, in CreateInput) (*Task, error) {
	in.normalize()
	if err := in.validate(); err != nil {
		return nil, err
	}
	return s.store.Create(ctx, userID, in.Title, in.Description)
}

func (s *Service) Get(ctx context.Context, userID, id int64) (*Task, error) {
	return s.store.GetByID(ctx, userID, id)
}

func (s *Service) List(ctx context.Context, userID int64, f ListFilter) ([]Task, Metadata, error) {
	if err := f.validate(); err != nil {
		return nil, Metadata{}, err
	}

	tasks, total, err := s.store.List(ctx, userID, f)
	if err != nil {
		return nil, Metadata{}, err
	}

	meta := Metadata{
		Page:       f.Page,
		PageSize:   f.PageSize,
		Total:      total,
		TotalPages: (total + f.PageSize - 1) / f.PageSize,
	}
	return tasks, meta, nil
}

func (s *Service) Update(ctx context.Context, userID, id int64, in UpdateInput) (*Task, error) {
	in.normalize()
	if err := in.validate(); err != nil {
		return nil, err
	}
	return s.store.Update(ctx, userID, id, in)
}

func (s *Service) Delete(ctx context.Context, userID, id int64) error {
	return s.store.Delete(ctx, userID, id)
}
