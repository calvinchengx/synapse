package rules

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	synerr "github.com/calvinchengx/synapse/internal/errors"
)

// RegistryEntry represents a rule available in the community registry.
type RegistryEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Category    string `json:"category"`
	Author      string `json:"author"`
}

// Registry provides access to the community rule registry.
type Registry struct {
	client   *http.Client
	cacheDir string
	cacheTTL time.Duration
}

// NewRegistry creates a Registry with caching support.
func NewRegistry(cacheDir string) *Registry {
	return &Registry{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		cacheDir: cacheDir,
		cacheTTL: 1 * time.Hour,
	}
}

// Search returns registry entries matching the keyword.
func (r *Registry) Search(keyword string) ([]RegistryEntry, error) {
	entries, err := r.loadIndex()
	if err != nil {
		return nil, err
	}

	q := strings.ToLower(keyword)
	var results []RegistryEntry
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e.Name), q) ||
			strings.Contains(strings.ToLower(e.Description), q) ||
			strings.Contains(strings.ToLower(e.Category), q) {
			results = append(results, e)
		}
	}
	return results, nil
}

// FetchRule downloads a single rule file from the given URL.
func (r *Registry) FetchRule(url string) (string, error) {
	resp, err := r.client.Get(url)
	if err != nil {
		return "", synerr.NewExternalError("github", "fetch", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", synerr.NewExternalError("github", "fetch",
			fmt.Errorf("HTTP %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", synerr.NewExternalError("github", "read", err)
	}

	return string(data), nil
}

// loadIndex returns the registry index, using cache if available and fresh.
func (r *Registry) loadIndex() ([]RegistryEntry, error) {
	// Try cache first
	if entries, err := r.readCache(); err == nil {
		return entries, nil
	}

	// Fetch from remote
	entries, err := r.fetchIndex()
	if err != nil {
		// Try stale cache as fallback
		if stale, cacheErr := r.readCacheIgnoreTTL(); cacheErr == nil {
			return stale, nil
		}
		return nil, err
	}

	// Write cache
	r.writeCache(entries)
	return entries, nil
}

// fetchIndex fetches the registry index from GitHub.
func (r *Registry) fetchIndex() ([]RegistryEntry, error) {
	// This would fetch from a real GitHub API or registry endpoint.
	// For now, return empty — the actual URL will be configured later.
	return []RegistryEntry{}, nil
}

func (r *Registry) cachePath() string {
	return filepath.Join(r.cacheDir, "registry.json")
}

func (r *Registry) readCache() ([]RegistryEntry, error) {
	return r.readCacheWithTTL(r.cacheTTL)
}

func (r *Registry) readCacheIgnoreTTL() ([]RegistryEntry, error) {
	return r.readCacheWithTTL(0)
}

func (r *Registry) readCacheWithTTL(ttl time.Duration) ([]RegistryEntry, error) {
	path := r.cachePath()
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if ttl > 0 && time.Since(info.ModTime()) > ttl {
		return nil, fmt.Errorf("cache expired")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []RegistryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *Registry) writeCache(entries []RegistryEntry) error {
	if err := os.MkdirAll(filepath.Dir(r.cachePath()), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	return os.WriteFile(r.cachePath(), data, 0644)
}
