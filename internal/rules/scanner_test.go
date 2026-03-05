package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantDesc    string
		wantKW      []string
		wantContent string
	}{
		{
			name: "standard frontmatter",
			input: `---
description: Test rule for security
keywords:
  - security
  - auth
---

# Security Rule

Content here.`,
			wantDesc:    "Test rule for security",
			wantKW:      []string{"security", "auth"},
			wantContent: "# Security Rule\n\nContent here.",
		},
		{
			name:        "no frontmatter",
			input:       "# Just a heading\n\nSome content.",
			wantDesc:    "",
			wantKW:      nil,
			wantContent: "# Just a heading\n\nSome content.",
		},
		{
			name: "frontmatter with tools and model",
			input: `---
description: Code reviewer agent
tools: Read, Grep, Glob
model: opus
---

You are a code reviewer.`,
			wantDesc:    "Code reviewer agent",
			wantContent: "You are a code reviewer.",
		},
		{
			name: "empty frontmatter",
			input: `---
---

Content only.`,
			wantDesc:    "",
			wantContent: "Content only.",
		},
		{
			name: "unclosed frontmatter",
			input: `---
description: unclosed
This is content`,
			wantDesc:    "",
			wantContent: "---\ndescription: unclosed\nThis is content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, content, err := parseFrontmatter(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fm.Description != tt.wantDesc {
				t.Errorf("description = %q, want %q", fm.Description, tt.wantDesc)
			}
			if tt.wantKW != nil {
				if len(fm.Keywords) != len(tt.wantKW) {
					t.Errorf("keywords len = %d, want %d", len(fm.Keywords), len(tt.wantKW))
				}
				for i, kw := range tt.wantKW {
					if i < len(fm.Keywords) && fm.Keywords[i] != kw {
						t.Errorf("keyword[%d] = %q, want %q", i, fm.Keywords[i], kw)
					}
				}
			}
			if content != tt.wantContent {
				t.Errorf("content = %q, want %q", content, tt.wantContent)
			}
		})
	}
}

func TestInferRuleType(t *testing.T) {
	tests := []struct {
		relPath  string
		expected RuleType
	}{
		{"rules/security.md", RuleTypeRule},
		{"skills/react.md", RuleTypeSkill},
		{"commands/commit.md", RuleTypeCommand},
		{"agents/code-reviewer.md", RuleTypeAgent},
		{"contexts/dev.md", RuleTypeContext},
		{"other/unknown.md", RuleTypeRule},
		{"standalone.md", RuleTypeRule},
	}

	for _, tt := range tests {
		t.Run(tt.relPath, func(t *testing.T) {
			got := inferRuleType(tt.relPath)
			if got != tt.expected {
				t.Errorf("inferRuleType(%q) = %s, want %s", tt.relPath, got, tt.expected)
			}
		})
	}
}

func TestScanDir(t *testing.T) {
	root := t.TempDir()

	// Create directory structure
	dirs := []string{"rules", "skills", "agents"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Write test files
	files := map[string]string{
		"rules/security.md": `---
description: Security best practices
keywords:
  - security
  - auth
---

# Security

Always validate input.`,
		"rules/testing.md": `---
description: Testing guidelines
---

# Testing

Write tests.`,
		"skills/react.md": `---
description: React best practices
keywords:
  - react
  - frontend
---

# React`,
		"agents/reviewer.md": `---
description: Code reviewer
tools: Read, Grep
model: opus
---

Review code.`,
	}

	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Also create a non-.md file that should be ignored
	if err := os.WriteFile(filepath.Join(root, "rules/README.txt"), []byte("ignore me"), 0644); err != nil {
		t.Fatal(err)
	}

	rules, err := ScanDir(root)
	if err != nil {
		t.Fatalf("ScanDir error: %v", err)
	}

	if len(rules) != 4 {
		t.Fatalf("expected 4 rules, got %d", len(rules))
	}

	// Check that types are correctly inferred
	typeCounts := make(map[RuleType]int)
	for _, r := range rules {
		typeCounts[r.Type]++
	}
	if typeCounts[RuleTypeRule] != 2 {
		t.Errorf("expected 2 rules, got %d", typeCounts[RuleTypeRule])
	}
	if typeCounts[RuleTypeSkill] != 1 {
		t.Errorf("expected 1 skill, got %d", typeCounts[RuleTypeSkill])
	}
	if typeCounts[RuleTypeAgent] != 1 {
		t.Errorf("expected 1 agent, got %d", typeCounts[RuleTypeAgent])
	}

	// Check a specific rule
	for _, r := range rules {
		if r.Filename == "security.md" {
			if r.Frontmatter.Description != "Security best practices" {
				t.Errorf("security.md description = %q", r.Frontmatter.Description)
			}
			if len(r.Frontmatter.Keywords) != 2 {
				t.Errorf("security.md keywords len = %d, want 2", len(r.Frontmatter.Keywords))
			}
			if r.Hash == "" {
				t.Error("security.md hash should not be empty")
			}
		}
	}
}

func TestScanDirEmpty(t *testing.T) {
	root := t.TempDir()
	rules, err := ScanDir(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 0 {
		t.Errorf("expected 0 rules in empty dir, got %d", len(rules))
	}
}

func TestScanDirNonexistent(t *testing.T) {
	_, err := ScanDir("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestParseRuleFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "test.md")

	content := `---
description: A test rule
keywords:
  - test
alwaysApply: true
---

# Test Rule

This is a test.`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	rule, err := ParseRuleFile(root, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rule.Filename != "test.md" {
		t.Errorf("filename = %q, want test.md", rule.Filename)
	}
	if rule.Frontmatter.Description != "A test rule" {
		t.Errorf("description = %q", rule.Frontmatter.Description)
	}
	if !rule.Frontmatter.AlwaysApply {
		t.Error("expected alwaysApply = true")
	}
	if rule.Content != "# Test Rule\n\nThis is a test." {
		t.Errorf("content = %q", rule.Content)
	}
}
