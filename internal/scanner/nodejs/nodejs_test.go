package nodejs

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
		{"package.json", true},
		{"package-lock.json", true},
		{"node_modules/package.json", true},
		{"go.mod", false},
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

func TestScanPackageJSON(t *testing.T) {
	scanner := NewScanner()
	
	// Use the test fixture
	fixturesDir := "../../../test/fixtures"
	packageJSONPath := filepath.Join(fixturesDir, "package.json")
	
	dependencies, err := scanner.Scan(packageJSONPath)
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
		
		if dep.PackageType != "npm" {
			t.Errorf("Expected package type 'npm', got '%s'", dep.PackageType)
		}
		
		if dep.FilePath != packageJSONPath {
			t.Errorf("Expected file path '%s', got '%s'", packageJSONPath, dep.FilePath)
		}
	}

	expectedDeps := []string{"express", "lodash", "moment", "jest", "eslint"}
	for _, expectedDep := range expectedDeps {
		if !found[expectedDep] {
			t.Errorf("Expected to find dependency '%s'", expectedDep)
		}
	}
}

func TestParseLicense(t *testing.T) {
	scanner := NewScanner()
	
	testCases := []struct {
		license  interface{}
		expected string
	}{
		{"MIT", "MIT"},
		{"Apache-2.0", "Apache-2.0"},
		{"", "UNKNOWN"},
		{nil, "UNKNOWN"},
		{map[string]interface{}{"type": "BSD-3-Clause"}, "BSD-3-Clause"},
		{[]interface{}{map[string]interface{}{"type": "GPL-3.0"}}, "GPL-3.0"},
	}

	for _, tc := range testCases {
		result := scanner.parseLicense(tc.license)
		if result != tc.expected {
			t.Errorf("parseLicense(%v) = %s, expected %s", tc.license, result, tc.expected)
		}
	}
}