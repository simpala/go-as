package go_as

import (
	"context"
	"fmt"
)

// Reconnector is responsible for reconnecting the final results to the main LLM instance.
type Reconnector struct {
	llmClient *LLMClient
}

// NewReconnector creates a new instance of the Reconnector.
func NewReconnector(llmClient *LLMClient) *Reconnector {
	return &Reconnector{
		llmClient: llmClient,
	}
}

// Reconnect takes the conversation history and returns the final response from the LLM.
func (r *Reconnector) Reconnect(ctx context.Context, history []Message) (string, error) {
	// Find the last assistant message with content
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		if msg.Role == "assistant" && msg.Content != "" {
			return msg.Content, nil
		}
	}
	return "", fmt.Errorf("no final answer found in conversation history")
}
