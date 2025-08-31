package golang

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"license-audit/pkg/types"
)

type Scanner struct{}

type GoMod struct {
	Module  Module    `json:"Module"`
	Require []Require `json:"Require"`
}

type Module struct {
	Path string `json:"Path"`
}

type Require struct {
	Path     string `json:"Path"`
	Version  string `json:"Version"`
	Indirect bool   `json:"Indirect"`
}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Name() string {
	return "golang"
}

func (s *Scanner) Detect(path string) bool {
	fileName := filepath.Base(path)
	return fileName == "go.mod" || fileName == "go.sum"
}

func (s *Scanner) Scan(path string) ([]types.Dependency, error) {
	fileName := filepath.Base(path)

	switch fileName {
	case "go.mod":
		return s.scanGoMod(path)
	case "go.sum":
		return s.scanGoSum(path)
	default:
		return nil, fmt.Errorf("unsupported Go file: %s", fileName)
	}
}

func (s *Scanner) scanGoMod(path string) ([]types.Dependency, error) {
	// First try to use 'go list' command if available
	if deps, err := s.scanWithGoList(path); err == nil {
		return deps, nil
	}

	// Fallback to parsing go.mod directly
	return s.parseGoMod(path)
}

func (s *Scanner) scanWithGoList(path string) ([]types.Dependency, error) {
	dir := filepath.Dir(path)

	// Check if go command is available
	if _, err := exec.LookPath("go"); err != nil {
		return nil, fmt.Errorf("go command not found")
	}

	cmd := exec.Command("go", "list", "-m", "-json", "all")
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run 'go list': %w", err)
	}

	var dependencies []types.Dependency
	decoder := json.NewDecoder(strings.NewReader(string(output)))

	for decoder.More() {
		var mod GoMod
		if err := decoder.Decode(&mod); err != nil {
			continue // Skip invalid entries
		}

		// Skip the main module
		if mod.Module.Path == "" {
			continue
		}

		dep := types.Dependency{
			Name:        mod.Module.Path,
			Version:     "", // Will be set below if available
			LicenseType: "UNKNOWN",
			PackageType: "go",
			FilePath:    path,
		}

		// Try to get license information
		if licenseText, licenseType := s.getLicenseInfo(mod.Module.Path, dir); licenseType != "" {
			dep.LicenseType = licenseType
			dep.LicenseText = licenseText
		}

		dependencies = append(dependencies, dep)
	}

	return dependencies, nil
}

func (s *Scanner) parseGoMod(path string) ([]types.Dependency, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open go.mod: %w", err)
	}
	defer file.Close()

	var dependencies []types.Dependency
	scanner := bufio.NewScanner(file)
	inRequireBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}

		if inRequireBlock && line == ")" {
			inRequireBlock = false
			continue
		}

		if strings.HasPrefix(line, "require ") || inRequireBlock {
			dep := s.parseRequireLine(line, path)
			if dep.Name != "" {
				dependencies = append(dependencies, dep)
			}
		}
	}

	return dependencies, scanner.Err()
}

func (s *Scanner) parseRequireLine(line, filePath string) types.Dependency {
	// Remove "require " prefix if present
	line = strings.TrimPrefix(line, "require ")
	line = strings.TrimSpace(line)

	// Skip invalid lines
	if line == "" || line == "(" || line == ")" {
		return types.Dependency{}
	}

	parts := strings.Fields(line)
	if len(parts) < 2 {
		return types.Dependency{}
	}

	name := parts[0]
	version := parts[1]

	// Remove any trailing comments or indirect markers
	if idx := strings.Index(version, "//"); idx != -1 {
		version = strings.TrimSpace(version[:idx])
	}

	dep := types.Dependency{
		Name:        name,
		Version:     version,
		LicenseType: "UNKNOWN",
		PackageType: "go",
		FilePath:    filePath,
	}

	// Try to get license information
	if licenseText, licenseType := s.getLicenseInfo(name, filepath.Dir(filePath)); licenseType != "" {
		dep.LicenseType = licenseType
		dep.LicenseText = licenseText
	}

	return dep
}

