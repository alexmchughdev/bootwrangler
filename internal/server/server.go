// Package server provides an HTTP server that serves a build directory for
// local provisioning.  It supports optional install-result callbacks and
// records every request for inspection by callers.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// RequestLog is one HTTP request record.
type RequestLog struct {
	Method     string
	Path       string
	StatusCode int
	RemoteAddr string
}

// CallbackEvent is one install-result callback received from an installer.
type CallbackEvent struct {
	ProfileName string `json:"profile_name"`
	Status      string `json:"status"` // "started" | "success" | "failure"
	Message     string `json:"message"`
}

// Server serves a build directory over HTTP.
type Server struct {
	root    string
	addr    string
	httpSrv *http.Server

	mu       sync.Mutex
	logs     []RequestLog
	events   []CallbackEvent
	listener net.Listener
}

// New creates a server that serves root over HTTP.
// root must be an absolute path to an existing directory.
func New(root, addr string) (*Server, error) {
	if !filepath.IsAbs(root) {
		return nil, fmt.Errorf("server: root must be an absolute path, got %q", root)
	}
	cleaned := filepath.Clean(root)
	s := &Server{
		root: cleaned,
		addr: addr,
	}
	mux := http.NewServeMux()
	mux.Handle("/callback", http.HandlerFunc(s.handleCallback))
	mux.Handle("/", http.HandlerFunc(s.handleFile))
	s.httpSrv = &http.Server{
		Handler: s.loggingMiddleware(s.traversalMiddleware(mux)),
	}
	return s, nil
}

// Start begins serving in a goroutine.  It returns immediately once the
// listener is bound so that Addr() returns the actual address.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("server start: %w", err)
	}
	s.listener = ln
	go func() {
		_ = s.httpSrv.Serve(ln)
	}()
	return nil
}

// Stop shuts the server down gracefully.
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server stop: %w", err)
	}
	return nil
}

// Addr returns the address the server is listening on.  Useful after :0
// port selection.
func (s *Server) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}

// Logs returns all request logs collected since start.
func (s *Server) Logs() []RequestLog {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]RequestLog, len(s.logs))
	copy(out, s.logs)
	return out
}

// Events returns all callback events received.
func (s *Server) Events() []CallbackEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]CallbackEvent, len(s.events))
	copy(out, s.events)
	return out
}

// traversalMiddleware rejects any request whose raw URI contains ".." path
// segments.  It runs before http.ServeMux cleans the path so that traversal
// sequences are never silently normalised.
func (s *Server) traversalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.RequestURI
		// Strip query string.
		if i := strings.IndexByte(raw, '?'); i >= 0 {
			raw = raw[:i]
		}
		// Decode percent-encoded sequences so that %2F..%2F is caught too.
		if decoded, err := pathUnescape(raw); err == nil {
			raw = decoded
		}
		for _, seg := range strings.Split(raw, "/") {
			if seg == ".." {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware wraps a handler and records every request.
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rec, r)
		s.mu.Lock()
		s.logs = append(s.logs, RequestLog{
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: rec.code,
			RemoteAddr: r.RemoteAddr,
		})
		s.mu.Unlock()
	})
}

// handleFile serves static files from the root directory.
func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	urlPath := r.URL.Path

	// If a raw path is present (URL-encoded), decode it ourselves so we can
	// inspect it for traversal sequences before net/http normalises it.
	if r.URL.RawPath != "" {
		decoded, err := pathUnescape(r.URL.RawPath)
		if err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		urlPath = decoded
	}

	// Reject any path that contains ".." segments before cleaning.
	// We check the raw URL segments so that paths like "/../etc/passwd"
	// (which filepath.Clean would silently absorb) are still refused.
	for _, seg := range strings.Split(urlPath, "/") {
		if seg == ".." {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	// Build the absolute target and verify it is inside root.
	cleanURL := filepath.Clean(urlPath)
	target := filepath.Clean(filepath.Join(s.root, cleanURL))
	rootWithSep := s.root + string(filepath.Separator)
	if target != s.root && !strings.HasPrefix(target, rootWithSep) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Serve the file directly (bypassing http.ServeFile's redirect behaviour
	// for index.html and URL-path cleaning).
	f, err := os.Open(target)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
}

// handleCallback accepts POST /callback with a JSON CallbackEvent body.
func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var ev CallbackEvent
	if err := json.Unmarshal(body, &ev); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.events = append(s.events, ev)
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// pathUnescape decodes a percent-encoded URL path segment by segment.
func pathUnescape(s string) (string, error) {
	var out strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '%' {
			if i+2 >= len(s) {
				return "", errors.New("invalid percent encoding")
			}
			hi := unhex(s[i+1])
			lo := unhex(s[i+2])
			if hi < 0 || lo < 0 {
				return "", errors.New("invalid percent encoding")
			}
			out.WriteByte(byte(hi<<4 | lo))
			i += 3
		} else {
			out.WriteByte(s[i])
			i++
		}
	}
	return out.String(), nil
}

func unhex(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

// statusRecorder wraps ResponseWriter to capture the written status code.
type statusRecorder struct {
	http.ResponseWriter
	code int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.code = code
	r.ResponseWriter.WriteHeader(code)
}
