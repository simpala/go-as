package go_as

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"sync"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	mcpcore "github.com/mark3labs/mcp-go/mcp"
)

// AgentCommand represents a command to be sent to an MCP agent.
type AgentCommand struct {
	AgentAlias string
	ToolName   string
	Args       interface{}
}

// MCPClient manages a single connection to an MCP agent via stdin/stdout.
type MCPClient struct {
	alias        string
	client       *mcpclient.Client
	cmd          *exec.Cmd
	logger       *slog.Logger
	mu           sync.Mutex
	callToolFunc func(ctx context.Context, toolName string, args interface{}) (*mcpcore.CallToolResult, error)
}

// NewMCPClient creates a new MCPClient and starts the agent process.
func NewMCPClient(alias string, command string, args []string, logger *slog.Logger) (*MCPClient, error) {
	if command == "" {
		return nil, fmt.Errorf("command cannot be empty for MCP client %s", alias)
	}

	mcpClient, err := mcpclient.NewStdioMCPClient(command, nil, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create stdio MCP client: %w", err)
	}

	client := &MCPClient{
		alias:  alias,
		cmd:    nil, // cmd is managed by transport, so we don't need it here
		client: mcpClient,
		logger: logger,
	}

	// Initialize the MCP client
	initRequest := mcpcore.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcpcore.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcpcore.Implementation{
		Name:    "go-as-orchestrator",
		Version: "1.0.0",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Add a timeout for initialization
	defer cancel()

	_, err = mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		client.Close() // Close the client if initialization fails
		return nil, fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	logger.Info("MCP client connected and initialized", "alias", alias, "command", command)

	return client, nil
}

// Close closes the client connection and stops the agent process.
func (c *MCPClient) Close() error {
	if c.client != nil {
		c.client.Close()
	}
	// cmd is managed by transport, so no need to kill it here
	return nil
}

// CallTool calls a tool on the MCP agent.
func (c *MCPClient) CallTool(ctx context.Context, toolName string, args interface{}) (*mcpcore.CallToolResult, error) {
	if c.callToolFunc != nil {
		return c.callToolFunc(ctx, toolName, args)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Marshal arguments to JSON
	argsBytes, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool arguments: %w", err)
	}

	result, err := c.client.CallTool(ctx, mcpcore.CallToolRequest{Params: mcpcore.CallToolParams{Name: toolName, Arguments: json.RawMessage(argsBytes)}})
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %w", err)
	}

	return result, nil
}

// GetTools makes an RPC call to the MCP agent to discover its supported tools.
func (c *MCPClient) GetTools(ctx context.Context) ([]mcpcore.Tool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tools, err := c.client.ListTools(ctx, mcpcore.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	return tools.Tools, nil
}
