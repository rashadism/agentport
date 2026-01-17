package agent

import (
	"context"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/anthropic"
	"charm.land/fantasy/providers/google"
	"charm.land/fantasy/providers/openai"

	"rca.agent/test/internal/config"
)

// providerType represents a supported LLM provider
type providerType string

// TODO: Test and add more providers
const (
	providerOpenAI    providerType = "openai"
	providerAnthropic providerType = "anthropic"
	providerGoogle    providerType = "google"
)

// modelInfo holds parsed model information
type modelInfo struct {
	Provider providerType
	ModelID  string
}

// parseModel parses a model string that may include provider prefix.
// Examples:
//   - "gpt-5.2" -> infers openai
//   - "openai:o4-mini" -> explicit openai
//   - "claude-opus-4.5" -> infers anthropic
//   - "anthropic:claude-sonnet-4.5" -> explicit anthropic
//   - "gemini-3-pro" -> infers google
func parseModel(model string) modelInfo {
	// Check for explicit provider prefix (provider:model)
	if idx := strings.Index(model, ":"); idx > 0 {
		providerStr := model[:idx]
		modelID := model[idx+1:]
		return modelInfo{
			Provider: providerType(providerStr),
			ModelID:  modelID,
		}
	}

	// Infer provider from model name
	provider := inferProvider(model)
	return modelInfo{
		Provider: provider,
		ModelID:  model,
	}
}

// inferProvider infers the provider from a model name based on common prefixes.
// Returns empty string if provider cannot be inferred.
//
// Inference rules:
//   - gpt-*, o1*, o3*, o4*, chatgpt* -> openai
//   - claude* -> anthropic
//   - gemini* -> google
func inferProvider(model string) providerType {
	modelLower := strings.ToLower(model)

	switch {
	case strings.HasPrefix(modelLower, "gpt-"),
		strings.HasPrefix(modelLower, "o1"),
		strings.HasPrefix(modelLower, "o3"),
		strings.HasPrefix(modelLower, "o4"),
		strings.HasPrefix(modelLower, "chatgpt"):
		return providerOpenAI

	case strings.HasPrefix(modelLower, "claude"):
		return providerAnthropic

	case strings.HasPrefix(modelLower, "gemini"):
		return providerGoogle

	default:
		return ""
	}
}

// initLanguageModel initializes a language model from a model string.
// The model string can be:
//   - Just the model ID: "gpt-5.2" (provider will be inferred)
//   - Provider:model format: "openai:o3" (explicit provider)
func initLanguageModel(ctx context.Context, model string, cfg *config.Config) (fantasy.LanguageModel, error) {
	info := parseModel(model)

	p, err := buildProvider(info.Provider, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build provider %s: %w", info.Provider, err)
	}

	lm, err := p.LanguageModel(ctx, info.ModelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get language model %s: %w", info.ModelID, err)
	}

	return lm, nil
}

// buildProvider creates a Fantasy provider based on the provider type
func buildProvider(pt providerType, cfg *config.Config) (fantasy.Provider, error) {
	switch pt {
	case providerOpenAI:
		return openai.New(openai.WithAPIKey(cfg.RCALLMAPIKey))

	case providerAnthropic:
		return anthropic.New(anthropic.WithAPIKey(cfg.RCALLMAPIKey))

	case providerGoogle:
		return google.New(google.WithGeminiAPIKey(cfg.RCALLMAPIKey))

	default:
		return nil, fmt.Errorf("unsupported provider: %s", pt)
	}
}
