package jwtauth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestValidator_ValidateHMAC_OK(t *testing.T) {
	secret := []byte("test-secret-at-least-32-bytes-long!!")
	user := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tenant := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	client := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	token, err := SignTestToken(secret, jwt.MapClaims{
		"user_id":   user.String(),
		"tenant_id": tenant.String(),
		"client_id": client.String(),
		"roles":     []string{"player"},
		"iss":       "replay-api",
		"aud":       "match-making-api",
	}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	v := NewValidator(Config{HMACSecret: secret, Issuer: "replay-api", Audience: "match-making-api"})
	got, err := v.ValidateHMAC(token)
	if err != nil {
		t.Fatalf("ValidateHMAC: %v", err)
	}
	if got.UserID != user || got.TenantID != tenant || got.ClientID != client {
		t.Fatalf("claims mismatch: %+v", got)
	}
	if len(got.Roles) != 1 || got.Roles[0] != "player" {
		t.Fatalf("roles: %v", got.Roles)
	}
}

func TestValidator_ValidateHMAC_Expired(t *testing.T) {
	secret := []byte("test-secret-at-least-32-bytes-long!!")
	token, err := SignTestToken(secret, jwt.MapClaims{
		"sub":       "11111111-1111-1111-1111-111111111111",
		"tenant_id": "22222222-2222-2222-2222-222222222222",
		"client_id": "33333333-3333-3333-3333-333333333333",
	}, -time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	v := NewValidator(Config{HMACSecret: secret})
	_, err = v.ValidateHMAC(token)
	if !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("want ErrExpiredToken, got %v", err)
	}
}

func TestValidator_ValidateHMAC_MissingTenant(t *testing.T) {
	secret := []byte("test-secret-at-least-32-bytes-long!!")
	token, err := SignTestToken(secret, jwt.MapClaims{
		"user_id":   "11111111-1111-1111-1111-111111111111",
		"client_id": "33333333-3333-3333-3333-333333333333",
	}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	v := NewValidator(Config{HMACSecret: secret})
	_, err = v.ValidateHMAC(token)
	if !errors.Is(err, ErrMissingTenant) {
		t.Fatalf("want ErrMissingTenant, got %v", err)
	}
}

func TestValidator_ValidateHMAC_InvalidSignature(t *testing.T) {
	token, err := SignTestToken([]byte("correct-secret-at-least-32-bytes!!"), jwt.MapClaims{
		"user_id":   "11111111-1111-1111-1111-111111111111",
		"tenant_id": "22222222-2222-2222-2222-222222222222",
		"client_id": "33333333-3333-3333-3333-333333333333",
	}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	v := NewValidator(Config{HMACSecret: []byte("wrong-secret-at-least-32-bytes-xx")})
	_, err = v.ValidateHMAC(token)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("want ErrInvalidToken, got %v", err)
	}
}
