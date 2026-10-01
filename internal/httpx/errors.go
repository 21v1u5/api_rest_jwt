package httpx

import (
	"log/slog"
	"net/http"
)

type errorBody struct {
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func Error(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]errorBody{
		"error": {Message: message},
	})
}

func ValidationError(w http.ResponseWriter, details map[string]string) {
	WriteJSON(w, http.StatusUnprocessableEntity, map[string]errorBody{
		"error": {Message: "validation failed", Details: details},
	})
}

func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("internal error",
		"err", err,
		"method", r.Method,
		"path", r.URL.Path,
	)
	Error(w, http.StatusInternalServerError, "internal server error")
}
