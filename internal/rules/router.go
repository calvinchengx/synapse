package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/calvinchengx/synapse/internal/llm"
)

// LLMCompleter is the interface the router uses to call an LLM.
type LLMCompleter interface {
	Complete(ctx context.Context, messages []llm.Message) (string, error)
}

// Router selects relevant rules for a given prompt using a tiered strategy.
type Router struct {
	llm          LLMCompleter
	alwaysActive []string
	maxActive    int
}

// NewRouter creates a Router. llm may be nil (keyword-only mode).
func NewRouter(llmClient LLMCompleter, alwaysActive []string, maxActive int) *Router {
	if maxActive <= 0 {
		maxActive = 10
	}
	return &Router{
		llm:          llmClient,
		alwaysActive: alwaysActive,
		maxActive:    maxActive,
	}
}

// Route selects the most relevant rules for a prompt.
// Tier 1: AI-based selection via LLM (if configured).
// Tier 2: Keyword matching fallback.
// Always-active rules are always included.
func (r *Router) Route(ctx context.Context, prompt string, available []Rule) RouterResult {
	// Start with always-active rules
	alwaysSet := make(map[string]bool)
	var alwaysRules []Rule
	for _, rule := range available {
		for _, name := range r.alwaysActive {
			if rule.Filename == name {
				alwaysSet[rule.Filename] = true
				alwaysRules = append(alwaysRules, rule)
			}
		}
	}

	// Tier 1: Try AI-based selection
	if r.llm != nil {
		selected, err := r.routeTier1(ctx, prompt, available)
		if err == nil && len(selected) > 0 {
			return r.buildResult(alwaysRules, selected, alwaysSet, "semantic", 1)
		}
		// Fall through to Tier 2 on any error
	}

	// Tier 2: Keyword matching
	selected := r.routeTier2(prompt, available)
	return r.buildResult(alwaysRules, selected, alwaysSet, "keyword", 2)
}

// routeTier1 uses the LLM to select relevant rule files.
func (r *Router) routeTier1(ctx context.Context, prompt string, available []Rule) ([]Rule, error) {
	// Build the file metadata for the LLM
	var fileDescs []string
	for _, rule := range available {
		desc := fmt.Sprintf("- %s: %s", rule.Filename, rule.Frontmatter.Description)
		if len(rule.Frontmatter.Keywords) > 0 {
			desc += fmt.Sprintf(" [keywords: %s]", strings.Join(rule.Frontmatter.Keywords, ", "))
		}
		fileDescs = append(fileDescs, desc)
	}

	systemPrompt := `You are a rule selector. Given a user prompt and a list of available rule files with descriptions, select the most relevant rules for the prompt. Return ONLY a JSON array of filenames. Example: ["security.md", "react.md"]`

	userPrompt := fmt.Sprintf("User prompt: %s\n\nAvailable rules:\n%s\n\nSelect the most relevant rules (max %d). Return only a JSON array of filenames.",
		prompt, strings.Join(fileDescs, "\n"), r.maxActive)

	response, err := r.llm.Complete(ctx, []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		return nil, err
	}

	// Parse JSON array from response
	filenames, err := parseFilenameArray(response)
	if err != nil {
		return nil, fmt.Errorf("parsing LLM response: %w", err)
	}

	// Map filenames back to rules
	ruleMap := make(map[string]Rule)
	for _, rule := range available {
		ruleMap[rule.Filename] = rule
	}

	var selected []Rule
	for _, name := range filenames {
		if rule, ok := ruleMap[name]; ok {
			selected = append(selected, rule)
		}
	}

	return selected, nil
}

// parseFilenameArray extracts a JSON array of strings from the LLM response.
// Handles responses that may contain extra text around the JSON.
func parseFilenameArray(response string) ([]string, error) {
	// Try direct parse first
	var filenames []string
	if err := json.Unmarshal([]byte(response), &filenames); err == nil {
		return filenames, nil
	}

	// Try to find JSON array in the response
	start := strings.Index(response, "[")
	end := strings.LastIndex(response, "]")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(response[start:end+1]), &filenames); err == nil {
			return filenames, nil
		}
	}

	return nil, fmt.Errorf("no valid JSON array found in response")
}

// routeTier2 uses keyword matching to select rules.
func (r *Router) routeTier2(prompt string, available []Rule) []Rule {
	promptLower := strings.ToLower(prompt)
	words := strings.Fields(promptLower)

	scores := make(map[string]int)
	ruleMap := make(map[string]Rule)

	for _, rule := range available {
		ruleMap[rule.Filename] = rule
		score := 0

		// Check keyword map
		for _, word := range words {
			if matchedFiles, ok := keywordMap[word]; ok {
				for _, f := range matchedFiles {
					if f == rule.Filename {
						score += 3
					}
				}
			}
		}

		// Check frontmatter keywords
		for _, kw := range rule.Frontmatter.Keywords {
			kwLower := strings.ToLower(kw)
			for _, word := range words {
				if word == kwLower || strings.Contains(kwLower, word) || strings.Contains(word, kwLower) {
					score += 2
				}
			}
		}

		// Check description
		descLower := strings.ToLower(rule.Frontmatter.Description)
		for _, word := range words {
			if len(word) > 2 && strings.Contains(descLower, word) {
				score++
			}
		}

		if score > 0 {
			scores[rule.Filename] = score
		}
	}

	// Sort by score descending, take top maxActive
	type scored struct {
		filename string
		score    int
	}
	var sorted []scored
	for name, score := range scores {
		sorted = append(sorted, scored{name, score})
	}
	// Simple selection sort (small N)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].score > sorted[i].score {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	var selected []Rule
	limit := r.maxActive
	if limit > len(sorted) {
		limit = len(sorted)
	}
	for i := 0; i < limit; i++ {
		if rule, ok := ruleMap[sorted[i].filename]; ok {
			selected = append(selected, rule)
		}
	}

	return selected
}

