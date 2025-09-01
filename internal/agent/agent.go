package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/landanqrew/mermaid-agent-documenter/internal/providers"
	"github.com/landanqrew/mermaid-agent-documenter/internal/tools"
)

// Structured output envelope types
type OutputType string

const (
	OutputTypeToolCall      OutputType = "tool_call"
	OutputTypeFinal         OutputType = "final"
	OutputTypeClarification OutputType = "clarification"
)

type StructuredOutput struct {
	Type       OutputType             `json:"type"`
	Tool       string                 `json:"tool,omitempty"`
	Args       map[string]interface{} `json:"args,omitempty"`
	Manifest   map[string]interface{} `json:"manifest,omitempty"`
	Questions  []string               `json:"questions,omitempty"`
	Confidence float64                `json:"confidence"`
	Rationale  string                 `json:"rationale"`
}

type MermaidDocumenterAgent struct {
	Provider         providers.LLMProvider
	Config           *AgentConfig
	RunID            string
	StepCount        int
	Transcript       string
	consecutiveFails int
}

type AgentConfig struct {
	Provider            string
	Model               string
	APIKey              string
	MaxSteps            int
	TimeoutSec          int
	TokenBudget         int
	CostCeilingUsd      float64
	ConfidenceThreshold float64
	OutputDir           string
	LogsDir             string
	RedactPII           bool
	StoreChainOfThought bool
	DocumentationTypes  []string
}

func NewMermaidDocumenterAgent(config *AgentConfig) *MermaidDocumenterAgent {
	return &MermaidDocumenterAgent{
		Provider:  providers.GetProvider(config.Provider),
		Config:    config,
		RunID:     uuid.New().String(),
		StepCount: 0,
	}
}

func (a *MermaidDocumenterAgent) SetTranscript(transcript string) {
	a.Transcript = transcript
}

