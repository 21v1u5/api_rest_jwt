package auth

import (
	"net/http"
	"strings"

	"github.com/21v1u5/api_rest_jwt/internal/httpx"
	"github.com/21v1u5/api_rest_jwt/internal/reqctx"
)

func (m *TokenManager) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || tokenStr == "" {
			unauthorized(w, "missing or malformed authorization header")
			return
		}

		userID, err := m.ParseAccess(tokenStr)
		if err != nil {
			unauthorized(w, "invalid or expired token")
			return
		}

		ctx := reqctx.WithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	httpx.Error(w, http.StatusUnauthorized, message)
}
