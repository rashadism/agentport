package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"charm.land/fantasy"

	"github.com/rashad/fantasydemo/internal/auth"
	"github.com/rashad/fantasydemo/internal/config"
	"github.com/rashad/fantasydemo/internal/handler"
	"github.com/rashad/fantasydemo/internal/mcp"
	"github.com/rashad/fantasydemo/internal/provider"
)

// Server represents the HTTP server
type Server struct {
	cfg        *config.Config
	mcpManager *mcp.Manager
	agent      fantasy.Agent
	httpServer *http.Server
}

// New creates a new server instance
func New(cfg *config.Config, systemPrompt string) (*Server, error) {
	ctx := context.Background()

	// Get model name with default
	modelName := cfg.RCAModelName
	if modelName == "" {
		modelName = "gpt-4o" // Default model
	}

	// Initialize language model using provider inference
	model, err := provider.InitLanguageModel(ctx, modelName, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize language model: %w", err)
	}

	slog.Info("Initialized model", "model", modelName)

	// Initialize MCP manager
	mcpManager := mcp.NewManager()

	// Get MCP server configs
	mcpServerConfigs := cfg.GetMCPServers()

	// Convert to mcp.Config slice
	var mcpConfigs []mcp.Config
	for _, s := range mcpServerConfigs {
		slog.Debug("MCP server configured", "name", s.Name, "url", s.URL)
		mcpConfigs = append(mcpConfigs, mcp.Config{
			Name:          s.Name,
			URL:           s.URL,
			Headers:       s.Headers,
			TLSSkipVerify: cfg.TLSInsecureSkipVerify,
		})
	}

	// Add OAuth headers if configured
	if cfg.IsOAuthConfigured() {
		token, err := auth.FetchOAuthToken(ctx, cfg)
		if err != nil {
			slog.Warn("Failed to fetch OAuth token", "error", err)
		} else {
			for i := range mcpConfigs {
				if mcpConfigs[i].Headers == nil {
					mcpConfigs[i].Headers = make(map[string]string)
				}
				mcpConfigs[i].Headers["Authorization"] = "Bearer " + token
			}
			slog.Debug("OAuth token acquired")
		}
	}

	// Initialize MCP connections
	mcpManager.Initialize(ctx, mcpConfigs)

	// Get all tools from MCP servers
	allTools := mcpManager.GetAllTools(ctx)
	slog.Info("MCP tools collected", "count", len(allTools))

	// Create agent
	agent := fantasy.NewAgent(
		model,
		fantasy.WithSystemPrompt(systemPrompt),
		fantasy.WithTools(allTools...),
	)

	return &Server{
		cfg:        cfg,
		mcpManager: mcpManager,
		agent:      agent,
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

	// Close MCP connections
	if err := s.mcpManager.Close(); err != nil {
		slog.Error("MCP close error", "error", err)
	}

	// Shutdown HTTP server
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	var req handler.AnalyzeRequest
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

	resp := handler.AnalyzeResponse{
		Response:   result.Response.Content.Text(),
		TotalSteps: len(result.Steps),
		Usage: handler.UsageInfo{
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
	json.NewEncoder(w).Encode(handler.AnalyzeResponse{Error: message})
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
