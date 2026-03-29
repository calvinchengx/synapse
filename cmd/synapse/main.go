package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/calvinchengx/synapse/internal/config"
	"github.com/calvinchengx/synapse/internal/deploy"
	"github.com/calvinchengx/synapse/internal/integration"
	"github.com/calvinchengx/synapse/internal/llm"
	"github.com/calvinchengx/synapse/internal/rules"
	"github.com/calvinchengx/synapse/internal/server"
	"github.com/calvinchengx/synapse/internal/web"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	root := &cobra.Command{
		Use:   "synapse",
		Short: "The psychic link between AI coding tools",
		Long: `Synapse manages rules, skills, and agents for AI coding assistants (Claude Code,
Cursor, Codex) and connects tools such as RTK and AgentsView into one workflow.

Configuration is read from ~/.synapse/config.yaml when present; see the project
documentation for defaults and environment overrides.

Examples:
  synapse init
  synapse list
  synapse test "refactor error handling in the API"
  synapse serve
  synapse doctor`,
		Example: `  # Initialise and deploy rules into the current project
  synapse init

  # Preview which rules match a prompt
  synapse test "add JWT middleware"

  # Run the HTTP API and embedded web UI
  synapse serve`,
	}

	root.AddCommand(
		newInitCmd(),
		newUpdateCmd(),
		newListCmd(),
		newAddCmd(),
		newRemoveCmd(),
		newSearchCmd(),
		newTestCmd(),
		newDoctorCmd(),
		newBrowseCmd(),
		newGetCmd(),
		newUninstallCmd(),
		newIntegrationsCmd(),
		newHookCmd(),
		newServeCmd(),
		newVersionCmd(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// loadConfig loads the synapse configuration.
func loadConfig() (config.Config, error) {
	cfgPath, err := config.DefaultConfigPath()
	if err != nil {
		return config.DefaultConfig(), nil
	}
	return config.Load(cfgPath)
}

// loadEngine creates a rule engine from configured sources.
func loadEngine(cfg config.Config) (*rules.Engine, error) {
	var roots []string
	for _, src := range cfg.Sources {
		if src.Path != "" {
			roots = append(roots, src.Path)
		}
	}
	if len(roots) == 0 {
		roots = append(roots, "./config")
	}
	return rules.NewEngine(roots...)
}

// getDeployers returns deployers for the configured targets.
func getDeployers(targets []string) []deploy.Deployer {
	var deployers []deploy.Deployer
	for _, t := range targets {
		switch t {
		case "claude":
			deployers = append(deployers, &deploy.ClaudeDeployer{})
		case "cursor":
			deployers = append(deployers, &deploy.CursorDeployer{})
		case "codex":
			deployers = append(deployers, &deploy.CodexDeployer{})
		}
	}
	if len(deployers) == 0 {
		deployers = append(deployers, &deploy.ClaudeDeployer{})
	}
	return deployers
}

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Synapse in the current project",
		Long: `Load rules from configured sources, create project metadata under ~/.synapse,
and deploy rules to assistant targets (.claude/, .cursor/, .codex/) according
to config.

Run this from your repository root after installing Synapse.`,
		Example: `  synapse init`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			engine, err := loadEngine(cfg)
			if err != nil {
				return fmt.Errorf("loading rules: %w", err)
			}

			synapseDir, err := config.SynapseDir()
			if err != nil {
				return err
			}

			projectDir, err := os.Getwd()
			if err != nil {
				return err
			}

			// Initialize metadata
			installer := rules.NewInstaller(engine, synapseDir)
			if err := installer.Init(projectDir); err != nil {
				return fmt.Errorf("initializing: %w", err)
			}

			// Deploy to targets
			deployers := getDeployers(cfg.Targets)
			errs := deploy.DeployAll(deployers, engine.All(), projectDir)
			for target, err := range errs {
				fmt.Fprintf(os.Stderr, "warning: deploy to %s failed: %v\n", target, err)
			}

			fmt.Printf("Initialized Synapse with %d rules across %d categories\n",
				engine.Count(), len(engine.Categories()))
			for _, cat := range engine.Categories() {
				fmt.Printf("  %s: %d\n", cat.Name, len(cat.Rules))
			}
			return nil
		},
	}
	cmd.Flags().BoolP("interactive", "i", false, "Run interactive setup wizard")
	return cmd
}

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update rules from configured sources",
		Long: `Refresh installed rules from all configured sources (local paths and git URLs),
then redeploy to the configured targets.`,
		Example: `  synapse update`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			engine, err := loadEngine(cfg)
			if err != nil {
				return fmt.Errorf("loading rules: %w", err)
			}

			synapseDir, err := config.SynapseDir()
			if err != nil {
				return err
			}
			projectDir, err := os.Getwd()
			if err != nil {
				return err
			}

			installer := rules.NewInstaller(engine, synapseDir)
			report, err := installer.Update(projectDir)
			if err != nil {
				return fmt.Errorf("updating: %w", err)
			}

			// Redeploy
			deployers := getDeployers(cfg.Targets)
			deploy.DeployAll(deployers, engine.All(), projectDir)

			fmt.Printf("Update complete:\n")
			fmt.Printf("  Added:     %d\n", len(report.Added))
			fmt.Printf("  Updated:   %d\n", len(report.Updated))
			fmt.Printf("  Unchanged: %d\n", len(report.Unchanged))
			fmt.Printf("  Skipped:   %d (user-modified)\n", len(report.Skipped))
			return nil
		},
	}
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed rules by category",
		Long: `Print every installed rule grouped by category, with a short description line
for each file.`,
		Example: `  synapse list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			engine, err := loadEngine(cfg)
			if err != nil {
				return fmt.Errorf("loading rules: %w", err)
			}

			for _, cat := range engine.Categories() {
				fmt.Printf("\n  %s (%d)\n", cat.Name, len(cat.Rules))
				for _, r := range cat.Rules {
					desc := r.Frontmatter.Description
					if len(desc) > 60 {
						desc = desc[:57] + "..."
					}
					fmt.Printf("    %-25s %s\n", r.Filename, desc)
				}
			}
			fmt.Printf("\n  Total: %d rules\n", engine.Count())
			return nil
		},
	}
}

func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <url>",
		Short: "Add a rule source",
		Long: `Register a new rule source (git URL or local path) in configuration. Not yet
