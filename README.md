# go-as

This Go module provides a framework for orchestrating interactions with Model Context Protocol (MCP) agents. It allows for dynamic connection to MCP servers, generation of execution plans based on user queries, and execution of those plans by calling tools exposed by the connected agents.

## Features

- **MCP Agent Management**: Connect to and manage multiple MCP agents.
- **LLM-driven Tool Calling**: Utilizes Large Language Models to intelligently select and execute tools based on natural language queries.

- **Query Orchestration**: Decompose user queries into executable plans.
- **Tool Execution**: Call tools exposed by MCP agents and process their results.
- **Extensible**: Designed to be extended with different LLM clients and planning strategies.
- **HTTP Server**: Provides an HTTP server to expose the orchestrator via a REST API.

## Installation

To use this module in your Go project, you can add it to your `go.mod` file:

```bash
go get your-repo/go-as
```

If you are developing locally, you can use a `replace` directive in your `go.mod` file:

```go
module your-project

go 1.23

require (
	go-as v0.0.0-00010101000000-000000000000
)

replace go-as => ../go-as
```

## Usage

### Running the Server

To run the HTTP server, you can use the following code:

```go
package main

import (
	"log/slog"
	"os"

	go_as "go-as"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	orchestrator, err := go_has.NewOrchestrator(&go_has.OrchestratorConfig{}, logger)
	if err != nil {
		logger.Error("failed to create orchestrator", "error", err)
		os.Exit(1)
	}

	// Connect to MCP agents here
	// Example:
	// filesysMCPExecPath := filepath.Join(wd, "..", "filesys_mcp", "filesys_mcp_exec")
	// if err := orchestrator.ManageMCP(&go_has.MCPConfig{
	// 	Alias: "fs",
	// 	Address: filesysMCPExecPath,
	// }); err != nil {
	// 	slog.Error("failed to connect to mcp", "error", err)
	// 	os.Exit(1)
	// }

	server := go_has.NewServer(orchestrator, logger)
	if err := server.Start(":8080"); err != nil {
		logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
```

### Sending a Request

You can send a request to the server using `curl`:

```bash
curl -X POST http://localhost:8080/orchestrate -d '{"query": "list the files in the current directory"}'
```

## API Reference

### `NewOrchestrator(config *OrchestratorConfig, logger *slog.Logger) (*Orchestrator, error)`

Initializes a new `Orchestrator` instance.

- `config`: Configuration for the orchestrator.
- `logger`: A `*slog.Logger` instance for logging.

### `(*Orchestrator) ExecuteTask(request *OrchestrationRequest, updateChan chan<- OrchestrationUpdate)`

Starts an orchestration job. It sends updates to the provided `updateChan` which provides real-time feedback on the task's progress and results.

- `request`: An `OrchestrationRequest` containing the user's query.
- `updateChan`: A channel to send `OrchestrationUpdate` messages.

### `(*Orchestrator) ManageMCP(config *MCPConfig) error`

Manages MCP connections.

- `config`: `MCPConfig` containing the alias and address of the MCP agent.

### `NewServer(orchestrator *Orchestrator, logger *slog.Logger) *Server`

Initializes a new `Server` instance.

- `orchestrator`: An instance of the `Orchestrator`.
- `logger`: A `*slog.Logger` instance for logging.

### `(*Server) Start(addr string) error`

Starts the HTTP server.

- `addr`: The address to listen on (e.g., ":8080").

## Configuration

### `OrchestratorConfig`

```go
type OrchestratorConfig struct {
	// Currently empty, but can be extended for LLM API keys, endpoints, etc.
}
```

### `MCPConfig`

```go
type MCPConfig struct {
	Alias   string `json:"alias"`
	Address string `json:"address"` // The address of the MCP agent (e.g., "tcp://localhost:8080")
}
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.