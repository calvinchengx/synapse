package rules

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRegistrySearch(t *testing.T) {
	cacheDir := t.TempDir()

	// Pre-populate cache
	entries := []RegistryEntry{
		{Name: "security-rules", Description: "Security best practices", Category: "security"},
		{Name: "react-patterns", Description: "React component patterns", Category: "frontend"},
		{Name: "go-best-practices", Description: "Go coding conventions", Category: "language"},
	}
	data, _ := json.Marshal(entries)
	os.MkdirAll(filepath.Join(cacheDir, "cache"), 0755)
	os.WriteFile(filepath.Join(cacheDir, "cache", "registry.json"), data, 0644)

	reg := NewRegistry(filepath.Join(cacheDir, "cache"))

	tests := []struct {
		query string
		want  int
	}{
		{"security", 1},
		{"react", 1},
		{"go", 1},
		{"frontend", 1},
		{"nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results, err := reg.Search(tt.query)
			if err != nil {
				t.Fatalf("Search error: %v", err)
			}
			if len(results) != tt.want {
				t.Errorf("Search(%q) = %d results, want %d", tt.query, len(results), tt.want)
			}
		})
	}
}

func TestRegistryFetchRule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("---\ndescription: Test rule\n---\n\n# Test\n\nContent."))
	}))
	defer server.Close()

	reg := NewRegistry(t.TempDir())
	content, err := reg.FetchRule(server.URL + "/test.md")
	if err != nil {
		t.Fatalf("FetchRule error: %v", err)
	}
	if content == "" {
		t.Error("expected non-empty content")
	}
}

func TestRegistryFetchRuleNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	reg := NewRegistry(t.TempDir())
	_, err := reg.FetchRule(server.URL + "/missing.md")
	if err == nil {
		t.Error("expected error for 404")
	}
}

func TestRegistryCacheWriteAndRead(t *testing.T) {
	cacheDir := t.TempDir()
	reg := NewRegistry(cacheDir)

	entries := []RegistryEntry{
		{Name: "test", Description: "test rule"},
	}
	if err := reg.writeCache(entries); err != nil {
		t.Fatalf("writeCache error: %v", err)
	}

	read, err := reg.readCache()
	if err != nil {
		t.Fatalf("readCache error: %v", err)
	}
	if len(read) != 1 || read[0].Name != "test" {
		t.Errorf("cache round-trip failed: %v", read)
	}
}
