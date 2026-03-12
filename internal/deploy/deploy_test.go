package deploy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/calvinchengx/synapse/internal/rules"
)

func testRules() []rules.Rule {
	return []rules.Rule{
		{
			Filename: "security.md",
			Type:     rules.RuleTypeRule,
			Frontmatter: rules.Frontmatter{
				Description: "Security best practices",
				Keywords:    []string{"security", "auth"},
			},
			Content: "# Security\n\nAlways validate input.",
		},
		{
			Filename: "react.md",
			Type:     rules.RuleTypeSkill,
			Frontmatter: rules.Frontmatter{
				Description: "React best practices",
				AlwaysApply: true,
			},
			Content: "# React\n\nUse functional components.",
		},
		{
			Filename: "reviewer.md",
			Type:     rules.RuleTypeAgent,
			Frontmatter: rules.Frontmatter{
				Description: "Code reviewer",
				Name:        "code-reviewer",
				Tools:       "Read, Grep, Glob",
				Model:       "opus",
			},
			Content: "You are a code reviewer.",
		},
	}
}

func TestClaudeDeploy(t *testing.T) {
	dir := t.TempDir()
	d := &ClaudeDeployer{}

	if err := d.Deploy(testRules(), dir); err != nil {
		t.Fatalf("Deploy error: %v", err)
	}

	// Check rules directory exists
	rulesDir := filepath.Join(dir, ".claude", "rules")
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		t.Fatalf("reading rules dir: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 rule files, got %d", len(entries))
	}

	// Check security.md content
	data, err := os.ReadFile(filepath.Join(rulesDir, "security.md"))
	if err != nil {
		t.Fatalf("reading security.md: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "description: Security best practices") {
		t.Error("security.md missing description frontmatter")
	}
	if !strings.Contains(content, "Always validate input.") {
		t.Error("security.md missing content")
	}

	// Check rules-inactive directory exists
	if _, err := os.Stat(filepath.Join(dir, ".claude", "rules-inactive")); err != nil {
		t.Error("rules-inactive directory should exist")
	}

	// Check settings.json
	settingsData, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("reading settings.json: %v", err)
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(settingsData, &settings); err != nil {
		t.Fatalf("parsing settings.json: %v", err)
	}
	if _, ok := settings["hooks"]; !ok {
		t.Error("settings.json missing hooks")
	}
}

func TestClaudeDeployerName(t *testing.T) {
	d := &ClaudeDeployer{}
	if d.Name() != "claude" {
		t.Errorf("Name() = %s, want claude", d.Name())
	}
}

func TestCursorDeploy(t *testing.T) {
	dir := t.TempDir()
	d := &CursorDeployer{}

	if err := d.Deploy(testRules(), dir); err != nil {
		t.Fatalf("Deploy error: %v", err)
	}

	// Check rules directory
	rulesDir := filepath.Join(dir, ".cursor", "rules")
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		t.Fatalf("reading rules dir: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 .mdc files, got %d", len(entries))
	}

	// Check .mdc extension
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".mdc") {
			t.Errorf("expected .mdc extension, got %s", e.Name())
		}
	}

	// Check react.mdc content
	data, err := os.ReadFile(filepath.Join(rulesDir, "react.mdc"))
	if err != nil {
		t.Fatalf("reading react.mdc: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "alwaysApply: true") {
		t.Error("react.mdc missing alwaysApply")
	}
	if !strings.Contains(content, "Use functional components.") {
		t.Error("react.mdc missing content")
	}
}

func TestCursorDeployerName(t *testing.T) {
	d := &CursorDeployer{}
	if d.Name() != "cursor" {
		t.Errorf("Name() = %s, want cursor", d.Name())
	}
}

func TestCodexDeploy(t *testing.T) {
	dir := t.TempDir()
	d := &CodexDeployer{}

	if err := d.Deploy(testRules(), dir); err != nil {
		t.Fatalf("Deploy error: %v", err)
	}

	// Check AGENTS.md exists
	agentsPath := filepath.Join(dir, ".codex", "AGENTS.md")
	data, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "# Agent Instructions") {
		t.Error("AGENTS.md missing header")
	}
	if !strings.Contains(content, "## security.md") {
		t.Error("AGENTS.md missing security.md section")
	}
	if !strings.Contains(content, "## react.md") {
		t.Error("AGENTS.md missing react.md section")
	}
	if !strings.Contains(content, "## reviewer.md") {
		t.Error("AGENTS.md missing reviewer.md section")
	}
	if !strings.Contains(content, "Always validate input.") {
		t.Error("AGENTS.md missing security content")
	}
}

func TestCodexDeployerName(t *testing.T) {
	d := &CodexDeployer{}
	if d.Name() != "codex" {
		t.Errorf("Name() = %s, want codex", d.Name())
	}
}

func TestDeployAll(t *testing.T) {
	dir := t.TempDir()
	deployers := []Deployer{
		&ClaudeDeployer{},
		&CursorDeployer{},
		&CodexDeployer{},
	}

	errs := DeployAll(deployers, testRules(), dir)
	if len(errs) > 0 {
		t.Errorf("DeployAll errors: %v", errs)
	}

	// Verify all targets created
	for _, path := range []string{
		filepath.Join(dir, ".claude", "rules"),
		filepath.Join(dir, ".cursor", "rules"),
		filepath.Join(dir, ".codex", "AGENTS.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("missing: %s", path)
		}
	}
}

func TestBuildRuleContentNoFrontmatter(t *testing.T) {
	rule := rules.Rule{
		Filename: "simple.md",
		Content:  "# Simple\n\nJust content.",
	}
	content := buildRuleContent(rule)
	if content != "# Simple\n\nJust content." {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestBuildRuleContentWithFrontmatter(t *testing.T) {
	rule := rules.Rule{
		Filename: "full.md",
		Frontmatter: rules.Frontmatter{
			Description: "Full rule",
			Keywords:    []string{"a", "b"},
			Tools:       "Read, Write",
			Model:       "opus",
			AlwaysApply: true,
		},
		Content: "Body content.",
	}
	content := buildRuleContent(rule)
	if !strings.Contains(content, "description: Full rule") {
		t.Error("missing description")
	}
	if !strings.Contains(content, "tools: Read, Write") {
		t.Error("missing tools")
	}
	if !strings.Contains(content, "model: opus") {
		t.Error("missing model")
	}
	if !strings.Contains(content, "alwaysApply: true") {
		t.Error("missing alwaysApply")
	}
	if !strings.Contains(content, "Body content.") {
		t.Error("missing body content")
	}
}
