package rules

import (
	"context"
	"fmt"
	"testing"

	"github.com/calvinchengx/synapse/internal/llm"
)

// mockLLM implements LLMCompleter for testing.
type mockLLM struct {
	response string
	err      error
}

func (m *mockLLM) Complete(ctx context.Context, messages []llm.Message) (string, error) {
	return m.response, m.err
}

func makeRouterTestRules() []Rule {
	return []Rule{
		{Filename: "essential.md", Type: RuleTypeRule, Frontmatter: Frontmatter{
			Description: "Essential coding rules",
		}},
		{Filename: "security.md", Type: RuleTypeRule, Frontmatter: Frontmatter{
			Description: "Security best practices",
			Keywords:    []string{"security", "auth", "xss"},
		}},
		{Filename: "react.md", Type: RuleTypeSkill, Frontmatter: Frontmatter{
			Description: "React best practices",
			Keywords:    []string{"react", "frontend", "component"},
		}},
		{Filename: "testing.md", Type: RuleTypeRule, Frontmatter: Frontmatter{
			Description: "Testing guidelines",
			Keywords:    []string{"test", "unit", "integration"},
		}},
		{Filename: "database.md", Type: RuleTypeSkill, Frontmatter: Frontmatter{
			Description: "Database best practices",
			Keywords:    []string{"sql", "database", "query"},
		}},
		{Filename: "commit.md", Type: RuleTypeCommand, Frontmatter: Frontmatter{
			Description: "Git commit conventions",
			Keywords:    []string{"git", "commit"},
		}},
		{Filename: "python.md", Type: RuleTypeSkill, Frontmatter: Frontmatter{
			Description: "Python best practices",
			Keywords:    []string{"python", "pip"},
		}},
	}
}

func TestRouterTier1Success(t *testing.T) {
	mock := &mockLLM{response: `["react.md", "testing.md"]`}
	router := NewRouter(mock, []string{"essential.md", "security.md"}, 10)
	rules := makeRouterTestRules()

	result := router.Route(context.Background(), "write a react component with tests", rules)

	if result.Method != "semantic" {
		t.Errorf("method = %s, want semantic", result.Method)
	}
	if result.Tier != 1 {
		t.Errorf("tier = %d, want 1", result.Tier)
	}

	// Should include always-active + LLM-selected
	names := ruleNames(result.Selected)
	assertContains(t, names, "essential.md")
	assertContains(t, names, "security.md")
	assertContains(t, names, "react.md")
	assertContains(t, names, "testing.md")
}

func TestRouterTier1Fallback(t *testing.T) {
	// LLM returns error → fall back to Tier 2
	mock := &mockLLM{err: fmt.Errorf("connection refused")}
	router := NewRouter(mock, []string{"essential.md", "security.md"}, 10)
	rules := makeRouterTestRules()

	result := router.Route(context.Background(), "write a react component", rules)

	if result.Method != "keyword" {
		t.Errorf("method = %s, want keyword", result.Method)
	}
	if result.Tier != 2 {
		t.Errorf("tier = %d, want 2", result.Tier)
	}

	names := ruleNames(result.Selected)
	assertContains(t, names, "essential.md")
	assertContains(t, names, "security.md")
	assertContains(t, names, "react.md")
}

func TestRouterTier1MalformedJSON(t *testing.T) {
	// LLM returns garbage → fall back to Tier 2
	mock := &mockLLM{response: "I think you should use react.md and testing.md"}
	router := NewRouter(mock, []string{"essential.md"}, 10)
	rules := makeRouterTestRules()

	result := router.Route(context.Background(), "write react tests", rules)

	if result.Method != "keyword" {
		t.Errorf("method = %s, want keyword (LLM returned non-JSON)", result.Method)
	}
}

func TestRouterTier2Keywords(t *testing.T) {
	router := NewRouter(nil, []string{"essential.md"}, 10)
	rules := makeRouterTestRules()

	tests := []struct {
		prompt   string
		expected []string
	}{
		{
			prompt:   "write a react component",
			expected: []string{"essential.md", "react.md"},
		},
		{
			prompt:   "fix the security vulnerability",
			expected: []string{"essential.md", "security.md"},
		},
		{
			prompt:   "write database migrations",
			expected: []string{"essential.md", "database.md"},
		},
		{
			prompt:   "git commit the changes",
			expected: []string{"essential.md", "commit.md"},
		},
		{
			prompt:   "write python tests",
			expected: []string{"essential.md", "python.md", "testing.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.prompt, func(t *testing.T) {
			result := router.Route(context.Background(), tt.prompt, rules)
			names := ruleNames(result.Selected)
			for _, exp := range tt.expected {
				assertContains(t, names, exp)
			}
		})
	}
}

func TestRouterKoreanKeywords(t *testing.T) {
	router := NewRouter(nil, nil, 10)
	rules := makeRouterTestRules()

	result := router.Route(context.Background(), "리액트 컴포넌트 작성", rules)
	names := ruleNames(result.Selected)
	assertContains(t, names, "react.md")
}

func TestRouterAlwaysActive(t *testing.T) {
	router := NewRouter(nil, []string{"essential.md", "security.md"}, 10)
	rules := makeRouterTestRules()

	// Even a prompt with no keyword matches should include always-active rules
	result := router.Route(context.Background(), "hello world", rules)
	names := ruleNames(result.Selected)
	assertContains(t, names, "essential.md")
	assertContains(t, names, "security.md")
}

func TestRouterNoDuplicates(t *testing.T) {
	// security.md is both always-active and keyword-matched
	router := NewRouter(nil, []string{"security.md"}, 10)
	rules := makeRouterTestRules()

	result := router.Route(context.Background(), "fix the security vulnerability", rules)
	names := ruleNames(result.Selected)

	// Count occurrences of security.md
	count := 0
	for _, n := range names {
		if n == "security.md" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("security.md appears %d times, want 1", count)
	}
}

func TestParseFilenameArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     int
		wantErr  bool
	}{
		{"clean JSON", `["a.md", "b.md"]`, 2, false},
		{"JSON with text", `Here are the files: ["a.md", "b.md"] that match`, 2, false},
		{"no JSON", "I think a.md and b.md", 0, true},
		{"empty array", "[]", 0, false},
		{"single item", `["security.md"]`, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFilenameArray(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != tt.want {
				t.Errorf("got %d filenames, want %d", len(result), tt.want)
			}
		})
	}
}

func TestRouterMaxActive(t *testing.T) {
	router := NewRouter(nil, nil, 2)

	// Create many rules that all match
	var rules []Rule
	for i := 0; i < 10; i++ {
		rules = append(rules, Rule{
			Filename: fmt.Sprintf("rule%d.md", i),
			Type:     RuleTypeRule,
			Frontmatter: Frontmatter{
				Keywords: []string{"test"},
			},
		})
	}

	result := router.Route(context.Background(), "test", rules)
	if len(result.Selected) > 2 {
		t.Errorf("selected %d rules, want max 2", len(result.Selected))
	}
}

// helpers

func ruleNames(rules []Rule) []string {
	names := make([]string, len(rules))
	for i, r := range rules {
		names[i] = r.Filename
	}
	return names
}

func assertContains(t *testing.T, slice []string, item string) {
	t.Helper()
	for _, s := range slice {
		if s == item {
			return
		}
	}
	t.Errorf("expected %v to contain %q", slice, item)
}
