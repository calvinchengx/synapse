package rules

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// dirToRuleType maps directory names to RuleType.
var dirToRuleType = map[string]RuleType{
	"rules":    RuleTypeRule,
	"skills":   RuleTypeSkill,
	"commands": RuleTypeCommand,
	"agents":   RuleTypeAgent,
	"contexts": RuleTypeContext,
}

// ScanDir recursively discovers .md files in the given root directory and
// parses their YAML frontmatter. Returns all discovered rules.
func ScanDir(root string) ([]Rule, error) {
	var rules []Rule

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		rule, err := ParseRuleFile(root, path)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}
		rules = append(rules, rule)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scanning %s: %w", root, err)
	}
	return rules, nil
}

// ParseRuleFile reads a .md file, extracts YAML frontmatter, and returns a Rule.
func ParseRuleFile(root, path string) (Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Rule{}, fmt.Errorf("reading file: %w", err)
	}

	relPath, err := filepath.Rel(root, path)
	if err != nil {
		relPath = filepath.Base(path)
	}

	fm, content, err := parseFrontmatter(string(data))
	if err != nil {
		return Rule{}, fmt.Errorf("parsing frontmatter: %w", err)
	}

	hash := fmt.Sprintf("%x", md5.Sum(data))

	ruleType := inferRuleType(relPath)

	return Rule{
		Path:        path,
		RelPath:     relPath,
		Filename:    filepath.Base(path),
		Type:        ruleType,
		Frontmatter: fm,
		Content:     content,
		Hash:        hash,
	}, nil
}

// parseFrontmatter extracts YAML frontmatter delimited by --- lines.
// Returns the parsed frontmatter, the remaining content, and any error.
func parseFrontmatter(text string) (Frontmatter, string, error) {
	var fm Frontmatter

	scanner := bufio.NewScanner(strings.NewReader(text))

	// Check for opening ---
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		// No frontmatter, entire text is content
		return fm, text, nil
	}

	// Collect frontmatter lines until closing ---
	var fmLines []string
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			found = true
			break
		}
		fmLines = append(fmLines, line)
	}
	if !found {
		// No closing ---, treat entire text as content
		return fm, text, nil
	}

	// Parse YAML frontmatter
	fmText := strings.Join(fmLines, "\n")
	if err := yaml.Unmarshal([]byte(fmText), &fm); err != nil {
		return fm, "", fmt.Errorf("invalid YAML: %w", err)
	}

	// Remaining content after frontmatter
	var contentLines []string
	for scanner.Scan() {
		contentLines = append(contentLines, scanner.Text())
	}
	content := strings.Join(contentLines, "\n")

	return fm, strings.TrimSpace(content), nil
}

// inferRuleType determines the RuleType from the relative path.
func inferRuleType(relPath string) RuleType {
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	if len(parts) > 0 {
		if rt, ok := dirToRuleType[parts[0]]; ok {
			return rt
		}
	}
	return RuleTypeRule // default
}
