package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

// MatchValidationRule defines a validation rule for match creation.
// Validation runs before producing MatchCreated (PotentialMatchFound → validation → MatchCreated).
type MatchValidationRule interface {
	Validate(ctx context.Context, params MatchValidationParams) error
}

// MatchValidationParams holds the context for match validation.
type MatchValidationParams struct {
	Pair            *pairing_entities.Pair
	Envelope        *schemas.EventEnvelope
	Payload         *schemas.PlayerQueuedPayload
	ResourceOwnerID string
	TenantID        string
	ClientID        string
}

// MatchValidator runs all validation rules before MatchCreated is produced.
type MatchValidator struct {
	rules []MatchValidationRule
}

// NewMatchValidator creates a validator with the default rules.
func NewMatchValidator() *MatchValidator {
	return &MatchValidator{
		rules: []MatchValidationRule{
			&ResourceOwnershipRule{},
			&RequiredFieldsRule{},
			&PairIntegrityRule{},
		},
	}
}

// Validate runs all rules. Returns the first error encountered.
func (v *MatchValidator) Validate(ctx context.Context, params MatchValidationParams) error {
	for _, rule := range v.rules {
		if err := rule.Validate(ctx, params); err != nil {
			return err
		}
	}
	return nil
}

// ResourceOwnershipRule ensures resource_owner_id is present (Epic §9).
type ResourceOwnershipRule struct{}

func (r *ResourceOwnershipRule) Validate(_ context.Context, params MatchValidationParams) error {
	if strings.TrimSpace(params.ResourceOwnerID) == "" {
		return fmt.Errorf("match validation: resource_owner_id is required for MatchCreated")
	}
	return nil
}

// RequiredFieldsRule ensures tenant_id and client_id are present for downstream access control.
type RequiredFieldsRule struct{}

// Validate ensures tenant_id and client_id are present for downstream access control.
func (r *RequiredFieldsRule) Validate(_ context.Context, params MatchValidationParams) error {
	if strings.TrimSpace(params.TenantID) == "" {
		return fmt.Errorf("match validation: tenant_id is required for MatchCreated")
	}
	if strings.TrimSpace(params.ClientID) == "" {
		return fmt.Errorf("match validation: client_id is required for MatchCreated")
	}
	return nil
}

// PairIntegrityRule ensures the pair exists and has players.
type PairIntegrityRule struct{}

// Validate ensures the pair exists and has players.
func (r *PairIntegrityRule) Validate(_ context.Context, params MatchValidationParams) error {
	if params.Pair == nil {
		return fmt.Errorf("match validation: pair is nil")
	}
	if len(params.Pair.Match) == 0 {
		return fmt.Errorf("match validation: pair has no players")
	}
	if params.Pair.ID == uuid.Nil {
		return fmt.Errorf("match validation: pair has no match_id")
	}
	return nil
}
