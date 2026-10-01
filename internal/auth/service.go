package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/21v1u5/api_rest_jwt/internal/user"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type UserStore interface {
	GetByEmail(ctx context.Context, email string) (*user.User, error)
}

type Service struct {
	users     UserStore
	tokens    *TokenManager
	dummyHash []byte
}

func NewService(users UserStore, tokens *TokenManager) *Service {
	dummy, err := bcrypt.GenerateFromPassword([]byte("timing-attack-protection"), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("generate dummy hash: %v", err))
	}
	return &Service{users: users, tokens: tokens, dummyHash: dummy}
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (s *Service) Login(ctx context.Context, in LoginInput) (*TokenResponse, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))

	u, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, user.ErrNotFound) {
		bcrypt.CompareHashAndPassword(s.dummyHash, []byte(in.Password))
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("compare password: %w", err)
	}

	access, err := s.tokens.GenerateAccess(u.ID)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken: access,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.tokens.AccessTTL().Seconds()),
	}, nil
}
