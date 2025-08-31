package docker

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
	return "docker"
}

func (s *Scanner) Detect(path string) bool {
	fileName := filepath.Base(path)
	return fileName == "Dockerfile" || fileName == "dockerfile" ||
		strings.HasSuffix(fileName, ".dockerfile")
}

func (s *Scanner) Scan(path string) ([]types.Dependency, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Dockerfile: %w", err)
	}
	defer file.Close()

	var dependencies []types.Dependency
	scanner := bufio.NewScanner(file)

	// Regex patterns for different package managers in Docker
	patterns := map[string]*regexp.Regexp{
		"apt": regexp.MustCompile(`(?i)apt-get\s+install.*?(\S+)`),
		"yum": regexp.MustCompile(`(?i)yum\s+install.*?(\S+)`),
		"apk": regexp.MustCompile(`(?i)apk\s+add.*?(\S+)`),
		"npm": regexp.MustCompile(`(?i)npm\s+install\s+(?:.*?)?(\S+)`),
		"pip": regexp.MustCompile(`(?i)pip\s+install.*?(\S+)`),
		"gem": regexp.MustCompile(`(?i)gem\s+install.*?(\S+)`),
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Look for RUN commands
		if strings.HasPrefix(strings.ToUpper(line), "RUN ") {
			lineDeps := s.parseRunCommand(line, patterns, path)
			dependencies = append(dependencies, lineDeps...)
		}

		// Look for FROM commands (base images)
		if strings.HasPrefix(strings.ToUpper(line), "FROM ") {
			if dep := s.parseFromCommand(line, path); dep.Name != "" {
				dependencies = append(dependencies, dep)
			}
		}
	}

	return dependencies, scanner.Err()
}

func (s *Scanner) parseRunCommand(line string, patterns map[string]*regexp.Regexp, filePath string) []types.Dependency {
	var dependencies []types.Dependency

	// Remove RUN prefix
	command := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "RUN"), "run"))

	for pkgManager, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(command, -1)
		for _, match := range matches {
			if len(match) > 1 {
				packageName := strings.TrimSpace(match[1])

				// Clean up package name
				packageName = strings.Trim(packageName, "\"'")
				if packageName == "" || strings.Contains(packageName, " ") {
					continue
				}

				dep := types.Dependency{
					Name:        packageName,
					Version:     "UNKNOWN",
					LicenseType: "UNKNOWN",
					PackageType: pkgManager,
					FilePath:    filePath,
				}

				dependencies = append(dependencies, dep)
			}
		}
	}

	return dependencies
}

func (s *Scanner) parseFromCommand(line string, filePath string) types.Dependency {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return types.Dependency{}
	}

	image := parts[1]

	// Skip scratch and empty images
	if image == "scratch" || image == "" {
		return types.Dependency{}
	}

	name := image
	version := "latest"

	// Parse image:tag format
	if strings.Contains(image, ":") {
		parts := strings.SplitN(image, ":", 2)
		name = parts[0]
		version = parts[1]
	}

	return types.Dependency{
		Name:        name,
		Version:     version,
		LicenseType: "UNKNOWN",
		PackageType: "docker-image",
		FilePath:    filePath,
	}
}
