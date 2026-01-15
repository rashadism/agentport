package agent

import (
	"context"
	"log/slog"

	"charm.land/fantasy"

	"rca.agent/test/internal/auth"
	"rca.agent/test/internal/config"
	"rca.agent/test/internal/mcp"
	"rca.agent/test/internal/provider"
)

// Agent holds the fantasy agent and its dependencies.
type Agent struct {
	fantasy.Agent
	MCPManager *mcp.Manager
}

// New creates a new Agent with MCP tools.
func New(ctx context.Context, cfg *config.Config, systemPrompt string) (*Agent, error) {
	// Initialize language model
	modelName := cfg.RCAModelName

	model, err := provider.InitLanguageModel(ctx, modelName, cfg)
	if err != nil {
		return nil, err
	}
	slog.Info("Initialized model", "model", modelName)

	// Initialize MCP manager
	mcpManager := mcp.NewManager()
	mcpConfigs := buildMCPConfigs(ctx, cfg)
	mcpManager.Initialize(ctx, mcpConfigs)

	// Get filtered tools
	tools := mcpManager.GetAllTools(ctx)

	// Create agent
	agent := fantasy.NewAgent(
		model,
		fantasy.WithSystemPrompt(systemPrompt),
		fantasy.WithTools(tools...),
	)

	return &Agent{
		Agent:      agent,
		MCPManager: mcpManager,
	}, nil
}

// Close cleans up resources.
func (a *Agent) Close() error {
	return a.MCPManager.Close()
}

func buildMCPConfigs(ctx context.Context, cfg *config.Config) []mcp.Config {
	var configs []mcp.Config

	for _, s := range cfg.GetMCPServers() {
		slog.Debug("MCP server configured", "name", s.Name, "url", s.URL)
		configs = append(configs, mcp.Config{
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
			for i := range configs {
				if configs[i].Headers == nil {
					configs[i].Headers = make(map[string]string)
				}
				configs[i].Headers["Authorization"] = "Bearer " + token
			}
			slog.Debug("OAuth token acquired")
		}
	}

	return configs
}
