package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Server is a simple HTTP server for previewing the generated site
type Server struct {
	directory string
	port      string
}

// New creates a new preview server
func New(directory, port string) *Server {
	return &Server{
		directory: directory,
		port:      port,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	handler, err := s.CreateHandler()
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              s.GetAddress(),
		Handler:           handler,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return srv.ListenAndServe()
}

// loggingMiddleware logs incoming requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// cacheMiddleware adds cache headers for static assets
func (s *Server) cacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Disable caching for development
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// Add CORS headers for local development
		w.Header().Set("Access-Control-Allow-Origin", "*")

		next.ServeHTTP(w, r)
	})
}

// CreateHandler creates and returns the HTTP handler without starting the server.
// This is useful for testing and for embedding the server in other applications.
func (s *Server) CreateHandler() (http.Handler, error) {
	// Check if directory exists
	if _, err := os.Stat(s.directory); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", s.directory)
	}

	// Get absolute path
	absPath, err := filepath.Abs(s.directory)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Create file server with directory listing disabled for security
	fs := http.FileServer(http.Dir(absPath))

	// Wrap with middleware
	return s.loggingMiddleware(s.cacheMiddleware(fs)), nil
}

// GetAddress returns the server address in the format :port
func (s *Server) GetAddress() string {
	return fmt.Sprintf(":%s", s.port)
}

// GetDirectory returns the directory being served
func (s *Server) GetDirectory() string {
	return s.directory
}
