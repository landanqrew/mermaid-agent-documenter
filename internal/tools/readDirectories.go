package tools

import (
	"os"
	"path/filepath"
)

type ReadDirectoriesTool struct{}

func (t *ReadDirectoriesTool) Name() string {
	return "readDirectories"
}

func (t *ReadDirectoriesTool) Description() string {
	return "List files and directories in a given path"
}

func (t *ReadDirectoriesTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type": "string",
				"description": "Path to directory to list contents of",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadDirectoriesTool) Execute(args map[string]interface{}) ToolResult {
	path, ok := args["path"].(string)
	if !ok {
		return ToolResult{
			Success: false,
			Error:   "Missing or invalid 'path' argument",
		}
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return ToolResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	var directories []string
	var files []string

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			directories = append(directories, fullPath)
		} else {
			files = append(files, fullPath)
		}
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"directories": directories,
			"files":       files,
		},
	}
}
