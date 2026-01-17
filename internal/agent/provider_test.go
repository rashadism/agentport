package agent

import (
	"testing"
)

func TestInferProvider(t *testing.T) {
	tests := []struct {
		model    string
		expected providerType
	}{
		// OpenAI
		{"gpt-5.2", providerOpenAI},
		{"gpt-5.1", providerOpenAI},
		{"gpt-4.1", providerOpenAI},
		{"gpt-4.1-mini", providerOpenAI},
		{"gpt-4o", providerOpenAI},
		{"o3", providerOpenAI},
		{"o3-pro", providerOpenAI},
		{"o4-mini", providerOpenAI},
		{"chatgpt-4o-latest", providerOpenAI},

		// Anthropic
		{"claude-opus-4.5", providerAnthropic},
		{"claude-sonnet-4.5", providerAnthropic},
		{"claude-haiku-4.5", providerAnthropic},
		{"claude-opus-4", providerAnthropic},
		{"claude-sonnet-4", providerAnthropic},

		// Google
		{"gemini-3-pro", providerGoogle},
		{"gemini-3-flash", providerGoogle},
		{"gemini-2.5-pro", providerGoogle},
		{"gemini-2.5-flash", providerGoogle},

		// Unknown
		{"unknown-model", ""},
		{"llama-3-70b", ""},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			got := inferProvider(tt.model)
			if got != tt.expected {
				t.Errorf("inferProvider(%q) = %q, want %q", tt.model, got, tt.expected)
			}
		})
	}
}

func TestParseModel(t *testing.T) {
	tests := []struct {
		model        string
		wantProvider providerType
		wantModelID  string
	}{
		// Explicit provider prefix
		{"openai:gpt-5.2", providerOpenAI, "gpt-5.2"},
		{"anthropic:claude-opus-4.5", providerAnthropic, "claude-opus-4.5"},
		{"google:gemini-3-pro", providerGoogle, "gemini-3-pro"},

		// No prefix - verify passthrough (inference tested separately)
		{"some-model", "", "some-model"},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			got := parseModel(tt.model)
			if got.Provider != tt.wantProvider {
				t.Errorf("parseModel(%q).Provider = %q, want %q", tt.model, got.Provider, tt.wantProvider)
			}
			if got.ModelID != tt.wantModelID {
				t.Errorf("parseModel(%q).ModelID = %q, want %q", tt.model, got.ModelID, tt.wantModelID)
			}
		})
	}
}
