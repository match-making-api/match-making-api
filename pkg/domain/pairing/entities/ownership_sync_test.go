package entities

import "testing"

func TestPlayerRating_EnsureResourceOwner(t *testing.T) {
	t.Parallel()
	r := &PlayerRating{
		TenantID:        "11111111-1111-1111-1111-111111111111",
		ClientID:        "22222222-2222-2222-2222-222222222222",
		ResourceOwnerID: "33333333-3333-3333-3333-333333333333",
	}
	if err := r.ValidateOwnership(); err != nil {
		t.Fatal(err)
	}
	if r.ResourceOwner.TenantID.String() != r.TenantID {
		t.Fatalf("tenant mismatch")
	}
}

func TestMatchResult_ValidateOwnership_Missing(t *testing.T) {
	t.Parallel()
	m := &MatchResult{}
	if err := m.ValidateOwnership(); err == nil {
		t.Fatal("expected error")
	}
}
