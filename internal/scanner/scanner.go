package scanner

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"license-audit/internal/ignore"
	"license-audit/internal/scanner/nodejs"
	"license-audit/internal/scanner/golang"
	"license-audit/internal/scanner/docker"
	"license-audit/internal/scanner/python"
	"license-audit/internal/scanner/ruby"
	"license-audit/internal/scanner/java"
	"license-audit/pkg/types"
)

type Scanner struct {
	config  *types.Config
	matcher *ignore.Matcher
	scanners []PackageScanner
}

type PackageScanner interface {
	Name() string
	Detect(path string) bool
	Scan(path string) ([]types.Dependency, error)
}

func New(config *types.Config) *Scanner {
	s := &Scanner{
		config:   config,
		scanners: []PackageScanner{},
	}

	// Initialize ignore matcher
	matcher, err := ignore.NewMatcher(config.IgnoreFile)
	if err != nil {
		fmt.Printf("Warning: failed to load ignore file: %v\n", err)
		matcher, _ = ignore.NewMatcher("")
	}
	s.matcher = matcher

	// Register scanners based on configuration
	if config.Scanners.NodeJS {
		s.scanners = append(s.scanners, nodejs.NewScanner())
	}
	if config.Scanners.Go {
		s.scanners = append(s.scanners, golang.NewScanner())
	}
	if config.Scanners.Docker {
		s.scanners = append(s.scanners, docker.NewScanner())
	}
	if config.Scanners.Python {
		s.scanners = append(s.scanners, python.NewScanner())
	}
	if config.Scanners.Ruby {
		s.scanners = append(s.scanners, ruby.NewScanner())
	}
	if config.Scanners.Java {
		s.scanners = append(s.scanners, java.NewScanner())
	}

	return s
}

func (s *Scanner) Scan() (*types.ScanResult, error) {
	result := &types.ScanResult{
		Timestamp:    time.Now(),
		Dependencies: []types.Dependency{},
		Issues:       []types.AuditIssue{},
		Summary: types.Summary{
			LicenseBreakdown: make(map[string]int),
			PackageBreakdown: make(map[string]int),
			IssueBreakdown:   make(map[string]int),
		},
	}

	for _, scanPath := range s.config.ScanPaths {
		result.ScanPath = scanPath
		if err := s.scanPath(scanPath, result); err != nil {
			return nil, fmt.Errorf("error scanning path %s: %w", scanPath, err)
		}
	}

	// Apply license overrides
	s.applyLicenseOverrides(result.Dependencies)

	// Calculate summary
	s.calculateSummary(result)

	return result, nil
}

func (s *Scanner) scanPath(scanPath string, result *types.ScanResult) error {
	return filepath.WalkDir(scanPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if path should be ignored
		relPath, _ := filepath.Rel(scanPath, path)
		if s.matcher.ShouldIgnore(filepath.Dir(relPath), d.Name()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories for file-based scanning
		if d.IsDir() {
			return nil
		}

		// Try each scanner
		for _, scanner := range s.scanners {
			if scanner.Detect(path) {
				deps, err := scanner.Scan(path)
				if err != nil {
					fmt.Printf("Warning: %s scanner failed for %s: %v\n", scanner.Name(), path, err)
					continue
				}

				// Set file path for each dependency
				for i := range deps {
					deps[i].FilePath = path
				}

				result.Dependencies = append(result.Dependencies, deps...)
			}
		}

		return nil
	})
}

func (s *Scanner) applyLicenseOverrides(dependencies []types.Dependency) {
	for i := range dependencies {
		if overrideLicense, exists := s.config.LicenseOverrides[dependencies[i].Name]; exists {
			dependencies[i].LicenseType = overrideLicense
		}
	}
}

func (s *Scanner) calculateSummary(result *types.ScanResult) {
	result.Summary.TotalDependencies = len(result.Dependencies)

	for _, dep := range result.Dependencies {
		// License breakdown
		license := dep.LicenseType
		if license == "" {
			license = "UNKNOWN"
		}
		result.Summary.LicenseBreakdown[license]++

		// Package type breakdown
		result.Summary.PackageBreakdown[dep.PackageType]++
	}
}