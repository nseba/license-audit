package ignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Matcher struct {
	patterns []pattern
}

type pattern struct {
	pattern string
	negate  bool
	isDir   bool
}

func NewMatcher(ignoreFile string) (*Matcher, error) {
	m := &Matcher{
		patterns: []pattern{},
	}

	if ignoreFile == "" {
		ignoreFile = ".licignore"
	}

	if _, err := os.Stat(ignoreFile); os.IsNotExist(err) {
		return m, nil // Return empty matcher if ignore file doesn't exist
	}

	file, err := os.Open(ignoreFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		p := pattern{
			pattern: line,
			negate:  false,
			isDir:   false,
		}

		// Handle negation
		if strings.HasPrefix(line, "!") {
			p.negate = true
			p.pattern = line[1:]
		}

		// Handle directory patterns
		if strings.HasSuffix(p.pattern, "/") {
			p.isDir = true
			p.pattern = strings.TrimSuffix(p.pattern, "/")
		}

		m.patterns = append(m.patterns, p)
	}

	return m, scanner.Err()
}

func (m *Matcher) ShouldIgnore(path, name string) bool {
	matched := false

	for _, p := range m.patterns {
		if m.matchPattern(p, path, name) {
			if p.negate {
				matched = false
			} else {
				matched = true
			}
		}
	}

	return matched
}

func (m *Matcher) matchPattern(p pattern, path, name string) bool {
	target := name

	// If pattern contains '/', match against full path
	if strings.Contains(p.pattern, "/") {
		target = path
	}

	// Simple glob matching
	matched, _ := filepath.Match(p.pattern, target)
	if matched {
		return true
	}

	// Check if it's a directory match
	if p.isDir {
		if info, err := os.Stat(filepath.Join(path, name)); err == nil && info.IsDir() {
			matched, _ := filepath.Match(p.pattern, name)
			return matched
		}
	}

	// Handle wildcard patterns
	if strings.Contains(p.pattern, "*") {
		matched, _ := filepath.Match(p.pattern, target)
		return matched
	}

	// Exact match
	return p.pattern == target
}

// Default patterns that are commonly ignored
func GetDefaultIgnorePatterns() []string {
	return []string{
		"node_modules/",
		".git/",
		".svn/",
		".hg/",
		"vendor/",
		"build/",
		"dist/",
		"target/",
		"bin/",
		"obj/",
		"*.tmp",
		"*.temp",
		".DS_Store",
		"Thumbs.db",
	}
}

func CreateDefaultIgnoreFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	patterns := GetDefaultIgnorePatterns()
	for _, pattern := range patterns {
		if _, err := file.WriteString(pattern + "\n"); err != nil {
			return err
		}
	}

	return nil
}
