package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestFindBinaryExists(t *testing.T) {
	// "go" should be in PATH
	path, err := FindBinary("go")
	if err != nil {
		t.Skipf("go not in PATH: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path for go binary")
	}
}

func TestFindBinaryNotExists(t *testing.T) {
	_, err := FindBinary("nonexistent-binary-xyz123")
	if err == nil {
		t.Error("expected error for nonexistent binary")
	}
}

func TestCheckDataFiles(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "exists.db")
	os.WriteFile(existing, []byte("data"), 0644)

	found, missing := CheckDataFiles([]string{existing, filepath.Join(dir, "missing.db")})
	if len(found) != 1 {
		t.Errorf("expected 1 found, got %d", len(found))
	}
	if len(missing) != 1 {
		t.Errorf("expected 1 missing, got %d", len(missing))
	}
}

func TestProbeHTTPReachable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Extract port from server URL
	port := 0
	for i := len(server.URL) - 1; i >= 0; i-- {
		if server.URL[i] == ':' {
			port, _ = strconv.Atoi(server.URL[i+1:])
			break
		}
	}

	if !ProbeHTTP(port, 2*time.Second) {
		t.Error("expected ProbeHTTP to return true for running server")
	}
}

func TestProbeHTTPUnreachable(t *testing.T) {
	// Port 1 is unlikely to be listening
	if ProbeHTTP(1, 500*time.Millisecond) {
		t.Error("expected ProbeHTTP to return false for unreachable port")
	}
}

func TestDiscoverWithBinary(t *testing.T) {
	manifest := ToolManifest{
		ID:   "go-test",
		Name: "Go",
		Detection: Detection{
			Binary: "go",
		},
	}

	status := Discover(manifest)
	if !status.BinaryFound {
		t.Skip("go not found in PATH")
	}
	if !status.Installed {
		t.Error("expected Installed=true when binary found")
	}
	if status.BinaryPath == "" {
		t.Error("expected non-empty BinaryPath")
	}
}

func TestDiscoverMissingBinary(t *testing.T) {
	manifest := ToolManifest{
		ID:   "missing",
		Name: "Missing",
		Detection: Detection{
			Binary: "nonexistent-binary-xyz123",
		},
	}

	status := Discover(manifest)
	if status.BinaryFound {
		t.Error("expected BinaryFound=false for missing binary")
	}
	if status.Installed {
		t.Error("expected Installed=false for missing binary")
	}
}

func TestDiscoverAll(t *testing.T) {
	manifests := []ToolManifest{
		{ID: "a", Detection: Detection{Binary: "go"}},
		{ID: "b", Detection: Detection{Binary: "nonexistent-xyz"}},
	}

	statuses := DiscoverAll(manifests)
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{8080, "8080"},
		{58080, "58080"},
	}
	for _, tt := range tests {
		if got := itoa(tt.input); got != tt.expected {
			t.Errorf("itoa(%d) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
