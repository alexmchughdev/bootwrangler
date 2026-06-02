package images

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// CacheDir returns the default image cache directory.
func CacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".bootwrangler", "cache", "images")
	}
	return filepath.Join(home, ".bootwrangler", "cache", "images")
}

// CachedPath returns the local cache path for a catalogue entry image.
func CachedPath(cacheDir, id, version, arch string, img ArchImage) string {
	ext := "." + string(img.Type)
	filename := strings.Join([]string{id, version, arch}, "-") + ext
	return filepath.Join(cacheDir, filename)
}

// Download downloads one image URL to destPath, showing progress via progressFn.
// progressFn receives (bytesWritten, totalBytes) — totalBytes may be -1 if unknown.
func Download(url, destPath string, progressFn func(written, total int64)) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("download: mkdir: %w", err)
	}

	resp, err := http.Get(url) //nolint:gosec // URL comes from validated catalogue
	if err != nil {
		return fmt.Errorf("download: GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: %s returned HTTP %d", url, resp.StatusCode)
	}

	f, err := os.CreateTemp(filepath.Dir(destPath), "bootwrangler-download-*")
	if err != nil {
		return fmt.Errorf("download: temp file: %w", err)
	}
	tmpPath := f.Name()
	defer os.Remove(tmpPath)

	var written int64
	buf := make([]byte, 32*1024)
	total := resp.ContentLength
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				f.Close()
				return fmt.Errorf("download: write: %w", werr)
			}
			written += int64(n)
			if progressFn != nil {
				progressFn(written, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			f.Close()
			return fmt.Errorf("download: read: %w", err)
		}
	}
	f.Close()

	return os.Rename(tmpPath, destPath)
}

// VerifySHA256 checks that the file at path matches the expected hex-encoded SHA256.
func VerifySHA256(path, expectedHex string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("verify: open %s: %w", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("verify: hash %s: %w", path, err)
	}

	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, expectedHex) {
		return fmt.Errorf("verify: checksum mismatch for %s: expected %s, got %s",
			filepath.Base(path), expectedHex, got)
	}
	return nil
}

// CacheStatus describes whether an image is locally cached.
type CacheStatus struct {
	Cached   bool
	Path     string
	Verified bool
}

// CheckCache returns the cache status for one image.
func CheckCache(cacheDir, id, version, arch string, img ArchImage) CacheStatus {
	path := CachedPath(cacheDir, id, version, arch, img)
	if _, err := os.Stat(path); err != nil {
		return CacheStatus{Path: path}
	}
	return CacheStatus{Cached: true, Path: path}
}
