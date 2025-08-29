package audit

import (
	"strings"

	"license-audit/pkg/types"
)

type Auditor struct {
	config *types.Config
}

func New(config *types.Config) *Auditor {
	return &Auditor{
		config: config,
	}
}

func (a *Auditor) Audit(dependencies []types.Dependency) []types.AuditIssue {
	var issues []types.AuditIssue

	for _, dep := range dependencies {
		// Check for dangerous licenses
		if a.isDangerousLicense(dep.LicenseType) {
			issue := types.AuditIssue{
				Severity:   "error",
				Type:       "dangerous_license",
				Message:    a.getDangerousLicenseMessage(dep.LicenseType),
				Dependency: dep,
				Suggestion: a.getDangerousLicenseSuggestion(dep.LicenseType),
			}
			issues = append(issues, issue)
		}

		// Check for unclear licenses
		if a.isUnclearLicense(dep.LicenseType) {
			issue := types.AuditIssue{
				Severity:   "warning",
				Type:       "unclear_license",
				Message:    a.getUnclearLicenseMessage(dep.LicenseType),
				Dependency: dep,
				Suggestion: "Review the dependency's repository or documentation to determine the correct license",
			}
			issues = append(issues, issue)
		}

		// Check for potential tainted licenses
		if a.isPotentiallyTaintedLicense(dep) {
			issue := types.AuditIssue{
				Severity:   "warning",
				Type:       "tainted_license",
				Message:    "Dependency may have tainted or conflicting license terms",
				Dependency: dep,
				Suggestion: "Carefully review the license terms and consult with legal counsel if necessary",
			}
			issues = append(issues, issue)
		}

		// Check for missing license information
		if dep.LicenseType == "UNKNOWN" && dep.LicenseText == "" {
			issue := types.AuditIssue{
				Severity:   "warning",
				Type:       "missing_license",
				Message:    "No license information found for this dependency",
				Dependency: dep,
				Suggestion: "Check the dependency's repository, package registry, or documentation for license information",
			}
			issues = append(issues, issue)
		}
	}

	return issues
}

func (a *Auditor) GetIssueBreakdown(issues []types.AuditIssue) map[string]int {
	breakdown := make(map[string]int)
	
	for _, issue := range issues {
		key := issue.Severity + "_" + issue.Type
		breakdown[key]++
		breakdown[issue.Severity]++
		breakdown[issue.Type]++
	}
	
	return breakdown
}

func (a *Auditor) isDangerousLicense(licenseType string) bool {
	if licenseType == "" || licenseType == "UNKNOWN" {
		return false
	}

	licenseType = strings.ToUpper(licenseType)
	
	for _, dangerous := range a.config.DangerousLicenses {
		if strings.ToUpper(dangerous) == licenseType {
			return true
		}
	}
	
	return false
}

func (a *Auditor) isUnclearLicense(licenseType string) bool {
	licenseType = strings.ToUpper(licenseType)
	
	for _, unclear := range a.config.UnclearLicenses {
		if strings.ToUpper(unclear) == licenseType {
			return true
		}
	}
	
	return false
}

func (a *Auditor) isPotentiallyTaintedLicense(dep types.Dependency) bool {
	// Check for common indicators of tainted licenses
	licenseText := strings.ToLower(dep.LicenseText)
	licenseType := strings.ToLower(dep.LicenseType)
	
	// Multiple license mentions that might conflict
	if strings.Contains(licenseText, "dual license") || 
	   strings.Contains(licenseText, "multiple license") ||
	   (strings.Contains(licenseType, "gpl") && strings.Contains(licenseText, "commercial")) {
		return true
	}
	
	// Check for custom or modified licenses
	if strings.Contains(licenseText, "modified") || 
	   strings.Contains(licenseText, "custom") ||
	   strings.Contains(licenseText, "proprietary") {
		return true
	}
	
	return false
}

func (a *Auditor) getDangerousLicenseMessage(licenseType string) string {
	messages := map[string]string{
		"GPL-2.0":    "GPL-2.0 is a copyleft license that may require releasing your source code under the same license",
		"GPL-3.0":    "GPL-3.0 is a copyleft license that may require releasing your source code under the same license",
		"AGPL-3.0":   "AGPL-3.0 has strong copyleft requirements including network use provisions",
		"LGPL-2.1":   "LGPL-2.1 may require releasing modifications to the library under the same license",
		"LGPL-3.0":   "LGPL-3.0 may require releasing modifications to the library under the same license",
		"CDDL-1.0":   "CDDL-1.0 has copyleft requirements that may conflict with proprietary code",
		"CDDL-1.1":   "CDDL-1.1 has copyleft requirements that may conflict with proprietary code",
		"EPL-1.0":    "EPL-1.0 has copyleft requirements for modifications and derivative works",
		"EPL-2.0":    "EPL-2.0 has copyleft requirements for modifications and derivative works",
		"CPL-1.0":    "CPL-1.0 has copyleft requirements that may affect your code",
		"OSL-3.0":    "OSL-3.0 has strong copyleft requirements including network distribution",
		"QPL-1.0":    "QPL-1.0 has specific requirements for commercial use",
	}
	
	if message, exists := messages[licenseType]; exists {
		return message
	}
	
	return "This license may have restrictions that could affect your project"
}

func (a *Auditor) getDangerousLicenseSuggestion(licenseType string) string {
	suggestions := map[string]string{
		"GPL-2.0":    "Consider using MIT, Apache-2.0, or BSD licensed alternatives",
		"GPL-3.0":    "Consider using MIT, Apache-2.0, or BSD licensed alternatives", 
		"AGPL-3.0":   "Consider using MIT, Apache-2.0, or BSD licensed alternatives",
		"LGPL-2.1":   "Ensure you comply with LGPL requirements or find MIT/Apache alternatives",
		"LGPL-3.0":   "Ensure you comply with LGPL requirements or find MIT/Apache alternatives",
		"CDDL-1.0":   "Consider using Apache-2.0 or MIT licensed alternatives",
		"CDDL-1.1":   "Consider using Apache-2.0 or MIT licensed alternatives",
		"EPL-1.0":    "Consider using Apache-2.0 or MIT licensed alternatives",
		"EPL-2.0":    "Consider using Apache-2.0 or MIT licensed alternatives",
		"CPL-1.0":    "Consider using Apache-2.0 or MIT licensed alternatives",
		"OSL-3.0":    "Consider using Apache-2.0 or MIT licensed alternatives",
		"QPL-1.0":    "Review commercial use requirements or find alternatives",
	}
	
	if suggestion, exists := suggestions[licenseType]; exists {
		return suggestion
	}
	
	return "Review the license terms carefully and consult with legal counsel if necessary"
}

func (a *Auditor) getUnclearLicenseMessage(licenseType string) string {
	messages := map[string]string{
		"UNKNOWN":     "No license information could be determined for this dependency",
		"UNLICENSED":  "This dependency is explicitly marked as unlicensed",
		"PROPRIETARY": "This dependency uses a proprietary license",
		"COMMERCIAL":  "This dependency requires a commercial license",
		"":            "No license information found",
	}
	
	if message, exists := messages[licenseType]; exists {
		return message
	}
	
	return "License information is unclear or ambiguous"
}