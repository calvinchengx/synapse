package rules

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallerInit(t *testing.T) {
	metaDir := t.TempDir()
	projectDir := t.TempDir()

	rules := []Rule{
		{Filename: "security.md", RelPath: "rules/security.md", Hash: "abc123", Content: "security"},
		{Filename: "react.md", RelPath: "skills/react.md", Hash: "def456", Content: "react"},
	}
	engine := NewEngineFromRules(rules)
	installer := NewInstaller(engine, metaDir)

	if err := installer.Init(projectDir); err != nil {
		t.Fatalf("Init error: %v", err)
	}

	// Verify meta.json was created
	metaPath := filepath.Join(metaDir, "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("reading meta.json: %v", err)
	}

	var meta DotrulesMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("parsing meta.json: %v", err)
	}

	if meta.Version != "1.0.0" {
		t.Errorf("version = %s, want 1.0.0", meta.Version)
	}
	if len(meta.Files) != 2 {
		t.Errorf("expected 2 files in metadata, got %d", len(meta.Files))
	}
	if meta.Files["security.md"].Hash != "abc123" {
		t.Error("security.md hash mismatch")
	}
}

func TestInstallerUpdate(t *testing.T) {
	metaDir := t.TempDir()
	projectDir := t.TempDir()

	// Create .claude/rules/ with a deployed file
	rulesDir := filepath.Join(projectDir, ".claude", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Initial rules — hash must match md5 of the deployed content "v1"
	deployedContent := "v1"
	deployedHash := fmt.Sprintf("%x", md5.Sum([]byte(deployedContent)))
	initialRules := []Rule{
		{Filename: "security.md", RelPath: "rules/security.md", Hash: deployedHash, Content: deployedContent},
	}
	engine := NewEngineFromRules(initialRules)
	installer := NewInstaller(engine, metaDir)

	if err := installer.Init(projectDir); err != nil {
		t.Fatal(err)
	}

	// Write the deployed file matching the tracked hash
	os.WriteFile(filepath.Join(rulesDir, "security.md"), []byte(deployedContent), 0644)

	// Update with new version
	updatedRules := []Rule{
		{Filename: "security.md", RelPath: "rules/security.md", Hash: "hash2", Content: "v2"},
		{Filename: "testing.md", RelPath: "rules/testing.md", Hash: "hash3", Content: "new"},
	}
	engine2 := NewEngineFromRules(updatedRules)
	installer2 := NewInstaller(engine2, metaDir)

	report, err := installer2.Update(projectDir)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	if len(report.Updated) != 1 || report.Updated[0] != "security.md" {
		t.Errorf("expected security.md updated, got %v", report.Updated)
	}
	if len(report.Added) != 1 || report.Added[0] != "testing.md" {
		t.Errorf("expected testing.md added, got %v", report.Added)
	}
}

func TestInstallerUpdateSkipsUserModified(t *testing.T) {
	metaDir := t.TempDir()
	projectDir := t.TempDir()

	rulesDir := filepath.Join(projectDir, ".claude", "rules")
	os.MkdirAll(rulesDir, 0755)

	// Init with a rule — use consistent hash
	originalContent := "v1"
	originalHash := fmt.Sprintf("%x", md5.Sum([]byte(originalContent)))
	initialRules := []Rule{
		{Filename: "security.md", RelPath: "rules/security.md", Hash: originalHash, Content: originalContent},
	}
	engine := NewEngineFromRules(initialRules)
	installer := NewInstaller(engine, metaDir)
	installer.Init(projectDir)

	// Deploy the file, then user modifies it
	os.WriteFile(filepath.Join(rulesDir, "security.md"), []byte("user modified content"), 0644)

	// Try to update with new version
	updatedRules := []Rule{
		{Filename: "security.md", RelPath: "rules/security.md", Hash: "newhash", Content: "v2"},
	}
	engine2 := NewEngineFromRules(updatedRules)
	installer2 := NewInstaller(engine2, metaDir)

	report, err := installer2.Update(projectDir)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	if len(report.Skipped) != 1 {
		t.Errorf("expected 1 skipped file, got %d", len(report.Skipped))
	}
	if len(report.Updated) != 0 {
		t.Errorf("expected 0 updated files, got %d", len(report.Updated))
	}
}

func TestInstallerUninstall(t *testing.T) {
	metaDir := t.TempDir()
	projectDir := t.TempDir()

	rulesDir := filepath.Join(projectDir, ".claude", "rules")
	os.MkdirAll(rulesDir, 0755)

	rules := []Rule{
		{Filename: "security.md", RelPath: "rules/security.md", Hash: "hash1"},
	}
	engine := NewEngineFromRules(rules)
	installer := NewInstaller(engine, metaDir)
	installer.Init(projectDir)

	// Create the deployed file
	os.WriteFile(filepath.Join(rulesDir, "security.md"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(projectDir, ".claude", "settings.json"), []byte("{}"), 0644)

	if err := installer.Uninstall(projectDir); err != nil {
		t.Fatalf("Uninstall error: %v", err)
	}

	// Verify files are removed
	if _, err := os.Stat(filepath.Join(rulesDir, "security.md")); !os.IsNotExist(err) {
		t.Error("security.md should be removed")
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".claude", "settings.json")); !os.IsNotExist(err) {
		t.Error("settings.json should be removed")
	}
	if _, err := os.Stat(filepath.Join(metaDir, "meta.json")); !os.IsNotExist(err) {
		t.Error("meta.json should be removed")
	}
}

func TestInstallerUninstallNoMeta(t *testing.T) {
	metaDir := t.TempDir()
	projectDir := t.TempDir()

	engine := NewEngineFromRules(nil)
	installer := NewInstaller(engine, metaDir)

	// Should not error when no metadata exists
	if err := installer.Uninstall(projectDir); err != nil {
		t.Fatalf("Uninstall error: %v", err)
	}
}
