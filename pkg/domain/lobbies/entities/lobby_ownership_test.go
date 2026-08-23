package entities

import (
	"testing"

	"github.com/google/uuid"
)

func TestLobby_EnsureResourceOwner_FromFlat(t *testing.T) {
	t.Parallel()
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	client := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	creator := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	l := &Lobby{TenantID: tenant, ClientID: client, CreatorID: creator}
	if err := l.ValidateOwnership(); err != nil {
		t.Fatal(err)
	}
	if l.ResourceOwner.TenantID != tenant || l.ResourceOwner.ClientID != client || l.ResourceOwner.UserID != creator {
		t.Fatalf("resource_owner not synced: %+v", l.ResourceOwner)
	}
}

func TestLobby_ValidateOwnership_Missing(t *testing.T) {
	t.Parallel()
	l := &Lobby{}
	if err := l.ValidateOwnership(); err == nil {
		t.Fatal("expected missing ownership error")
	}
}
