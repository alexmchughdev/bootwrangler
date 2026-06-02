package images_test

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexmchughdev/bootwrangler/internal/images"
)

func TestVerifySHA256(t *testing.T) {
	dir := t.TempDir()
	content := []byte("test image data")
	path := filepath.Join(dir, "test.iso")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	h := sha256.Sum256(content)
	expected := hex.EncodeToString(h[:])

	if err := images.VerifySHA256(path, expected); err != nil {
		t.Errorf("expected verify to pass: %v", err)
	}
	if err := images.VerifySHA256(path, "deadbeef"); err == nil {
		t.Error("expected verify to fail on bad checksum")
	}
}

func TestDownload(t *testing.T) {
	content := []byte("fake iso content for testing")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(content)
	}))
	defer srv.Close()

	dir := t.TempDir()
	destPath := filepath.Join(dir, "downloaded.iso")

	var lastWritten int64
	err := images.Download(srv.URL+"/test.iso", destPath, func(written, total int64) {
		lastWritten = written
	})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}

	got, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
	if lastWritten == 0 {
		t.Error("progress callback was never called")
	}
}

func TestCheckCache(t *testing.T) {
	dir := t.TempDir()

	img := images.ArchImage{Type: "iso"}
	status := images.CheckCache(dir, "test-id", "1.0", "x86_64", img)
	if status.Cached {
		t.Error("expected not cached initially")
	}

	// Create the file.
	if err := os.WriteFile(status.Path, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	status2 := images.CheckCache(dir, "test-id", "1.0", "x86_64", img)
	if !status2.Cached {
		t.Error("expected cached after file created")
	}
}
