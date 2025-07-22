package go_as

// OrchestrationRequest represents a request to the Orchestrator.
type OrchestrationRequest struct {
	Query string
	// Add other request fields here
}

// OrchestrationUpdate represents an update or result from the Orchestrator.
type OrchestrationUpdate struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Error   error  `json:"error,omitempty"`
}
