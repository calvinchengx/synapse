package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadManifests loads tool manifests from up to 3 paths:
// 1. Built-in (bundled in binary)
// 2. User global (~/.synapse/integrations/)
// 3. Project-local (.synapse/integrations/)
func LoadManifests(paths ...string) ([]ToolManifest, error) {
	var all []ToolManifest
	seen := make(map[string]bool)

	for _, p := range paths {
		manifests, err := loadFromPath(p)
		if err != nil {
			continue // skip missing/broken paths
		}
		for _, m := range manifests {
			if !seen[m.ID] {
				seen[m.ID] = true
				all = append(all, m)
			}
		}
	}
	return all, nil
}

// loadFromPath loads manifests from a directory or single YAML file.
func loadFromPath(path string) ([]ToolManifest, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return loadManifestFile(path)
	}

	// Load all .yaml and .yml files in the directory
	var all []ToolManifest
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			manifests, err := loadManifestFile(filepath.Join(path, name))
			if err != nil {
				continue
			}
			all = append(all, manifests...)
		}
	}
	return all, nil
}

// loadManifestFile parses a single YAML manifest file.
func loadManifestFile(path string) ([]ToolManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var mf ManifestFile
	if err := yaml.Unmarshal(data, &mf); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return mf.Tools, nil
}

// ExpandPath expands ~ to the user's home directory.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
