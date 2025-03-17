package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	ErrCacheFileNotFound = errors.New("cache file not found")
)

// Cache represents a local cache for storing and retrieving data.
// It binds a local file path and a remote data source URL
type Cache[T any] struct {
	data T            // stores cached data
	mu   sync.RWMutex // mutex for concurrent access

	cacheFilePath string // local file path for cache storage
	remoteUrl     string // URL of the remote data source

	modTime    time.Time    // last modified time of the cached data
	httpClient *http.Client // HTTP client with timeout
}

// NewCache initializes and loads the cache from disk or remote source
func NewCache[T any](cacheFilePath, remoteUrl string) (*Cache[T], error) {
	timeout := 30 * time.Second // Set HTTP request timeout to 30 seconds

	cache := &Cache[T]{
		cacheFilePath: cacheFilePath,
		remoteUrl:     remoteUrl,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}

	err := cache.loadFromDisk()
	if err != nil {
		// local cache missing or invalid, fetch from remote
		_, err = cache.Update()
		if err != nil {
			return nil, fmt.Errorf("error building cache: %v", err)
		}
	}

	return cache, nil
}

func (c *Cache[T]) Data() T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data
}

// Update refreshes the cache by fetching remote data and saving it to disk
// Returns true if the data was updated, false if it was already up-to-date
func (c *Cache[T]) Update() (bool, error) {
	updated, err := c.fetch()
	if err != nil {
		return false, err
	}

	if updated {
		// Save to disk
		err = c.saveToDisk()
		if err != nil {
			return updated, err
		}
	}

	return updated, nil
}

// fetch fetches the remote data and updates the cache
// Returns true if the data was updated, false if it was already up-to-date
func (c *Cache[T]) fetch() (bool, error) {
	// Create HTTP request with If-Modified-Since header to reduce unnecessary downloads
	req, err := http.NewRequest("GET", c.remoteUrl, nil)
	if err != nil {
		return false, err
	}
	if !c.modTime.IsZero() {
		req.Header.Set("If-Modified-Since", c.modTime.Format(http.TimeFormat))
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotModified:
		return false, nil
	case http.StatusOK:
		return true, c.parseResponse(resp)
	default:
		return true, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
	}
}

// parseResponse parses the response body and updates the cache
func (c *Cache[T]) parseResponse(resp *http.Response) error {
	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var bodyData T
	err = json.Unmarshal(body, &bodyData)
	if err != nil {
		return err
	}
	// Lock to prevent concurrent updates
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = bodyData

	// Update last modified time from response headers
	lastModified := resp.Header.Get("Last-Modified")
	if lastModified != "" {
		c.modTime, err = time.Parse(http.TimeFormat, lastModified)
		if err != nil {
			return err
		}
	}

	return nil
}

// saveToDisk persists the current cache data to the local file system
func (c *Cache[T]) saveToDisk() error {
	// Create directory structure if needed
	err := os.MkdirAll(filepath.Dir(c.cacheFilePath), 0755)
	if err != nil {
		return err
	}

	// Serialize data to JSON
	file, err := json.Marshal(c.data)
	if err != nil {
		return err
	}

	// Write to file with proper permissions
	err = os.WriteFile(c.cacheFilePath, file, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache[T]) loadFromDisk() error {
	fileInfo, err := os.Stat(c.cacheFilePath)
	if err != nil {
		// Return if file does not exist
		return ErrCacheFileNotFound
	}
	// Update last modified time
	c.modTime = fileInfo.ModTime()

	// Read the cache file.
	file, err := os.ReadFile(c.cacheFilePath)
	if err != nil {
		return fmt.Errorf("error read file from cache: %v", err)
	}

	// Unmarshal the cache file.
	var fileData T
	err = json.Unmarshal(file, &fileData)
	if err != nil {
		return fmt.Errorf("error json unmarshal from cache: %v", err)
	}
	c.data = fileData

	return nil
}
