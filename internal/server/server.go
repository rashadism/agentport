package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"charm.land/fantasy"

	"rca.agent/test/internal/agent"
	"rca.agent/test/internal/config"
)

// AnalyzeRequest represents the request body for /analyze
type AnalyzeRequest struct {
	Prompt string `json:"prompt"`
}

// AnalyzeResponse represents the response body for /analyze
type AnalyzeResponse struct {
	Response   string    `json:"response"`
	TotalSteps int       `json:"total_steps"`
	Usage      UsageInfo `json:"usage"`
	Error      string    `json:"error,omitempty"`
}

// UsageInfo represents token usage information
type UsageInfo struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	TotalTokens  int64 `json:"total_tokens"`
}

// Server represents the HTTP server
type Server struct {
	cfg        *config.Config
	agent      *agent.Agent
	httpServer *http.Server
}

// New creates a new server instance
func New(cfg *config.Config, systemPrompt string) (*Server, error) {
	a, err := agent.New(context.Background(), cfg, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return &Server{
		cfg:   cfg,
		agent: a,
	}, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("POST /analyze", s.handleAnalyze)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.cfg.ServerPort),
		Handler:      mux,
		ReadTimeout:  s.cfg.ReadTimeout,
		WriteTimeout: s.cfg.WriteTimeout,
	}

	slog.Info("Server starting", "port", s.cfg.ServerPort)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("Server shutting down")

	if err := s.agent.Close(); err != nil {
		slog.Error("Agent close error", "error", err)
	}

	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Prompt == "" {
		s.writeError(w, http.StatusBadRequest, "prompt is required")
		return
	}

	// Create context with analysis timeout
	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.AnalysisTimeout)
	defer cancel()

	startTime := time.Now()

	slog.Info("Starting analysis", "prompt", truncate(req.Prompt, 100))

	result, err := s.agent.Stream(ctx, fantasy.AgentStreamCall{
		Prompt: req.Prompt,
		OnAgentStart: func() {
			slog.Debug("Agent started")
		},
		OnAgentFinish: func(result *fantasy.AgentResult) error {
			slog.Debug("Agent finished", "steps", len(result.Steps), "tokens", result.TotalUsage.TotalTokens)
			return nil
		},
		OnStepStart: func(step int) error {
			slog.Debug("Model step", "step", step)
			return nil
		},
		OnToolCall: func(toolCall fantasy.ToolCallContent) error {
			slog.Debug("Tool call", "tool", toolCall.ToolName)
			return nil
		},
	})

	if err != nil {
		slog.Error("Analysis error", "error", err)
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	duration := time.Since(startTime)
	slog.Info("Analysis completed", "duration", duration)

	resp := AnalyzeResponse{
		Response:   result.Response.Content.Text(),
		TotalSteps: len(result.Steps),
		Usage: UsageInfo{
			InputTokens:  result.TotalUsage.InputTokens,
			OutputTokens: result.TotalUsage.OutputTokens,
			TotalTokens:  result.TotalUsage.TotalTokens,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(AnalyzeResponse{Error: message})
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
