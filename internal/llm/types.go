package llm

import "fmt"

// StreamRequest is the request payload for initiating an LLM stream.
type StreamRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
}

// StreamChunk is a batch of streamed content. It is the payload for
// llm:chunk and llm:done events sent from the Go backend to the frontend.
type StreamChunk struct {
	Text        string `json:"text"`
	Done        bool   `json:"done"`
	TotalTokens int    `json:"totalTokens,omitempty"`
}

// StreamError represents a classified error from the streaming API.
// It implements the error interface so it can be returned from Stream().
type StreamError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *StreamError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Credentials is the JSON structure for the fallback credentials file
// at ~/.config/ais/credentials.json. The file must have 0600 permissions.
type Credentials struct {
	AnthropicAPIKey string `json:"anthropic_api_key"`
}

// Error code constants for StreamError.Code values.
const (
	ErrCodeNetwork   = "network"
	ErrCodeAuth      = "auth"
	ErrCodeRateLimit = "rate_limit"
	ErrCodeCancelled = "cancelled"
	ErrCodeAPI       = "api"
)

// Model identifiers for supported Claude models.
// These match the model IDs from the Anthropic API.
const (
	ModelHaiku   = "claude-haiku-4-5"
	ModelSonnet  = "claude-sonnet-5"
	ModelOpus    = "claude-opus-5"
	DefaultModel = ModelHaiku
)

// AvailableModels is the ordered list of supported models (cheapest first).
var AvailableModels = []string{ModelHaiku, ModelSonnet, ModelOpus}

// MaxTokens is the maximum number of output tokens per stream request.
const MaxTokens = 4096

// batchInterval is the duration between chunk emissions to the frontend.
// Multiple SSE events arriving within this window are concatenated into
// a single StreamChunk, keeping the frontend event rate at ~20/sec.
const batchIntervalMs = 50
