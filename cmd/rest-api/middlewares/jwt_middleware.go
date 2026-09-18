package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/leet-gaming/match-making-api/pkg/common"
	jwtauth "github.com/leet-gaming/match-making-api/pkg/domain/iam/jwt"
)

// JWTAuthMiddleware validates Authorization: Bearer JWTs on sync HTTP entry points.
// When a Bearer token is present it must be valid; identity is written to context.
// When absent, the request continues so RID-based ResourceContextMiddleware can authenticate.
type JWTAuthMiddleware struct {
	validator *jwtauth.Validator
}

// NewJWTAuthMiddleware builds middleware from env (JWT_HMAC_SECRET, JWT_ISSUER, JWT_AUDIENCE).
func NewJWTAuthMiddleware() *JWTAuthMiddleware {
	return &JWTAuthMiddleware{validator: jwtauth.NewValidator(jwtauth.ConfigFromEnv())}
}

// NewJWTAuthMiddlewareWithValidator is for tests.
func NewJWTAuthMiddlewareWithValidator(v *jwtauth.Validator) *JWTAuthMiddleware {
	return &JWTAuthMiddleware{validator: v}
}

func (m *JWTAuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			next.ServeHTTP(w, r)
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeAuthError(w, http.StatusUnauthorized, "invalid_authorization", "Authorization must be Bearer <token>")
			return
		}

		claims, err := m.validator.ValidateHMAC(parts[1])
		if err != nil {
			status, code, msg := mapJWTError(err)
			slog.WarnContext(r.Context(), "jwt validation failed", "error", err, "code", code)
			writeAuthError(w, status, code, msg)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, common.TenantIDKey, claims.TenantID)
		ctx = context.WithValue(ctx, common.ClientIDKey, claims.ClientID)
		ctx = context.WithValue(ctx, common.GroupIDKey, claims.GroupID)
		ctx = context.WithValue(ctx, common.UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
		ctx = context.WithValue(ctx, common.AudienceKey, common.UserAudienceIDKey)

		slog.InfoContext(ctx, "jwt authenticated",
			"user_id", claims.UserID,
			"tenant_id", claims.TenantID,
			"client_id", claims.ClientID,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func mapJWTError(err error) (status int, code, message string) {
	switch {
	case errors.Is(err, jwtauth.ErrExpiredToken):
		return http.StatusUnauthorized, "token_expired", "JWT expired"
	case errors.Is(err, jwtauth.ErrMissingToken):
		return http.StatusUnauthorized, "token_missing", "Bearer token required"
	case errors.Is(err, jwtauth.ErrMissingSecret):
		return http.StatusUnauthorized, "auth_misconfigured", "JWT validation is not configured"
	case errors.Is(err, jwtauth.ErrMissingSubject),
		errors.Is(err, jwtauth.ErrMissingTenant),
		errors.Is(err, jwtauth.ErrMissingClient):
		return http.StatusForbidden, "token_claims_invalid", err.Error()
	default:
		return http.StatusUnauthorized, "token_invalid", "JWT validation failed"
	}
}

func writeAuthError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
}
