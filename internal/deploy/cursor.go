package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/calvinchengx/synapse/internal/rules"
)

// CursorDeployer deploys rules to .cursor/rules/ as .mdc files.
type CursorDeployer struct{}

func (d *CursorDeployer) Name() string { return "cursor" }

// Deploy converts .md rules to .mdc format and writes them to
// <targetDir>/.cursor/rules/.
func (d *CursorDeployer) Deploy(ruleSet []rules.Rule, targetDir string) error {
	rulesDir := filepath.Join(targetDir, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return fmt.Errorf("creating .cursor/rules: %w", err)
	}

	for _, rule := range ruleSet {
		mdcFilename := strings.TrimSuffix(rule.Filename, ".md") + ".mdc"
		dest := filepath.Join(rulesDir, mdcFilename)
		content := convertToMDC(rule)
		if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", mdcFilename, err)
		}
	}

	return nil
}

// convertToMDC converts a Rule to Cursor's .mdc format.
// MDC uses a slightly different frontmatter structure.
func convertToMDC(rule rules.Rule) string {
	fm := rule.Frontmatter

	var front string
	front = "---\n"
	if fm.Description != "" {
		front += fmt.Sprintf("description: %s\n", fm.Description)
	}
	if fm.AlwaysApply {
		front += "alwaysApply: true\n"
	}
	front += "---\n\n"

	return front + rule.Content
}
