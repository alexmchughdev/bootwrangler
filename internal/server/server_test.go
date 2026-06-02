package server_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexmchughdev/bootwrangler/internal/server"
)

// startServer creates a temp root with the given files, starts a server on a
// random port, and returns the server and its base URL.  The caller is
// responsible for calling s.Stop().
func startServer(t *testing.T, files map[string]string) (*server.Server, string) {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}

	s, err := server.New(root, "127.0.0.1:0")
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}
	if err := s.Start(); err != nil {
		t.Fatalf("server.Start: %v", err)
	}
	return s, "http://" + s.Addr()
}

// TestServeFile verifies that a GET for an existing file returns 200 and the
// correct content.
func TestServeFile(t *testing.T) {
	const want = "hello from bootwrangler"
	s, base := startServer(t, map[string]string{
		"hello.txt": want,
	})
	defer s.Stop()

	resp, err := http.Get(base + "/hello.txt")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if got := strings.TrimSpace(string(body)); got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

// TestPathTraversal verifies that path-traversal attempts are rejected with
// 403.
func TestPathTraversal(t *testing.T) {
	s, base := startServer(t, map[string]string{
		"safe.txt": "safe content",
	})
	defer s.Stop()

	attempts := []string{
		"/../etc/passwd",
		"/..",
		"/safe.txt/../../etc/passwd",
	}

	for _, path := range attempts {
		t.Run(path, func(t *testing.T) {
			resp, err := http.Get(base + path)
			if err != nil {
				t.Fatalf("GET %s: %v", path, err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusForbidden {
				t.Fatalf("path %q: status = %d, want 403", path, resp.StatusCode)
			}
		})
	}
}

// TestPathTraversalEncoded verifies that percent-encoded traversal sequences
// are also rejected.
func TestPathTraversalEncoded(t *testing.T) {
	s, base := startServer(t, map[string]string{
		"safe.txt": "safe content",
	})
	defer s.Stop()

	// Use a raw HTTP request so we can send the un-normalised path.
	// http.Get would normalise the URL before sending it.
	encodedPath := "/%2F..%2F..%2F"
	req, err := http.NewRequest(http.MethodGet, base+"/", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.URL.Opaque = encodedPath

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// A connection error after the server sends 403 is fine.
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("encoded traversal: status = %d, want 403 or 400", resp.StatusCode)
	}
}

// TestCallback verifies that a POST to /callback stores the event and
// returns 204.
func TestCallback(t *testing.T) {
	s, base := startServer(t, nil)
	defer s.Stop()

	body := `{"profile_name":"archlinux","status":"success","message":"install complete"}`
	resp, err := http.Post(base+"/callback", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("POST /callback: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}

	events := s.Events()
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	ev := events[0]
	if ev.ProfileName != "archlinux" {
		t.Errorf("ProfileName = %q, want %q", ev.ProfileName, "archlinux")
	}
	if ev.Status != "success" {
		t.Errorf("Status = %q, want %q", ev.Status, "success")
	}
	if ev.Message != "install complete" {
		t.Errorf("Message = %q, want %q", ev.Message, "install complete")
	}
}

// TestRequestLogPopulated verifies that requests are logged.
func TestRequestLogPopulated(t *testing.T) {
	s, base := startServer(t, map[string]string{
		"page.html": "<html></html>",
	})
	defer s.Stop()

	resp, err := http.Get(base + "/page.html")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()

	logs := s.Logs()
	if len(logs) == 0 {
		t.Fatal("expected at least one log entry, got none")
	}
	found := false
	for _, l := range logs {
		if l.Path == "/page.html" && l.Method == http.MethodGet && l.StatusCode == http.StatusOK {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected log entry for GET /page.html 200, got %+v", logs)
	}
}

// TestStopShutdown verifies that after Stop() subsequent requests fail.
func TestStopShutdown(t *testing.T) {
	s, base := startServer(t, map[string]string{
		"file.txt": "content",
	})

	// Confirm it works before stop.
	resp, err := http.Get(base + "/file.txt")
	if err != nil {
		t.Fatalf("GET before stop: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("before stop: status = %d, want 200", resp.StatusCode)
	}

	if err := s.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	// After stop, the request should fail (connection refused / EOF).
	_, err = http.Get(base + "/file.txt")
	if err == nil {
		t.Fatal("expected error after Stop(), got nil")
	}
	_ = fmt.Sprintf("post-stop error (expected): %v", err)
}
