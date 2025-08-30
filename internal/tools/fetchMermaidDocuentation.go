package tools

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type FetchMermaidDocumentationTool struct{}

func (t *FetchMermaidDocumentationTool) Name() string {
	return "fetchMermaidDocumentation"
}

func (t *FetchMermaidDocumentationTool) Description() string {
	return "Fetch Mermaid documentation and syntax information"
}

func (t *FetchMermaidDocumentationTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"topic": map[string]interface{}{
				"type":        "string",
				"description": "Specific Mermaid topic to search for (optional)",
			},
			"version": map[string]interface{}{
				"type":        "string",
				"description": "Mermaid version to get docs for (optional)",
			},
		},
	}
}

func (t *FetchMermaidDocumentationTool) Execute(args map[string]interface{}) ToolResult {
	var topic string

	if t, exists := args["topic"]; exists {
		if topicStr, ok := t.(string); ok {
			topic = topicStr
		}
	}

	// Version parameter is accepted but not used in current implementation
	if _, exists := args["version"]; exists {
		// Could be used for version-specific documentation in the future
	}

	// For now, we'll fetch from the official Mermaid documentation
	baseURL := "https://mermaid.js.org"

	var url string
	var content string
	var err error

	if topic != "" {
		// Try to construct a documentation URL for the topic
		url = fmt.Sprintf("%s/config/diagrams-and-syntaxes/%s.html", baseURL, strings.ToLower(topic))
		content, err = fetchURL(url)
		if err != nil {
			// Fallback to general documentation
			url = baseURL + "/config/diagrams-and-syntaxes.html"
			content, err = fetchURL(url)
		}
	} else {
		// Fetch general Mermaid documentation
		url = baseURL + "/config/diagrams-and-syntaxes.html"
		content, err = fetchURL(url)
	}

	if err != nil {
		return ToolResult{
			Success: false,
			Error:   "Failed to fetch Mermaid documentation: " + err.Error(),
		}
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"url":     url,
			"content": content,
		},
	}
}

func fetchURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
