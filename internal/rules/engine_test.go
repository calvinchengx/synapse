package rules

import (
	"testing"
)

func makeTestRules() []Rule {
	return []Rule{
		{Filename: "security.md", Type: RuleTypeRule, Frontmatter: Frontmatter{
			Description: "Security best practices",
			Keywords:    []string{"security", "auth", "xss"},
		}},
		{Filename: "testing.md", Type: RuleTypeRule, Frontmatter: Frontmatter{
			Description: "Testing guidelines",
			Keywords:    []string{"test", "unit", "integration"},
		}},
		{Filename: "react.md", Type: RuleTypeSkill, Frontmatter: Frontmatter{
			Description: "React best practices",
			Keywords:    []string{"react", "frontend", "component"},
		}},
		{Filename: "commit.md", Type: RuleTypeCommand, Frontmatter: Frontmatter{
			Description: "Git commit conventions",
			Keywords:    []string{"git", "commit"},
		}},
		{Filename: "code-reviewer.md", Type: RuleTypeAgent, Frontmatter: Frontmatter{
			Description: "Code review specialist",
			Keywords:    []string{"review", "quality"},
		}},
		{Filename: "dev.md", Type: RuleTypeContext, Frontmatter: Frontmatter{
			Description: "Development context",
		}},
	}
}

func TestEngineAll(t *testing.T) {
	e := NewEngineFromRules(makeTestRules())
	if got := e.Count(); got != 6 {
		t.Errorf("Count() = %d, want 6", got)
	}
}

func TestEngineByType(t *testing.T) {
	e := NewEngineFromRules(makeTestRules())

	tests := []struct {
		ruleType RuleType
		want     int
	}{
		{RuleTypeRule, 2},
		{RuleTypeSkill, 1},
		{RuleTypeCommand, 1},
		{RuleTypeAgent, 1},
		{RuleTypeContext, 1},
	}

	for _, tt := range tests {
		t.Run(string(tt.ruleType), func(t *testing.T) {
			got := e.ByType(tt.ruleType)
			if len(got) != tt.want {
				t.Errorf("ByType(%s) = %d rules, want %d", tt.ruleType, len(got), tt.want)
			}
		})
	}
}

func TestEngineCategories(t *testing.T) {
	e := NewEngineFromRules(makeTestRules())
	cats := e.Categories()

	if len(cats) != 5 {
		t.Fatalf("expected 5 categories, got %d", len(cats))
	}

	// Check category names
	expectedNames := []string{"rules", "skills", "commands", "agents", "contexts"}
	for i, cat := range cats {
		if cat.Name != expectedNames[i] {
			t.Errorf("category[%d].Name = %s, want %s", i, cat.Name, expectedNames[i])
		}
	}
}

func TestEngineSearch(t *testing.T) {
	e := NewEngineFromRules(makeTestRules())

	tests := []struct {
		query string
		want  int
	}{
		{"security", 1},
		{"react", 1},
		{"test", 1}, // testing.md (filename match)
		{"commit", 1},
		{"review", 1},
		{"nonexistent", 0},
		{"best practices", 2}, // security + react descriptions
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := e.Search(tt.query)
			if len(got) != tt.want {
				names := make([]string, len(got))
				for i, r := range got {
					names[i] = r.Filename
				}
				t.Errorf("Search(%q) = %d rules %v, want %d", tt.query, len(got), names, tt.want)
			}
		})
	}
}

func TestEngineGet(t *testing.T) {
	e := NewEngineFromRules(makeTestRules())

	r, ok := e.Get("security.md")
	if !ok {
		t.Fatal("expected to find security.md")
	}
	if r.Frontmatter.Description != "Security best practices" {
		t.Errorf("description = %q", r.Frontmatter.Description)
	}

	_, ok = e.Get("nonexistent.md")
	if ok {
		t.Error("expected not to find nonexistent.md")
	}
}

func TestNewEngine(t *testing.T) {
	// Test with empty directory
	root := t.TempDir()
	e, err := NewEngine(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Count() != 0 {
		t.Errorf("expected 0 rules, got %d", e.Count())
	}
}

func TestEngineCategoriesEmpty(t *testing.T) {
	e := NewEngineFromRules(nil)
	cats := e.Categories()
	if len(cats) != 0 {
		t.Errorf("expected 0 categories for empty engine, got %d", len(cats))
	}
}
