package deploy

import (
	"github.com/calvinchengx/synapse/internal/rules"
)

// Deployer writes rules to a target AI tool's directory.
type Deployer interface {
	// Deploy writes the given rules to the target directory.
	Deploy(ruleSet []rules.Rule, targetDir string) error
	// Name returns the deployer's target name (e.g., "claude", "cursor", "codex").
	Name() string
}

// DeployAll runs all deployers against the given rules and project directory.
func DeployAll(deployers []Deployer, ruleSet []rules.Rule, projectDir string) map[string]error {
	errs := make(map[string]error)
	for _, d := range deployers {
		if err := d.Deploy(ruleSet, projectDir); err != nil {
			errs[d.Name()] = err
		}
	}
	return errs
}
