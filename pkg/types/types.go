package types

import "time"

type Dependency struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	LicenseType string `json:"license_type"`
	LicenseText string `json:"license_text,omitempty"`
	Repository  string `json:"repository,omitempty"`
	Homepage    string `json:"homepage,omitempty"`
	LicenseURL  string `json:"license_url,omitempty"`
	PackageType string `json:"package_type"` // npm, go, docker, etc.
	FilePath    string `json:"file_path"`    // where this dependency was found
}

type AuditIssue struct {
	Severity   string     `json:"severity"` // warning, error, info
	Type       string     `json:"type"`     // dangerous_license, unclear_license, tainted_license
	Message    string     `json:"message"`
	Dependency Dependency `json:"dependency"`
	Suggestion string     `json:"suggestion,omitempty"`
}

type ScanResult struct {
	Timestamp    time.Time    `json:"timestamp"`
	ScanPath     string       `json:"scan_path"`
	Dependencies []Dependency `json:"dependencies"`
	Issues       []AuditIssue `json:"issues"`
	Summary      Summary      `json:"summary"`
}

type Summary struct {
	TotalDependencies int            `json:"total_dependencies"`
	LicenseBreakdown  map[string]int `json:"license_breakdown"`
	PackageBreakdown  map[string]int `json:"package_breakdown"`
	IssueBreakdown    map[string]int `json:"issue_breakdown"`
}

type Config struct {
	ScanPaths         []string          `toml:"scan_paths"`
	OutputFormat      string            `toml:"output_format"` // json, markdown
	OutputFile        string            `toml:"output_file"`
	IgnoreFile        string            `toml:"ignore_file"`       // default: .licignore
	ConfigPaths       []string          `toml:"config_paths"`      // additional config file paths
	LicenseOverrides  map[string]string `toml:"license_overrides"` // package_name -> license
	DangerousLicenses []string          `toml:"dangerous_licenses"`
	UnclearLicenses   []string          `toml:"unclear_licenses"`
	EnableAudit       bool              `toml:"enable_audit"`
	Scanners          ScannerConfig     `toml:"scanners"`
}

type ScannerConfig struct {
	NodeJS bool `toml:"nodejs"`
	Go     bool `toml:"go"`
	Docker bool `toml:"docker"`
	Python bool `toml:"python"`
	Ruby   bool `toml:"ruby"`
	Java   bool `toml:"java"`
}
