package go_as

import (
	"encoding/json"
)

// JSONRPCRequest represents a generic JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"` // ID is optional for notifications
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a generic JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error object.
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// CallToolParams represents the parameters for a "tools/call" request.
type CallToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// CallToolResult represents the result of a "tools/call" response.
// This is a simplified version for the client side.
type CallToolResult struct {
	IsError           bool            `json:"isError"`
	Content           []Content       `json:"content"`
	StructuredContent json.RawMessage `json:"structuredContent"`
}

// Content represents a generic content block within a tool result.
type Content struct {
	Text string `json:"text,omitempty"`
	// Add other content types as needed (e.g., Image, HTML)
}

// ToolDefinition represents the structure of a tool definition as returned by "tools/list"
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// ToolListResult represents the result of a "tools/list" response.
type ToolListResult struct {
	Tools []ToolDefinition `json:"tools"`
}

// Notification represents a generic JSON-RPC 2.0 notification.
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}
