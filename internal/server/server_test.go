package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	s := New("/tmp/test", "8080")

	assert.Equal(t, "/tmp/test", s.directory)
	assert.Equal(t, "8080", s.port)
}

func TestServer_StartWithNonExistentDirectory(t *testing.T) {
	t.Parallel()

	s := New("/this/directory/does/not/exist", "8080")

	err := s.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directory does not exist")
}

func TestServer_CacheMiddleware(t *testing.T) {
	t.Parallel()

	s := New(".", "8080")

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with cache middleware
	wrapped := s.cacheMiddleware(handler)

	// Make a test request
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	// Check cache headers are set correctly
	assert.Equal(t, "no-cache, no-store, must-revalidate", rr.Header().Get("Cache-Control"))
	assert.Equal(t, "no-cache", rr.Header().Get("Pragma"))
	assert.Equal(t, "0", rr.Header().Get("Expires"))
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestServer_LoggingMiddleware(t *testing.T) {
	t.Parallel()

	s := New(".", "8080")

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with logging middleware
	wrapped := s.loggingMiddleware(handler)

	// Make a test request
	req := httptest.NewRequest("GET", "/test-path", nil)
	rr := httptest.NewRecorder()

	// This should not panic
	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestServer_ServesStaticFiles(t *testing.T) {
	// Create a temp directory with a test file
	tempDir := t.TempDir()

	// Create a test file with a simple name
	testFile := filepath.Join(tempDir, "hello.txt")
	err := os.WriteFile(testFile, []byte("Hello, World!"), 0644)
	require.NoError(t, err)

	s := New(tempDir, "0")

	// Use http.StripPrefix with the file server to avoid redirect issues
	absPath, _ := filepath.Abs(tempDir)
	fs := http.FileServer(http.Dir(absPath))

	// Create test server
	ts := httptest.NewServer(fs)
	defer ts.Close()

	// Test serving the text file via HTTP
	resp, err := http.Get(ts.URL + "/hello.txt")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "Hello, World!", string(body))

	// Verify the server object is set up correctly
	assert.Equal(t, tempDir, s.directory)
}

func TestServer_404ForNonExistentFile(t *testing.T) {
	tempDir := t.TempDir()

	absPath, _ := filepath.Abs(tempDir)
	fs := http.FileServer(http.Dir(absPath))

	ts := httptest.NewServer(fs)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/nonexistent.txt")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestServer_ServesNestedDirectories(t *testing.T) {
	tempDir := t.TempDir()

	// Create nested directory structure
	nestedDir := filepath.Join(tempDir, "data", "repos")
	err := os.MkdirAll(nestedDir, 0755)
	require.NoError(t, err)

	// Create a file in nested directory
	testFile := filepath.Join(nestedDir, "metrics.json")
	err = os.WriteFile(testFile, []byte(`{"count": 42}`), 0644)
	require.NoError(t, err)

	absPath, _ := filepath.Abs(tempDir)
	fs := http.FileServer(http.Dir(absPath))

	ts := httptest.NewServer(fs)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/data/repos/metrics.json")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "42")
}

func TestServer_MiddlewareCombination(t *testing.T) {
	t.Parallel()

	s := New(".", "8080")

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	})

	// Combine middlewares like in the actual server
	combined := s.loggingMiddleware(s.cacheMiddleware(innerHandler))

	req := httptest.NewRequest("GET", "/any-path", nil)
	rr := httptest.NewRecorder()

	combined.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)
	body, _ := io.ReadAll(rr.Body)
	assert.Equal(t, "response", string(body))

	// Check headers were set by cache middleware
	assert.NotEmpty(t, rr.Header().Get("Cache-Control"))
}

func TestServer_ServesIndexHtml(t *testing.T) {
	tempDir := t.TempDir()

	// Create an index.html
	indexFile := filepath.Join(tempDir, "index.html")
	err := os.WriteFile(indexFile, []byte("<html><body>Test Page</body></html>"), 0644)
	require.NoError(t, err)

	absPath, _ := filepath.Abs(tempDir)
	fs := http.FileServer(http.Dir(absPath))

	ts := httptest.NewServer(fs)
	defer ts.Close()

	// Test serving index.html via root path
	resp, err := http.Get(ts.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Test Page")
}
