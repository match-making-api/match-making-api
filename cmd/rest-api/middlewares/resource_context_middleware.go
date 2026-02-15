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
		VerifyRID:    verifyRID,
		OperationMap: make(map[string]string),
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
		// SECURITY: Do NOT set UserIDKey to uuid.New(). It must come from verified RID token.
		// Setting it to uuid.Nil ensures handlers that check for auth will reject unauthenticated requests.
		ctx = context.WithValue(ctx, common.UserIDKey, uuid.Nil)

		rid := r.Header.Get(controllers.ResourceOwnerIDHeaderKey)
		if rid == "" {
			slog.WarnContext(ctx, "missing resource owner id", "ResourceOwnerIDHeaderKey", controllers.ResourceOwnerIDHeaderKey)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ridUUID, parseErr := uuid.Parse(rid)
		if parseErr != nil {
			slog.ErrorContext(ctx, "invalid rid format", controllers.ResourceOwnerIDHeaderKey, rid)
			http.Error(w, `{"error":"invalid resource owner id"}`, http.StatusBadRequest)
			return
		}

		reso, err := m.VerifyRID.Exec(ctx, ridUUID, operationID)
		if err != nil {
			slog.ErrorContext(ctx, "unable to verify rid", controllers.ResourceOwnerIDHeaderKey, rid)
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		if reso != nil && !reso.GetIsValid() {
			slog.WarnContext(ctx, "rid verification failed", "reason", reso.GetReason())
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// SECURITY: Extract identity from the verified RID token's ResourceOwner.
		// The RID token is the single source of truth — all identity fields come from it.
		// DO NOT use headers (X-User-ID etc.) for identity — they are client-controlled.
		if ro := reso.GetResourceOwner(); ro != nil {
			if tenantID, err := uuid.Parse(ro.GetTenantId()); err == nil && tenantID != uuid.Nil {
				ctx = context.WithValue(ctx, common.TenantIDKey, tenantID)
			}
			if clientID, err := uuid.Parse(ro.GetClientId()); err == nil && clientID != uuid.Nil {
				ctx = context.WithValue(ctx, common.ClientIDKey, clientID)
			}
			if groupID, err := uuid.Parse(ro.GetGroupId()); err == nil && groupID != uuid.Nil {
				ctx = context.WithValue(ctx, common.GroupIDKey, groupID)
			}
			if userID, err := uuid.Parse(ro.GetUserId()); err == nil && userID != uuid.Nil {
				ctx = context.WithValue(ctx, common.UserIDKey, userID)
			}
		}

		// Set intended audience from verified RID token
		ctx = context.WithValue(ctx, common.AudienceKey, common.IntendedAudienceKey(reso.GetIntendedAudience()))

		// Mark request as authenticated
		ctx = context.WithValue(ctx, common.AuthenticatedKey, true)

		slog.InfoContext(ctx, "resource owner verified",
			"user_id", ctx.Value(common.UserIDKey),
			"group_id", ctx.Value(common.GroupIDKey),
			"audience", reso.GetIntendedAudience(),
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *ResourceContextMiddleware) RegisterOperation(path, operationID string) {
	m.OperationMap[path] = operationID
}
