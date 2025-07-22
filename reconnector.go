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
	messages := append(history, Message{
		Role:    "user",
		Content: "Please provide a summary or a final answer based on the conversation history.",
	})

	llmResponse, err := r.llmClient.CallChatCompletionWithToolChoice(ctx, messages, nil, "none")
	if err != nil {
		return "", fmt.Errorf("failed to get final response from LLM: %w", err)
	}

	return llmResponse.Choices[0].Message.Content, nil
}
