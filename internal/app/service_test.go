package app

import "testing"

func TestServiceHealth(t *testing.T) {
	t.Parallel()

	service := NewService()

	got := service.Health()

	if got.Status != "ok" {
		t.Fatalf("Health().Status = %q, want %q", got.Status, "ok")
	}
	if got.Version == "" {
		t.Fatal("Health().Version is empty")
	}
}
