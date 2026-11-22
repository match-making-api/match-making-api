package mapping

import (
	"testing"

	"github.com/google/uuid"
)

func TestResolvePartyID_SoloDefault(t *testing.T) {
	t.Parallel()
	player := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	party, solo := ResolvePartyID(player, "")
	if !solo || party != player {
		t.Fatalf("expected solo party=player, got solo=%v party=%s", solo, party)
	}
}

func TestResolvePartyID_SharedParty(t *testing.T) {
	t.Parallel()
	player := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	partyStr := "22222222-2222-2222-2222-222222222222"
	party, solo := ResolvePartyID(player, partyStr)
	if solo || party.String() != partyStr {
		t.Fatalf("expected shared party, got solo=%v party=%s", solo, party)
	}
}

func TestMapQueueJoin(t *testing.T) {
	t.Parallel()
	player := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	game := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	m := MapQueueJoin(player, game, "br", "t1", "c1", "ro1", "")
	if m.PartyID != player || !m.Solo || m.RegionSlug != "br" {
		t.Fatalf("unexpected mapping: %+v", m)
	}
}

func TestMapQueueJoin_SharedPartyAndLeaveParity(t *testing.T) {
	t.Parallel()
	player := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	game := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	party := "cccccccc-cccc-cccc-cccc-cccccccccccc"
	join := MapQueueJoin(player, game, "br", "t1", "c1", "ro1", party)
	leave := MapQueueJoin(player, game, "br", "t1", "c1", "ro1", party)
	if join.Solo || leave.Solo {
		t.Fatalf("expected shared party on join/leave")
	}
	if join.PartyID != leave.PartyID || join.PartyID.String() != party {
		t.Fatalf("join/leave party mismatch: join=%s leave=%s", join.PartyID, leave.PartyID)
	}
	if join.TenantID != leave.TenantID || join.ClientID != leave.ClientID || join.ResourceOwnerID != leave.ResourceOwnerID {
		t.Fatalf("ownership mismatch on leave mapping")
	}
}
