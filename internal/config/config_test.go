package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected default host 127.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.LLM.Provider != "litellm" {
		t.Errorf("expected default LLM provider litellm, got %s", cfg.LLM.Provider)
	}
	if cfg.LLM.Timeout != 10*time.Second {
		t.Errorf("expected default LLM timeout 10s, got %s", cfg.LLM.Timeout)
	}
	if len(cfg.Router.AlwaysActive) != 2 {
		t.Errorf("expected 2 always-active rules, got %d", len(cfg.Router.AlwaysActive))
	}
	if cfg.Integrations.RTK.Enabled != true {
		t.Error("expected RTK integration enabled by default")
	}
	if cfg.Integrations.AgentsView.Enabled != true {
		t.Error("expected AgentsView integration enabled by default")
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	// Should return defaults
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
}

func TestLoadValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	yaml := `
server:
  host: "0.0.0.0"
  port: 9090
llm:
  provider: anthropic
  model: claude-3-haiku-20240307
  timeout: 5s
  max_retries: 3
targets:
  - claude
  - cursor
router:
  always_active:
    - essential.md
  max_active_rules: 5
integrations:
  rtk:
    enabled: false
  agentsview:
    enabled: true
    api_url: "http://localhost:9999"
`
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.LLM.Provider != "anthropic" {
		t.Errorf("expected provider anthropic, got %s", cfg.LLM.Provider)
	}
	if cfg.LLM.Model != "claude-3-haiku-20240307" {
		t.Errorf("expected model claude-3-haiku-20240307, got %s", cfg.LLM.Model)
	}
	if cfg.LLM.Timeout != 5*time.Second {
		t.Errorf("expected timeout 5s, got %s", cfg.LLM.Timeout)
	}
	if cfg.LLM.MaxRetries != 3 {
		t.Errorf("expected max_retries 3, got %d", cfg.LLM.MaxRetries)
	}
	if len(cfg.Targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(cfg.Targets))
	}
	if cfg.Router.MaxActiveRules != 5 {
		t.Errorf("expected max_active_rules 5, got %d", cfg.Router.MaxActiveRules)
	}
	if cfg.Integrations.RTK.Enabled {
		t.Error("expected RTK disabled")
	}
	if cfg.Integrations.AgentsView.APIURL != "http://localhost:9999" {
		t.Errorf("expected agentsview API URL http://localhost:9999, got %s", cfg.Integrations.AgentsView.APIURL)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestEnvOverrides(t *testing.T) {
	t.Setenv("SYNAPSE_HOST", "10.0.0.1")
	t.Setenv("SYNAPSE_PORT", "3000")
	t.Setenv("SYNAPSE_LLM_PROVIDER", "openai")
	t.Setenv("SYNAPSE_LLM_BASE_URL", "http://my-proxy:4000")
	t.Setenv("SYNAPSE_LLM_MODEL", "gpt-4o-mini")

	// Load from nonexistent file to get defaults + env overrides
	cfg, err := Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Host != "10.0.0.1" {
		t.Errorf("expected host 10.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 3000 {
		t.Errorf("expected port 3000, got %d", cfg.Server.Port)
	}
	if cfg.LLM.Provider != "openai" {
		t.Errorf("expected provider openai, got %s", cfg.LLM.Provider)
	}
	if cfg.LLM.BaseURL != "http://my-proxy:4000" {
		t.Errorf("expected base_url http://my-proxy:4000, got %s", cfg.LLM.BaseURL)
	}
	if cfg.LLM.Model != "gpt-4o-mini" {
		t.Errorf("expected model gpt-4o-mini, got %s", cfg.LLM.Model)
	}
}

func TestEnvOverrideInvalidPort(t *testing.T) {
	t.Setenv("SYNAPSE_PORT", "notanumber")

	cfg, err := Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should keep default port when env var is invalid
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080 for invalid env, got %d", cfg.Server.Port)
	}
}
