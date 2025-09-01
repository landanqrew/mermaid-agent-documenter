package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFileContentsTool_ValidatePath(t *testing.T) {
	tool := &WriteFileContentsTool{}

	// Get home directory for testing
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	// Create a temporary project directory for testing
	tempProjectDir := filepath.Join(homeDir, "mermaid-agent-documenter", "test-project")
	err = os.MkdirAll(tempProjectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp project directory: %v", err)
	}
	defer os.RemoveAll(tempProjectDir)

	// Create a temporary config file with our test project
	configDir := filepath.Join(homeDir, "mermaid-agent-documenter")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	configContent := `{"currentProject": {"name": "test-project", "rootDir": "` + strings.ReplaceAll(tempProjectDir, `\`, `\\`) + `"}}`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	defer os.Remove(configPath)

	tests := []struct {
		name        string
		path        string
		expectError bool
		description string
	}{
		{
			name:        "valid_mermaid_config_path",
			path:        filepath.Join(homeDir, "mermaid-agent-documenter", "config.json"),
			expectError: false,
			description: "Should allow access to files in ~/mermaid-agent-documenter/",
		},
		{
			name:        "valid_project_subdirectory",
			path:        filepath.Join(tempProjectDir, "out", "test.md"),
			expectError: false,
			description: "Should allow access to files in current project directory",
		},
		{
			name:        "valid_project_root",
			path:        filepath.Join(tempProjectDir, "test.md"),
			expectError: false,
			description: "Should allow access to files directly in project root",
		},
		{
			name:        "invalid_system_path",
			path:        "/etc/passwd",
			expectError: true,
			description: "Should reject access to system files",
		},
		{
			name:        "invalid_home_subdirectory",
			path:        filepath.Join(homeDir, "Documents", "test.md"),
			expectError: true,
			description: "Should reject access to other home subdirectories",
		},
		{
			name:        "invalid_parent_directory",
			path:        filepath.Join(homeDir, "..", "test.md"),
			expectError: true,
			description: "Should reject access to parent directories",
		},
		{
			name:        "invalid_absolute_path",
			path:        "/tmp/test.md",
			expectError: true,
			description: "Should reject access to /tmp directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tool.validatePath(tt.path)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s (%s), but got none", tt.path, tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for %s (%s), but got: %v", tt.path, tt.description, err)
			}
		})
	}
}

func TestWriteFileContentsTool_Execute_ValidPath(t *testing.T) {
	tool := &WriteFileContentsTool{}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	// Create a test file in the allowed directory
	testFile := filepath.Join(homeDir, "mermaid-agent-documenter", "test_write.md")

	// Clean up any existing file
	os.Remove(testFile)
	defer os.Remove(testFile)

	args := map[string]interface{}{
		"path":      testFile,
		"content":   "test content for unit test",
		"overwrite": "allow",
	}

	result := tool.Execute(args)

	if result.Success != true {
		t.Errorf("Expected successful execution, but got error: %s", result.Error)
	}

	// Verify the file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("Expected file to be created at %s, but it doesn't exist", testFile)
	}

	// Verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Failed to read created file: %v", err)
	}
	if string(content) != "test content for unit test" {
		t.Errorf("Expected content 'test content for unit test', got '%s'", string(content))
	}
}

func TestWriteFileContentsTool_Execute_InvalidPath(t *testing.T) {
	tool := &WriteFileContentsTool{}

	args := map[string]interface{}{
		"path":      "/etc/test_write_invalid.md",
		"content":   "this should not be written",
		"overwrite": "allow",
	}

	result := tool.Execute(args)

	if result.Success != false {
		t.Errorf("Expected execution to fail for invalid path, but it succeeded")
	}

	if !strings.Contains(result.Error, "outside allowed directories") {
		t.Errorf("Expected error about path being outside allowed directories, got: %s", result.Error)
	}

	// Verify the file was NOT created
	if _, err := os.Stat("/etc/test_write_invalid.md"); !os.IsNotExist(err) {
		t.Errorf("Expected file to NOT be created at /etc/test_write_invalid.md, but it exists")
		os.Remove("/etc/test_write_invalid.md") // Clean up if it was created
	}
}

func TestWriteFileContentsTool_Execute_MissingPath(t *testing.T) {
	tool := &WriteFileContentsTool{}

	args := map[string]interface{}{
		"content": "test content",
	}

	result := tool.Execute(args)

	if result.Success != false {
		t.Errorf("Expected execution to fail with missing path, but it succeeded")
	}

	if !strings.Contains(result.Error, "Missing or invalid 'path' argument") {
		t.Errorf("Expected error about missing path argument, got: %s", result.Error)
	}
}

func TestWriteFileContentsTool_Execute_MissingContent(t *testing.T) {
	tool := &WriteFileContentsTool{}

	homeDir, _ := os.UserHomeDir()
	testFile := filepath.Join(homeDir, "mermaid-agent-documenter", "test_missing_content.md")

	args := map[string]interface{}{
		"path": testFile,
	}

	result := tool.Execute(args)

	if result.Success != false {
		t.Errorf("Expected execution to fail with missing content, but it succeeded")
	}

	if !strings.Contains(result.Error, "Missing or invalid 'content' argument") {
		t.Errorf("Expected error about missing content argument, got: %s", result.Error)
	}
}