func (s *Scanner) scanGoSum(path string) ([]types.Dependency, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open go.sum: %w", err)
	}
	defer file.Close()

	dependencyMap := make(map[string]types.Dependency)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		name := parts[0]
		version := parts[1]

		// Skip /go.mod entries and focus on actual dependencies
		if strings.HasSuffix(version, "/go.mod") {
			continue
		}

		// Only keep the main version entry, not the hash entries
		if strings.Contains(version, "h1:") {
			continue
		}

		key := name + "@" + version
		if _, exists := dependencyMap[key]; !exists {
			dep := types.Dependency{
				Name:        name,
				Version:     version,
				LicenseType: "UNKNOWN",
				PackageType: "go",
				FilePath:    path,
			}

			// Try to get license information
			if licenseText, licenseType := s.getLicenseInfo(name, filepath.Dir(path)); licenseType != "" {
				dep.LicenseType = licenseType
				dep.LicenseText = licenseText
			}

			dependencyMap[key] = dep
		}
	}

	// Convert map to slice
	var dependencies []types.Dependency
	for _, dep := range dependencyMap {
		dependencies = append(dependencies, dep)
	}

	return dependencies, scanner.Err()
}

func (s *Scanner) getLicenseInfo(modulePath, workDir string) (string, string) {
	// Try to use go mod download to get module info
	if _, err := exec.LookPath("go"); err == nil {
		return s.getLicenseFromGoMod(modulePath, workDir)
	}

	// Try to find in vendor directory
	return s.getLicenseFromVendor(modulePath, workDir)
}

func (s *Scanner) getLicenseFromGoMod(modulePath, workDir string) (string, string) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", modulePath)
	cmd.Dir = workDir

	output, err := cmd.Output()
	if err != nil {
		return "", "UNKNOWN"
	}

	moduleDir := strings.TrimSpace(string(output))
	if moduleDir == "" {
		return "", "UNKNOWN"
	}

	return s.readLicenseFromDir(moduleDir)
}

func (s *Scanner) getLicenseFromVendor(modulePath, workDir string) (string, string) {
	vendorPath := filepath.Join(workDir, "vendor", modulePath)
	return s.readLicenseFromDir(vendorPath)
}

func (s *Scanner) readLicenseFromDir(dir string) (string, string) {
	licenseFiles := []string{
		"LICENSE", "LICENSE.txt", "LICENSE.md", "LICENSE.rst",
		"license", "license.txt", "license.md", "license.rst",
		"COPYING", "COPYING.txt", "COPYRIGHT", "COPYRIGHT.txt",
	}

	for _, fileName := range licenseFiles {
		licensePath := filepath.Join(dir, fileName)
		if data, err := os.ReadFile(licensePath); err == nil {
			licenseText := string(data)
			licenseType := s.detectLicenseType(licenseText)
			return licenseText, licenseType
		}
	}

	return "", "UNKNOWN"
}

func (s *Scanner) detectLicenseType(licenseText string) string {
	text := strings.ToLower(licenseText)

	// Common license patterns
	patterns := map[string]string{
		"mit license":                 "MIT",
		"apache license, version 2.0": "Apache-2.0",
		"apache license version 2.0":  "Apache-2.0",
		"bsd 3-clause":                "BSD-3-Clause",
		"bsd 2-clause":                "BSD-2-Clause",
		"gnu general public license":  "GPL-3.0",
		"gnu lesser general public":   "LGPL-3.0",
		"mozilla public license":      "MPL-2.0",
		"isc license":                 "ISC",
		"unlicense":                   "Unlicense",
	}

	for pattern, license := range patterns {
		if strings.Contains(text, pattern) {
			return license
		}
	}

	return "UNKNOWN"
}
