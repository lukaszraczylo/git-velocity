package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileCache_Basic(t *testing.T) {
	// Create temp directory for cache
	tempDir := t.TempDir()

	cache, err := NewFileCache(tempDir, time.Hour)
	require.NoError(t, err)

	// Test Set and Get
	cache.Set("test-key", "test-value")

	value, ok := cache.Get("test-key")
	assert.True(t, ok)
	assert.Equal(t, "test-value", value)
}

func TestFileCache_GetNonExistent(t *testing.T) {
	tempDir := t.TempDir()

	cache, err := NewFileCache(tempDir, time.Hour)
	require.NoError(t, err)

	value, ok := cache.Get("non-existent")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestFileCache_Expiration(t *testing.T) {
	tempDir := t.TempDir()

	// Use a very short TTL
	cache, err := NewFileCache(tempDir, 50*time.Millisecond)
	require.NoError(t, err)

	cache.Set("expire-key", "expire-value")

	// Should be available immediately
	value, ok := cache.Get("expire-key")
	assert.True(t, ok)
	assert.Equal(t, "expire-value", value)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired now
	value, ok = cache.Get("expire-key")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestFileCache_Delete(t *testing.T) {
	tempDir := t.TempDir()

	cache, err := NewFileCache(tempDir, time.Hour)
	require.NoError(t, err)

	cache.Set("delete-key", "delete-value")

	// Verify it exists
	_, ok := cache.Get("delete-key")
	assert.True(t, ok)

	// Delete it
	cache.Delete("delete-key")

	// Should be gone
	value, ok := cache.Get("delete-key")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestFileCache_Clear(t *testing.T) {
	tempDir := t.TempDir()

	cache, err := NewFileCache(tempDir, time.Hour)
	require.NoError(t, err)

	// Add multiple entries
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Clear the cache
	err = cache.Clear()
	require.NoError(t, err)

	// All should be gone
	_, ok := cache.Get("key1")
	assert.False(t, ok)
	_, ok = cache.Get("key2")
	assert.False(t, ok)
	_, ok = cache.Get("key3")
	assert.False(t, ok)
}

func TestFileCache_ComplexValues(t *testing.T) {
	tempDir := t.TempDir()

	cache, err := NewFileCache(tempDir, time.Hour)
	require.NoError(t, err)

	// Test with map
	mapValue := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}
	cache.Set("map-key", mapValue)

	retrieved, ok := cache.Get("map-key")
	assert.True(t, ok)
	assert.Equal(t, mapValue, retrieved)

	// Test with slice
	sliceValue := []interface{}{"a", "b", "c"}
	cache.Set("slice-key", sliceValue)

	retrieved, ok = cache.Get("slice-key")
	assert.True(t, ok)
	assert.Equal(t, sliceValue, retrieved)
}

func TestFileCache_CreateDirectory(t *testing.T) {
	// Test that NewFileCache creates directory if it doesn't exist
	tempDir := filepath.Join(t.TempDir(), "nested", "cache", "dir")

	cache, err := NewFileCache(tempDir, time.Hour)
	require.NoError(t, err)

	// Verify directory was created
	info, err := os.Stat(tempDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Should be usable
	cache.Set("key", "value")
	value, ok := cache.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "value", value)
}

func TestMemoryCache_Basic(t *testing.T) {
	t.Parallel()

	cache := NewMemoryCache(time.Hour)

	// Test Set and Get
	cache.Set("test-key", "test-value")

	value, ok := cache.Get("test-key")
	assert.True(t, ok)
	assert.Equal(t, "test-value", value)
}

func TestMemoryCache_GetNonExistent(t *testing.T) {
	t.Parallel()

	cache := NewMemoryCache(time.Hour)

	value, ok := cache.Get("non-existent")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestMemoryCache_Expiration(t *testing.T) {
	t.Parallel()

	cache := NewMemoryCache(50 * time.Millisecond)

	cache.Set("expire-key", "expire-value")

	// Should be available immediately
	value, ok := cache.Get("expire-key")
	assert.True(t, ok)
	assert.Equal(t, "expire-value", value)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired now
	value, ok = cache.Get("expire-key")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestMemoryCache_Delete(t *testing.T) {
	t.Parallel()

	cache := NewMemoryCache(time.Hour)

	cache.Set("delete-key", "delete-value")

	// Verify it exists
	_, ok := cache.Get("delete-key")
	assert.True(t, ok)

	// Delete it
	cache.Delete("delete-key")

	// Should be gone
	value, ok := cache.Get("delete-key")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestMemoryCache_Clear(t *testing.T) {
	t.Parallel()

	cache := NewMemoryCache(time.Hour)

	// Add multiple entries
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Clear the cache
	err := cache.Clear()
	require.NoError(t, err)

	// All should be gone
	_, ok := cache.Get("key1")
	assert.False(t, ok)
	_, ok = cache.Get("key2")
	assert.False(t, ok)
	_, ok = cache.Get("key3")
	assert.False(t, ok)
}

func TestNoopCache_AlwaysReturnsFalse(t *testing.T) {
	t.Parallel()

	cache := NewNoopCache()

	// Set something
	cache.Set("key", "value")

	// Get should return false
	value, ok := cache.Get("key")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestNoopCache_DeleteAndClear(t *testing.T) {
	t.Parallel()

	cache := NewNoopCache()

	// These should not panic or error
	cache.Delete("key")
	err := cache.Clear()
	assert.NoError(t, err)
}

func TestFileCache_KeyToPath(t *testing.T) {
	t.Parallel()

	cache := &FileCache{directory: "/tmp/cache"}

	path1 := cache.keyToPath("key1")
	path2 := cache.keyToPath("key2")
	path1Again := cache.keyToPath("key1")

	// Different keys should produce different paths
	assert.NotEqual(t, path1, path2)

	// Same key should produce same path
	assert.Equal(t, path1, path1Again)

	// Path should end with .gob
	assert.Contains(t, path1, ".gob")
}

func TestCacheInterface(t *testing.T) {
	t.Parallel()

	// Ensure all cache types implement the interface
	var _ Cache = (*FileCache)(nil)
	var _ Cache = (*MemoryCache)(nil)
	var _ Cache = (*NoopCache)(nil)
}
