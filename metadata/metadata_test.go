package metadata

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testMetadata = MetadataMap{
	"example-module": Metadata{
		VersionMappings: []*VersionMapping{
			{
				CVersion:   "v1.0",
				GoVersions: []string{"go1.20"},
			},
		},
	},
}

// TestNewMetadataMgr verifies that the MetadataMgr is successfully created with valid remote data.
func TestNewMetadataMgr(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/llpkgstore.json" {
			w.Header().Set("Last-Modified", "Sat, 01 Jan 2022 00:00:00 GMT")
			json.NewEncoder(w).Encode(testMetadata)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Save original URL and override with mock server URL
	originalURL := remoteMetadataURL
	defer func() { remoteMetadataURL = originalURL }()
	remoteMetadataURL = server.URL + "/llpkgstore.json"

	tmpDir := t.TempDir()
	mgr, err := NewMetadataMgr(tmpDir)
	require.NoError(t, err, "Failed to create MetadataMgr instance")

	// Verify successful initialization
	assert.NotNil(t, mgr, "Metadata manager should not be nil")
	assert.NotNil(t, mgr.cache, "Cache instance should be initialized")
}

// TestMetadataMgr_AllMetadata verifies that AllMetadata() returns the complete metadata after update.
func TestMetadataMgr_AllMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/llpkgstore.json" {
			json.NewEncoder(w).Encode(testMetadata)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Override remote URL to mock server
	originalURL := remoteMetadataURL
	defer func() { remoteMetadataURL = originalURL }()
	remoteMetadataURL = server.URL + "/llpkgstore.json"

	tmpDir := t.TempDir()
	mgr, _ := NewMetadataMgr(tmpDir)

	// Force metadata update from mock server
	err := mgr.update()
	require.NoError(t, err, "Metadata update should succeed")

	// Retrieve and validate all metadata
	data, err := mgr.AllMetadata()
	require.NoError(t, err, "Failed to retrieve metadata")
	assert.Equal(t, testMetadata, data, "Returned metadata should match test data")
}

// TestNewMetadataMgr_LocalAndRemoteFailure verifies MetadataMgr creation fails when both local cache and remote are unavailable
func TestNewMetadataMgr_LocalAndRemoteFailure(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, cachedMetadataFileName)
	err := os.WriteFile(cachePath, []byte(`{ invalid json `), 0644)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	originalURL := remoteMetadataURL
	defer func() { remoteMetadataURL = originalURL }()
	remoteMetadataURL = server.URL + "/llpkgstore.json"

	_, err = NewMetadataMgr(tmpDir)
	require.Error(t, err)
}

// TestMetadataMgr_MetadataByName_Existing validates MetadataByName returns correct metadata for existing modules
func TestMetadataMgr_MetadataByName_Existing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/llpkgstore.json" {
			json.NewEncoder(w).Encode(testMetadata)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	originalURL := remoteMetadataURL
	defer func() { remoteMetadataURL = originalURL }()
	remoteMetadataURL = server.URL + "/llpkgstore.json"

	tmpDir := t.TempDir()
	mgr, _ := NewMetadataMgr(tmpDir)

	metadata, err := mgr.MetadataByName("example-module")
	require.NoError(t, err)
	assert.Equal(t, testMetadata["example-module"], metadata)
}

// TestMetadataMgr_MetadataByName_NotFound ensures MetadataByName returns error for non-existent modules
func TestMetadataMgr_MetadataByName_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/llpkgstore.json" {
			json.NewEncoder(w).Encode(MetadataMap{})
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	originalURL := remoteMetadataURL
	defer func() { remoteMetadataURL = originalURL }()
	remoteMetadataURL = server.URL + "/llpkgstore.json"

	tmpDir := t.TempDir()
	mgr, _ := NewMetadataMgr(tmpDir)

	_, err := mgr.MetadataByName("nonexistent-module")
	assert.ErrorIs(t, err, ErrMetadataNotInCache)
}

// TestMetadataMgr_ModuleExists checks ModuleExists correctly identifies existing/non-existent modules
func TestMetadataMgr_ModuleExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/llpkgstore.json" {
			json.NewEncoder(w).Encode(testMetadata)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	originalURL := remoteMetadataURL
	defer func() { remoteMetadataURL = originalURL }()
	remoteMetadataURL = server.URL + "/llpkgstore.json"

	tmpDir := t.TempDir()
	mgr, _ := NewMetadataMgr(tmpDir)

	exists, err := mgr.ModuleExists("example-module")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = mgr.ModuleExists("nonexistent-module")
	require.NoError(t, err)
	assert.False(t, exists)
}

// TestMetadataMgr_UpdateError verifies Update() returns error for failed remote requests
func TestMetadataMgr_UpdateError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/llpkgstore.json" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	originalURL := remoteMetadataURL
	defer func() { remoteMetadataURL = originalURL }()
	remoteMetadataURL = server.URL + "/llpkgstore.json"

	testMetadataJSON, err := json.Marshal(testMetadata)
	require.NoError(t, err)

	tmpDir := t.TempDir()
	// ensure metadataMgr can be created
	os.WriteFile(filepath.Join(tmpDir, "llpkgstore.json"), testMetadataJSON, 0644)
	mgr, err := NewMetadataMgr(tmpDir)
	require.NoError(t, err)

	err = mgr.update()
	assert.Error(t, err)
}

// Test invalid remote data scenario
// TestNewMetadataMgr_InvalidRemoteData verifies creation fails with invalid remote JSON
func TestNewMetadataMgr_InvalidRemoteData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/llpkgstore.json" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`invalid json`))
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	originalURL := remoteMetadataURL
	defer func() { remoteMetadataURL = originalURL }()
	remoteMetadataURL = server.URL + "/llpkgstore.json"

	tmpDir := t.TempDir()
	_, err := NewMetadataMgr(tmpDir)
	require.Error(t, err)
}

// Test recovery from corrupted cache file
// TestNewMetadataMgr_BadCacheFile checks recovery from corrupted cache by fetching remote data
func TestNewMetadataMgr_BadCacheFile(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, cachedMetadataFileName)
	err := os.WriteFile(cachePath, []byte(`{ invalid json `), 0644)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/llpkgstore.json" {
			json.NewEncoder(w).Encode(testMetadata)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	originalURL := remoteMetadataURL
	defer func() { remoteMetadataURL = originalURL }()
	remoteMetadataURL = server.URL + "/llpkgstore.json"

	mgr, err := NewMetadataMgr(tmpDir)
	require.NoError(t, err)

	data, err := mgr.AllMetadata()
	require.NoError(t, err)
	assert.Equal(t, testMetadata, data)
}
