package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFileContentsTool_ValidatePath(t *testing.T) {
	tool := &ReadFileContentsTool{}

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
			description: "Should allow reading files in ~/mermaid-agent-documenter/",
		},
		{
			name:        "valid_project_subdirectory",
			path:        filepath.Join(tempProjectDir, "transcripts", "test.txt"),
			expectError: false,
			description: "Should allow reading files in current project directory",
		},
		{
			name:        "valid_project_root",
			path:        filepath.Join(tempProjectDir, "README.md"),
			expectError: false,
			description: "Should allow reading files directly in project root",
		},
		{
			name:        "invalid_system_path",
			path:        "/etc/passwd",
			expectError: true,
			description: "Should reject reading system files",
		},
		{
			name:        "invalid_home_subdirectory",
			path:        filepath.Join(homeDir, "Documents", "secret.txt"),
			expectError: true,
			description: "Should reject reading other home subdirectories",
		},
		{
			name:        "invalid_parent_directory",
			path:        filepath.Join(homeDir, "..", "sensitive.txt"),
			expectError: true,
			description: "Should reject reading parent directories",
		},
		{
			name:        "invalid_absolute_path",
			path:        "/tmp/secret.txt",
			expectError: true,
			description: "Should reject reading /tmp directory",
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

func TestReadFileContentsTool_Execute_ValidFile(t *testing.T) {
	tool := &ReadFileContentsTool{}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	// Create a test file in the allowed directory
	testFile := filepath.Join(homeDir, "mermaid-agent-documenter", "test_read.md")
	testContent := "This is test content for reading."
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	args := map[string]interface{}{
		"path": testFile,
	}

	result := tool.Execute(args)

	if result.Success != true {
		t.Errorf("Expected successful execution, but got error: %s", result.Error)
	}

	// Verify the content was read correctly
	data, ok := result.Data.(map[string]interface{})
	if !ok || data == nil {
		t.Errorf("Expected data in result to be a map, but got %T", result.Data)
		return
	}

	content, ok := data["content"].(string)
	if !ok {
		t.Errorf("Expected content to be a string, but got %T", data["content"])
		return
	}

	if content != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, content)
	}

	path, ok := data["path"].(string)
	if !ok || path != testFile {
		t.Errorf("Expected path '%s', got '%s'", testFile, path)
	}
}

func TestReadFileContentsTool_Execute_InvalidPath(t *testing.T) {
	tool := &ReadFileContentsTool{}

	args := map[string]interface{}{
		"path": "/etc/passwd",
	}

	result := tool.Execute(args)

	if result.Success != false {
		t.Errorf("Expected execution to fail for invalid path, but it succeeded")
	}

	if !strings.Contains(result.Error, "outside allowed directories") {
		t.Errorf("Expected error about path being outside allowed directories, got: %s", result.Error)
	}
}

func TestReadFileContentsTool_Execute_NonexistentFile(t *testing.T) {
	tool := &ReadFileContentsTool{}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	nonexistentFile := filepath.Join(homeDir, "mermaid-agent-documenter", "does_not_exist.md")

	args := map[string]interface{}{
		"path": nonexistentFile,
	}

	result := tool.Execute(args)

	if result.Success != false {
		t.Errorf("Expected execution to fail for nonexistent file, but it succeeded")
	}

	// The error should be about the file not existing, not about path validation
	if strings.Contains(result.Error, "outside allowed directories") {
		t.Errorf("Expected file not found error, but got path validation error: %s", result.Error)
	}
}

func TestReadFileContentsTool_Execute_MissingPath(t *testing.T) {
	tool := &ReadFileContentsTool{}

	args := map[string]interface{}{
		"maxBytes": 100,
	}

	result := tool.Execute(args)

	if result.Success != false {
		t.Errorf("Expected execution to fail with missing path, but it succeeded")
	}

	if !strings.Contains(result.Error, "Missing or invalid 'path' argument") {
		t.Errorf("Expected error about missing path argument, got: %s", result.Error)
	}
}

func TestReadFileContentsTool_Execute_WithMaxBytes(t *testing.T) {
	tool := &ReadFileContentsTool{}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	// Create a test file with known content
	testFile := filepath.Join(homeDir, "mermaid-agent-documenter", "test_maxbytes.txt")
	testContent := "This is a longer test content that we can limit with maxBytes parameter."
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	args := map[string]interface{}{
		"path":     testFile,
		"maxBytes": 20, // Limit to first 20 bytes
	}

	result := tool.Execute(args)

	if result.Success != true {
		t.Errorf("Expected successful execution, but got error: %s", result.Error)
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok || data == nil {
		t.Errorf("Expected data in result to be a map, but got %T", result.Data)
		return
	}

	content, ok := data["content"].(string)
	if !ok {
		t.Errorf("Expected content to be a string, but got %T", data["content"])
		return
	}

	// Content should be truncated to maxBytes
	expectedContent := testContent[:20]
	if content != expectedContent {
		t.Errorf("Expected truncated content '%s', got '%s'", expectedContent, content)
	}

	// Should indicate truncation
	truncated, ok := data["truncated"].(bool)
	if !ok || truncated != true {
		t.Errorf("Expected truncated to be true, got %v", data["truncated"])
	}
}
