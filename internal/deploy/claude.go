package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/calvinchengx/synapse/internal/rules"
)

// ClaudeDeployer deploys rules to .claude/ directory.
type ClaudeDeployer struct{}

func (d *ClaudeDeployer) Name() string { return "claude" }

// Deploy copies .md rule files to <targetDir>/.claude/rules/ and generates
// a settings.json hook configuration pointing to `synapse hook`.
func (d *ClaudeDeployer) Deploy(ruleSet []rules.Rule, targetDir string) error {
	rulesDir := filepath.Join(targetDir, ".claude", "rules")
	inactiveDir := filepath.Join(targetDir, ".claude", "rules-inactive")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return fmt.Errorf("creating .claude/rules: %w", err)
	}
	if err := os.MkdirAll(inactiveDir, 0755); err != nil {
		return fmt.Errorf("creating .claude/rules-inactive: %w", err)
	}

	// Copy all rules to the rules directory
	for _, rule := range ruleSet {
		dest := filepath.Join(rulesDir, rule.Filename)
		content := buildRuleContent(rule)
		if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", rule.Filename, err)
		}
	}

	// Write settings.json with hook configuration
	if err := d.writeSettings(targetDir); err != nil {
		return fmt.Errorf("writing settings.json: %w", err)
	}

	return nil
}

// writeSettings creates .claude/settings.json with the synapse hook.
func (d *ClaudeDeployer) writeSettings(targetDir string) error {
	settingsPath := filepath.Join(targetDir, ".claude", "settings.json")

	settings := map[string]interface{}{
		"hooks": map[string]interface{}{
			"UserPromptSubmit": []map[string]interface{}{
				{
					"hooks": []map[string]interface{}{
						{
							"type":    "command",
							"command": "synapse hook",
						},
					},
					"timeout": 120,
				},
			},
		},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsPath, data, 0644)
}

// buildRuleContent reconstructs the .md file with frontmatter.
func buildRuleContent(rule rules.Rule) string {
	// If the rule has frontmatter, reconstruct the full file
	fm := rule.Frontmatter
	if fm.Description == "" && fm.Name == "" && len(fm.Keywords) == 0 {
		return rule.Content
	}

	var front string
	front = "---\n"
	if fm.Description != "" {
		front += fmt.Sprintf("description: %s\n", fm.Description)
	}
	if fm.Name != "" {
		front += fmt.Sprintf("name: %s\n", fm.Name)
	}
	if len(fm.Keywords) > 0 {
		front += "keywords:\n"
		for _, kw := range fm.Keywords {
			front += fmt.Sprintf("  - %s\n", kw)
		}
	}
	if fm.Tools != "" {
		front += fmt.Sprintf("tools: %s\n", fm.Tools)
	}
	if fm.Model != "" {
		front += fmt.Sprintf("model: %s\n", fm.Model)
	}
	if fm.AlwaysApply {
		front += "alwaysApply: true\n"
	}
	front += "---\n\n"

	return front + rule.Content
}
