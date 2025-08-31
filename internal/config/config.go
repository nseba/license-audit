package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"license-audit/pkg/types"
)

const (
	DefaultConfigFileName = ".license-audit.toml"
	GlobalConfigFileName  = ".license-audit.toml"
)

func Load(configPath string) (*types.Config, error) {
	cfg := getDefaultConfig()

	// Load global config from home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		globalConfigPath := filepath.Join(homeDir, GlobalConfigFileName)
		if _, err := os.Stat(globalConfigPath); err == nil {
			if err := loadConfigFile(globalConfigPath, cfg); err != nil {
				return nil, fmt.Errorf("error loading global config: %w", err)
			}
		}
	}

	// Load project config from current directory
	projectConfigPath := DefaultConfigFileName
	if _, err := os.Stat(projectConfigPath); err == nil {
		if err := loadConfigFile(projectConfigPath, cfg); err != nil {
			return nil, fmt.Errorf("error loading project config: %w", err)
		}
	}

	// Load specified config file (overrides others)
	if configPath != "" {
		if err := loadConfigFile(configPath, cfg); err != nil {
			return nil, fmt.Errorf("error loading specified config file: %w", err)
		}
	}

	// Validate and set defaults
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func loadConfigFile(path string, cfg *types.Config) error {
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return fmt.Errorf("failed to decode TOML file %s: %w", path, err)
	}
	return nil
}

func getDefaultConfig() *types.Config {
	return &types.Config{
		ScanPaths:        []string{"."},
		OutputFormat:     "json",
		OutputFile:       "",
		IgnoreFile:       ".licignore",
		ConfigPaths:      []string{},
		LicenseOverrides: map[string]string{},
		DangerousLicenses: []string{
			"GPL-2.0",
			"GPL-3.0",
			"AGPL-3.0",
			"LGPL-2.1",
			"LGPL-3.0",
			"CDDL-1.0",
			"CDDL-1.1",
			"CPL-1.0",
			"EPL-1.0",
			"EPL-2.0",
			"OSL-3.0",
			"QPL-1.0",
		},
		UnclearLicenses: []string{
			"UNKNOWN",
			"UNLICENSED",
			"PROPRIETARY",
			"COMMERCIAL",
			"",
		},
		EnableAudit: true,
		Scanners: types.ScannerConfig{
			NodeJS: true,
			Go:     true,
			Docker: true,
			Python: true,
			Ruby:   true,
			Java:   true,
		},
	}
}

func validateConfig(cfg *types.Config) error {
	if len(cfg.ScanPaths) == 0 {
		cfg.ScanPaths = []string{"."}
	}

	if cfg.OutputFormat != "json" && cfg.OutputFormat != "markdown" {
		return fmt.Errorf("output format must be 'json' or 'markdown', got '%s'", cfg.OutputFormat)
	}

	// Validate scan paths exist
	for _, path := range cfg.ScanPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("scan path does not exist: %s", path)
		}
	}

	return nil
}

func SaveDefault(path string) error {
	cfg := getDefaultConfig()

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config to TOML: %w", err)
	}

	return nil
}
