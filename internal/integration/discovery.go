package integration

import (
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// DiscoverAll probes all manifested tools and returns their status.
func DiscoverAll(manifests []ToolManifest) []ToolStatus {
	statuses := make([]ToolStatus, len(manifests))
	for i, m := range manifests {
		statuses[i] = Discover(m)
	}
	return statuses
}

// Discover checks whether a single tool is installed and accessible.
func Discover(manifest ToolManifest) ToolStatus {
	status := ToolStatus{Tool: manifest}

	// Check binary
	if manifest.Detection.Binary != "" {
		path, err := FindBinary(manifest.Detection.Binary)
		if err == nil {
			status.BinaryFound = true
			status.BinaryPath = path
			status.Installed = true
		}
	}

	// Check data files
	platform := runtime.GOOS
	for _, df := range manifest.Detection.DataFiles {
		resolved := df.ResolvedPath(platform)
		if resolved == "" {
			continue
		}
		expanded := ExpandPath(resolved)
		if _, err := os.Stat(expanded); err == nil {
			status.DataFilesFound = append(status.DataFilesFound, expanded)
		} else {
			status.DataFilesMissing = append(status.DataFilesMissing, expanded)
		}
	}

	// Check HTTP endpoints
	for _, port := range manifest.Detection.Ports {
		reachable := ProbeHTTP(port, 3*time.Second)
		status.APIReachable = &reachable
		break // only check first port
	}

	return status
}

// FindBinary searches for a binary in PATH.
func FindBinary(name string) (string, error) {
	return exec.LookPath(name)
}

// CheckDataFiles checks which paths exist and which are missing.
func CheckDataFiles(paths []string) (found, missing []string) {
	for _, p := range paths {
		expanded := ExpandPath(p)
		if _, err := os.Stat(expanded); err == nil {
			found = append(found, expanded)
		} else {
			missing = append(missing, expanded)
		}
	}
	return
}

// ProbeHTTP checks if an HTTP server is listening on the given port.
func ProbeHTTP(port int, timeout time.Duration) bool {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get("http://localhost:" + itoa(port))
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}