func (a *MermaidDocumenterAgent) Run(ctx context.Context) error {
	systemPrompt := a.buildSystemPrompt()

	conversation := []map[string]interface{}{
		{
			"role":    "system",
			"content": systemPrompt,
		},
		{
			"role":    "user",
			"content": fmt.Sprintf("Please analyze this application transcript and generate Mermaid documentation:\n\n%s", a.Transcript),
		},
	}

	for a.StepCount < a.Config.MaxSteps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Build the conversation string for the LLM
		conversationStr := a.buildConversationString(conversation)

		// Call the LLM
		response, err := a.Provider.GenerateContent(ctx, conversationStr, a.Config.Model, a.Config.APIKey)
		if err != nil {
			return fmt.Errorf("LLM call failed: %w", err)
		}

		// Parse the structured output
		output, err := a.parseStructuredOutput(response)
		if err != nil {
			return fmt.Errorf("failed to parse LLM response: %w", err)
		}

		// Log the interaction
		a.logInteraction(conversation, response, output)

		// Handle the output based on type
		switch output.Type {
		case OutputTypeToolCall:
			if output.Confidence < a.Config.ConfidenceThreshold {
				// Ask for clarification instead of executing low-confidence tool calls
				conversation = append(conversation, map[string]interface{}{
					"role":    "assistant",
					"content": response,
				})
				conversation = append(conversation, map[string]interface{}{
					"role":    "user",
					"content": "Your confidence is below the threshold. Please provide clarification or reconsider your approach.",
				})
				continue
			}

			// Modify file paths to use output directory if they're relative
			modifiedArgs := a.modifyFilePaths(output.Args)

			// Execute the tool
			result := tools.ExecuteTool(output.Tool, a.argsToJSON(modifiedArgs))

			if result.Success && result.Data != nil {
				fmt.Printf("✅ Tool completed successfully\n")
				a.consecutiveFails = 0 // Reset failure counter on success
			} else if !result.Success {
				fmt.Printf("❌ Tool failed: %s\n", result.Error)
				a.consecutiveFails++

				// If too many consecutive failures, force final manifest
				if a.consecutiveFails >= 3 {
					fmt.Printf("⚠️  Too many consecutive failures (%d), forcing final manifest\n", a.consecutiveFails)
					return nil // This will trigger final manifest processing
				}

				// If the tool failed, add error context to guide the next action
				errorMsg := fmt.Sprintf("Tool execution failed: %s. ", result.Error)
				if strings.Contains(result.Error, "Mermaid CLI error") {
					errorMsg += "This is likely due to invalid Mermaid syntax. Check your diagram syntax, especially ER diagrams which should use semicolons (;) not commas (,) to separate attributes. "
				}
				errorMsg += "Please fix the issue and try again, or return a final manifest if you cannot resolve it. You MUST respond with valid JSON tool calls or final manifest."

				conversation = append(conversation, map[string]interface{}{
					"role":    "system",
					"content": errorMsg,
				})
			}

			resultStr := fmt.Sprintf("Tool result: %v", result)

			conversation = append(conversation, map[string]interface{}{
				"role":    "assistant",
				"content": response,
			})
			conversation = append(conversation, map[string]interface{}{
				"role":    "user",
				"content": resultStr,
			})

		case OutputTypeFinal:
			if output.Confidence >= a.Config.ConfidenceThreshold {
				// Process the final manifest
				a.processFinalManifest(output.Manifest)
				return nil
			} else {
				// Ask for clarification
				conversation = append(conversation, map[string]interface{}{
					"role":    "assistant",
					"content": response,
				})
				conversation = append(conversation, map[string]interface{}{
					"role":    "user",
					"content": "Your confidence is below the threshold. Please provide clarification or reconsider your approach.",
				})
				continue
			}

		case OutputTypeClarification:
			// Handle clarification request
			fmt.Printf("Agent needs clarification:\n")
			for _, question := range output.Questions {
				fmt.Printf("- %s\n", question)
			}
			return fmt.Errorf("clarification needed")

		default:
			fmt.Printf("⚠️  Unknown output type: %s\n", output.Type)
			// For unknown types, try to continue with the next step
			fmt.Printf("🔄 Continuing with next step...\n")
			continue
		}

		a.StepCount++
	}

	return fmt.Errorf("maximum steps (%d) exceeded", a.Config.MaxSteps)
}

func (a *MermaidDocumenterAgent) buildSystemPrompt() string {
	content := "## Summary\\n\\nThe transcript describes a GoCarWash application.\\n\\n```mermaid\\ngraph TD\\n    A[User] --> B[App]\\n```"

	basePrompt := `You are Mermaid Documenter Agent.

TASK: Create documentation with Mermaid diagrams and generate SVG images.

REQUIRED SEQUENCE:
1. FIRST: Use writeFileContents to create summary.md with VALID Mermaid diagrams
2. SECOND: Use generateMermaidImage to convert the Markdown file to SVG images
3. THIRD: Return final manifest ONLY after both files are created

FILE PATH REQUIREMENTS:
- ALWAYS use the EXACT filename you created in writeFileContents (e.g., "summary.md")
- Do NOT use relative paths or modify the filename

MERMAID SYNTAX RULES:
- For ER diagrams: Use simple attribute names without types: Site {id; name}
- Avoid complex ER relationships - use simple ||--o{ syntax
- For sequence diagrams: Use simple participant names without spaces
- Keep syntax simple and avoid special characters
- Test syntax mentally: Would this parse correctly?

ERROR HANDLING:
- If generateMermaidImage fails, the error message will contain specific syntax issues
- Fix the identified syntax problems and try again
- Focus on the sequence diagram first if ER diagram fails

IMPORTANT: You MUST call generateMermaidImage as a separate tool call after creating the Markdown file. Do NOT claim SVG generation in the final manifest unless you actually called the generateMermaidImage tool.

MERMAID DIAGRAM BEST PRACTICES:
- Use simple sequence diagrams when possible - they are most reliable
- Avoid complex ER diagrams with data types (use simple attribute names only)
- Limit files to ONE diagram type to avoid parsing conflicts
- For ER diagrams: Use format "Entity { attribute1 attribute2 }" without types or semicolons
- For relationships: Use simple "Entity1 -- Entity2 : description" format
- Test diagrams mentally: Would this parse correctly in Mermaid?`

	// Add OpenAI-specific instructions for tool calling sequence
	if a.Config.Provider == "openai" {
		basePrompt += `

OPENAI-SPECIFIC INSTRUCTIONS:
- ALWAYS follow this EXACT sequence: writeFileContents -> generateMermaidImage -> final manifest
- NEVER call generateMermaidImage before creating the file with writeFileContents
- NEVER skip steps or combine tool calls in a single response
- If you receive an error about file not existing, create the file first before generating images
- Wait for tool results before proceeding to the next step`
	}

	basePrompt += `

Return ONLY JSON:

TOOL CALL 1 (create documentation):
{"type":"tool_call","tool":"writeFileContents","args":{"path":"summary.md","content":"` + content + `","overwrite":"allow"},"confidence":0.95,"rationale":"creating documentation"}

TOOL CALL 2 (generate images):
{"type":"tool_call","tool":"generateMermaidImage","args":{"inputFile":"summary.md","outputFile":"summary","format":"svg"},"confidence":0.95,"rationale":"generating SVG images"}

FINAL RESULT (only after both steps complete):
{"type":"final","manifest":{"summary.md":"created","summary.svg":"generated"},"confidence":0.95,"rationale":"documentation complete"}`

	if len(a.Config.DocumentationTypes) > 0 {
		basePrompt = strings.Replace(basePrompt, "summary", strings.Join(a.Config.DocumentationTypes, "_"), 1)
	}

	return basePrompt
}