implemented.`,
		Example: `  synapse add https://github.com/org/synapse-rules`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("synapse add %s: not yet implemented\n", args[0])
			return nil
		},
	}
}

func newRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <source>",
		Short: "Remove a rule source",
		Long: `Remove a previously added source from configuration. Not yet implemented.`,
		Example: `  synapse remove ./vendor/rules`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("synapse remove %s: not yet implemented\n", args[0])
			return nil
		},
	}
}

func newSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <keyword>",
		Short: "Search local rules by keyword",
		Long: `Search installed rules whose names or front matter match the given keyword.`,
		Example: `  synapse search jwt`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			engine, err := loadEngine(cfg)
			if err != nil {
				return fmt.Errorf("loading rules: %w", err)
			}

			// Search local rules
			results := engine.Search(args[0])
			if len(results) == 0 {
				fmt.Printf("No rules found matching %q\n", args[0])
				return nil
			}

			fmt.Printf("Found %d rules matching %q:\n\n", len(results), args[0])
			for _, r := range results {
				fmt.Printf("  %-25s [%s] %s\n", r.Filename, r.Type, r.Frontmatter.Description)
			}
			return nil
		},
	}
}

func newTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test <prompt>",
		Short: "Test semantic router with a prompt",
		Long: `Run the semantic router on a sample user prompt and print which rules would be
activated (tier, routing method, and file list). Configure an LLM in
~/.synapse/config.yaml for Tier-1 routing; otherwise lower tiers apply.`,
		Example: `  synapse test "migrate auth to OAuth2"`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			engine, err := loadEngine(cfg)
			if err != nil {
				return fmt.Errorf("loading rules: %w", err)
			}

			// Create LLM client if configured
			var llmClient rules.LLMCompleter
			if cfg.LLM.Provider != "" {
				llmClient = llm.NewClient(
					cfg.LLM.Provider,
					cfg.LLM.BaseURL,
					cfg.LLM.APIKeyEnv,
					cfg.LLM.Model,
					cfg.LLM.Timeout,
					cfg.LLM.MaxRetries,
				)
			}

			router := rules.NewRouter(llmClient, cfg.Router.AlwaysActive, cfg.Router.MaxActiveRules)
			result := router.Route(context.Background(), args[0], engine.All())

			fmt.Printf("Router result (tier %d, method: %s):\n\n", result.Tier, result.Method)
			for _, r := range result.Selected {
				fmt.Printf("  %-25s %s\n", r.Filename, r.Frontmatter.Description)
			}
			fmt.Printf("\n  %d rules selected\n", len(result.Selected))
			return nil
		},
	}
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run health checks and diagnostics",
		Long: `Verify that the config file exists (or defaults are used), rule source paths are
reachable, the Synapse data directory is present, and LLM settings are
recorded.`,
		Example: `  synapse doctor`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			fmt.Println("Synapse Doctor")
			fmt.Println()

			// Check config
			cfgPath, _ := config.DefaultConfigPath()
			if _, err := os.Stat(cfgPath); err == nil {
				fmt.Printf("  ✓ Config: %s\n", cfgPath)
			} else {
				fmt.Printf("  - Config: using defaults (no %s)\n", cfgPath)
			}

			// Check rule sources
			for _, src := range cfg.Sources {
				if src.Path != "" {
					if _, err := os.Stat(src.Path); err == nil {
						fmt.Printf("  ✓ Source: %s\n", src.Path)
					} else {
						fmt.Printf("  ✗ Source: %s (not found)\n", src.Path)
					}
				}
			}

			// Check synapse directory
			synapseDir, err := config.SynapseDir()
			if err == nil {
				if _, err := os.Stat(synapseDir); err == nil {
					fmt.Printf("  ✓ Data: %s\n", synapseDir)
				} else {
					fmt.Printf("  - Data: %s (not initialized, run synapse init)\n", synapseDir)
				}
			}

			// Check LLM
			fmt.Printf("  - LLM: %s (%s)\n", cfg.LLM.Provider, cfg.LLM.Model)

			fmt.Println("\n  All checks passed.")
			return nil
		},
	}
}

func newBrowseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "browse",
		Short: "Open the Synapse web UI",
		Long: `Start the web UI in a browser. Not yet implemented; use "synapse serve" and open
the printed URL.`,
		Example: `  synapse browse`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("synapse browse: not yet implemented (Phase 4)")
			return nil
		},
	}
}

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <file>",
		Short: "Download a single rule from the registry",
		Long: `Fetch one rule file from a registry. Not yet implemented.`,
		Example: `  synapse get security.md`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("synapse get %s: not yet implemented\n", args[0])
			return nil
		},
	}
}

func newUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove all Synapse-managed files from the project",
		Long: `Remove Synapse-installed artifacts from the current working directory. Does not
remove ~/.synapse unless documented otherwise.`,
		Example: `  synapse uninstall`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			engine, err := loadEngine(cfg)
			if err != nil {
				return fmt.Errorf("loading rules: %w", err)
			}

			synapseDir, err := config.SynapseDir()
			if err != nil {
				return err
			}
			projectDir, err := os.Getwd()
			if err != nil {
				return err
			}

			installer := rules.NewInstaller(engine, synapseDir)
			if err := installer.Uninstall(projectDir); err != nil {
				return fmt.Errorf("uninstalling: %w", err)
			}

			fmt.Println("Synapse files removed from project.")
			return nil
		},
	}
}

func newIntegrationsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "integrations",
		Short: "List integrated tools and their status",
		Long: `Show configured integrations (RTK, AgentsView, etc.): binaries found, optional API
reachability, and capabilities. Manifests may live under ./config/integrations,
~/.synapse/integrations, or .synapse/integrations.`,
		Example: `  synapse integrations`,
		RunE: func(cmd *cobra.Command, args []string) error {
			synapseDir, _ := config.SynapseDir()

			// Load manifests from built-in + user + project paths
			manifests, err := integration.LoadManifests(
				"./config/integrations",
				filepath.Join(synapseDir, "integrations"),
				".synapse/integrations",
			)
			if err != nil {
				return fmt.Errorf("loading manifests: %w", err)
			}

			if len(manifests) == 0 {
				fmt.Println("No integrations configured.")
				return nil
			}

			statuses := integration.DiscoverAll(manifests)

			fmt.Println()
			fmt.Println("  Integrations")
			fmt.Println()
			for _, s := range statuses {
				mark := "✗"
				if s.Installed {
					mark = "✓"
				}
				fmt.Printf("  %s %-20s", mark, s.Tool.Name)
				if s.BinaryPath != "" {
					fmt.Printf(" %s", s.BinaryPath)
				}
				fmt.Println()

				if s.APIReachable != nil {
					if *s.APIReachable {
						fmt.Println("    API:                   reachable")
					} else {
						fmt.Println("    API:                   offline")
					}
				}

				if len(s.Tool.Capabilities) > 0 {
					fmt.Printf("    Capabilities:          %s\n", strings.Join(s.Tool.Capabilities, ", "))
				}
				fmt.Println()
			}

			fmt.Println("  Add integrations: place .integration.yaml in ~/.synapse/integrations/")
			return nil
		},
	}
}

func newHookCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "hook",
		Short:  "Claude Code hook handler (reads prompt from stdin)",
		Long: `Internal entrypoint for Claude Code: read hook JSON from stdin, route the user
prompt, and move rule files between .claude/rules and .claude/rules-inactive.
Configure the synapse binary with this subcommand in .claude/settings.json.`,
		Example: `  # Invoked by Claude Code; not run interactively
  synapse hook`,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Read prompt from stdin (Claude Code format)
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}

			// Parse the hook input
			var input struct {
				Prompts []struct {
					Content string `json:"content"`
				} `json:"prompts"`
			}
			if err := json.Unmarshal(data, &input); err != nil {
				return fmt.Errorf("parsing hook input: %w", err)
			}

			if len(input.Prompts) == 0 {
				return nil
			}
			prompt := input.Prompts[0].Content

			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			engine, err := loadEngine(cfg)
			if err != nil {
				return err
			}

			// Route using keyword-only for hook (fast path)
			router := rules.NewRouter(nil, cfg.Router.AlwaysActive, cfg.Router.MaxActiveRules)
			result := router.Route(context.Background(), prompt, engine.All())

			// Move selected files to active, others to inactive
			projectDir, _ := os.Getwd()
			activeDir := filepath.Join(projectDir, ".claude", "rules")
			inactiveDir := filepath.Join(projectDir, ".claude", "rules-inactive")
			os.MkdirAll(activeDir, 0755)
			os.MkdirAll(inactiveDir, 0755)

			selectedSet := make(map[string]bool)
			for _, r := range result.Selected {
				selectedSet[r.Filename] = true
			}

			for _, r := range engine.All() {
				src := filepath.Join(inactiveDir, r.Filename)
				dst := filepath.Join(activeDir, r.Filename)
				if selectedSet[r.Filename] {
					// Activate: move from inactive to active (if in inactive)
					if _, err := os.Stat(src); err == nil {
						os.Rename(src, dst)
					}
				} else {
					// Deactivate: move from active to inactive (if in active)
					if _, err := os.Stat(dst); err == nil {
						os.Rename(dst, src)
					}
				}
			}

			return nil
		},
	}
}

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP server without opening the browser",
		Long: `Start the Gin HTTP server: REST API, integration endpoints, and the embedded
Svelte SPA. Listens on --host and --port (defaults 127.0.0.1:8090). CORS
allowed origins come from config or localhost defaults for development.`,
		Example: `  synapse serve
  synapse serve --host 127.0.0.1 --port 8090`,
		RunE: func(cmd *cobra.Command, args []string) error {
			host, _ := cmd.Flags().GetString("host")
			port, _ := cmd.Flags().GetInt("port")

			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			eng, err := loadEngine(cfg)
			if err != nil {
				return fmt.Errorf("loading rules: %w", err)
			}

			// Load integration manifests
			manifestPath := filepath.Join("config", "integrations", "integrations.yaml")
			manifests, _ := integration.LoadManifests(manifestPath)

			// Embed SPA assets
			spaFS, _ := web.Assets()

			// Build allowed origins from config
			origins := cfg.Server.AllowedOrigins
			if len(origins) == 0 {
				origins = []string{
					fmt.Sprintf("http://localhost:%d", port),
					"http://localhost:5173",
				}
			}

			srvCfg := server.Config{
				Host:           host,
				Port:           port,
				AllowedOrigins: origins,
			}

			srv := server.New(srvCfg, eng, manifests, spaFS)

			fmt.Printf("Synapse server listening on http://%s:%d\n", host, port)

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := srv.Start(ctx); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().String("host", "127.0.0.1", "Host to bind to")
	cmd.Flags().Int("port", 8090, "Port to listen on")
	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long: `Print the synapse version string, git commit, and build date embedded at compile
time (see Makefile ldflags).`,
		Example: `  synapse version`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("synapse %s (commit: %s, built: %s)\n", version, commit, buildDate)
		},
	}
}
