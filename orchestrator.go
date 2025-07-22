package go_as

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

// Orchestrator is the main struct for the module.
type Orchestrator struct {
	config     *OrchestratorConfig
	logger     *slog.Logger
	mcpClients map[string]*MCPClient // Use a map of MCPClient
	llmClient  *LLMClient
}

// NewOrchestrator creates a new instance of the orchestrator.
func NewOrchestrator(config *OrchestratorConfig, logger *slog.Logger) (*Orchestrator, error) {
	llmConfig := &LLMClientConfig{
		ServerURL: GetLLMServerURL(),
		ModelName: GetLLMModelName(),
		Timeout:   GetLLMTimeout(),
	}
	return &Orchestrator{
		config:     config,
		logger:     logger,
		mcpClients: make(map[string]*MCPClient), // Initialize the map
		llmClient:  NewLLMClient(llmConfig, logger),
	}, nil
}

// ExecuteTask executes an orchestration task based on the request.
func (o *Orchestrator) ExecuteTask(request *OrchestrationRequest, updateChan chan<- OrchestrationUpdate) {
	defer close(updateChan)

	o.logger.Info("Orchestrator: Starting task execution.", "query", request.Query)

	// 1. Fetch available tools from connected MCP agents
	var availableTools []Tool
	o.logger.Info("Orchestrator: Fetching available tools from MCP agents.")
	for alias, client := range o.mcpClients {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Get the tool definitions from the MCP agent
		mcpTools, err := client.GetTools(ctx)
		if err != nil {
			o.logger.Error("Orchestrator: Failed to get tools from MCP agent", "alias", alias, "error", err)
			continue
		}
		o.logger.Info("Orchestrator: Found tools for agent", "alias", alias, "count", len(mcpTools))

		// Convert MCP tool definitions to LLM-compatible Tool format
		for _, mcpTool := range mcpTools {

			// Construct the full tool name as "agentAlias.toolName"
			fullToolName := fmt.Sprintf("%s.%s", alias, mcpTool.Name)

			// We need to convert it to a raw JSON string for the LLM Tool.Parameters field
			paramsBytes, err := json.Marshal(mcpTool.InputSchema)
			if err != nil {
				o.logger.Error("Orchestrator: Failed to marshal tool input schema", "tool", fullToolName, "error", err)
				continue
			}

			availableTools = append(availableTools, Tool{
				Type: "function",
				Function: ToolFunction{
					Name:        fullToolName,
					Description: mcpTool.Description,
					Parameters:  json.RawMessage(paramsBytes),
				},
			})
		}
	}

	if len(availableTools) == 0 {
		updateChan <- OrchestrationUpdate{Type: "error", Content: "No tools available from connected agents.", Error: fmt.Errorf("no tools available")}
		o.logger.Error("Orchestrator: No tools available from connected agents.")
		return
	}
	o.logger.Info("Orchestrator: Total available tools", "count", len(availableTools))

	// 2. Create and execute the agent
	o.logger.Info("Orchestrator: Creating and executing agent.")
	agent := NewAgent(o.llmClient, o.mcpClients, o.logger, availableTools)
	finalResult, err := agent.Execute(context.Background(), request.Query)
	if err != nil {
		updateChan <- OrchestrationUpdate{Type: "error", Content: fmt.Sprintf("Agent execution failed: %v", err), Error: err}
		o.logger.Error("Orchestrator: Agent execution failed.", "error", err)
		return
	}

	updateChan <- OrchestrationUpdate{Type: "result", Content: finalResult}
	o.logger.Info("Orchestrator: Task completed successfully.", "result", finalResult)
}

// ManageMCP manages the lifecycle and configuration of an MCP.
func (o *Orchestrator) ManageMCP(config *MCPConfig) error {
	o.logger.Info("Orchestrator: Connecting MCP", "alias", config.Alias, "command", config.Command)
	client, err := NewMCPClient(config.Alias, config.Command, config.Args, o.logger)
	if err != nil {
		return fmt.Errorf("failed to create MCP client for %s: %w", config.Alias, err)
	}
	o.mcpClients[config.Alias] = client
	o.logger.Info("Orchestrator: MCP connected successfully.", "alias", config.Alias)
	return nil
}
