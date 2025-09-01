package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type WriteFileContentsTool struct{}

// validatePath checks if the given path is within allowed directories
func (t *WriteFileContentsTool) validatePath(path string) error {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Allowed base directories
	allowedDirs := []string{
		filepath.Join(homeDir, "mermaid-agent-documenter"), // ~/mermaid-agent-documenter/
	}

	// Add current project directory if available
	configPath := filepath.Join(homeDir, "mermaid-agent-documenter", "config.json")
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err == nil {
			var cfg struct {
				CurrentProject *struct {
					RootDir string `json:"rootDir"`
				} `json:"currentProject,omitempty"`
			}
			if err := json.Unmarshal(data, &cfg); err == nil && cfg.CurrentProject != nil {
				allowedDirs = append(allowedDirs, cfg.CurrentProject.RootDir)
			}
		}
	}

	// Check if the path is within one of the allowed directories
	for _, allowedDir := range allowedDirs {
		absAllowedDir, err := filepath.Abs(allowedDir)
		if err != nil {
			continue // Skip invalid allowed directories
		}

		// Check if absPath is within or equal to absAllowedDir
		relPath, err := filepath.Rel(absAllowedDir, absPath)
		if err != nil {
			continue // Path is not relative to this allowed directory
		}

		// If relPath doesn't start with ".." it's within the allowed directory
		if !strings.HasPrefix(relPath, "..") {
			return nil // Path is valid
		}
	}

	return fmt.Errorf("path '%s' is outside allowed directories. File operations are only allowed within ~/mermaid-agent-documenter/ or the current project directory", path)
}

func (t *WriteFileContentsTool) Name() string {
	return "writeFileContents"
}

func (t *WriteFileContentsTool) Description() string {
	return "Write content to a file"
}

func (t *WriteFileContentsTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to write",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write to the file",
			},
			"createDirs": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to create parent directories if they don't exist",
			},
			"overwrite": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"explicit", "allow"},
				"description": "Overwrite behavior: 'explicit' requires confirmation, 'allow' allows overwriting",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteFileContentsTool) Execute(args map[string]interface{}) ToolResult {
	path, ok := args["path"].(string)
	if !ok {
		return ToolResult{
			Success: false,
			Error:   "Missing or invalid 'path' argument",
		}
	}

	// Validate that the path is within allowed directories
	if err := t.validatePath(path); err != nil {
		return ToolResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	content, ok := args["content"].(string)
	if !ok {
		return ToolResult{
			Success: false,
			Error:   "Missing or invalid 'content' argument",
		}
	}

	// Debug: print what we're trying to write
	fmt.Printf("📝 Writing to: %s (%d chars)\n", path, len(content))

	createDirs := true
	if cd, exists := args["createDirs"]; exists {
		if cdBool, ok := cd.(bool); ok {
			createDirs = cdBool
		}
	}

	overwrite := "allow" // Default to allow for agent workflow
	if ow, exists := args["overwrite"]; exists {
		if owStr, ok := ow.(string); ok && (owStr == "explicit" || owStr == "allow") {
			overwrite = owStr
		}
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return ToolResult{
				Success: false,
				Error:   "Failed to get home directory: " + err.Error(),
			}
		}
		path = strings.Replace(path, "~", home, 1)
	}

	// Create directories if requested
	if createDirs {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return ToolResult{
				Success: false,
				Error:   "Failed to create directories: " + err.Error(),
			}
		}
	}

	// Check if file exists and handle overwrite policy
	if _, err := os.Stat(path); err == nil {
		if overwrite == "explicit" {
			return ToolResult{
				Success: false,
				Error:   "File exists and overwrite is set to 'explicit'. Use overwrite='allow' to overwrite.",
			}
		}
	}

	// Write the file
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return ToolResult{
			Success: false,
			Error:   "Failed to write file: " + err.Error(),
		}
	}
	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"path":         path,
			"bytesWritten": len(content),
		},
	}
}
