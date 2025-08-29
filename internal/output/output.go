package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"license-audit/pkg/types"
)

type Formatter interface {
	Write(result *types.ScanResult, outputPath string) error
}

type JSONFormatter struct{}
type MarkdownFormatter struct{}

func New(format string) (Formatter, error) {
	switch format {
	case "json":
		return &JSONFormatter{}, nil
	case "markdown":
		return &MarkdownFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}

func (f *JSONFormatter) Write(result *types.ScanResult, outputPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if outputPath == "" || outputPath == "-" {
		fmt.Print(string(data))
		return nil
	}

	return os.WriteFile(outputPath, data, 0644)
}

func (f *MarkdownFormatter) Write(result *types.ScanResult, outputPath string) error {
	var sb strings.Builder

	// Header
	sb.WriteString("# License Audit Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n", result.Timestamp.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Scan Path:** %s\n\n", result.ScanPath))

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Dependencies:** %d\n", result.Summary.TotalDependencies))
	sb.WriteString(fmt.Sprintf("- **Issues Found:** %d\n\n", len(result.Issues)))

	// License Breakdown
	if len(result.Summary.LicenseBreakdown) > 0 {
		sb.WriteString("### License Distribution\n\n")
		sb.WriteString("| License | Count |\n")
		sb.WriteString("|---------|-------|\n")
		for license, count := range result.Summary.LicenseBreakdown {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", license, count))
		}
		sb.WriteString("\n")
	}

	// Package Type Breakdown
	if len(result.Summary.PackageBreakdown) > 0 {
		sb.WriteString("### Package Types\n\n")
		sb.WriteString("| Package Type | Count |\n")
		sb.WriteString("|--------------|-------|\n")
		for pkgType, count := range result.Summary.PackageBreakdown {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", pkgType, count))
		}
		sb.WriteString("\n")
	}

	// Issues Section
	if len(result.Issues) > 0 {
		sb.WriteString("## Issues\n\n")
		
		// Group issues by severity
		errorIssues := filterIssuesBySeverity(result.Issues, "error")
		warningIssues := filterIssuesBySeverity(result.Issues, "warning")
		
		if len(errorIssues) > 0 {
			sb.WriteString("### 🚨 Errors\n\n")
			f.writeIssues(&sb, errorIssues)
		}
		
		if len(warningIssues) > 0 {
			sb.WriteString("### ⚠️ Warnings\n\n")
			f.writeIssues(&sb, warningIssues)
		}
	}

	// Dependencies Section
	sb.WriteString("## Dependencies\n\n")
	sb.WriteString("| Name | Version | License | Type | File Path |\n")
	sb.WriteString("|------|---------|---------|------|-----------|\n")
	
	for _, dep := range result.Dependencies {
		name := dep.Name
		version := dep.Version
		license := dep.LicenseType
		pkgType := dep.PackageType
		filePath := dep.FilePath
		
		// Escape pipe characters in the data
		name = strings.ReplaceAll(name, "|", "\\|")
		version = strings.ReplaceAll(version, "|", "\\|")
		license = strings.ReplaceAll(license, "|", "\\|")
		
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n", 
			name, version, license, pkgType, filePath))
	}

	content := sb.String()

	if outputPath == "" || outputPath == "-" {
		fmt.Print(content)
		return nil
	}

	return os.WriteFile(outputPath, []byte(content), 0644)
}

func (f *MarkdownFormatter) writeIssues(sb *strings.Builder, issues []types.AuditIssue) {
	for _, issue := range issues {
		sb.WriteString(fmt.Sprintf("#### %s\n\n", issue.Dependency.Name))
		sb.WriteString(fmt.Sprintf("- **Type:** %s\n", cases.Title(language.English).String(strings.ReplaceAll(issue.Type, "_", " "))))
		sb.WriteString(fmt.Sprintf("- **Message:** %s\n", issue.Message))
		sb.WriteString(fmt.Sprintf("- **Package:** %s@%s (%s)\n", 
			issue.Dependency.Name, issue.Dependency.Version, issue.Dependency.PackageType))
		sb.WriteString(fmt.Sprintf("- **File:** %s\n", issue.Dependency.FilePath))
		
		if issue.Suggestion != "" {
			sb.WriteString(fmt.Sprintf("- **Suggestion:** %s\n", issue.Suggestion))
		}
		
		if issue.Dependency.Repository != "" {
			sb.WriteString(fmt.Sprintf("- **Repository:** %s\n", issue.Dependency.Repository))
		}
		
		if issue.Dependency.Homepage != "" {
			sb.WriteString(fmt.Sprintf("- **Homepage:** %s\n", issue.Dependency.Homepage))
		}
		
		sb.WriteString("\n")
	}
}

func filterIssuesBySeverity(issues []types.AuditIssue, severity string) []types.AuditIssue {
	var filtered []types.AuditIssue
	for _, issue := range issues {
		if issue.Severity == severity {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}