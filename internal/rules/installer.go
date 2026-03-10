package rules

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Installer manages rule installation, updates, and removal.
type Installer struct {
	engine  *Engine
	metaDir string // ~/.synapse/
}

// NewInstaller creates an Installer.
func NewInstaller(engine *Engine, metaDir string) *Installer {
	return &Installer{engine: engine, metaDir: metaDir}
}

// Init initializes Synapse in the project directory. Creates ~/.synapse/
// and writes initial metadata.
func (inst *Installer) Init(projectDir string) error {
	if err := os.MkdirAll(inst.metaDir, 0755); err != nil {
		return fmt.Errorf("creating synapse dir: %w", err)
	}

	// Write initial metadata
	meta := DotrulesMeta{
		Version:   "1.0.0",
		Files:     make(map[string]FileInfo),
		UpdatedAt: time.Now(),
	}

	for _, rule := range inst.engine.All() {
		meta.Files[rule.Filename] = FileInfo{
			RelPath:    rule.RelPath,
			Hash:       rule.Hash,
			DeployedAt: time.Now(),
			Source:     "built-in",
		}
	}

	return inst.writeMeta(meta)
}

// Update compares current rule hashes with deployed files and updates
// any that haven't been modified by the user.
func (inst *Installer) Update(projectDir string) (UpdateReport, error) {
	meta, err := inst.readMeta()
	if err != nil {
		return UpdateReport{}, fmt.Errorf("reading metadata: %w", err)
	}

	report := UpdateReport{}

	for _, rule := range inst.engine.All() {
		existing, tracked := meta.Files[rule.Filename]
		if !tracked {
			// New rule
			meta.Files[rule.Filename] = FileInfo{
				RelPath:    rule.RelPath,
				Hash:       rule.Hash,
				DeployedAt: time.Now(),
				Source:     "built-in",
			}
			report.Added = append(report.Added, rule.Filename)
			continue
		}

		if existing.Hash == rule.Hash {
			report.Unchanged = append(report.Unchanged, rule.Filename)
			continue
		}

		// Check if user modified the deployed file
		deployedPath := filepath.Join(projectDir, ".claude", "rules", rule.Filename)
		if isUserModified(deployedPath, existing.Hash) {
			report.Skipped = append(report.Skipped, rule.Filename)
			continue
		}

		// Safe to update
		meta.Files[rule.Filename] = FileInfo{
			RelPath:    rule.RelPath,
			Hash:       rule.Hash,
			DeployedAt: time.Now(),
			Source:     existing.Source,
		}
		report.Updated = append(report.Updated, rule.Filename)
	}

	meta.UpdatedAt = time.Now()
	if err := inst.writeMeta(meta); err != nil {
		return report, fmt.Errorf("writing metadata: %w", err)
	}

	return report, nil
}

// Uninstall removes all Synapse-managed files from the project.
func (inst *Installer) Uninstall(projectDir string) error {
	// Remove .claude/rules/ managed files
	meta, err := inst.readMeta()
	if err != nil {
		// No metadata means nothing to uninstall
		return nil
	}

	rulesDir := filepath.Join(projectDir, ".claude", "rules")
	for filename := range meta.Files {
		path := filepath.Join(rulesDir, filename)
		os.Remove(path) // ignore errors for missing files
	}

	// Remove settings.json hook
	os.Remove(filepath.Join(projectDir, ".claude", "settings.json"))

	// Remove metadata
	os.Remove(filepath.Join(inst.metaDir, "meta.json"))

	return nil
}

// UpdateReport summarizes what changed during an update.
type UpdateReport struct {
	Added     []string
	Updated   []string
	Unchanged []string
	Skipped   []string // user-modified files not overwritten
}

func (inst *Installer) writeMeta(meta DotrulesMeta) error {
	path := filepath.Join(inst.metaDir, "meta.json")
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (inst *Installer) readMeta() (DotrulesMeta, error) {
	path := filepath.Join(inst.metaDir, "meta.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return DotrulesMeta{Files: make(map[string]FileInfo)}, err
	}
	var meta DotrulesMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return DotrulesMeta{Files: make(map[string]FileInfo)}, err
	}
	if meta.Files == nil {
		meta.Files = make(map[string]FileInfo)
	}
	return meta, nil
}

// isUserModified checks if the file at path has been modified by the user
// (i.e., its current hash differs from the expected hash).
func isUserModified(path, expectedHash string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false // file doesn't exist = not user-modified
	}
	currentHash := fmt.Sprintf("%x", md5.Sum(data))
	return currentHash != expectedHash
}
