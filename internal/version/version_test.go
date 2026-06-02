package version

import "testing"

func TestCurrent(t *testing.T) {
	original := Value
	t.Cleanup(func() {
		Value = original
	})
	Value = "1.2.3"

	if got := Current(); got != "1.2.3" {
		t.Fatalf("Current() = %q, want %q", got, "1.2.3")
	}
}
