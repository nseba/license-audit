package config

import (
	"os"
	"path/filepath"
	"testing"

	"license-audit/pkg/types"
)

func TestLoadDefaultConfig(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.OutputFormat != "json" {
		t.Errorf("Expected output format 'json', got '%s'", cfg.OutputFormat)
	}

	if !cfg.EnableAudit {
		t.Error("Expected audit to be enabled by default")
	}

	if !cfg.Scanners.NodeJS {
		t.Error("Expected NodeJS scanner to be enabled by default")
	}
}

func TestLoadConfigFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.toml")
	
	configContent := `
output_format = "markdown"
enable_audit = false

[scanners]
nodejs = false
go = true
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.OutputFormat != "markdown" {
		t.Errorf("Expected output format 'markdown', got '%s'", cfg.OutputFormat)
	}

	if cfg.EnableAudit {
		t.Error("Expected audit to be disabled")
	}

	if cfg.Scanners.NodeJS {
		t.Error("Expected NodeJS scanner to be disabled")
	}

	if !cfg.Scanners.Go {
		t.Error("Expected Go scanner to be enabled")
	}
}

func TestValidateConfig(t *testing.T) {
	cfg := &types.Config{
		OutputFormat: "invalid",
		ScanPaths:    []string{},
	}

	err := validateConfig(cfg)
	if err == nil {
		t.Error("Expected validation error for invalid output format")
	}

	cfg.OutputFormat = "json"
	err = validateConfig(cfg)
	if err != nil {
		t.Errorf("Expected no error after fixing output format, got %v", err)
	}

	if len(cfg.ScanPaths) != 1 || cfg.ScanPaths[0] != "." {
		t.Error("Expected default scan path to be set")
	}
}