func (a *MermaidDocumenterAgent) buildConversationString(conversation []map[string]interface{}) string {
	var sb strings.Builder
	for _, msg := range conversation {
		sb.WriteString(fmt.Sprintf("%s: %s\n", msg["role"], msg["content"]))
	}
	return sb.String()
}

func (a *MermaidDocumenterAgent) parseStructuredOutput(response string) (*StructuredOutput, error) {
	response = strings.TrimSpace(response)

	// First, try to detect if this is an API error response
	if a.isAPIErrorResponse(response) {
		return nil, fmt.Errorf("API error in response: %s", response)
	}

	// Clean the response by removing markdown code blocks
	response = a.cleanMarkdownCodeBlocks(response)

	// Try to extract the first valid JSON object from the response
	jsonObjects := a.extractJSONObject(response)
	if len(jsonObjects) == 0 {
		return nil, fmt.Errorf("no valid JSON objects found in response: %s", response)
	}

	// Parse the first JSON object
	var output StructuredOutput
	firstObject := jsonObjects[0]

	// Try to fix common JSON issues before parsing
	firstObject = a.fixCommonJSONIssues(firstObject)

	if err := json.Unmarshal([]byte(firstObject), &output); err != nil {
		// If JSON parsing fails, provide more context and debugging info
		fmt.Printf("🔍 JSON Parsing Debug:\n")
		fmt.Printf("  📄 Raw response length: %d characters\n", len(response))
		fmt.Printf("  📄 First object length: %d characters\n", len(firstObject))
		fmt.Printf("  📄 First object preview: %s...\n", firstObject[:min(200, len(firstObject))])
		fmt.Printf("  ❌ JSON Error: %v\n", err)

		return nil, fmt.Errorf("failed to parse response as structured output JSON: %w. First object: %s", err, firstObject)
	}

	// Validate the parsed output has required fields
	if output.Type == "" {
		return nil, fmt.Errorf("parsed output missing required 'type' field")
	}

	return &output, nil
}

