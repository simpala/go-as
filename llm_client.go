package go_has

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// LLMClientConfig holds configuration for the LLM client.
type LLMClientConfig struct {
	ServerURL string
	ModelName string
	Timeout   time.Duration
}

// Message represents a message in the chat completion.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call made by the LLM.
type ToolCall struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall represents a function call within a tool call.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string of arguments
}

// Tool represents a tool definition for the LLM.
type Tool struct {
	Type     string `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction represents the function details of a tool.
type ToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters"` // JSON Schema
}

// LLMClient interacts with the LLM API.
type LLMClient struct {
	config *LLMClientConfig
	logger *slog.Logger
	client *http.Client
}

// NewLLMClient creates a new LLMClient.
func NewLLMClient(config *LLMClientConfig, logger *slog.Logger) *LLMClient {
	client := &http.Client{Timeout: config.Timeout}
	return &LLMClient{
		config: config,
		logger: logger,
		client: client,
	}
}

// ChatCompletionRequest represents the request body for chat completions.
type ChatCompletionRequest struct {
	Model      string      `json:"model"`
	Messages   []Message   `json:"messages"`
	Tools      []Tool      `json:"tools,omitempty"`
	ToolChoice interface{} `json:"tool_choice,omitempty"`
	Stream     bool        `json:"stream,omitempty"`
}

// ChatCompletionResponse represents the response body for chat completions.
type ChatCompletionResponse struct {
	Choices []struct {
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
		Index        int     `json:"index"`
	} `json:"choices"`
}

// ChatCompletionStreamChunk represents a chunk in a streaming chat completion.
type ChatCompletionStreamChunk struct {
	Choices []struct {
		Delta Delta `json:"delta"`
	} `json:"choices"`
}

// Delta represents a change in content in a streaming response.
type Delta struct {
	Content   string `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// CallChatCompletion sends a chat completion request to the LLM.
func (c *LLMClient) CallChatCompletion(ctx context.Context, messages []Message, tools []Tool) (*ChatCompletionResponse, error) {
	return c.CallChatCompletionWithToolChoice(ctx, messages, tools, nil)
}

// CallChatCompletionWithToolChoice sends a chat completion request to the LLM with a tool choice.
func (c *LLMClient) CallChatCompletionWithToolChoice(ctx context.Context, messages []Message, tools []Tool, toolChoice interface{}) (*ChatCompletionResponse, error) {
	requestBody, err := json.Marshal(ChatCompletionRequest{
		Model:      c.config.ModelName,
		Messages:   messages,
		Tools:      tools,
		ToolChoice: toolChoice,
		Stream:     false,
	})
	if err != nil {
		return nil, fmt.Errorf("could not marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.ServerURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	c.logger.Info("Sending LLM request", "url", c.config.ServerURL, "model", c.config.ModelName)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-OK status: %d, body: %s", resp.StatusCode, respBody)
	}

	var llmResponse ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResponse); err != nil {
		return nil, fmt.Errorf("could not decode response body: %w", err)
	}

	if len(llmResponse.Choices) == 0 {
		return nil, fmt.Errorf("no choices in LLM response")
	}

	return &llmResponse, nil
}

// StreamChatCompletion sends a streaming chat completion request to the LLM.
func (c *LLMClient) StreamChatCompletion(ctx context.Context, messages []Message, tools []Tool, chunkChan chan<- string) error {
	requestBody, err := json.Marshal(ChatCompletionRequest{
		Model:    c.config.ModelName,
		Messages: messages,
		Tools:    tools,
		Stream:   true,
	})
	if err != nil {
		return fmt.Errorf("could not marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.ServerURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	c.logger.Info("Sending streaming LLM request", "url", c.config.ServerURL, "model", c.config.ModelName)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("non-OK status: %d, body: %s", resp.StatusCode, respBody)
	}

	reader := bufio.NewReader(resp.Body)
	var buffer []byte

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		chunk := make([]byte, 512) // Read in chunks
		n, readErr := reader.Read(chunk)

		if n > 0 {
			buffer = append(buffer, chunk[:n]...)
			for {
				line, extractErr := extractLine(&buffer)
				if extractErr == io.EOF {
					break
				}
				if extractErr != nil {
					return fmt.Errorf("error extracting line from stream: %w", extractErr)
				}

				line = strings.TrimSpace(line)
				if line == "data: [DONE]" {
					return nil // Stream finished
				}

				if strings.HasPrefix(line, "data:") {
					jsonStr := strings.TrimPrefix(line, "data:")
					jsonStr = strings.TrimSpace(jsonStr)
					var streamChunk ChatCompletionStreamChunk
					if unmarshalErr := json.Unmarshal([]byte(jsonStr), &streamChunk); unmarshalErr != nil {
						c.logger.Warn("Warning: Error unmarshaling JSON chunk", "error", unmarshalErr, "data", jsonStr)
						continue
					}
					for _, choice := range streamChunk.Choices {
						if choice.Delta.Content != "" {
							chunkChan <- choice.Delta.Content
						}
					}
				}
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				return nil // End of stream
			}
			return fmt.Errorf("error reading stream: %w", readErr)
		}
	}
}

func extractLine(buffer *[]byte) (string, error) {
	idx := bytes.IndexByte(*buffer, '\n')
	if idx == -1 {
		return "", io.EOF
	}
	line := (*buffer)[:idx+1]
	*buffer = (*buffer)[idx+1:]
	return string(line), nil
}

// GetLLMModelName retrieves the LLM model name from environment variable or returns default.
func GetLLMModelName() string {
	modelName := os.Getenv("LLM_MODEL")
	if modelName == "" {
		modelName = "llama3.1" // Default model name
	}
	return modelName
}

// GetLLMServerURL retrieves the LLM server URL from environment variable or returns default.
func GetLLMServerURL() string {
	serverURL := os.Getenv("LLM_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://127.0.0.1:8084/v1/chat/completions" // Default server URL
	}
	return serverURL
}

// GetLLMTimeout retrieves the LLM timeout from environment variable or returns default.
func GetLLMTimeout() time.Duration {
	timeoutStr := os.Getenv("LLM_TIMEOUT_SECONDS")
	timeout := 60 * time.Second // Default timeout
	if timeoutStr != "" {
		if seconds, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = time.Duration(seconds) * time.Second
		}
	}
	return timeout
}
