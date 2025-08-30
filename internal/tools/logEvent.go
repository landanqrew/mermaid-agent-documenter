package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type LogEventTool struct{}

func (t *LogEventTool) Name() string {
	return "logEvent"
}

func (t *LogEventTool) Description() string {
	return "Log an event with level, message, and optional data"
}

func (t *LogEventTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"level": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"debug", "info", "warn", "error"},
				"description": "Log level",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Log message",
			},
			"data": map[string]interface{}{
				"type":        "object",
				"description": "Optional additional data to log",
			},
		},
		"required": []string{"level", "message"},
	}
}

func (t *LogEventTool) Execute(args map[string]interface{}) ToolResult {
	level, ok := args["level"].(string)
	if !ok {
		return ToolResult{
			Success: false,
			Error:   "Missing or invalid 'level' argument",
		}
	}

	message, ok := args["message"].(string)
	if !ok {
		return ToolResult{
			Success: false,
			Error:   "Missing or invalid 'message' argument",
		}
	}

	// Validate level
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[level] {
		return ToolResult{
			Success: false,
			Error:   "Invalid log level. Must be one of: debug, info, warn, error",
		}
	}

	// Get log directory
	home, err := os.UserHomeDir()
	if err != nil {
		return ToolResult{
			Success: false,
			Error:   "Failed to get home directory: " + err.Error(),
		}
	}

	logDir := filepath.Join(home, "mermaid-agent-documenter", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return ToolResult{
			Success: false,
			Error:   "Failed to create log directory: " + err.Error(),
		}
	}

	// Create log entry
	logEntry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     level,
		"message":   message,
	}

	if data, exists := args["data"]; exists {
		logEntry["data"] = data
	}

	// Write to logs.jsonl
	logFile := filepath.Join(logDir, "events.jsonl")
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return ToolResult{
			Success: false,
			Error:   "Failed to open log file: " + err.Error(),
		}
	}
	defer file.Close()

	logJSON, err := json.Marshal(logEntry)
	if err != nil {
		return ToolResult{
			Success: false,
			Error:   "Failed to marshal log entry: " + err.Error(),
		}
	}

	if _, err := file.WriteString(string(logJSON) + "\n"); err != nil {
		return ToolResult{
			Success: false,
			Error:   "Failed to write log entry: " + err.Error(),
		}
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"logged": true,
		},
	}
}
