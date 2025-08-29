package ruby

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"license-audit/pkg/types"
)

type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Name() string {
	return "ruby"
}

func (s *Scanner) Detect(path string) bool {
	fileName := filepath.Base(path)
	return fileName == "Gemfile" || fileName == "Gemfile.lock" || strings.HasSuffix(fileName, ".gemspec")
}

func (s *Scanner) Scan(path string) ([]types.Dependency, error) {
	fileName := filepath.Base(path)
	
	switch {
	case fileName == "Gemfile":
		return s.scanGemfile(path)
	case fileName == "Gemfile.lock":
		return s.scanGemfileLock(path)
	case strings.HasSuffix(fileName, ".gemspec"):
		return s.scanGemspec(path)
	default:
		return nil, fmt.Errorf("unsupported Ruby file: %s", fileName)
	}
}

func (s *Scanner) scanGemfile(path string) ([]types.Dependency, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Gemfile: %w", err)
	}
	defer file.Close()

	var dependencies []types.Dependency
	scanner := bufio.NewScanner(file)
	
	// Regex to match gem declarations
	gemPattern := regexp.MustCompile(`^\s*gem\s+['"]([^'"]+)['"](?:\s*,\s*['"]([^'"]+)['"])?`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		matches := gemPattern.FindStringSubmatch(line)
		if len(matches) >= 2 {
			name := matches[1]
			version := "UNKNOWN"
			if len(matches) > 2 && matches[2] != "" {
				version = matches[2]
			}

			dep := types.Dependency{
				Name:        name,
				Version:     version,
				LicenseType: "UNKNOWN",
				PackageType: "ruby",
				FilePath:    path,
			}

			dependencies = append(dependencies, dep)
		}
	}

	return dependencies, scanner.Err()
}

func (s *Scanner) scanGemfileLock(path string) ([]types.Dependency, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Gemfile.lock: %w", err)
	}
	defer file.Close()

	var dependencies []types.Dependency
	scanner := bufio.NewScanner(file)
	inSpecsSection := false
	
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		
		if trimmed == "GEM" {
			continue
		}
		
		if strings.HasPrefix(trimmed, "specs:") {
			inSpecsSection = true
			continue
		}
		
		if inSpecsSection {
			if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") {
				// This is a direct dependency line
				parts := strings.Fields(trimmed)
				if len(parts) >= 2 {
					name := parts[0]
					version := strings.Trim(parts[1], "()")
					
					dep := types.Dependency{
						Name:        name,
						Version:     version,
						LicenseType: "UNKNOWN",
						PackageType: "ruby",
						FilePath:    path,
					}
					
					dependencies = append(dependencies, dep)
				}
			}
			
			if !strings.HasPrefix(line, " ") && trimmed != "" {
				// End of specs section
				break
			}
		}
	}

	return dependencies, scanner.Err()
}

func (s *Scanner) scanGemspec(path string) ([]types.Dependency, error) {
	// Basic implementation for .gemspec files
	return []types.Dependency{}, nil
}