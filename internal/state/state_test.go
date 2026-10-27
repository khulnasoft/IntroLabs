package state

import (
	"path/filepath"
	"testing"
)

func TestStoreCreateAndGet(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	inst := store.Create("web/intro-lab", "1.0.0", "/path/to/template", "compose.yaml", "labkit-web-intro-lab", []int{8081}, []string{"http://127.0.0.1:8081"})

	if inst.LabID != "web/intro-lab" {
		t.Errorf("expected labID web/intro-lab, got %s", inst.LabID)
	}
	if inst.Status != "running" {
		t.Errorf("expected status running, got %s", inst.Status)
	}
	if len(inst.Ports) != 1 || inst.Ports[0] != 8081 {
		t.Errorf("expected ports [8081], got %v", inst.Ports)
	}

	got := store.Get("web/intro-lab")
	if got == nil {
		t.Fatal("expected to get created instance")
	}
	if got.Status != "running" {
		t.Errorf("expected status running, got %s", got.Status)
	}
}

func TestStoreUpdateStatus(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	store.Create("web/intro-lab", "1.0.0", "/path", "compose.yaml", "proj", []int{8081}, []string{"http://127.0.0.1:8081"})

	if err := store.UpdateStatus("web/intro-lab", "stopped"); err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	inst := store.Get("web/intro-lab")
	if inst.Status != "stopped" {
		t.Errorf("expected status stopped, got %s", inst.Status)
	}
	if inst.StoppedAt == nil {
		t.Error("expected stoppedAt to be set")
	}
}

func TestStoreRemove(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	store.Create("web/intro-lab", "1.0.0", "/path", "compose.yaml", "proj", []int{8081}, []string{"http://127.0.0.1:8081"})

	if err := store.Remove("web/intro-lab"); err != nil {
		t.Fatalf("failed to remove: %v", err)
	}

	if store.Get("web/intro-lab") != nil {
		t.Error("expected instance to be removed")
	}

	// Removing again should error
	if err := store.Remove("web/intro-lab"); err == nil {
		t.Error("expected error removing non-existent instance")
	}
}

func TestStoreList(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	store.Create("web/intro-lab", "1.0.0", "/path", "compose.yaml", "proj", []int{8081}, []string{"http://127.0.0.1:8081"})
	store.Create("api/auth-lab", "2.0.0", "/path2", "compose.yaml", "proj2", []int{8082}, []string{"http://127.0.0.1:8082"})

	instances := store.List()
	if len(instances) != 2 {
		t.Errorf("expected 2 instances, got %d", len(instances))
	}

	running := store.Running()
	if len(running) != 2 {
		t.Errorf("expected 2 running instances, got %d", len(running))
	}
}

func TestStorePersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()

	// Create first store and add an instance
	store1, err := NewStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	store1.Create("web/intro-lab", "1.0.0", "/path", "compose.yaml", "proj", []int{8081}, []string{"http://127.0.0.1:8081"})

	// Create a second store pointing to same dir - should load existing state
	store2, err := NewStore(dir)
	if err != nil {
		t.Fatalf("failed to create second store: %v", err)
	}

	inst := store2.Get("web/intro-lab")
	if inst == nil {
		t.Fatal("expected instance to persist across store instances")
	}
	if inst.Status != "running" {
		t.Errorf("expected status running, got %s", inst.Status)
	}
}

var _ = filepath.Join // keep import used if needed