// cleanMarkdownCodeBlocks removes markdown code block formatting from the response
func (a *MermaidDocumenterAgent) cleanMarkdownCodeBlocks(response string) string {
	response = strings.TrimSpace(response)

	// Handle various markdown code block formats
	if strings.HasPrefix(response, "```json") {
		// Remove opening marker
		response = strings.TrimPrefix(response, "```json")
		// Remove closing marker if present
		response = strings.TrimSuffix(response, "```")
	} else if strings.HasPrefix(response, "```") {
		// Remove generic code block markers
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
	}

	return strings.TrimSpace(response)
}

// extractJSONObject extracts individual JSON objects from a concatenated JSON string
func (a *MermaidDocumenterAgent) extractJSONObject(response string) []string {
	var objects []string

	// First, try to parse the entire response as a single JSON object
	var temp interface{}
	if err := json.Unmarshal([]byte(response), &temp); err == nil {
		// If it parses successfully, return it as the only object
		return []string{response}
	}

	// If that fails, try a simpler approach: split by "}{" and add back the braces
	if strings.Contains(response, "}{") {
		parts := strings.Split(response, "}{")

		for i, part := range parts {
			var obj string
			if i == 0 {
				// First part: add opening brace
				obj = part + "}"
			} else if i == len(parts)-1 {
				// Last part: add closing brace
				obj = "{" + part
			} else {
				// Middle parts: add both braces
				obj = "{" + part + "}"
			}

			// Test if this is valid JSON
			var temp interface{}
			if err := json.Unmarshal([]byte(obj), &temp); err == nil {
				objects = append(objects, obj)
			}
		}
	}

	// If splitting didn't work, try the brace-counting approach as fallback
	if len(objects) == 0 {
		objects = a.extractJSONObjectBraceCounting(response)
	}

	return objects
}

