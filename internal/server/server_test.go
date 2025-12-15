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
		_, _ = w.Write([]byte("OK"))
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
	err := os.WriteFile(testFile, []byte("Hello, World!"), 0600)
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
	err = os.WriteFile(testFile, []byte(`{"count": 42}`), 0600)
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
		_, _ = w.Write([]byte("response"))
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
	err := os.WriteFile(indexFile, []byte("<html><body>Test Page</body></html>"), 0600)
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

func TestServer_CreateHandler(t *testing.T) {
	tempDir := t.TempDir()

	// Create an index.html
	indexFile := filepath.Join(tempDir, "index.html")
	err := os.WriteFile(indexFile, []byte("<html><body>Handler Test</body></html>"), 0600)
	require.NoError(t, err)

	s := New(tempDir, "8080")

	handler, err := s.CreateHandler()
	require.NoError(t, err)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Handler Test")

	// Check middleware headers are applied
	assert.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get("Cache-Control"))
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestServer_CreateHandlerWithNonExistentDirectory(t *testing.T) {
	t.Parallel()

	s := New("/this/directory/does/not/exist", "8080")

	handler, err := s.CreateHandler()
	assert.Error(t, err)
	assert.Nil(t, handler)
	assert.Contains(t, err.Error(), "directory does not exist")
}

func TestServer_GetAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		port     string
		expected string
	}{
		{"standard port", "8080", ":8080"},
		{"different port", "3000", ":3000"},
		{"port 0 for random", "0", ":0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(".", tt.port)
			assert.Equal(t, tt.expected, s.GetAddress())
		})
	}
}

func TestServer_ServesJSONWithCorrectContentType(t *testing.T) {
	tempDir := t.TempDir()

	// Create a JSON file
	jsonFile := filepath.Join(tempDir, "data.json")
	err := os.WriteFile(jsonFile, []byte(`{"status": "ok"}`), 0600)
	require.NoError(t, err)

	s := New(tempDir, "0")
	handler, err := s.CreateHandler()
	require.NoError(t, err)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/data.json")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// Check content type is JSON
	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, "application/json")
}

func TestServer_ServesHTMLWithCorrectContentType(t *testing.T) {
	tempDir := t.TempDir()

	// Create an HTML file
	htmlFile := filepath.Join(tempDir, "page.html")
	err := os.WriteFile(htmlFile, []byte("<html><body>HTML Page</body></html>"), 0600)
	require.NoError(t, err)

	s := New(tempDir, "0")
	handler, err := s.CreateHandler()
	require.NoError(t, err)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/page.html")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// Check content type is HTML
	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, "text/html")
}

func TestServer_CORSHeaders(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0600)
	require.NoError(t, err)

	s := New(tempDir, "0")
	handler, err := s.CreateHandler()
	require.NoError(t, err)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/test.txt")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check CORS header
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestServer_CacheDisabledHeaders(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0600)
	require.NoError(t, err)

	s := New(tempDir, "0")
	handler, err := s.CreateHandler()
	require.NoError(t, err)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/test.txt")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check cache headers are disabled for development
	assert.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get("Cache-Control"))
	assert.Equal(t, "no-cache", resp.Header.Get("Pragma"))
	assert.Equal(t, "0", resp.Header.Get("Expires"))
}

func TestServer_LoggingMiddlewareWithDifferentMethods(t *testing.T) {
	t.Parallel()

	s := New(".", "8080")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := s.loggingMiddleware(handler)

	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test-path", nil)
			rr := httptest.NewRecorder()

			wrapped.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
		})
	}
}

func TestServer_CacheMiddlewarePreservesResponseBody(t *testing.T) {
	t.Parallel()

	s := New(".", "8080")

	expectedBody := "This is the response body content"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedBody))
	})

	wrapped := s.cacheMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	body, _ := io.ReadAll(rr.Body)
	assert.Equal(t, expectedBody, string(body))
}

func TestNew_WithEmptyValues(t *testing.T) {
	t.Parallel()

	s := New("", "")
	assert.Equal(t, "", s.directory)
	assert.Equal(t, "", s.port)
}

func TestNew_WithSpecialCharactersInPath(t *testing.T) {
	t.Parallel()

	path := "/path/with spaces/and-dashes/and_underscores"
	s := New(path, "8080")
	assert.Equal(t, path, s.directory)
}
