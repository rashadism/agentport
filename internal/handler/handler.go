package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"rca.agent/test/internal/agent"
)

// AnalysisService defines the interface for analysis operations.
type AnalysisService interface {
	Analyze(ctx context.Context, prompt string) (*agent.AnalysisResult, error)
}

// Handler handles HTTP requests.
type Handler struct {
	analysis AnalysisService
	timeout  time.Duration
}

// New creates a new handler.
func New(analysis AnalysisService, timeout time.Duration) *Handler {
	return &Handler{
		analysis: analysis,
		timeout:  timeout,
	}
}

// RegisterRoutes registers all routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /analyze", h.Analyze)
}

// Health handles health check requests.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Analyze handles analysis requests.
func (h *Handler) Analyze(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Prompt == "" {
		h.writeError(w, http.StatusBadRequest, "prompt is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	startTime := time.Now()

	result, err := h.analysis.Analyze(ctx, req.Prompt)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Info("Analysis completed", "duration", time.Since(startTime))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
