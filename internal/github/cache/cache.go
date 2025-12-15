package cache

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cache defines the interface for caching
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)
	Clear() error
}

// FileCache implements file-based caching
type FileCache struct {
	directory string
	ttl       time.Duration
	mu        sync.RWMutex
}

// cacheEntry wraps a cached value with expiration
type cacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// NewFileCache creates a new file-based cache
func NewFileCache(directory string, ttl time.Duration) (*FileCache, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(directory, 0750); err != nil {
		return nil, err
	}

	return &FileCache{
		directory: directory,
		ttl:       ttl,
	}, nil
}

// Get retrieves a value from the cache
func (c *FileCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	path := c.keyToPath(key)

	file, err := os.Open(path) // #nosec G304 -- path is internally generated hash
	if err != nil {
		return nil, false
	}
	defer file.Close()

	var entry cacheEntry
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&entry); err != nil {
		return nil, false
	}

	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		_ = os.Remove(path)
		return nil, false
	}

	return entry.Value, true
}

// Set stores a value in the cache
func (c *FileCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := cacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}

	path := c.keyToPath(key)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return
	}

	file, err := os.Create(path) // #nosec G304 -- path is internally generated hash
	if err != nil {
		return
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	_ = encoder.Encode(entry)
}

// Delete removes a value from the cache
func (c *FileCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := c.keyToPath(key)
	_ = os.Remove(path)
}

// Clear removes all cached values
func (c *FileCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return os.RemoveAll(c.directory)
}

// keyToPath converts a cache key to a file path
func (c *FileCache) keyToPath(key string) string {
	hash := sha256.Sum256([]byte(key))
	filename := hex.EncodeToString(hash[:8]) + ".gob"
	return filepath.Join(c.directory, filename)
}

// NoopCache is a cache that doesn't cache anything
type NoopCache struct{}

// NewNoopCache creates a new no-op cache
func NewNoopCache() *NoopCache {
	return &NoopCache{}
}

// Get always returns false
func (c *NoopCache) Get(key string) (interface{}, bool) {
	return nil, false
}

// Set does nothing
func (c *NoopCache) Set(key string, value interface{}) {}

// Delete does nothing
func (c *NoopCache) Delete(key string) {}

// Clear does nothing
func (c *NoopCache) Clear() error {
	return nil
}

// Register types for gob encoding
func init() {
	// Register common types that might be cached
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})
}
