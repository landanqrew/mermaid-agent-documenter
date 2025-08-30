package tools

import (
	"os"
	"strconv"
)

type ReadFileContentsTool struct{}

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

	var maxBytes int64 = -1 // read all by default
	if mb, exists := args["maxBytes"]; exists {
		if mbFloat, ok := mb.(float64); ok {
			maxBytes = int64(mbFloat)
		} else if mbStr, ok := mb.(string); ok {
			if parsed, err := strconv.ParseInt(mbStr, 10, 64); err == nil {
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