package images

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var sha256Pattern = regexp.MustCompile(`(?i)\b[a-f0-9]{64}\b`)

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

// DownloadAndVerify downloads an image and verifies it against catalogue
// checksum metadata when a checksum URL is available.
func DownloadAndVerify(img ArchImage, destPath string, progressFn func(written, total int64)) error {
	clearCachedImage := func() {
		_ = os.Remove(destPath)
		_ = os.Remove(VerifiedMarkerPath(destPath))
	}

	_ = os.Remove(VerifiedMarkerPath(destPath))
	if err := Download(img.URL, destPath, progressFn); err != nil {
		return err
	}

	if img.ChecksumURL == "" {
		return nil
	}

	checksumText, err := FetchText(img.ChecksumURL)
	if err != nil {
		clearCachedImage()
		return fmt.Errorf("download: checksum: %w", err)
	}
	expected, err := ExtractSHA256(checksumText, img.URL)
	if err != nil {
		clearCachedImage()
		return fmt.Errorf("download: checksum: %w", err)
	}
	if err := VerifySHA256(destPath, expected); err != nil {
		clearCachedImage()
		return err
	}
	if err := MarkVerified(destPath, expected); err != nil {
		clearCachedImage()
		return err
	}
	return nil
}

// FetchText downloads a small text resource such as a checksum file.
func FetchText(rawURL string) (string, error) {
	resp, err := http.Get(rawURL) //nolint:gosec // URL comes from validated catalogue
	if err != nil {
		return "", fmt.Errorf("GET %s: %w", rawURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s returned HTTP %d", rawURL, resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if err != nil {
		return "", fmt.Errorf("read %s: %w", rawURL, err)
	}
	return string(data), nil
}

// ExtractSHA256 finds the expected hash for imageURL in a distro checksum file.
func ExtractSHA256(checksumText, imageURL string) (string, error) {
	filename := imageFilename(imageURL)
	var hashes []string
	for _, line := range strings.Split(checksumText, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		matches := sha256Pattern.FindAllString(line, -1)
		for _, match := range matches {
			hash := strings.ToLower(match)
			hashes = append(hashes, hash)
			if filename == "" || strings.Contains(line, filename) {
				return hash, nil
			}
		}
	}
	if len(hashes) == 1 {
		return hashes[0], nil
	}
	if filename == "" {
		return "", fmt.Errorf("no SHA256 checksum found")
	}
	return "", fmt.Errorf("no SHA256 checksum found for %s", filename)
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

// MarkVerified records the checksum that successfully verified the cached image.
func MarkVerified(imagePath, sha256Hex string) error {
	return os.WriteFile(VerifiedMarkerPath(imagePath), []byte(strings.ToLower(sha256Hex)+"\n"), 0o644)
}

// VerifiedMarkerPath returns the sidecar file used to mark a verified cache entry.
func VerifiedMarkerPath(imagePath string) string {
	return imagePath + ".sha256"
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
	verified := false
	if data, err := os.ReadFile(VerifiedMarkerPath(path)); err == nil {
		verified = sha256Pattern.MatchString(string(data))
	}
	return CacheStatus{Cached: true, Path: path, Verified: verified}
}

func imageFilename(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Path == "" {
		return filepath.Base(rawURL)
	}
	return path.Base(parsed.Path)
}
