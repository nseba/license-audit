package python

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"license-audit/pkg/types"
)

type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Name() string {
	return "python"
}

func (s *Scanner) Detect(path string) bool {
	fileName := filepath.Base(path)
	return fileName == "requirements.txt" || fileName == "setup.py" || 
		   fileName == "pyproject.toml" || fileName == "Pipfile"
}

func (s *Scanner) Scan(path string) ([]types.Dependency, error) {
	fileName := filepath.Base(path)
	
	switch fileName {
	case "requirements.txt":
		return s.scanRequirements(path)
	case "setup.py", "pyproject.toml", "Pipfile":
		return s.scanOtherFormats(path)
	default:
		return nil, fmt.Errorf("unsupported Python file: %s", fileName)
	}
}

func (s *Scanner) scanRequirements(path string) ([]types.Dependency, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open requirements.txt: %w", err)
	}
	defer file.Close()

	var dependencies []types.Dependency
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		dep := s.parseRequirement(line, path)
		if dep.Name != "" {
			dependencies = append(dependencies, dep)
		}
	}

	return dependencies, scanner.Err()
}

func (s *Scanner) parseRequirement(line, filePath string) types.Dependency {
	// Remove options like -e, --editable
	if strings.HasPrefix(line, "-e ") || strings.HasPrefix(line, "--editable ") {
		return types.Dependency{} // Skip editable installs
	}

	// Parse package==version format
	name := line
	version := "UNKNOWN"
	
	for _, sep := range []string{"==", ">=", "<=", "~=", ">", "<"} {
		if strings.Contains(line, sep) {
			parts := strings.SplitN(line, sep, 2)
			name = strings.TrimSpace(parts[0])
			if len(parts) > 1 {
				version = strings.TrimSpace(parts[1])
			}
			break
		}
	}

	// Clean package name
	name = strings.Trim(name, "\"'")
	if name == "" {
		return types.Dependency{}
	}

	return types.Dependency{
		Name:        name,
		Version:     version,
		LicenseType: "UNKNOWN",
		PackageType: "python",
		FilePath:    filePath,
	}
}

func (s *Scanner) scanOtherFormats(path string) ([]types.Dependency, error) {
	// Basic implementation - just return empty for now
	// Could be enhanced to parse setup.py, pyproject.toml, etc.
	return []types.Dependency{}, nil
}