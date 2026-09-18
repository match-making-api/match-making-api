package jwtauth

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims carries player identity and resource ownership from a sync JWT.
type Claims struct {
	UserID   uuid.UUID
	TenantID uuid.UUID
	ClientID uuid.UUID
	GroupID  uuid.UUID
	Roles    []string
	Raw      jwt.MapClaims
}

// Config controls HMAC JWT validation (issuer/audience optional).
type Config struct {
	HMACSecret []byte
	Issuer     string
	Audience   string
}

// ConfigFromEnv loads JWT settings from environment.
// JWT_HMAC_SECRET is required for validation; empty disables Bearer acceptance.
func ConfigFromEnv() Config {
	return Config{
		HMACSecret: []byte(strings.TrimSpace(os.Getenv("JWT_HMAC_SECRET"))),
		Issuer:     strings.TrimSpace(os.Getenv("JWT_ISSUER")),
		Audience:   strings.TrimSpace(os.Getenv("JWT_AUDIENCE")),
	}
}

var (
	ErrMissingToken     = errors.New("missing bearer token")
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token expired")
	ErrMissingSecret    = errors.New("jwt hmac secret not configured")
	ErrMissingSubject   = errors.New("token missing user subject")
	ErrMissingTenant    = errors.New("token missing tenant_id")
	ErrMissingClient    = errors.New("token missing client_id")
)

// Validator validates player JWTs for sync HTTP entry points.
type Validator struct {
	cfg Config
}

// NewValidator creates a Validator.
func NewValidator(cfg Config) *Validator {
	return &Validator{cfg: cfg}
}

// ValidateHMAC parses and validates an HMAC-signed JWT, extracting ownership claims.
func (v *Validator) ValidateHMAC(tokenString string) (*Claims, error) {
	if len(v.cfg.HMACSecret) == 0 {
		return nil, ErrMissingSecret
	}
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, ErrMissingToken
	}

	opts := []jwt.ParserOption{jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name, jwt.SigningMethodHS384.Name, jwt.SigningMethodHS512.Name})}
	if v.cfg.Issuer != "" {
		opts = append(opts, jwt.WithIssuer(v.cfg.Issuer))
	}
	if v.cfg.Audience != "" {
		opts = append(opts, jwt.WithAudience(v.cfg.Audience))
	}

	parsed, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return v.cfg.HMACSecret, nil
	}, opts...)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !parsed.Valid {
		return nil, ErrInvalidToken
	}

	mapClaims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	userID, err := claimUUID(mapClaims, "user_id", "sub")
	if err != nil || userID == uuid.Nil {
		return nil, ErrMissingSubject
	}
	tenantID, err := claimUUID(mapClaims, "tenant_id")
	if err != nil || tenantID == uuid.Nil {
		return nil, ErrMissingTenant
	}
	clientID, err := claimUUID(mapClaims, "client_id")
	if err != nil || clientID == uuid.Nil {
		return nil, ErrMissingClient
	}
	groupID, _ := claimUUID(mapClaims, "group_id")

	roles := claimStringSlice(mapClaims, "roles")

	return &Claims{
		UserID:   userID,
		TenantID: tenantID,
		ClientID: clientID,
		GroupID:  groupID,
		Roles:    roles,
		Raw:      mapClaims,
	}, nil
}

func claimUUID(claims jwt.MapClaims, keys ...string) (uuid.UUID, error) {
	for _, key := range keys {
		raw, ok := claims[key]
		if !ok || raw == nil {
			continue
		}
		s, ok := raw.(string)
		if !ok {
			return uuid.Nil, fmt.Errorf("claim %s is not a string", key)
		}
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		id, err := uuid.Parse(s)
		if err != nil {
			return uuid.Nil, fmt.Errorf("claim %s: %w", key, err)
		}
		return id, nil
	}
	return uuid.Nil, nil
}

func claimStringSlice(claims jwt.MapClaims, key string) []string {
	raw, ok := claims[key]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	default:
		return nil
	}
}

// MustExpireSoon is a test helper: ensures exp is set when signing test tokens.
func SignTestToken(secret []byte, claims jwt.MapClaims, ttl time.Duration) (string, error) {
	if claims == nil {
		claims = jwt.MapClaims{}
	}
	now := time.Now()
	claims["iat"] = now.Unix()
	claims["exp"] = now.Add(ttl).Unix()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}
