package translations

import (
	"os"
	"strings"
	"testing"
)

func TestTranslationHelper(t *testing.T) {
	// Setup: Create a temporary config file and cleanup afterwards
	tmpFile, err := os.CreateTemp("", "github-mcp-server-config.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFileName := tmpFile.Name()
	defer os.Remove(tmpFileName)
	defer tmpFile.Close()

	// Test basic functionality and caching
	t.Run("basic functionality and caching", func(t *testing.T) {
		// Save original environment and restore it after test
		oldEnv := os.Getenv("GITHUB_MCP_TEST_KEY")
		defer os.Setenv("GITHUB_MCP_TEST_KEY", oldEnv)

		// Set environment variable for test
		os.Setenv("GITHUB_MCP_TEST_KEY", "env value")

		helper, cleanup := TranslationHelper()
		defer cleanup()

		// First call should get value from environment
		value1 := helper("TEST_KEY", "default value")
		if value1 != "env value" {
			t.Errorf("Expected environment value 'env value', got %q", value1)
		}

		// Change environment, but value should be cached
		os.Setenv("GITHUB_MCP_TEST_KEY", "new env value")
		value2 := helper("TEST_KEY", "default value")
		if value2 != "env value" {
			t.Errorf("Expected cached value 'env value', got %q", value2)
		}

		// Test with a different key that doesn't have an env var
		defaultKey := helper("DEFAULT_KEY", "just default")
		if defaultKey != "just default" {
			t.Errorf("Expected default value 'just default', got %q", defaultKey)
		}
	})

	// Test cleanup function which dumps to JSON
	t.Run("cleanup function", func(t *testing.T) {
		// Create a test-specific temporary directory for dumping the config file
		tempDir, err := os.MkdirTemp("", "translation-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Change to the temporary directory for the duration of this test
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temporary directory: %v", err)
		}
		defer os.Chdir(originalDir)

		// Set test environment variables
		os.Setenv("GITHUB_MCP_DUMP_TEST", "dump value")
		defer os.Unsetenv("GITHUB_MCP_DUMP_TEST")

		// Create and use the helper
		helper, cleanup := TranslationHelper()
		_ = helper("DUMP_TEST", "default dump")

		// Call the cleanup function which should dump to a file
		cleanup()

		// Check if the file was created
		_, err = os.Stat("github-mcp-server-config.json")
		if err != nil {
			t.Errorf("Expected config file to be created, got error: %v", err)
		}

		// Read and verify contents (basic check)
		content, err := os.ReadFile("github-mcp-server-config.json")
		if err != nil {
			t.Errorf("Failed to read dumped config file: %v", err)
		}

		if string(content) == "" {
			t.Error("Dumped config file is empty")
		}

		// Verify the content contains our key (a simple check)
		if !containsString(string(content), "DUMP_TEST") {
			t.Errorf("Dumped config doesn't contain expected key. Content: %s", content)
		}
	})
}

// Helper function to check if a string contains another string
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
