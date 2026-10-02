package task

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/21v1u5/api_rest_jwt/internal/httpx"
	"github.com/21v1u5/api_rest_jwt/internal/reqctx"
	"github.com/21v1u5/api_rest_jwt/internal/validator"
)

type service interface {
	Create(ctx context.Context, userID int64, in CreateInput) (*Task, error)
	Get(ctx context.Context, userID, id int64) (*Task, error)
	List(ctx context.Context, userID int64, f ListFilter) ([]Task, Metadata, error)
	Update(ctx context.Context, userID, id int64, in UpdateInput) (*Task, error)
	Delete(ctx context.Context, userID, id int64) error
}

type Handler struct {
	svc service
}

func NewHandler(svc service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux, requireAuth func(http.Handler) http.Handler) {
	mux.Handle("POST /tasks", requireAuth(http.HandlerFunc(h.create)))
	mux.Handle("GET /tasks", requireAuth(http.HandlerFunc(h.list)))
	mux.Handle("GET /tasks/{id}", requireAuth(http.HandlerFunc(h.get)))
	mux.Handle("PATCH /tasks/{id}", requireAuth(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /tasks/{id}", requireAuth(http.HandlerFunc(h.delete)))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(w, r)
	if !ok {
		return
	}

	var in CreateInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	t, err := h.svc.Create(r.Context(), userID, in)
	if err != nil {
		handleError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]*Task{"task": t})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(w, r)
	if !ok {
		return
	}

	qs := r.URL.Query()
	v := validator.New()
	f := ListFilter{
		Page:     httpx.QueryInt(qs, "page", 1, v),
		PageSize: httpx.QueryInt(qs, "page_size", 20, v),
		Done:     httpx.QueryBool(qs, "done", v),
	}
	if !v.Valid() {
		httpx.ValidationError(w, v.Errors)
		return
	}

	tasks, meta, err := h.svc.List(r.Context(), userID, f)
	if err != nil {
		handleError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"tasks":    tasks,
		"metadata": meta,
	})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(w, r)
	if !ok {
		return
	}
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	t, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		handleError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]*Task{"task": t})
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(w, r)
	if !ok {
		return
	}
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var in UpdateInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	t, err := h.svc.Update(r.Context(), userID, id, in)
	if err != nil {
		handleError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]*Task{"task": t})
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(w, r)
	if !ok {
		return
	}
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		handleError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func userIDFrom(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, ok := reqctx.UserID(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
	}
	return id, ok
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		httpx.Error(w, http.StatusBadRequest, "invalid task id")
		return 0, false
	}
	return id, true
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	var vErr *validator.ValidationError
	switch {
	case errors.As(err, &vErr):
		httpx.ValidationError(w, vErr.Fields)
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, "task not found")
	default:
		httpx.ServerError(w, r, err)
	}
}
