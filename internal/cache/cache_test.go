package cache

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// To avoid circular dependency
type MockMetadataMap map[string]MockMetadata

type MockMetadata struct {
	VersionMappings []MockVersionMapping `json:"versions"`
}

type MockVersionMapping struct {
	CVersion   string   `json:"c"`
	GoVersions []string `json:"go"`
}

// Test data
var testMetadata = MockMetadataMap{
	"example-module": MockMetadata{
		VersionMappings: []MockVersionMapping{
			{
				CVersion:   "v1.0",
				GoVersions: []string{"go1.20"},
			},
		},
	},
}

func TestCache_InitFromRemote(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Last-Modified", "Sat, 01 Jan 2022 00:00:00 GMT")
		json.NewEncoder(w).Encode(testMetadata)
	}))
	defer server.Close()

	// Create temporary directory
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "test_cache.json")

	// Initialize cache
	cache, err := NewCache[MockMetadataMap](cachePath, server.URL)
	require.NoError(t, err)

	// Validate data
	// Because MockMetadataMap is a map type, it needs to be converted to a comparable struct for assertion
	assert.Equal(t, testMetadata, cache.Data())
}

// Test loading cache from disk (skip remote request)
func TestCache_LoadFromDisk(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "existing_cache.json")

	// Manually write cache file
	err := os.WriteFile(cachePath, []byte(`{"example-module": {"versions": [{"c":"v1.0","go":["go1.20"]}]}}`), 0644)
	require.NoError(t, err)

	// Initialize cache (should load from local file directly)
	cache, err := NewCache[MockMetadataMap](cachePath, "https://unused.com")
	require.NoError(t, err)

	// Validate data
	assert.Equal(t, testMetadata, cache.Data())
}

// Test HTTP 304 Not Modified response
func TestCache_Fetch304(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-Modified-Since") == "old-time" {
			w.WriteHeader(http.StatusNotModified)
		} else {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(testMetadata)
		}
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.json")

	// Initial load
	cache, _ := NewCache[MockMetadataMap](cachePath, server.URL)

	// Set modTime to an old time
	cache.modTime = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	// Second fetch should return 304
	_, err := cache.Update()
	require.NoError(t, err)

	// Validate data remains unchanged
	assert.Equal(t, testMetadata, cache.Data())
}

// Test invalid JSON response
func TestCache_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	_, err := NewCache[MockMetadataMap](filepath.Join(tmpDir, "cache.json"), server.URL)
	require.Error(t, err) // Should return deserialization error
}

// Test concurrent updates and reads
func TestCache_ConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 500)
		json.NewEncoder(w).Encode(testMetadata)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.json")

	// Initialize cache
	cache, err := NewCache[MockMetadataMap](cachePath, server.URL)
	require.NoError(t, err)

	// Concurrently update and read cache, reader if i%4 != 0, writer if i%4 == 0
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%4 != 0 {
				_, _ = cache.Update()
			} else {
				_ = cache.Data()
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Validate data
	assert.Equal(t, testMetadata, cache.Data())

	// Validate modTime
	assert.True(t, time.Now().After(cache.modTime), "modTime should be updated after concurrent access")

	return
}
