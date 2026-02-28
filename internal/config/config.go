package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all Synapse configuration.
type Config struct {
	Sources      []Source          `yaml:"sources"`
	Targets      []string          `yaml:"targets"`
	Router       RouterConfig      `yaml:"router"`
	LLM          LLMConfig         `yaml:"llm"`
	Server       ServerConfig      `yaml:"server"`
	Integrations IntegrationsConfig `yaml:"integrations"`
}

// Source is a rule source (local directory or git repo).
type Source struct {
	Path   string `yaml:"path"`
	URL    string `yaml:"url"`
	Branch string `yaml:"branch"`
}

// RouterConfig controls the semantic router behavior.
type RouterConfig struct {
	AlwaysActive   []string `yaml:"always_active"`
	MaxActiveRules int      `yaml:"max_active_rules"`
}

// LLMConfig controls the LLM backend for Tier 1 routing.
type LLMConfig struct {
	Provider   string        `yaml:"provider"`
	BaseURL    string        `yaml:"base_url"`
	APIKeyEnv  string        `yaml:"api_key_env"`
	Model      string        `yaml:"model"`
	Timeout    time.Duration `yaml:"timeout"`
	MaxRetries int           `yaml:"max_retries"`
}

// ServerConfig controls the HTTP server.
type ServerConfig struct {
	Host           string   `yaml:"host"`
	Port           int      `yaml:"port"`
	AuthToken      string   `yaml:"auth_token"`
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedHosts   []string `yaml:"allowed_hosts"`
}

// IntegrationsConfig holds per-tool integration settings.
type IntegrationsConfig struct {
	RTK        IntegrationToggle `yaml:"rtk"`
	AgentsView AgentsViewConfig  `yaml:"agentsview"`
}

// IntegrationToggle enables or disables an integration.
type IntegrationToggle struct {
	Enabled bool `yaml:"enabled"`
}

// AgentsViewConfig holds AgentsView-specific settings.
type AgentsViewConfig struct {
	Enabled bool   `yaml:"enabled"`
	APIURL  string `yaml:"api_url"`
}

// DefaultConfig returns configuration with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Sources: []Source{
			{Path: "./config"},
		},
		Targets: []string{"claude"},
		Router: RouterConfig{
			AlwaysActive:   []string{"essential.md", "security.md"},
			MaxActiveRules: 10,
		},
		LLM: LLMConfig{
			Provider:   "litellm",
			BaseURL:    "http://localhost:4000",
			Model:      "claude-3-haiku",
			Timeout:    10 * time.Second,
			MaxRetries: 2,
		},
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
		},
		Integrations: IntegrationsConfig{
			RTK:        IntegrationToggle{Enabled: true},
			AgentsView: AgentsViewConfig{Enabled: true, APIURL: "http://localhost:58080"},
		},
	}
}

// Load reads configuration from the given YAML file path, falling back
// to defaults for any missing values. If the file does not exist, defaults
// are returned without error.
func Load(path string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return applyEnvOverrides(cfg), nil
		}
		return cfg, fmt.Errorf("reading config %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config %s: %w", path, err)
	}

	return applyEnvOverrides(cfg), nil
}

// applyEnvOverrides applies environment variable overrides to config.
func applyEnvOverrides(cfg Config) Config {
	if v := os.Getenv("SYNAPSE_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("SYNAPSE_PORT"); v != "" {
		var port int
		if _, err := fmt.Sscanf(v, "%d", &port); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("SYNAPSE_LLM_PROVIDER"); v != "" {
		cfg.LLM.Provider = v
	}
	if v := os.Getenv("SYNAPSE_LLM_BASE_URL"); v != "" {
		cfg.LLM.BaseURL = v
	}
	if v := os.Getenv("SYNAPSE_LLM_MODEL"); v != "" {
		cfg.LLM.Model = v
	}
	return cfg
}

// SynapseDir returns the path to ~/.synapse/.
func SynapseDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, ".synapse"), nil
}

// DefaultConfigPath returns the path to ~/.synapse/config.yaml.
func DefaultConfigPath() (string, error) {
	dir, err := SynapseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}