// fixCommonJSONIssues attempts to fix common JSON formatting issues
func (a *MermaidDocumenterAgent) fixCommonJSONIssues(jsonStr string) string {
	// Remove any trailing commas before closing braces/brackets
	jsonStr = strings.ReplaceAll(jsonStr, ",}", "}")
	jsonStr = strings.ReplaceAll(jsonStr, ",]", "]")

	// Ensure proper JSON structure
	jsonStr = strings.TrimSpace(jsonStr)

	return jsonStr
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// modifyFilePaths modifies file paths in tool arguments to use the output directory
func (a *MermaidDocumenterAgent) modifyFilePaths(args map[string]interface{}) map[string]interface{} {
	modifiedArgs := make(map[string]interface{})

	// Copy all original args
	for k, v := range args {
		modifiedArgs[k] = v
	}

	// Check for path arguments that need modification (handles both "path" and "inputFile")
	pathArgs := []string{"path", "inputFile"}
	for _, argName := range pathArgs {
		if pathVal, exists := args[argName]; exists {
			if pathStr, ok := pathVal.(string); ok {
				// If path is relative (doesn't start with / or ~), prepend output directory
				if !strings.HasPrefix(pathStr, "/") && !strings.HasPrefix(pathStr, "~") && !filepath.IsAbs(pathStr) {
					modifiedPath := filepath.Join(a.Config.OutputDir, pathStr)
					modifiedArgs[argName] = modifiedPath
				}
			}
		}
	}

	return modifiedArgs
}

// extractJSONObjectBraceCounting uses brace counting to extract JSON objects
func (a *MermaidDocumenterAgent) extractJSONObjectBraceCounting(response string) []string {
	var objects []string
	var currentObject strings.Builder
	braceCount := 0
	inString := false
	escapeNext := false

	for _, char := range response {
		currentObject.WriteRune(char)

		switch char {
		case '"':
			if !escapeNext {
				inString = !inString
			}
		case '\\':
			escapeNext = !escapeNext
			continue
		case '{':
			if !inString {
				braceCount++
			}
		case '}':
			if !inString {
				braceCount--
				if braceCount == 0 {
					// We've found a complete JSON object
					obj := strings.TrimSpace(currentObject.String())
					if obj != "" {
						objects = append(objects, obj)
					}
					currentObject.Reset()
				}
			}
		}

		if char != '\\' {
			escapeNext = false
		}
	}

	return objects
}

// completePartialJSONObject attempts to complete a partial JSON object
func (a *MermaidDocumenterAgent) completePartialJSONObject(partial string) string {
	// Count braces to see what's missing
	openBraces := strings.Count(partial, "{")
	closeBraces := strings.Count(partial, "}")

	if openBraces <= closeBraces {
		return "" // Not a partial object or already complete
	}

	// Add missing closing braces
	completed := partial
	for i := 0; i < openBraces-closeBraces; i++ {
		completed += "}"
	}

	// Test if it's now valid JSON
	var temp interface{}
	if err := json.Unmarshal([]byte(completed), &temp); err == nil {
		return completed
	}

	return "" // Couldn't complete it
}

// isAPIErrorResponse checks if the response appears to be an API error rather than our expected output
func (a *MermaidDocumenterAgent) isAPIErrorResponse(response string) bool {
	// Check for common API error patterns
	errorPatterns := []string{
		"Error 400",
		"Error 401",
		"Error 403",
		"Error 404",
		"API key not valid",
		"Model not found",
		"Invalid model",
		"unsupported model",
		"model does not exist",
		"INVALID_ARGUMENT",
		"PERMISSION_DENIED",
		"NOT_FOUND",
	}

	responseLower := strings.ToLower(response)
	for _, pattern := range errorPatterns {
		if strings.Contains(responseLower, strings.ToLower(pattern)) {
			return true
		}
	}

	// Check if it looks like a JSON error object
	if strings.HasPrefix(strings.TrimSpace(response), "{") {
		var temp map[string]interface{}
		if err := json.Unmarshal([]byte(response), &temp); err == nil {
			// It's valid JSON, check if it has error fields
			if _, hasError := temp["error"]; hasError {
				return true
			}
			if _, hasMessage := temp["message"]; hasMessage {
				if _, hasCode := temp["code"]; hasCode {
					return true
				}
			}
		}
	}

	return false
}

func (a *MermaidDocumenterAgent) argsToJSON(args map[string]interface{}) string {
	jsonBytes, _ := json.Marshal(args)
	return string(jsonBytes)
}

func (a *MermaidDocumenterAgent) logInteraction(conversation []map[string]interface{}, response string, output *StructuredOutput) {
	fmt.Printf("Step %d: %s (confidence: %.2f)\n", a.StepCount+1, output.Type, output.Confidence)

	// Skip logging if LogsDir is not set
	if a.Config.LogsDir == "" {
		return
	}

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(a.Config.LogsDir, 0755); err != nil {
		fmt.Printf("Warning: Failed to create logs directory: %v\n", err)
		return
	}

	// Create log entry
	logEntry := map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339),
		"run_id":      a.RunID,
		"step":        a.StepCount + 1,
		"provider":    a.Config.Provider,
		"model":       a.Config.Model,
		"output_type": output.Type,
		"confidence":  output.Confidence,
		"rationale":   output.Rationale,
	}

	// Add chain of thought if enabled
	if a.Config.StoreChainOfThought {
		logEntry["conversation"] = conversation
		logEntry["response"] = response
	}

	// Add tool information if applicable
	if output.Type == "tool_call" {
		logEntry["tool"] = output.Tool
		logEntry["args"] = output.Args
	}

	// Add final manifest if applicable
	if output.Type == "final" {
		logEntry["manifest"] = output.Manifest
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Printf("Warning: Failed to marshal log entry: %v\n", err)
		return
	}

	// Write to logs.jsonl file
	logFilePath := filepath.Join(a.Config.LogsDir, "logs.jsonl")
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Warning: Failed to open log file: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(string(jsonData) + "\n"); err != nil {
		fmt.Printf("Warning: Failed to write to log file: %v\n", err)
	}
}

func (a *MermaidDocumenterAgent) processFinalManifest(manifest map[string]interface{}) {
	// TODO: Process and validate the final manifest
	fmt.Printf("Processing final manifest: %v\n", manifest)
}
