package middlewares

import (
	"context"
	"net/http"
	"strings"
)

// AuthMiddleware validates incoming requests with a Bearer token.
type AuthMiddleware struct {
}

// NewAuthMiddleware creates a new instance of AuthMiddleware.
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// Handler is a middleware function that validates the Authorization header of incoming HTTP requests.
//
// Parameters:
//   - next: The next http.Handler in the chain to be called if authentication is successful.
//
// Returns:
//   - http.Handler: A new http.Handler that wraps the authentication logic around the next handler.
//     If authentication fails, it responds with an Unauthorized status. If successful, it calls the next handler.
func (am *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			http.Error(w, "no-auth", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authorizationHeader, "Bearer ")
		if len(bearerToken) != 2 {
			http.Error(w, "no-auth", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.Background()))
	})
}
