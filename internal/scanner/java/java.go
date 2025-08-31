package java

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"license-audit/pkg/types"
)

type Scanner struct{}

type POM struct {
	XMLName      xml.Name     `xml:"project"`
	Dependencies Dependencies `xml:"dependencies"`
}

type Dependencies struct {
	Dependency []Dependency `xml:"dependency"`
}

type Dependency struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
	Scope      string `xml:"scope"`
}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Name() string {
	return "java"
}

func (s *Scanner) Detect(path string) bool {
	fileName := filepath.Base(path)
	return fileName == "pom.xml" || fileName == "build.gradle" ||
		fileName == "build.gradle.kts"
}

func (s *Scanner) Scan(path string) ([]types.Dependency, error) {
	fileName := filepath.Base(path)

	switch {
	case fileName == "pom.xml":
		return s.scanPOM(path)
	case strings.HasPrefix(fileName, "build.gradle"):
		return s.scanGradle(path)
	default:
		return nil, fmt.Errorf("unsupported Java file: %s", fileName)
	}
}

func (s *Scanner) scanPOM(path string) ([]types.Dependency, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open pom.xml: %w", err)
	}
	defer file.Close()

	var pom POM
	if err := xml.NewDecoder(file).Decode(&pom); err != nil {
		return nil, fmt.Errorf("failed to parse pom.xml: %w", err)
	}

	var dependencies []types.Dependency
	for _, dep := range pom.Dependencies.Dependency {
		name := dep.GroupID + ":" + dep.ArtifactID
		version := dep.Version
		if version == "" {
			version = "UNKNOWN"
		}

		dependency := types.Dependency{
			Name:        name,
			Version:     version,
			LicenseType: "UNKNOWN",
			PackageType: "maven",
			FilePath:    path,
		}

		dependencies = append(dependencies, dependency)
	}

	return dependencies, nil
}

func (s *Scanner) scanGradle(path string) ([]types.Dependency, error) {
	// Basic implementation for Gradle files
	// This could be enhanced to actually parse Gradle build files
	return []types.Dependency{}, nil
}
