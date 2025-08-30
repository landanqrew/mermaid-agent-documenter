package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	Provider   providers.LLMProvider
	Config     *AgentConfig
	RunID      string
	StepCount  int
	Transcript string
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
	RedactPII           bool
	StoreChainOfThought bool
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

			// Execute the tool
			result := tools.ExecuteTool(output.Tool, a.argsToJSON(output.Args))
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
			return fmt.Errorf("unknown output type: %s", output.Type)
		}

		a.StepCount++
	}

	return fmt.Errorf("maximum steps (%d) exceeded", a.Config.MaxSteps)
}

func (a *MermaidDocumenterAgent) buildSystemPrompt() string {
	return `You are "Mermaid Documenter Agent". Read application transcripts and produce accurate Markdown docs with Mermaid diagrams. Prefer multiple small, focused diagrams. Use verbatim labels from the transcript where possible. Include 2–3 sentence context descriptions per file, plus assumptions and openQuestions only when necessary.

Diagram policy: allowed types sequence|flowchart|class|er|state|journey|graph; default direction LR; ≤ ~40 nodes; split by concern. Validate Mermaid syntax.

When confident (≥ 0.90), issue a single tool_call to writeFileContents per file with the complete Markdown content. Otherwise, emit a clarification with targeted questions.

Available tools:
- readDirectories(path): List files and directories
- readFileContents(path, maxBytes?): Read file contents
- writeFileContents(path, content, createDirs=true, overwrite="explicit"): Write files
- getUserInput(prompt): Get user input
- fetchMermaidDocumentation(topic?, version?): Get Mermaid docs
- logEvent(level, message, data?): Log events

Respond ONLY with JSON envelopes:
{"type":"tool_call","tool":"writeFileContents","args":{...},"confidence":0.92,"rationale":"..."}
{"type":"final","manifest":{...},"confidence":0.94,"rationale":"..."}
{"type":"clarification","questions":[...],"confidence":0.62}`
}

func (a *MermaidDocumenterAgent) buildConversationString(conversation []map[string]interface{}) string {
	var sb strings.Builder
	for _, msg := range conversation {
		sb.WriteString(fmt.Sprintf("%s: %s\n", msg["role"], msg["content"]))
	}
	return sb.String()
}

func (a *MermaidDocumenterAgent) parseStructuredOutput(response string) (*StructuredOutput, error) {
	// Try to parse as JSON directly
	var output StructuredOutput
	if err := json.Unmarshal([]byte(response), &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (a *MermaidDocumenterAgent) argsToJSON(args map[string]interface{}) string {
	jsonBytes, _ := json.Marshal(args)
	return string(jsonBytes)
}

func (a *MermaidDocumenterAgent) logInteraction(conversation []map[string]interface{}, response string, output *StructuredOutput) {
	// TODO: Implement proper logging to logs.jsonl
	fmt.Printf("Step %d: %s (confidence: %.2f)\n", a.StepCount+1, output.Type, output.Confidence)
}

func (a *MermaidDocumenterAgent) processFinalManifest(manifest map[string]interface{}) {
	// TODO: Process and validate the final manifest
	fmt.Printf("Processing final manifest: %v\n", manifest)
}
