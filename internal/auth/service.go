package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/21v1u5/api_rest_jwt/internal/user"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type UserStore interface {
	GetByEmail(ctx context.Context, email string) (*user.User, error)
}

type RefreshStore interface {
	Create(ctx context.Context, userID int64, hash string, expiresAt time.Time) error
	Rotate(ctx context.Context, oldHash, newHash string, expiresAt time.Time) (int64, error)
	Delete(ctx context.Context, hash string) error
	DeleteAllForUser(ctx context.Context, userID int64) error
}

type Service struct {
	users      UserStore
	refresh    RefreshStore
	tokens     *TokenManager
	refreshTTL time.Duration
	dummyHash  []byte
}

func NewService(users UserStore, refresh RefreshStore, tokens *TokenManager, refreshTTL time.Duration) *Service {
	dummy, err := bcrypt.GenerateFromPassword([]byte("timing-attack-protection"), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("generate dummy hash: %v", err))
	}
	return &Service{
		users:      users,
		refresh:    refresh,
		tokens:     tokens,
		refreshTTL: refreshTTL,
		dummyHash:  dummy,
	}
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
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

	return s.issueTokens(ctx, u.ID)

}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *Service) issueTokens(ctx context.Context, userID int64) (*TokenResponse, error) {
	plain, hash, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	if err := s.refresh.Create(ctx, userID, hash, time.Now().Add(s.refreshTTL)); err != nil {
		return nil, err
	}

	return s.buildResponse(userID, plain)
}

func (s *Service) Refresh(ctx context.Context, in RefreshInput) (*TokenResponse, error) {
	if in.RefreshToken == "" {
		return nil, ErrInvalidToken
	}

	plain, newHash, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	userID, err := s.refresh.Rotate(ctx, hashToken(in.RefreshToken), newHash, time.Now().Add(s.refreshTTL))
	if err != nil {
		return nil, err
	}

	return s.buildResponse(userID, plain)
}

func (s *Service) Logout(ctx context.Context, in RefreshInput) error {
	if in.RefreshToken == "" {
		return nil
	}
	return s.refresh.Delete(ctx, hashToken(in.RefreshToken))
}

func (s *Service) LogoutAll(ctx context.Context, userID int64) error {
	return s.refresh.DeleteAllForUser(ctx, userID)
}

func (s *Service) buildResponse(userID int64, refreshPlain string) (*TokenResponse, error) {
	access, err := s.tokens.GenerateAccess(userID)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  access,
		RefreshToken: refreshPlain,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.tokens.AccessTTL().Seconds()),
	}, nil
}
