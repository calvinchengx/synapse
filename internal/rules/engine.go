package rules

import (
	"fmt"
	"strings"
)

// Engine loads and manages rules from configured sources.
type Engine struct {
	rules []Rule
}

// NewEngine creates a new Engine by scanning the given root directories.
func NewEngine(roots ...string) (*Engine, error) {
	var allRules []Rule

	for _, root := range roots {
		rules, err := ScanDir(root)
		if err != nil {
			return nil, fmt.Errorf("scanning %s: %w", root, err)
		}
		allRules = append(allRules, rules...)
	}

	return &Engine{rules: allRules}, nil
}

// NewEngineFromRules creates an Engine from pre-loaded rules.
func NewEngineFromRules(rules []Rule) *Engine {
	return &Engine{rules: rules}
}

// All returns all loaded rules.
func (e *Engine) All() []Rule {
	return e.rules
}

// ByType returns all rules of the given type.
func (e *Engine) ByType(rt RuleType) []Rule {
	var result []Rule
	for _, r := range e.rules {
		if r.Type == rt {
			result = append(result, r)
		}
	}
	return result
}

// Categories groups rules by type and returns them as categories.
func (e *Engine) Categories() []Category {
	typeOrder := []RuleType{
		RuleTypeRule,
		RuleTypeSkill,
		RuleTypeCommand,
		RuleTypeAgent,
		RuleTypeContext,
	}

	var cats []Category
	for _, rt := range typeOrder {
		rules := e.ByType(rt)
		if len(rules) > 0 {
			cats = append(cats, Category{
				Name:  string(rt) + "s",
				Type:  rt,
				Rules: rules,
			})
		}
	}
	return cats
}

// Search returns rules whose filename, description, or keywords match the query.
func (e *Engine) Search(query string) []Rule {
	q := strings.ToLower(query)
	var result []Rule
	for _, r := range e.rules {
		if matchesQuery(r, q) {
			result = append(result, r)
		}
	}
	return result
}

// Get returns a rule by filename. Returns false if not found.
func (e *Engine) Get(filename string) (Rule, bool) {
	for _, r := range e.rules {
		if r.Filename == filename {
			return r, true
		}
	}
	return Rule{}, false
}

// Count returns the total number of loaded rules.
func (e *Engine) Count() int {
	return len(e.rules)
}

func matchesQuery(r Rule, query string) bool {
	if strings.Contains(strings.ToLower(r.Filename), query) {
		return true
	}
	if strings.Contains(strings.ToLower(r.Frontmatter.Description), query) {
		return true
	}
	for _, kw := range r.Frontmatter.Keywords {
		if strings.Contains(strings.ToLower(kw), query) {
			return true
		}
	}
	return false
}
