package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/21v1u5/api_rest_jwt/internal/httpx"
	"github.com/21v1u5/api_rest_jwt/internal/validator"
)

type service interface {
	Register(ctx context.Context, in RegisterInput) (*User, error)
}

type Handler struct {
	svc service
}

func NewHandler(svc service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", h.register)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var in RegisterInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	u, err := h.svc.Register(r.Context(), in)
	if err != nil {
		var vErr *validator.ValidationError
		switch {
		case errors.As(err, &vErr):
			httpx.ValidationError(w, vErr.Fields)
		case errors.Is(err, ErrEmailTaken):
			httpx.Error(w, http.StatusConflict, "email already registered")
		default:
			httpx.ServerError(w, r, err)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]*User{"user": u})
}
