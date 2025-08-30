package tools

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type GetUserInputTool struct{}

func (t *GetUserInputTool) Name() string {
	return "getUserInput"
}

func (t *GetUserInputTool) Description() string {
	return "Get interactive input from the user"
}

func (t *GetUserInputTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"prompt": map[string]interface{}{
				"type":        "string",
				"description": "Prompt message to display to the user",
			},
		},
		"required": []string{"prompt"},
	}
}

func (t *GetUserInputTool) Execute(args map[string]interface{}) ToolResult {
	prompt, ok := args["prompt"].(string)
	if !ok {
		return ToolResult{
			Success: false,
			Error:   "Missing or invalid 'prompt' argument",
		}
	}

	fmt.Print(prompt + " ")
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return ToolResult{
			Success: false,
			Error:   "Failed to read user input: " + err.Error(),
		}
	}

	// Trim whitespace and newlines
	answer = strings.TrimSpace(answer)

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"answer": answer,
		},
	}
}
