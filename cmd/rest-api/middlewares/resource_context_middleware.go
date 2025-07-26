package middlewares

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leet-gaming/match-making-api/cmd/rest-api/controllers"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/iam/ports/in"
)

// ResourceContextMiddleware handles resource context for incoming requests.
type ResourceContextMiddleware struct {
	VerifyRID    in.VerifyRIDKeyCommand
	OperationMap map[string]string
}

// NewResourceContextMiddleware creates a new ResourceContextMiddleware.
func NewResourceContextMiddleware(container *container.Container) *ResourceContextMiddleware {
	var verifyRID in.VerifyRIDKeyCommand
	err := container.Resolve(&verifyRID)

	if err != nil {
		slog.Error("unable to resolve VerifyRIDKeyCommand")
	}

	return &ResourceContextMiddleware{
		VerifyRID: verifyRID,
	}
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
		ctx := r.Context()
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		operationID := m.OperationMap[path]

		slog.InfoContext(r.Context(), "resource context middleware", "path", r.URL.Path, "method", r.Method, "rid", r.Header.Get(controllers.ResourceOwnerIDHeaderKey))
		ctx = context.WithValue(ctx, common.TenantIDKey, common.TeamPROTenantID)
		ctx = context.WithValue(ctx, common.ClientIDKey, common.TeamPROAppClientID)
		ctx = context.WithValue(ctx, common.GroupIDKey, uuid.New())
		ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())

		rid := r.Header.Get(controllers.ResourceOwnerIDHeaderKey)
		if rid == "" {
			slog.WarnContext(ctx, "missing resource owner id", "ResourceOwnerIDHeaderKey", controllers.ResourceOwnerIDHeaderKey)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		reso, err := m.VerifyRID.Exec(ctx, uuid.MustParse(rid), operationID)
		if err != nil {
			slog.ErrorContext(ctx, "unable to verify rid", controllers.ResourceOwnerIDHeaderKey, rid)
			http.Error(w, "unknown", http.StatusUnauthorized)
		}

		slog.InfoContext(ctx, "resource owner verified", "reso", reso)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *ResourceContextMiddleware) RegisterOperation(path, operationID string) {
	m.OperationMap[path] = operationID
}
