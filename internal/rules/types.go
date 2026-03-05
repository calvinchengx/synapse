package rules

import (
	"strings"
	"time"
)

// RuleType classifies a rule file by its directory origin.
type RuleType string

const (
	RuleTypeRule    RuleType = "rule"
	RuleTypeSkill   RuleType = "skill"
	RuleTypeCommand RuleType = "command"
	RuleTypeAgent   RuleType = "agent"
	RuleTypeContext RuleType = "context"
)

// Rule represents a single .md rule file with its parsed metadata.
type Rule struct {
	// Path is the absolute filesystem path to the .md file.
	Path string

	// RelPath is the path relative to the config root (e.g., "rules/security.md").
	RelPath string

	// Filename is the base filename (e.g., "security.md").
	Filename string

	// Type is derived from the parent directory.
	Type RuleType

	// Frontmatter holds parsed YAML frontmatter fields.
	Frontmatter Frontmatter

	// Content is the markdown body after frontmatter.
	Content string

	// Hash is the MD5 hex digest of the file content (for change detection).
	Hash string
}

// Frontmatter holds YAML metadata parsed from the top of a .md file.
type Frontmatter struct {
	Description string   `yaml:"description"`
	Keywords    []string `yaml:"keywords"`
	Tools       string   `yaml:"tools"` // comma-separated string (ai-nexus format)
	Model       string   `yaml:"model"`
	Category    string   `yaml:"category"`
	Name        string   `yaml:"name"`
	AlwaysApply bool     `yaml:"alwaysApply"`
}

// ToolsList returns the tools as a slice, splitting on comma.
func (f Frontmatter) ToolsList() []string {
	if f.Tools == "" {
		return nil
	}
	parts := strings.Split(f.Tools, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// Category groups rules by their type and optional subcategory.
type Category struct {
	Name  string
	Type  RuleType
	Rules []Rule
}

// FileInfo tracks a deployed rule file for change detection.
type FileInfo struct {
	RelPath    string    `json:"rel_path"`
	Hash       string    `json:"hash"`
	DeployedAt time.Time `json:"deployed_at"`
	Source     string    `json:"source"`
}

// DotrulesMeta is the installation metadata stored in ~/.synapse/meta.json.
type DotrulesMeta struct {
	Version   string            `json:"version"`
	Sources   []SourceEntry     `json:"sources"`
	Files     map[string]FileInfo `json:"files"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// SourceEntry represents a configured rule source.
type SourceEntry struct {
	Type   string `json:"type"`   // "local" or "git"
	Path   string `json:"path"`   // local path or git URL
	Branch string `json:"branch"` // git branch (empty for local)
}

// RouterResult holds the output of the semantic router.
type RouterResult struct {
	Selected []Rule
	Method   string // "semantic", "keyword", "integration"
	Tier     int    // 1, 2, or 3
}
