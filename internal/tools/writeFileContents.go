package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type WriteFileContentsTool struct{}

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
