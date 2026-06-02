package library_test

import (
	"strings"
	"testing"

	"github.com/alexmchughdev/bootwrangler/internal/library"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
)

func TestVersionedLibrary_InitAndCommit(t *testing.T) {
	dir := t.TempDir()
	vlib := library.NewVersioned(dir)
	if err := vlib.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	p := minimalProfile("versioned-profile")
	filename, err := vlib.Add(p)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	if err := vlib.CommitProfile(filename, "initial version"); err != nil {
		t.Fatalf("CommitProfile: %v", err)
	}

	entries, err := vlib.History(filename)
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one history entry after commit")
	}
	if entries[0].Message != "initial version" {
		t.Errorf("expected message %q, got %q", "initial version", entries[0].Message)
	}
}

func TestVersionedLibrary_MultipleVersions(t *testing.T) {
	dir := t.TempDir()
	vlib := library.NewVersioned(dir)
	if err := vlib.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	p := minimalProfile("multi-version")
	filename, err := vlib.Add(p)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := vlib.CommitProfile(filename, "first"); err != nil {
		t.Fatalf("CommitProfile first: %v", err)
	}

	p2 := minimalProfile("multi-version")
	p2.System.Hostname = "updated-host"
	if _, err := vlib.Add(p2); err != nil {
		t.Fatalf("Add v2: %v", err)
	}
	if err := vlib.CommitProfile(filename, "second"); err != nil {
		t.Fatalf("CommitProfile second: %v", err)
	}

	entries, err := vlib.History(filename)
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if len(entries) < 2 {
		t.Fatalf("expected 2 history entries, got %d", len(entries))
	}
}

func TestVersionedLibrary_DiffAndRestore(t *testing.T) {
	dir := t.TempDir()
	vlib := library.NewVersioned(dir)
	if err := vlib.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	p := minimalProfile("diff-test")
	filename, err := vlib.Add(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := vlib.CommitProfile(filename, "v1"); err != nil {
		t.Fatal(err)
	}

	p2 := minimalProfile("diff-test")
	p2.System.Hostname = "changed-host"
	if _, err := vlib.Add(p2); err != nil {
		t.Fatal(err)
	}
	if err := vlib.CommitProfile(filename, "v2"); err != nil {
		t.Fatal(err)
	}

	entries, err := vlib.History(filename)
	if err != nil || len(entries) < 2 {
		t.Fatalf("need 2 history entries for diff test, got %v, %v", entries, err)
	}

	diff, err := vlib.Diff(filename, entries[1].Hash, entries[0].Hash)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if !strings.Contains(diff, "changed-host") {
		t.Errorf("expected diff to contain 'changed-host', got:\n%s", diff)
	}

	if err := vlib.Restore(filename, entries[1].Hash); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	restored, err := vlib.Get("diff-test")
	if err != nil {
		t.Fatalf("Get after restore: %v", err)
	}
	if restored.System.Hostname != p.System.Hostname {
		t.Errorf("restore did not revert hostname: got %q, want %q",
			restored.System.Hostname, p.System.Hostname)
	}
}

func TestVersionedLibrary_IdempotentCommit(t *testing.T) {
	dir := t.TempDir()
	vlib := library.NewVersioned(dir)
	if err := vlib.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	p := minimalProfile("idem-test")
	filename, err := vlib.Add(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := vlib.CommitProfile(filename, "first"); err != nil {
		t.Fatal(err)
	}

	// Committing unchanged file should not error.
	if err := vlib.CommitProfile(filename, "no change"); err != nil {
		t.Errorf("CommitProfile on unchanged file should not error: %v", err)
	}

	entries, _ := vlib.History(filename)
	if len(entries) != 1 {
		t.Errorf("expected 1 commit for unchanged file, got %d", len(entries))
	}
}

// Compile-time check: VersionedLibrary embeds Library.
var _ interface {
	Add(profile.Profile) (string, error)
	List() ([]library.Entry, error)
} = (*library.VersionedLibrary)(nil)
