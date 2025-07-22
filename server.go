package go_as

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Server is the HTTP server for the go-as module.
type Server struct {
	orchestrator *Orchestrator
	logger       *slog.Logger
}

// NewServer creates a new instance of the Server.
func NewServer(orchestrator *Orchestrator, logger *slog.Logger) *Server {
	return &Server{
		orchestrator: orchestrator,
		logger:       logger,
	}
}

// Start starts the HTTP server.
func (s *Server) Start(addr string) error {
	http.HandleFunc("/orchestrate", s.handleOrchestrate)
	s.logger.Info("Server listening on", "addr", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleOrchestrate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req OrchestrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updateChan := make(chan OrchestrationUpdate)
	go s.orchestrator.ExecuteTask(&req, updateChan)

	for update := range updateChan {
		if update.Type == "result" || update.Type == "error" {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(update); err != nil {
				s.logger.Error("Failed to write response", "error", err)
			}
			return
		}
	}
}
