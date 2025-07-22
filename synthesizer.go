package go_as

import (
	"encoding/json"
	"fmt"
	"strings"

	mcpcore "github.com/mark3labs/mcp-go/mcp"
)

// Synthesizer is responsible for synthesizing the result of a tool call into a string.
type Synthesizer struct{}

// NewSynthesizer creates a new instance of the Synthesizer.
func NewSynthesizer() *Synthesizer {
	return &Synthesizer{}
}

// Synthesize takes a tool call result and returns a string representation of it.
func (s *Synthesizer) Synthesize(result *mcpcore.CallToolResult) (string, error) {
	if result.IsError {
		// Attempt to provide a more structured error message if the content is text.
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcpcore.TextContent); ok {
				return fmt.Sprintf("Tool returned an error: %s", textContent.Text), nil
			}
		}
		return "Tool returned an unspecified error.", nil
	}

	var contentBuilder strings.Builder
	for _, c := range result.Content {
		switch v := c.(type) {
		case mcpcore.TextContent:
			contentBuilder.WriteString(v.Text)
		default:
			// For any other content type, marshal it to a JSON string.
			// This is a good default for structured data that the LLM can often interpret.
			jsonContent, err := json.Marshal(c)
			if err != nil {
				return "", fmt.Errorf("failed to marshal content to JSON: %w", err)
			}
			contentBuilder.Write(jsonContent)
		}
	}

	return contentBuilder.String(), nil
}