// buildResult merges always-active rules with selected rules, deduplicating.
func (r *Router) buildResult(always, selected []Rule, alwaysSet map[string]bool, method string, tier int) RouterResult {
	var merged []Rule
	merged = append(merged, always...)
	for _, s := range selected {
		if !alwaysSet[s.Filename] {
			merged = append(merged, s)
		}
	}
	return RouterResult{
		Selected: merged,
		Method:   method,
		Tier:     tier,
	}
}

// keywordMap maps prompt keywords to relevant rule filenames.
// Keyword routing and tier fallbacks for the semantic router.
var keywordMap = map[string][]string{
	// Frontend
	"react":       {"react.md", "frontend.md"},
	"vue":         {"vue.md", "frontend.md"},
	"angular":     {"angular.md", "frontend.md"},
	"svelte":      {"svelte.md", "frontend.md"},
	"nextjs":      {"react.md", "frontend.md"},
	"next.js":     {"react.md", "frontend.md"},
	"frontend":    {"frontend.md"},
	"css":         {"frontend.md"},
	"html":        {"frontend.md"},
	"component":   {"react.md", "frontend.md"},
	"ui":          {"frontend.md"},

	// Languages
	"typescript":  {"typescript.md"},
	"javascript":  {"typescript.md"},
	"python":      {"python.md"},
	"go":          {"go.md"},
	"golang":      {"go.md"},
	"rust":        {"rust.md"},
	"java":        {"java.md"},
	"swift":       {"swift.md"},
	"kotlin":      {"kotlin.md"},
	"ruby":        {"ruby.md"},
	"php":         {"php.md"},
	"c#":          {"csharp.md"},
	"csharp":      {"csharp.md"},

	// Testing
	"test":        {"testing.md"},
	"testing":     {"testing.md"},
	"jest":        {"testing.md"},
	"pytest":      {"testing.md", "python.md"},
	"unittest":    {"testing.md"},
	"spec":        {"testing.md"},
	"coverage":    {"testing.md"},
	"tdd":         {"testing.md"},

	// DevOps / Infrastructure
	"docker":      {"docker.md", "devops.md"},
	"kubernetes":  {"kubernetes.md", "devops.md"},
	"k8s":         {"kubernetes.md", "devops.md"},
	"ci":          {"devops.md"},
	"cd":          {"devops.md"},
	"deploy":      {"devops.md"},
	"terraform":   {"devops.md"},
	"aws":         {"devops.md"},

	// Database
	"database":    {"database.md"},
	"sql":         {"database.md"},
	"postgres":    {"database.md"},
	"mysql":       {"database.md"},
	"mongodb":     {"database.md"},
	"redis":       {"database.md"},
	"migration":   {"database.md"},
	"query":       {"database.md"},

	// Security
	"security":    {"security.md"},
	"auth":        {"security.md"},
	"authentication": {"security.md"},
	"authorization": {"security.md"},
	"xss":         {"security.md"},
	"csrf":        {"security.md"},
	"injection":   {"security.md"},
	"encrypt":     {"security.md"},

	// Git / Version Control
	"commit":      {"commit.md"},
	"git":         {"commit.md"},
	"merge":       {"commit.md"},
	"branch":      {"commit.md"},
	"pr":          {"commit.md"},
	"review":      {"commit.md"},

	// Performance
	"performance": {"performance.md"},
	"optimize":    {"performance.md"},
	"profiling":   {"performance.md"},
	"cache":       {"performance.md"},
	"latency":     {"performance.md"},
	"memory":      {"performance.md"},

	// API
	"api":         {"api.md"},
	"rest":        {"api.md"},
	"graphql":     {"api.md"},
	"grpc":        {"api.md"},
	"endpoint":    {"api.md"},

	// Documentation
	"docs":        {"documentation.md"},
	"documentation": {"documentation.md"},
	"readme":      {"documentation.md"},
	"comment":     {"documentation.md"},

	// Refactoring
	"refactor":    {"refactoring.md"},
	"cleanup":     {"refactoring.md"},
	"simplify":    {"refactoring.md"},
	"rename":      {"refactoring.md"},

	// Debug
	"debug":       {"debugging.md"},
	"error":       {"debugging.md"},
	"bug":         {"debugging.md"},
	"fix":         {"debugging.md"},
	"troubleshoot": {"debugging.md"},

	// Korean keywords
	"리액트":      {"react.md", "frontend.md"},
	"테스트":      {"testing.md"},
	"보안":        {"security.md"},
	"데이터베이스": {"database.md"},
	"커밋":        {"commit.md"},
	"배포":        {"devops.md"},
	"성능":        {"performance.md"},
	"디버그":      {"debugging.md"},
	"리팩토링":    {"refactoring.md"},
	"문서":        {"documentation.md"},
}
