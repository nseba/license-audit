package audit

import (
	"testing"

	"license-audit/pkg/types"
)

func TestAuditDangerousLicenses(t *testing.T) {
	config := &types.Config{
		DangerousLicenses: []string{"GPL-3.0", "AGPL-3.0"},
		UnclearLicenses:   []string{"UNKNOWN", "PROPRIETARY"},
	}
	
	auditor := New(config)
	
	dependencies := []types.Dependency{
		{Name: "test-gpl", LicenseType: "GPL-3.0", PackageType: "npm"},
		{Name: "test-mit", LicenseType: "MIT", PackageType: "npm"},
		{Name: "test-unknown", LicenseType: "UNKNOWN", PackageType: "npm"},
		{Name: "test-proprietary", LicenseType: "PROPRIETARY", PackageType: "npm"},
	}
	
	issues := auditor.Audit(dependencies)
	
	// Should find 4 issues: 1 dangerous license + 2 unclear licenses + 1 missing license (for UNKNOWN)
	if len(issues) != 4 {
		t.Errorf("Expected 4 issues, got %d", len(issues))
	}
	
	// Check for dangerous license issue
	foundDangerous := false
	foundUnclear := 0
	
	for _, issue := range issues {
		switch issue.Type {
		case "dangerous_license":
			if issue.Dependency.Name == "test-gpl" && issue.Severity == "error" {
				foundDangerous = true
			}
		case "unclear_license":
			if issue.Severity == "warning" {
				foundUnclear++
			}
		}
	}
	
	if !foundDangerous {
		t.Error("Expected to find dangerous license issue for GPL-3.0")
	}
	
	if foundUnclear != 2 {
		t.Errorf("Expected to find 2 unclear license issues, got %d", foundUnclear)
	}
}

func TestIsDangerousLicense(t *testing.T) {
	config := &types.Config{
		DangerousLicenses: []string{"GPL-2.0", "GPL-3.0", "AGPL-3.0"},
	}
	
	auditor := New(config)
	
	testCases := []struct {
		license  string
		expected bool
	}{
		{"GPL-3.0", true},
		{"gpl-3.0", true}, // case insensitive
		{"MIT", false},
		{"Apache-2.0", false},
		{"UNKNOWN", false},
		{"", false},
	}
	
	for _, tc := range testCases {
		result := auditor.isDangerousLicense(tc.license)
		if result != tc.expected {
			t.Errorf("isDangerousLicense(%s) = %v, expected %v", tc.license, result, tc.expected)
		}
	}
}

func TestIsUnclearLicense(t *testing.T) {
	config := &types.Config{
		UnclearLicenses: []string{"UNKNOWN", "PROPRIETARY", "COMMERCIAL"},
	}
	
	auditor := New(config)
	
	testCases := []struct {
		license  string
		expected bool
	}{
		{"UNKNOWN", true},
		{"unknown", true}, // case insensitive
		{"PROPRIETARY", true},
		{"MIT", false},
		{"GPL-3.0", false},
		{"", false},
	}
	
	for _, tc := range testCases {
		result := auditor.isUnclearLicense(tc.license)
		if result != tc.expected {
			t.Errorf("isUnclearLicense(%s) = %v, expected %v", tc.license, result, tc.expected)
		}
	}
}

func TestIsPotentiallyTaintedLicense(t *testing.T) {
	auditor := New(&types.Config{})
	
	testCases := []struct {
		dep      types.Dependency
		expected bool
	}{
		{
			types.Dependency{
				LicenseType: "dual license",
				LicenseText: "This software is dual license",
			},
			true,
		},
		{
			types.Dependency{
				LicenseType: "GPL-3.0",
				LicenseText: "GPL license with commercial exception",
			},
			true,
		},
		{
			types.Dependency{
				LicenseType: "Custom",
				LicenseText: "This is a custom license",
			},
			true,
		},
		{
			types.Dependency{
				LicenseType: "MIT",
				LicenseText: "MIT License standard text",
			},
			false,
		},
	}
	
	for _, tc := range testCases {
		result := auditor.isPotentiallyTaintedLicense(tc.dep)
		if result != tc.expected {
			t.Errorf("isPotentiallyTaintedLicense(%+v) = %v, expected %v", tc.dep, result, tc.expected)
		}
	}
}

func TestGetIssueBreakdown(t *testing.T) {
	auditor := New(&types.Config{})
	
	issues := []types.AuditIssue{
		{Severity: "error", Type: "dangerous_license"},
		{Severity: "error", Type: "dangerous_license"},
		{Severity: "warning", Type: "unclear_license"},
		{Severity: "warning", Type: "tainted_license"},
	}
	
	breakdown := auditor.GetIssueBreakdown(issues)
	
	if breakdown["error"] != 2 {
		t.Errorf("Expected 2 errors, got %d", breakdown["error"])
	}
	
	if breakdown["warning"] != 2 {
		t.Errorf("Expected 2 warnings, got %d", breakdown["warning"])
	}
	
	if breakdown["dangerous_license"] != 2 {
		t.Errorf("Expected 2 dangerous license issues, got %d", breakdown["dangerous_license"])
	}
	
	if breakdown["error_dangerous_license"] != 2 {
		t.Errorf("Expected 2 error dangerous license issues, got %d", breakdown["error_dangerous_license"])
	}
}