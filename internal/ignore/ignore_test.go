package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShouldIgnore(t *testing.T) {
	// Create temporary ignore file
	tmpDir := t.TempDir()
	ignoreFile := filepath.Join(tmpDir, ".licignore")

	ignoreContent := `# Test ignore file
node_modules/
*.tmp
test-*
!test-important
`

	err := os.WriteFile(ignoreFile, []byte(ignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create ignore file: %v", err)
	}

	matcher, err := NewMatcher(ignoreFile)
	if err != nil {
		t.Fatalf("Failed to create matcher: %v", err)
	}

	testCases := []struct {
		path     string
		name     string
		expected bool
	}{
		{"./", "node_modules", true},
		{"./", "package.json", false},
		{"./", "test.tmp", true},
		{"./", "test-file", true},
		{"./", "test-important", false}, // negation
		{"./", "important-test", false},
	}

	for _, tc := range testCases {
		result := matcher.ShouldIgnore(tc.path, tc.name)
		if result != tc.expected {
			t.Errorf("ShouldIgnore(%s, %s) = %v, expected %v", tc.path, tc.name, result, tc.expected)
		}
	}
}

func TestEmptyIgnoreFile(t *testing.T) {
	matcher, err := NewMatcher("nonexistent.ignore")
	if err != nil {
		t.Fatalf("Expected no error for nonexistent file, got %v", err)
	}

	// Should not ignore anything
	result := matcher.ShouldIgnore("./", "any-file")
	if result {
		t.Error("Empty matcher should not ignore any files")
	}
}

func TestDefaultIgnorePatterns(t *testing.T) {
	patterns := GetDefaultIgnorePatterns()

	expectedPatterns := []string{"node_modules/", ".git/", "vendor/", "*.tmp"}

	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range patterns {
			if pattern == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected default pattern '%s' not found", expected)
		}
	}
}

func TestCreateDefaultIgnoreFile(t *testing.T) {
	tmpDir := t.TempDir()
	ignoreFile := filepath.Join(tmpDir, ".licignore")

	err := CreateDefaultIgnoreFile(ignoreFile)
	if err != nil {
		t.Fatalf("Failed to create default ignore file: %v", err)
	}

	// Check if file was created
	if _, err := os.Stat(ignoreFile); os.IsNotExist(err) {
		t.Error("Default ignore file was not created")
	}

	// Check content
	content, err := os.ReadFile(ignoreFile)
	if err != nil {
		t.Fatalf("Failed to read ignore file: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "node_modules/") {
		t.Error("Expected 'node_modules/' in default ignore file")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
