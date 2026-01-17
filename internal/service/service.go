package service

import (
	"context"

	"rca.agent/test/internal/agent"
	"rca.agent/test/internal/config"
)

// DefaultMaxSteps is the maximum number of agent steps.
const DefaultMaxSteps = 30

// DefaultSystemPrompt is the default system prompt for the analysis service.
const DefaultSystemPrompt = `
You are very helpful assistant in openchoreo. You can help users with queries regarding how to get something done in openchoreo. The entity hierarchy in openchoreo is as follows:
Organization -> Project -> Component

Use your tools as needed. CRITICAL: Always use the todos tool to keep track of your task, and always update it as you make progress.

When you are ready to respond, you MUST call the structured_output tool to submit your response.
`

// AnalysisOutput is the default structured output schema.
type AnalysisOutput struct {
	Summary     string    `json:"summary" description:"Brief summary of findings"`
	Findings    []Finding `json:"findings" description:"Detailed findings from the analysis"`
	Suggestions []string  `json:"suggestions,omitempty" description:"Recommended actions or next steps"`
}

// Finding represents a single finding.
type Finding struct {
	Entity      string `json:"entity" description:"The entity this finding relates to (org, project, component, etc.)"`
	Description string `json:"description" description:"Description of the finding"`
	Severity    string `json:"severity,omitempty" description:"Severity level" enum:"info,low,medium,high,critical"`
}

// AnalysisService provides analysis capabilities.
type AnalysisService struct {
	agent *agent.Agent
}

// NewAnalysisService creates a new analysis service.
func NewAnalysisService(ctx context.Context, cfg *config.Config) (*AnalysisService, error) {
	a, err := agent.New(ctx, cfg, agent.Options{
		SystemPrompt: DefaultSystemPrompt,
		OutputSchema: AnalysisOutput{},
		MaxSteps:     DefaultMaxSteps,
	})
	if err != nil {
		return nil, err
	}

	return &AnalysisService{agent: a}, nil
}

// Analyze runs an analysis with the given prompt.
func (s *AnalysisService) Analyze(ctx context.Context, prompt string) (*agent.AnalysisResult, error) {
	return s.agent.Analyze(ctx, prompt)
}

// Close cleans up resources.
func (s *AnalysisService) Close() error {
	return s.agent.Close()
}
