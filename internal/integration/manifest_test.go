package integration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifestFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")

	yaml := `
tools:
  - id: testtool
    name: TestTool
    description: A test tool
    homepage: https://example.com
    detection:
      binary: testtool
    capabilities:
      - testing
    dataEndpoints:
      - id: data
        type: sqlite
        path: "~/.testtool/data.db"
`
	os.WriteFile(path, []byte(yaml), 0644)

	manifests, err := loadManifestFile(path)
	if err != nil {
		t.Fatalf("loadManifestFile error: %v", err)
	}
	if len(manifests) != 1 {
		t.Fatalf("expected 1 manifest, got %d", len(manifests))
	}
	if manifests[0].ID != "testtool" {
		t.Errorf("id = %q, want testtool", manifests[0].ID)
	}
	if manifests[0].Name != "TestTool" {
		t.Errorf("name = %q, want TestTool", manifests[0].Name)
	}
	if len(manifests[0].Capabilities) != 1 {
		t.Errorf("expected 1 capability, got %d", len(manifests[0].Capabilities))
	}
}

func TestLoadManifests(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	// Write manifest in dir1
	os.WriteFile(filepath.Join(dir1, "tools.yaml"), []byte(`
tools:
  - id: tool1
    name: Tool1
    description: First tool
`), 0644)

	// Write manifest in dir2 with duplicate and new
	os.WriteFile(filepath.Join(dir2, "tools.yaml"), []byte(`
tools:
  - id: tool1
    name: Tool1Duplicate
    description: Should be skipped
  - id: tool2
    name: Tool2
    description: Second tool
`), 0644)

	manifests, err := LoadManifests(dir1, dir2)
	if err != nil {
		t.Fatalf("LoadManifests error: %v", err)
	}

	if len(manifests) != 2 {
		t.Fatalf("expected 2 manifests (deduped), got %d", len(manifests))
	}
	// First tool1 should win
	if manifests[0].Name != "Tool1" {
		t.Errorf("first manifest name = %q, want Tool1", manifests[0].Name)
	}
}

func TestLoadManifestsMissingPath(t *testing.T) {
	manifests, err := LoadManifests("/nonexistent/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(manifests) != 0 {
		t.Errorf("expected 0 manifests for missing path, got %d", len(manifests))
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(home, "test")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		got := ExpandPath(tt.input)
		if got != tt.expected {
			t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestDataFilePathResolvedPath(t *testing.T) {
	dfp := DataFilePath{
		Linux:   "/linux/path",
		Darwin:  "/darwin/path",
		Windows: "/windows/path",
		Path:    "/fallback/path",
	}

	if got := dfp.ResolvedPath("linux"); got != "/linux/path" {
		t.Errorf("linux = %q", got)
	}
	if got := dfp.ResolvedPath("darwin"); got != "/darwin/path" {
		t.Errorf("darwin = %q", got)
	}
	if got := dfp.ResolvedPath("windows"); got != "/windows/path" {
		t.Errorf("windows = %q", got)
	}
	if got := dfp.ResolvedPath("freebsd"); got != "/fallback/path" {
		t.Errorf("freebsd = %q", got)
	}
}

func TestDataFilePathFallback(t *testing.T) {
	dfp := DataFilePath{Path: "/common/path"}

	if got := dfp.ResolvedPath("linux"); got != "/common/path" {
		t.Errorf("linux fallback = %q", got)
	}
	if got := dfp.ResolvedPath("darwin"); got != "/common/path" {
		t.Errorf("darwin fallback = %q", got)
	}
}
