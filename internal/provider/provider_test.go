package provider

import (
	"testing"
)

func TestInferProvider(t *testing.T) {
	tests := []struct {
		model    string
		expected Type
	}{
		// OpenAI
		{"gpt-5.2", OpenAI},
		{"gpt-5.1", OpenAI},
		{"gpt-4.1", OpenAI},
		{"gpt-4.1-mini", OpenAI},
		{"gpt-4o", OpenAI},
		{"o3", OpenAI},
		{"o3-pro", OpenAI},
		{"o4-mini", OpenAI},
		{"chatgpt-4o-latest", OpenAI},

		// Anthropic
		{"claude-opus-4.5", Anthropic},
		{"claude-sonnet-4.5", Anthropic},
		{"claude-haiku-4.5", Anthropic},
		{"claude-opus-4", Anthropic},
		{"claude-sonnet-4", Anthropic},

		// Google
		{"gemini-3-pro", Google},
		{"gemini-3-flash", Google},
		{"gemini-2.5-pro", Google},
		{"gemini-2.5-flash", Google},

		// Unknown
		{"unknown-model", ""},
		{"llama-3-70b", ""},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			got := InferProvider(tt.model)
			if got != tt.expected {
				t.Errorf("InferProvider(%q) = %q, want %q", tt.model, got, tt.expected)
			}
		})
	}
}

func TestParseModel(t *testing.T) {
	tests := []struct {
		model        string
		wantProvider Type
		wantModelID  string
	}{
		// Explicit provider prefix
		{"openai:gpt-5.2", OpenAI, "gpt-5.2"},
		{"anthropic:claude-opus-4.5", Anthropic, "claude-opus-4.5"},
		{"google:gemini-3-pro", Google, "gemini-3-pro"},

		// No prefix - verify passthrough (inference tested separately)
		{"some-model", "", "some-model"},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			got := ParseModel(tt.model)
			if got.Provider != tt.wantProvider {
				t.Errorf("ParseModel(%q).Provider = %q, want %q", tt.model, got.Provider, tt.wantProvider)
			}
			if got.ModelID != tt.wantModelID {
				t.Errorf("ParseModel(%q).ModelID = %q, want %q", tt.model, got.ModelID, tt.wantModelID)
			}
		})
	}
}
