package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/21v1u5/api_rest_jwt/internal/httpx"
	"github.com/21v1u5/api_rest_jwt/internal/reqctx"
)

type service interface {
	Login(ctx context.Context, in LoginInput) (*TokenResponse, error)
	Refresh(ctx context.Context, in RefreshInput) (*TokenResponse, error)
	Logout(ctx context.Context, in RefreshInput) error
	LogoutAll(ctx context.Context, userID int64) error
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux, requireAuth func(http.Handler) http.Handler) {
	mux.HandleFunc("POST /auth/login", h.login)
	mux.HandleFunc("POST /auth/refresh", h.refresh)
	mux.HandleFunc("POST /auth/logout", h.logout)
	mux.Handle("POST /auth/logout-all", requireAuth(http.HandlerFunc(h.logoutAll)))
}

type Handler struct {
	svc service
}

func NewHandler(svc service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var in LoginInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	tokens, err := h.svc.Login(r.Context(), in)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpx.Error(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		httpx.ServerError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, tokens)
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var in RefreshInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	tokens, err := h.svc.Refresh(r.Context(), in)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			httpx.Error(w, http.StatusUnauthorized, "invalid or expired refresh token")
			return
		}
		httpx.ServerError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, tokens)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	var in RefreshInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.Logout(r.Context(), in); err != nil {
		httpx.ServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) logoutAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := reqctx.UserID(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.svc.LogoutAll(r.Context(), userID); err != nil {
		httpx.ServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
