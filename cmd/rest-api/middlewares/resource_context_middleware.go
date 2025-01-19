package middlewares

import (
	"context"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	common "github.com/leet-gaming/match-making-api/pkg/domain"
)

// ResourceContextMiddleware handles resource context for incoming requests.
type ResourceContextMiddleware struct {
}

// NewResourceContextMiddleware creates a new ResourceContextMiddleware.
func NewResourceContextMiddleware(container *container.Container) *ResourceContextMiddleware {
	return &ResourceContextMiddleware{}
}

// Handler is a middleware function that adds resource context to incoming HTTP requests.
//
// Parameters:
//   - next: The next handler in the middleware chain to be called after this middleware.
//
// Returns:
//   - http.Handler: A new HTTP handler that wraps the provided next handler with added context.
func (m *ResourceContextMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), common.TenantIDKey, common.TeamPROTenantID)
		ctx = context.WithValue(ctx, common.ClientIDKey, common.TeamPROAppClientID)
		ctx = context.WithValue(ctx, common.GroupIDKey, uuid.New())
		ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())

		rid := r.Header.Get("X-Resource-Owner-ID")
		if rid == "" {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
