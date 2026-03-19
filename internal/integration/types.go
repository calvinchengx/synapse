package integration

import "time"

// ToolManifest describes a third-party tool that Synapse can integrate with.
type ToolManifest struct {
	ID           string         `yaml:"id"`
	Name         string         `yaml:"name"`
	Description  string         `yaml:"description"`
	Homepage     string         `yaml:"homepage"`
	Detection    Detection      `yaml:"detection"`
	Capabilities []string       `yaml:"capabilities"`
	DataEndpoints []DataEndpoint `yaml:"dataEndpoints"`
}

// Detection describes how to find a tool on the system.
type Detection struct {
	Binary    string            `yaml:"binary"`
	DataFiles []DataFilePath    `yaml:"dataFiles"`
	Ports     []int             `yaml:"ports"`
}

// DataFilePath holds platform-specific paths for a tool's data files.
// It can be a simple string or a map of platform → path.
type DataFilePath struct {
	Linux   string `yaml:"linux"`
	Darwin  string `yaml:"darwin"`
	Windows string `yaml:"windows"`
	// Path is used when the path is the same across all platforms.
	Path string `yaml:"path"`
}

// ResolvedPath returns the data file path for the current platform.
func (d DataFilePath) ResolvedPath(platform string) string {
	switch platform {
	case "linux":
		if d.Linux != "" {
			return d.Linux
		}
	case "darwin":
		if d.Darwin != "" {
			return d.Darwin
		}
	case "windows":
		if d.Windows != "" {
			return d.Windows
		}
	}
	return d.Path
}

// DataEndpoint describes how to access a tool's data.
type DataEndpoint struct {
	ID      string `yaml:"id"`
	Type    string `yaml:"type"`    // "sqlite" or "http"
	Path    string `yaml:"path"`    // for sqlite
	BaseURL string `yaml:"baseUrl"` // for http
}

// ToolStatus holds the discovery result for a single tool.
type ToolStatus struct {
	Tool             ToolManifest
	Installed        bool
	BinaryFound      bool
	BinaryPath       string
	DataFilesFound   []string
	DataFilesMissing []string
	APIReachable     *bool
	Version          string
	Error            error
}

// ManifestFile is the top-level YAML structure for integration manifests.
type ManifestFile struct {
	Tools []ToolManifest `yaml:"tools"`
}

// TokenSavingsReport holds aggregated RTK token savings data.
type TokenSavingsReport struct {
	TotalTokensIn  int64
	TotalTokensOut int64
	SavingsPercent float64
	CommandCount   int
	TopCommands    []CommandSavings
	Period         TimePeriod
	DataStale      bool
}

// CommandSavings holds per-command token savings.
type CommandSavings struct {
	Command        string
	Count          int
	TokensSaved    int64
	AvgSavingsPct  float64
}

// TimePeriod defines a time range for queries.
type TimePeriod struct {
	From time.Time
	To   time.Time
	Days int
}

// SessionReport holds aggregated AgentsView session data.
type SessionReport struct {
	TotalSessions  int
	RecentSessions []SessionSummary
	ToolUsage      []ToolUsageEntry
	RecentProjects []string
	DataStale      bool
}

// SessionSummary holds a brief overview of an AgentsView session.
type SessionSummary struct {
	ID           string
	Project      string
	Agent        string
	StartedAt    time.Time
	EndedAt      *time.Time
	MessageCount int
}

// ToolUsageEntry holds tool usage frequency data.
type ToolUsageEntry struct {
	ToolName string
	Category string
	Count    int
	Percent  float64
}

// ActivityEntry is a single line in the JSONL activity log.
type ActivityEntry struct {
	Timestamp time.Time `json:"ts"`
	Event     string    `json:"event"`
	Rules     []string  `json:"rules,omitempty"`
	Method    string    `json:"method,omitempty"`
	Project   string    `json:"project,omitempty"`
}
