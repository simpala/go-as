package go_as

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mcpcore "github.com/mark3labs/mcp-go/mcp"
)

func TestAgentExecution(t *testing.T) {
	// Mock LLM Server
	mockLLMServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatCompletionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		var resp ChatCompletionResponse
		if strings.Contains(req.Messages[0].Content, "Nexus Orchestrator") {
			// Phase 1: Orchestrator (Planning)
			resp = ChatCompletionResponse{
				Choices: []struct {
					Message      Message `json:"message"`
					FinishReason string  `json:"finish_reason"`
					Index        int     `json:"index"`
				}{
					{
						Message: Message{
							Role:    "assistant",
							Content: `<plan>
1. List files in the current directory.
</plan>`,
							ToolCalls: []ToolCall{
								{
									ID:   "call_123",
									Type: "function",
									Function: FunctionCall{
										Name:      "fs.list_directory",
										Arguments: `{"path": "."}`,
									},
								},
							},
						},
						FinishReason: "tool_calls",
						Index:        0,
					},
				},
			}
		} else if strings.Contains(req.Messages[0].Content, "Nexus") {
			// Phase 2: Nexus (Execution)
			resp = ChatCompletionResponse{
				Choices: []struct {
					Message      Message `json:"message"`
					FinishReason string  `json:"finish_reason"`
					Index        int     `json:"index"`
				}{
					{
						Message: Message{
							Role:    "assistant",
							Content: "Final answer",
							ToolCalls: nil,
						},
						FinishReason: "stop",
						Index:        0,
					},
				},
			}
		} else {
			// Reconnector phase
			resp = ChatCompletionResponse{
				Choices: []struct {
					Message      Message `json:"message"`
					FinishReason string  `json:"finish_reason"`
					Index        int     `json:"index"`
				}{
					{
						Message: Message{
							Role:    "assistant",
							Content: "Final answer",
						},
						FinishReason: "stop",
						Index:        0,
					},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	}))
	defer mockLLMServer.Close()

	// Mock MCP Client
	mockMCPClient := &MCPClient{
		callToolFunc: func(ctx context.Context, toolName string, args interface{}) (*mcpcore.CallToolResult, error) {
			assert.Equal(t, "list_directory", toolName)
			assert.Equal(t, map[string]interface{}{"path": "."}, args)
			return &mcpcore.CallToolResult{
				IsError: false,
				Content: []mcpcore.Content{
					mcpcore.TextContent{Text: "file1.txt"},
				},
			}, nil
		},
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	llmConfig := &LLMClientConfig{
		ServerURL: mockLLMServer.URL,
		ModelName: "test-model",
		Timeout:   5 * time.Second,
	}
	llmClient := NewLLMClient(llmConfig, logger)

	mcpClients := map[string]*MCPClient{"fs": mockMCPClient}
	availableTools := []Tool{
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "fs.list_directory",
				Description: "Lists files in a directory.",
				Parameters:  json.RawMessage(`{"type": "object", "properties": {"path": {"type": "string"}}, "required": ["path"]}`),
			},
		},
	}

	agent := NewAgent(llmClient, mcpClients, logger, availableTools)
	finalResult, err := agent.Execute(context.Background(), "list files in current directory")

	require.NoError(t, err)
	assert.Equal(t, "Final answer", finalResult)
}

func TestExtractContentBetweenTags(t *testing.T) {
	testCases := []struct {
		name          string
		text          string
		startTag      string
		endTag        string
		expected      string
		expectedFound bool
	}{
		{
			name:          "Simple case",
			text:          "<plan>hello</plan>",
			startTag:      "<plan>",
			endTag:        "</plan>",
			expected:      "hello",
			expectedFound: true,
		},
		{
			name:          "Case insensitive",
			text:          "<PLAN>hello</PLAN>",
			startTag:      "<plan>",
			endTag:        "</plan>",
			expected:      "hello",
			expectedFound: true,
		},
		{
			name:          "With surrounding text",
			text:          "Here is the plan: <plan>hello</plan>\nThat is all.",
			startTag:      "<plan>",
			endTag:        "</plan>",
			expected:      "hello",			
			expectedFound: true,
		},
		{
			name:          "No tags",
			text:          "hello",
			startTag:      "<plan>",
			endTag:        "</plan>",
			expected:      "",
			expectedFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, found := extractContentBetweenTags(tc.text, tc.startTag, tc.endTag)
			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.expectedFound, found)
		})
	}
}