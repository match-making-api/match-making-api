package mongodb

import "testing"

func TestResourceOwnershipIndexModels_Names(t *testing.T) {
	t.Parallel()
	models := ResourceOwnershipIndexModels()
	if len(models) < 3 {
		t.Fatalf("expected at least 3 ownership indexes, got %d", len(models))
	}
	want := map[string]bool{
		"idx_ro_tenant_client": false,
		"idx_ro_tenant_user":   false,
		"idx_ro_tenant_group":  false,
	}
	for _, m := range models {
		if m.Options == nil || m.Options.Name == nil {
			t.Fatal("index missing name")
		}
		name := *m.Options.Name
		if _, ok := want[name]; ok {
			want[name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("missing index name %s", name)
		}
	}
}

func TestFlatTenantClientIndexModels(t *testing.T) {
	t.Parallel()
	models := FlatTenantClientIndexModels()
	if len(models) != 1 || models[0].Options == nil || models[0].Options.Name == nil {
		t.Fatal("expected named flat tenant/client index")
	}
	if *models[0].Options.Name != "idx_flat_tenant_client" {
		t.Fatalf("unexpected name %s", *models[0].Options.Name)
	}
}
