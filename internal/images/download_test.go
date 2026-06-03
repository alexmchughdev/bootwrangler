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

func TestExtractSHA256(t *testing.T) {
	const (
		ubuntuHash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
		fedoraHash = "1111111111111111111111111111111111111111111111111111111111111111"
		rockyHash  = "2222222222222222222222222222222222222222222222222222222222222222"
		singleHash = "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd"
	)

	tests := []struct {
		name         string
		checksumText string
		imageURL     string
		want         string
	}{
		{
			name:         "ubuntu debian hash filename",
			checksumText: ubuntuHash + " *ubuntu-24.04.2-live-server-amd64.iso\n",
			imageURL:     "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso",
			want:         ubuntuHash,
		},
		{
			name: "fedora filename and hash",
			checksumText: "-----BEGIN PGP SIGNED MESSAGE-----\n" +
				"SHA256 (Fedora-Server-dvd-x86_64-42-1.1.iso) = " + fedoraHash + "\n",
			imageURL: "https://download.fedoraproject.org/Fedora-Server-dvd-x86_64-42-1.1.iso",
			want:     fedoraHash,
		},
		{
			name: "rocky filename and hash",
			checksumText: "Rocky-9.5-x86_64-dvd.iso: " + rockyHash + "\n" +
				"Rocky-9.5-x86_64-minimal.iso: " + ubuntuHash + "\n",
			imageURL: "https://download.rockylinux.org/pub/rocky/9/isos/x86_64/Rocky-9.5-x86_64-dvd.iso",
			want:     rockyHash,
		},
		{
			name:         "single hash sha256 file",
			checksumText: singleHash + "\n",
			imageURL:     "https://example.com/custom.iso",
			want:         singleHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := images.ExtractSHA256(tt.checksumText, tt.imageURL)
			if err != nil {
				t.Fatalf("ExtractSHA256: %v", err)
			}
			if got != tt.want {
				t.Fatalf("ExtractSHA256 = %q, want %q", got, tt.want)
			}
		})
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

func TestDownloadAndVerifyCreatesMarkerAndVerifiedCacheStatus(t *testing.T) {
	content := []byte("verified iso content")
	expected := sha256Hex(content)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test.iso":
			_, _ = w.Write(content)
		case "/SHA256SUMS":
			_, _ = w.Write([]byte(expected + " *test.iso\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cacheDir := t.TempDir()
	img := images.ArchImage{
		Type:        images.ImageTypeISO,
		URL:         srv.URL + "/test.iso",
		ChecksumURL: srv.URL + "/SHA256SUMS",
	}
	destPath := images.CachedPath(cacheDir, "test-id", "1.0", "x86_64", img)

	var lastWritten int64
	err := images.DownloadAndVerify(img, destPath, func(written, total int64) {
		lastWritten = written
	})
	if err != nil {
		t.Fatalf("DownloadAndVerify: %v", err)
	}

	got, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Fatalf("downloaded content = %q, want %q", got, content)
	}
	if lastWritten != int64(len(content)) {
		t.Fatalf("last written progress = %d, want %d", lastWritten, len(content))
	}

	marker, err := os.ReadFile(images.VerifiedMarkerPath(destPath))
	if err != nil {
		t.Fatalf("verified marker: %v", err)
	}
	if string(marker) != expected+"\n" {
		t.Fatalf("verified marker = %q, want %q", marker, expected+"\n")
	}

	status := images.CheckCache(cacheDir, "test-id", "1.0", "x86_64", img)
	if !status.Cached {
		t.Fatal("expected cached after verified download")
	}
	if !status.Verified {
		t.Fatal("expected verified cache status after marker created")
	}
}

func TestDownloadAndVerifyChecksumMismatchRemovesImageAndMarker(t *testing.T) {
	content := []byte("corrupt iso content")
	badChecksum := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test.iso":
			_, _ = w.Write(content)
		case "/SHA256SUMS":
			_, _ = w.Write([]byte(badChecksum + " *test.iso\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	destPath := filepath.Join(dir, "test.iso")
	img := images.ArchImage{
		Type:        images.ImageTypeISO,
		URL:         srv.URL + "/test.iso",
		ChecksumURL: srv.URL + "/SHA256SUMS",
	}
	if err := images.MarkVerified(destPath, sha256Hex([]byte("previous image"))); err != nil {
		t.Fatalf("create stale marker: %v", err)
	}

	if err := images.DownloadAndVerify(img, destPath, nil); err == nil {
		t.Fatal("expected DownloadAndVerify to fail on checksum mismatch")
	}
	if _, err := os.Stat(destPath); !os.IsNotExist(err) {
		t.Fatalf("expected downloaded image to be removed, stat err = %v", err)
	}
	if _, err := os.Stat(images.VerifiedMarkerPath(destPath)); !os.IsNotExist(err) {
		t.Fatalf("expected verified marker to be absent, stat err = %v", err)
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
	if status2.Verified {
		t.Error("expected cached file without marker to be unverified")
	}
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
