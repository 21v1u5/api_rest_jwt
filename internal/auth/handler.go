package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/21v1u5/api_rest_jwt/internal/httpx"
)

type service interface {
	Login(ctx context.Context, in LoginInput) (*TokenResponse, error)
}

type Handler struct {
	svc service
}

func NewHandler(svc service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/login", h.login)
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
