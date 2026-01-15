package provider

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

// Type represents a supported LLM provider
type Type string

// TODO: Test and add more providers
const (
	OpenAI    Type = "openai"
	Anthropic Type = "anthropic"
	Google    Type = "google"
)

// ModelInfo holds parsed model information
type ModelInfo struct {
	Provider Type
	ModelID  string
}

// ParseModel parses a model string that may include provider prefix.
// Examples:
//   - "gpt-5.2" -> infers openai
//   - "openai:o4-mini" -> explicit openai
//   - "claude-opus-4.5" -> infers anthropic
//   - "anthropic:claude-sonnet-4.5" -> explicit anthropic
//   - "gemini-3-pro" -> infers google
func ParseModel(model string) ModelInfo {
	// Check for explicit provider prefix (provider:model)
	if idx := strings.Index(model, ":"); idx > 0 {
		providerStr := model[:idx]
		modelID := model[idx+1:]
		return ModelInfo{
			Provider: Type(providerStr),
			ModelID:  modelID,
		}
	}

	// Infer provider from model name
	provider := InferProvider(model)
	return ModelInfo{
		Provider: provider,
		ModelID:  model,
	}
}

// InferProvider infers the provider from a model name based on common prefixes.
// Returns empty string if provider cannot be inferred.
//
// Inference rules:
//   - gpt-*, o1*, o3*, o4*, chatgpt* -> openai
//   - claude* -> anthropic
//   - gemini* -> google
func InferProvider(model string) Type {
	modelLower := strings.ToLower(model)

	switch {
	case strings.HasPrefix(modelLower, "gpt-"),
		strings.HasPrefix(modelLower, "o1"),
		strings.HasPrefix(modelLower, "o3"),
		strings.HasPrefix(modelLower, "o4"),
		strings.HasPrefix(modelLower, "chatgpt"):
		return OpenAI

	case strings.HasPrefix(modelLower, "claude"):
		return Anthropic

	case strings.HasPrefix(modelLower, "gemini"):
		return Google

	default:
		return ""
	}
}

// InitLanguageModel initializes a language model from a model string.
// The model string can be:
//   - Just the model ID: "gpt-5.2" (provider will be inferred)
//   - Provider:model format: "openai:o3" (explicit provider)
func InitLanguageModel(ctx context.Context, model string, cfg *config.Config) (fantasy.LanguageModel, error) {
	info := ParseModel(model)

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
func buildProvider(providerType Type, cfg *config.Config) (fantasy.Provider, error) {
	switch providerType {
	case OpenAI:
		return openai.New(openai.WithAPIKey(cfg.RCALLMAPIKey))

	case Anthropic:
		return anthropic.New(anthropic.WithAPIKey(cfg.RCALLMAPIKey))

	case Google:
		return google.New(google.WithGeminiAPIKey(cfg.RCALLMAPIKey))

	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerType)
	}
}
