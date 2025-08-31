package tools

import (
	"encoding/json"
	"fmt"
)

type ToolResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type Tool interface {
	Name() string
	Description() string
	Execute(args map[string]interface{}) ToolResult
	Schema() map[string]interface{}
}

var toolRegistry = map[string]Tool{}

func RegisterTool(tool Tool) {
	toolRegistry[tool.Name()] = tool
}

func GetTool(name string) Tool {
	return toolRegistry[name]
}

func ListTools() map[string]Tool {
	return toolRegistry
}

func init() {
	RegisterTool(&ReadDirectoriesTool{})
	RegisterTool(&ReadFileContentsTool{})
	RegisterTool(&WriteFileContentsTool{})
	RegisterTool(&GetUserInputTool{})
	RegisterTool(&FetchMermaidDocumentationTool{})
	RegisterTool(&LogEventTool{})
	RegisterTool(&GenerateMermaidImageTool{})
}

// ExecuteTool executes a tool by name with JSON arguments
func ExecuteTool(toolName string, argsJSON string) ToolResult {
	tool := GetTool(toolName)
	if tool == nil {
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Tool '%s' not found", toolName),
		}
	}

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Invalid JSON arguments: %v", err),
		}
	}

	return tool.Execute(args)
}
