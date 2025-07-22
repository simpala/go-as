package go_as

// OrchestratorConfig holds configuration for the Orchestrator.
type OrchestratorConfig struct {
	// Add configuration fields here
}

// MCPConfig holds configuration for a Managed Compute Provider (MCP).
type MCPConfig struct {
	Alias   string
	Command string
	Args    []string
}
