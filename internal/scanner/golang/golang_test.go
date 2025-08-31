package golang

import (
	"path/filepath"
	"testing"
)

func TestDetect(t *testing.T) {
	scanner := NewScanner()

	testCases := []struct {
		path     string
		expected bool
	}{
		{"go.mod", true},
		{"go.sum", true},
		{"vendor/go.mod", true},
		{"package.json", false},
		{"requirements.txt", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := scanner.Detect(tc.path)
		if result != tc.expected {
			t.Errorf("Detect(%s) = %v, expected %v", tc.path, result, tc.expected)
		}
	}
}

func TestParseRequireLine(t *testing.T) {
	scanner := NewScanner()

	testCases := []struct {
		line     string
		expected string // expected name
	}{
		{"github.com/gin-gonic/gin v1.9.1", "github.com/gin-gonic/gin"},
		{"require github.com/stretchr/testify v1.8.4", "github.com/stretchr/testify"},
		{"github.com/bytedance/sonic v1.9.1 // indirect", "github.com/bytedance/sonic"},
		{"", ""},
		{"(", ""},
		{")", ""},
	}

	for _, tc := range testCases {
		result := scanner.parseRequireLine(tc.line, "test.go")
		if result.Name != tc.expected {
			t.Errorf("parseRequireLine(%s).Name = %s, expected %s", tc.line, result.Name, tc.expected)
		}

		if tc.expected != "" && result.PackageType != "go" {
			t.Errorf("Expected package type 'go', got '%s'", result.PackageType)
		}
	}
}

func TestScanGoMod(t *testing.T) {
	scanner := NewScanner()

	// Use the test fixture
	fixturesDir := "../../../test/fixtures"
	goModPath := filepath.Join(fixturesDir, "go.mod")

	dependencies, err := scanner.Scan(goModPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(dependencies) == 0 {
		t.Error("Expected to find dependencies")
	}

	// Check for specific dependencies
	found := make(map[string]bool)
	for _, dep := range dependencies {
		found[dep.Name] = true

		if dep.PackageType != "go" {
			t.Errorf("Expected package type 'go', got '%s'", dep.PackageType)
		}

		if dep.FilePath != goModPath {
			t.Errorf("Expected file path '%s', got '%s'", goModPath, dep.FilePath)
		}
	}

	expectedDeps := []string{"github.com/gin-gonic/gin", "github.com/stretchr/testify"}
	for _, expectedDep := range expectedDeps {
		if !found[expectedDep] {
			t.Errorf("Expected to find dependency '%s'", expectedDep)
		}
	}
}

func TestDetectLicenseType(t *testing.T) {
	scanner := NewScanner()

	testCases := []struct {
		licenseText string
		expected    string
	}{
		{"MIT License\n\nPermission is hereby granted...", "MIT"},
		{"Apache License, Version 2.0", "Apache-2.0"},
		{"BSD 3-Clause License", "BSD-3-Clause"},
		{"GNU General Public License", "GPL-3.0"},
		{"This is unlicense", "Unlicense"},
		{"Some random text", "UNKNOWN"},
		{"", "UNKNOWN"},
	}

	for _, tc := range testCases {
		result := scanner.detectLicenseType(tc.licenseText)
		if result != tc.expected {
			t.Errorf("detectLicenseType(%q) = %s, expected %s", tc.licenseText, result, tc.expected)
		}
	}
}
