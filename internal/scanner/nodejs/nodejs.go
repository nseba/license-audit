package nodejs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"license-audit/pkg/types"
)

type Scanner struct{}

type PackageJSON struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	License      interface{}            `json:"license"`
	Licenses     []LicenseInfo          `json:"licenses"`
	Repository   interface{}            `json:"repository"`
	Homepage     string                 `json:"homepage"`
	Dependencies map[string]string      `json:"dependencies"`
	DevDependencies map[string]string   `json:"devDependencies"`
}

type PackageLockJSON struct {
	Name         string                        `json:"name"`
	Version      string                        `json:"version"`
	Dependencies map[string]PackageLockEntry   `json:"dependencies"`
	Packages     map[string]PackageLockPackage `json:"packages"`
}

type PackageLockEntry struct {
	Version  string `json:"version"`
	Resolved string `json:"resolved"`
}

type PackageLockPackage struct {
	Name     string      `json:"name"`
	Version  string      `json:"version"`
	License  interface{} `json:"license"`
	Resolved string      `json:"resolved"`
}

type LicenseInfo struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Name() string {
	return "nodejs"
}

func (s *Scanner) Detect(path string) bool {
	fileName := filepath.Base(path)
	return fileName == "package.json" || fileName == "package-lock.json"
}

func (s *Scanner) Scan(path string) ([]types.Dependency, error) {
	fileName := filepath.Base(path)
	
	switch fileName {
	case "package.json":
		return s.scanPackageJSON(path)
	case "package-lock.json":
		return s.scanPackageLock(path)
	default:
		return nil, fmt.Errorf("unsupported Node.js file: %s", fileName)
	}
}

func (s *Scanner) scanPackageJSON(path string) ([]types.Dependency, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open package.json: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}

	var dependencies []types.Dependency

	// Process main dependencies
	for name, version := range pkg.Dependencies {
		dep := s.createDependency(name, version, path)
		dependencies = append(dependencies, dep)
	}

	// Process dev dependencies
	for name, version := range pkg.DevDependencies {
		dep := s.createDependency(name, version, path)
		dependencies = append(dependencies, dep)
	}

	return dependencies, nil
}

func (s *Scanner) scanPackageLock(path string) ([]types.Dependency, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open package-lock.json: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read package-lock.json: %w", err)
	}

	var lockFile PackageLockJSON
	if err := json.Unmarshal(data, &lockFile); err != nil {
		return nil, fmt.Errorf("failed to parse package-lock.json: %w", err)
	}

	var dependencies []types.Dependency

	// Handle npm v7+ format (packages)
	if len(lockFile.Packages) > 0 {
		for pkgPath, pkg := range lockFile.Packages {
			// Skip root package (empty string key)
			if pkgPath == "" {
				continue
			}
			
			name := pkg.Name
			if name == "" {
				// Extract name from path for scoped packages
				parts := strings.Split(pkgPath, "node_modules/")
				if len(parts) > 1 {
					name = parts[len(parts)-1]
				}
			}

			dep := types.Dependency{
				Name:        name,
				Version:     pkg.Version,
				LicenseType: s.parseLicense(pkg.License),
				PackageType: "npm",
				FilePath:    path,
			}

			dependencies = append(dependencies, dep)
		}
	} else {
		// Handle older npm format (dependencies)
		for name, entry := range lockFile.Dependencies {
			dep := types.Dependency{
				Name:        name,
				Version:     entry.Version,
				LicenseType: "UNKNOWN", // License info not available in old format
				PackageType: "npm",
				FilePath:    path,
			}

			dependencies = append(dependencies, dep)
		}
	}

	return dependencies, nil
}

func (s *Scanner) createDependency(name, version, filePath string) types.Dependency {
	dep := types.Dependency{
		Name:        name,
		Version:     version,
		LicenseType: "UNKNOWN",
		PackageType: "npm",
		FilePath:    filePath,
	}

	// Try to read license from node_modules if available
	nodeModulesPath := filepath.Join(filepath.Dir(filePath), "node_modules", name)
	if licenseText, licenseType := s.readLicenseFromNodeModules(nodeModulesPath); licenseType != "" {
		dep.LicenseType = licenseType
		dep.LicenseText = licenseText
	}

	return dep
}

func (s *Scanner) readLicenseFromNodeModules(pkgPath string) (string, string) {
	// Try to read package.json from node_modules
	pkgJSONPath := filepath.Join(pkgPath, "package.json")
	if file, err := os.Open(pkgJSONPath); err == nil {
		defer file.Close()
		
		data, err := io.ReadAll(file)
		if err == nil {
			var pkg PackageJSON
			if err := json.Unmarshal(data, &pkg); err == nil {
				licenseType := s.parseLicense(pkg.License)
				if licenseType != "UNKNOWN" {
					// Try to read license file
					licenseText := s.readLicenseFile(pkgPath)
					return licenseText, licenseType
				}
			}
		}
	}

	return "", "UNKNOWN"
}

func (s *Scanner) readLicenseFile(pkgPath string) string {
	licenseFiles := []string{"LICENSE", "LICENSE.txt", "LICENSE.md", "license", "license.txt", "license.md"}
	
	for _, fileName := range licenseFiles {
		licensePath := filepath.Join(pkgPath, fileName)
		if data, err := os.ReadFile(licensePath); err == nil {
			return string(data)
		}
	}

	return ""
}

func (s *Scanner) parseLicense(license interface{}) string {
	if license == nil {
		return "UNKNOWN"
	}

	switch v := license.(type) {
	case string:
		if v == "" {
			return "UNKNOWN"
		}
		return v
	case map[string]interface{}:
		if licType, ok := v["type"].(string); ok {
			return licType
		}
	case []interface{}:
		if len(v) > 0 {
			if firstLicense, ok := v[0].(map[string]interface{}); ok {
				if licType, ok := firstLicense["type"].(string); ok {
					return licType
				}
			}
		}
	}

	return "UNKNOWN"
}