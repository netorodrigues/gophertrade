package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// MockAuth is a simple middleware that extracts User-ID from headers
func MockAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("User-ID")
		if userID == "" {
			// For prototype/mock, we can allow anonymous or set a default
			userID = "anonymous"
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID retrieves the User-ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return "anonymous"
}
