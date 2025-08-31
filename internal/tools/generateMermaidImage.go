package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GenerateMermaidImageTool struct{}

func (t *GenerateMermaidImageTool) Name() string {
	return "generateMermaidImage"
}

func (t *GenerateMermaidImageTool) Description() string {
	return "Generate SVG/PNG images from Mermaid diagram files using Mermaid CLI"
}

func (t *GenerateMermaidImageTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"inputFile": map[string]interface{}{
				"type":        "string",
				"description": "Path to the Markdown file containing Mermaid diagrams",
			},
			"outputFile": map[string]interface{}{
				"type":        "string",
				"description": "Path for the output image file (without extension)",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"svg", "png", "pdf"},
				"description": "Output format: svg (default), png, or pdf",
				"default":     "svg",
			},
			"createDirs": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to create output directories if they don't exist",
				"default":     true,
			},
		},
		"required": []string{"inputFile", "outputFile"},
	}
}

func (t *GenerateMermaidImageTool) Execute(args map[string]interface{}) ToolResult {
	inputFile, ok := args["inputFile"].(string)
	if !ok {
		return ToolResult{
			Success: false,
			Error:   "Missing or invalid 'inputFile' argument",
		}
	}

	outputFile, ok := args["outputFile"].(string)
	if !ok {
		return ToolResult{
			Success: false,
			Error:   "Missing or invalid 'outputFile' argument",
		}
	}

	format := "svg" // default
	if fmt, exists := args["format"].(string); exists && (fmt == "svg" || fmt == "png" || fmt == "pdf") {
		format = fmt
	}

	createDirs := true
	if cd, exists := args["createDirs"]; exists {
		if cdBool, ok := cd.(bool); ok {
			createDirs = cdBool
		}
	}

	// Expand ~ to home directory
	if strings.HasPrefix(inputFile, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return ToolResult{
				Success: false,
				Error:   "Failed to get home directory: " + err.Error(),
			}
		}
		inputFile = strings.Replace(inputFile, "~", home, 1)
	}

	if strings.HasPrefix(outputFile, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return ToolResult{
				Success: false,
				Error:   "Failed to get home directory: " + err.Error(),
			}
		}
		outputFile = strings.Replace(outputFile, "~", home, 1)
	}

	// Check if input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Input file does not exist: %s", inputFile),
		}
	}

	// Create output directory if needed
	if createDirs {
		outputDir := filepath.Dir(outputFile)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return ToolResult{
				Success: false,
				Error:   "Failed to create output directories: " + err.Error(),
			}
		}
	}

	// Check if Mermaid CLI is available
	if _, err := exec.LookPath("mmdc"); err != nil {
		return ToolResult{
			Success: false,
			Error:   "Mermaid CLI (mmdc) is not installed. Install it with: npm install -g @mermaid-js/mermaid-cli",
		}
	}

	// Construct the full output path with extension
	fullOutputPath := outputFile
	if !strings.HasSuffix(fullOutputPath, "."+format) {
		fullOutputPath = fullOutputPath + "." + format
	}

	// Build Mermaid CLI command
	cmd := exec.Command("mmdc", "-i", inputFile, "-o", fullOutputPath)

	// Set environment variables if needed
	cmd.Env = os.Environ()

	// Execute the command
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check for common Mermaid CLI errors
		errorMsg := string(output)
		if strings.Contains(errorMsg, "No diagram found") {
			return ToolResult{
				Success: false,
				Error:   fmt.Sprintf("No Mermaid diagrams found in file: %s", inputFile),
			}
		}
		if strings.Contains(errorMsg, "Syntax error") {
			return ToolResult{
				Success: false,
				Error:   fmt.Sprintf("Mermaid syntax error in file %s: %s", inputFile, errorMsg),
			}
		}
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Mermaid CLI error: %v\nOutput: %s", err, errorMsg),
		}
	}

	// Verify the output file was created
	if _, err := os.Stat(fullOutputPath); os.IsNotExist(err) {
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Output file was not created: %s", fullOutputPath),
		}
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"inputFile":     inputFile,
			"outputFile":    fullOutputPath,
			"format":        format,
			"commandOutput": string(output),
		},
	}
}
