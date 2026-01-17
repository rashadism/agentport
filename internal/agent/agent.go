package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"charm.land/fantasy"

	"rca.agent/test/internal/auth"
	"rca.agent/test/internal/config"
	"rca.agent/test/internal/mcp"
	"rca.agent/test/internal/tools"
)

// Options configures the agent.
type Options struct {
	SystemPrompt string
	OutputSchema any // If set, enables structured output with this schema
	MaxSteps     int // Maximum number of agent steps
}

// AnalysisResult is the result of an analysis.
type AnalysisResult struct {
	Output     any    `json:"output,omitempty"` // Structured output (if OutputSchema was set)
	Text       string `json:"text,omitempty"`   // Raw text output
	TotalSteps int    `json:"total_steps"`
	Usage      Usage  `json:"usage"`
}

// Usage represents token usage information.
type Usage struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	TotalTokens  int64 `json:"total_tokens"`
}

// Agent holds the fantasy agent and its dependencies.
type Agent struct {
	agent        fantasy.Agent
	mcpManager   *mcp.Manager
	outputSchema any
}

// New creates a new Agent with MCP tools.
func New(ctx context.Context, cfg *config.Config, opts Options) (*Agent, error) {
	// Initialize language model
	modelName := cfg.RCAModelName

	model, err := initLanguageModel(ctx, modelName, cfg)
	if err != nil {
		return nil, err
	}
	slog.Info("Initialized model", "model", modelName)

	// Initialize MCP manager
	mcpManager := mcp.NewManager()
	mcpConfigs := buildMCPConfigs(ctx, cfg)
	mcpManager.Initialize(ctx, mcpConfigs)

	// Get and filter MCP tools
	mcpTools := filterMCPTools(mcpManager.GetAllTools(ctx))
	slog.Info("MCP tools filtered", "allowed", len(mcpTools))

	// Add native tools
	allTools := append(mcpTools, tools.NewTodosTool())

	// Build stop conditions
	stopConditions := []fantasy.StopCondition{fantasy.StepCountIs(opts.MaxSteps)}

	// Add structured output tool if schema provided (workaround until json mode is supported)
	if opts.OutputSchema != nil {
		allTools = append(allTools, tools.NewStructuredOutputTool(opts.OutputSchema))
		stopConditions = append(stopConditions, fantasy.HasToolCall(tools.StructuredOutputToolName))
	}

	agentOpts := []fantasy.AgentOption{
		fantasy.WithSystemPrompt(opts.SystemPrompt),
		fantasy.WithTools(allTools...),
		fantasy.WithStopConditions(stopConditions...),
	}

	agent := fantasy.NewAgent(model, agentOpts...)

	return &Agent{
		agent:        agent,
		mcpManager:   mcpManager,
		outputSchema: opts.OutputSchema,
	}, nil
}

// Analyze runs the analysis and returns a structured result.
func (a *Agent) Analyze(ctx context.Context, prompt string) (*AnalysisResult, error) {
	slog.Info("Starting analysis", "prompt", truncate(prompt, 100))

	result, err := a.agent.Stream(ctx, fantasy.AgentStreamCall{
		Prompt: prompt,
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
			var input any
			if err := json.Unmarshal([]byte(toolCall.Input), &input); err == nil {
				slog.Debug("Tool call",
					"tool", toolCall.ToolName,
					"id", toolCall.ToolCallID,
					"input", input)
			} else {
				slog.Debug("Tool call", "tool", toolCall.ToolName, "input", toolCall.Input)
			}
			return nil
		},
		OnToolResult: func(result fantasy.ToolResultContent) error {
			var text string
			if textResult, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentText](result.Result); ok {
				text = textResult.Text
			} else if errResult, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentError](result.Result); ok {
				text = errResult.Error.Error()
			} else {
				text = fmt.Sprintf("%v", result.Result)
			}

			var output any
			if err := json.Unmarshal([]byte(text), &output); err == nil {
				slog.Debug("Tool result",
					"tool", result.ToolName,
					"id", result.ToolCallID,
					"result", output)
			} else {
				if len(text) > 500 {
					text = text[:500] + "...[truncated]"
				}
				slog.Debug("Tool result", "tool", result.ToolName, "result", text)
			}
			return nil
		},
	})

	if err != nil {
		slog.Error("Analysis error", "error", err)
		return nil, err
	}

	return a.buildResult(result), nil
}

func (a *Agent) buildResult(result *fantasy.AgentResult) *AnalysisResult {
	analysisResult := &AnalysisResult{
		TotalSteps: len(result.Steps),
		Usage: Usage{
			InputTokens:  result.TotalUsage.InputTokens,
			OutputTokens: result.TotalUsage.OutputTokens,
			TotalTokens:  result.TotalUsage.TotalTokens,
		},
	}

	// Extract structured output if schema was provided
	if a.outputSchema != nil && len(result.Steps) > 0 {
		lastStep := result.Steps[len(result.Steps)-1]
		for _, content := range lastStep.Response.Content {
			if toolCall, ok := fantasy.AsContentType[fantasy.ToolCallContent](content); ok {
				if toolCall.ToolName == tools.StructuredOutputToolName {
					var output any
					if err := json.Unmarshal([]byte(toolCall.Input), &output); err == nil {
						analysisResult.Output = output
					}
					break
				}
			}
		}
	}

	// Fallback to text response
	if analysisResult.Output == nil {
		analysisResult.Text = result.Response.Content.Text()
	}

	return analysisResult
}

// Close cleans up resources.
func (a *Agent) Close() error {
	return a.mcpManager.Close()
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

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
