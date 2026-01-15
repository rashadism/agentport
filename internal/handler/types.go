package handler

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
