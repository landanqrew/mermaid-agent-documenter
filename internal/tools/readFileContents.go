package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ReadFileContentsTool struct{}

// validatePath checks if the given path is within allowed directories
func (t *ReadFileContentsTool) validatePath(path string) error {
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

func (t *ReadFileContentsTool) Name() string {
	return "readFileContents"
}

func (t *ReadFileContentsTool) Description() string {
	return "Read the contents of a file"
}

func (t *ReadFileContentsTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type": "string",
				"description": "Path to the file to read",
			},
			"maxBytes": map[string]interface{}{
				"type": "number",
				"description": "Maximum number of bytes to read (optional)",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFileContentsTool) Execute(args map[string]interface{}) ToolResult {
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

	var maxBytes int64 = -1 // read all by default
	if mb, exists := args["maxBytes"]; exists {
		switch v := mb.(type) {
		case float64:
			maxBytes = int64(v)
		case int:
			maxBytes = int64(v)
		case int64:
			maxBytes = v
		case string:
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				maxBytes = parsed
			}
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return ToolResult{
			Success: false,
			Error:   err.Error(),
		}
	}
	defer file.Close()

	var data []byte
	if maxBytes > 0 {
		data = make([]byte, maxBytes)
		n, err := file.Read(data)
		if err != nil && n == 0 {
			return ToolResult{
				Success: false,
				Error:   err.Error(),
			}
		}
		data = data[:n]
	} else {
		data, err = os.ReadFile(path)
		if err != nil {
			return ToolResult{
				Success: false,
				Error:   err.Error(),
			}
		}
	}

	truncated := maxBytes > 0 && int64(len(data)) >= maxBytes

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"path":      path,
			"content":   string(data),
			"truncated": truncated,
		},
	}
}