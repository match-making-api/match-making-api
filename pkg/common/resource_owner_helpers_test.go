package common

import "testing"

func TestResourceOwner_ValidateTenantClient(t *testing.T) {
	t.Parallel()
	ok := ResourceOwner{
		TenantID: ParseUUIDOrNil("11111111-1111-1111-1111-111111111111"),
		ClientID: ParseUUIDOrNil("22222222-2222-2222-2222-222222222222"),
	}
	if err := ok.ValidateTenantClient(); err != nil {
		t.Fatalf("expected valid ownership, got %v", err)
	}
	if err := (ResourceOwner{}).ValidateTenantClient(); err == nil {
		t.Fatal("expected error for empty ownership")
	}
}

func TestResourceOwnerFromFlat_RoundTrip(t *testing.T) {
	t.Parallel()
	tenant := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	client := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	user := "cccccccc-cccc-cccc-cccc-cccccccccccc"
	ro := ResourceOwnerFromFlat(tenant, client, "", user)
	if err := ro.ValidateTenantClient(); err != nil {
		t.Fatal(err)
	}
	gotT, gotC, _, gotU := ro.FlatStrings()
	if gotT != tenant || gotC != client || gotU != user {
		t.Fatalf("round-trip mismatch: %s %s %s", gotT, gotC, gotU)
	}
}
