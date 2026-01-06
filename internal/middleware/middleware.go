package middleware

import (
	"context"
	"encoding/json"

	"log/slog"
	"mastery-project/internal/repository"
	"net/http"
)

type AuthMiddleware struct {
	Session *repository.SessionRepository
}

func NewAuthMiddleware(sessionRepo *repository.SessionRepository) *AuthMiddleware {
	return &AuthMiddleware{
		Session: sessionRepo,
	}
}

type contextKey string

const userContextKey contextKey = "user"

func (m *AuthMiddleware) Protected(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sessionCookie, err := r.Cookie("sessionToken")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			err := json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    "UNAUTHORIZED",
				"message": "Missing or invalid session",
			})
			if err != nil {
				return
			}
			return
		}

		ctx := r.Context()

		user, err := m.Session.GetUserBySessionID(ctx, sessionCookie.Value)
		if err != nil {
			slog.Warn("invalid session", "err", err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			err := json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    "UNAUTHORIZED",
				"message": "Session expired or invalid",
			})
			if err != nil {
				return
			}
			return
		}

		// attach user to context
		ctx = context.WithValue(r.Context(), userContextKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
