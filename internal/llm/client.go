package llm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/vertex"
)

// Client wraps the Anthropic SDK client and provides streaming with
// 50ms chunk batching to keep the frontend event rate at ~20 events/sec.
type Client struct {
	apiClient anthropic.Client
	model     string
}

// NewClient creates a new LLM client with the given API key and model.
// If model is empty, DefaultModel ("claude-haiku-4-5") is used.
func NewClient(apiKey string, model string) *Client {
	if model == "" {
		model = DefaultModel
	}

	c := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &Client{
		apiClient: c,
		model:     model,
	}
}

// NewVertexClient creates a client that routes through Google Cloud Vertex AI.
// Auth uses Application Default Credentials (ADC) — run `gcloud auth application-default login`.
func NewVertexClient(ctx context.Context, region string, projectID string, model string) (client *Client, err error) {
	if model == "" {
		model = DefaultModel
	}
	if region == "" {
		return nil, fmt.Errorf("vertex region is required")
	}
	if projectID == "" {
		return nil, fmt.Errorf("vertex project ID is required")
	}

	defer func() {
		if r := recover(); r != nil {
			client = nil
			err = fmt.Errorf("vertex auth failed: %v", r)
		}
	}()

	c := anthropic.NewClient(
		vertex.WithGoogleAuth(ctx, region, projectID),
	)

	return &Client{
		apiClient: c,
		model:     model,
	}, nil
}

// newClientWithBaseURL creates a client pointing at a custom base URL.
// This is used by tests to redirect API calls to an httptest server.
// Retries are disabled to prevent test hangs on 429/5xx responses.
func newClientWithBaseURL(apiKey string, model string, baseURL string) *Client {
	if model == "" {
		model = DefaultModel
	}

	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithMaxRetries(0),
	}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	c := anthropic.NewClient(opts...)

	return &Client{
		apiClient: c,
		model:     model,
	}
}

// Stream opens a streaming connection to the Anthropic API and delivers
// content to the caller via the emit callback.
//
// Behavior:
//   - Creates a streaming Messages request using anthropic-sdk-go
//   - Batches incoming SSE text deltas at 50ms intervals before calling emit
//   - On completion, flushes remaining buffer and emits a final chunk with Done=true
//   - Respects context cancellation for user-initiated stream stop
//   - Returns a *StreamError on API failures (implements error interface)
//
// The emit callback receives StreamChunk values:
//   - During streaming: {Text: "...", Done: false}
//   - On completion:    {Text: "", Done: true, TotalTokens: N}
func (c *Client) Stream(ctx context.Context, req StreamRequest, emit func(StreamChunk)) error {
	model := req.Model
	if model == "" {
		model = c.model
	}

	params := anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: MaxTokens,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.Prompt)),
		},
	}

	stream := c.apiClient.Messages.NewStreaming(ctx, params)

	// Accumulate the full message for final token counts.
	message := anthropic.Message{}

	// Batching state: buffer collects text deltas between flush intervals.
	var mu sync.Mutex
	var buf strings.Builder
	var batchTimer *time.Timer

	// flush emits the current buffer as a StreamChunk and resets it.
	// Must be called with mu held. The emit call happens inside the lock
	// to prevent concurrent emissions from the timer and the main loop,
	// avoiding the Wails v2 event system data race (issue #2448).
	flush := func() {
		if buf.Len() > 0 {
			text := buf.String()
			buf.Reset()
			emit(StreamChunk{Text: text})
		}
	}

	// Process SSE events from the stream.
	for stream.Next() {
		event := stream.Current()
		message.Accumulate(event)

		switch ev := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			switch d := ev.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				mu.Lock()
				buf.WriteString(d.Text)
				if batchTimer == nil {
					batchTimer = time.AfterFunc(
						time.Duration(BatchIntervalMs)*time.Millisecond,
						func() {
							mu.Lock()
							flush()
							batchTimer = nil
							mu.Unlock()
						},
					)
				}
				mu.Unlock()
			}
		}
	}

	// Stop any pending batch timer and wait for it to complete.
	mu.Lock()
	if batchTimer != nil {
		batchTimer.Stop()
		batchTimer = nil
	}
	mu.Unlock()

	// Check for stream errors before flushing remaining content.
	if err := stream.Err(); err != nil {
		// Flush any partial content that arrived before the error.
		mu.Lock()
		flush()
		mu.Unlock()

		return classifyError(err)
	}

	// Flush remaining buffered content.
	mu.Lock()
	flush()
	mu.Unlock()

	// Emit the final done chunk with total token count.
	totalTokens := int(message.Usage.InputTokens + message.Usage.OutputTokens)
	emit(StreamChunk{
		Done:        true,
		TotalTokens: totalTokens,
	})

	return nil
}

// Translate performs a non-streaming translation of the given text to the
// target language. Uses Haiku for minimal latency regardless of the client's
// configured model. Returns the translated text or a classified error.
func (c *Client) Translate(ctx context.Context, text string, targetLang string) (string, error) {
	params := anthropic.MessageNewParams{
		Model:     ModelHaiku,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: fmt.Sprintf("Translate the following text to %s. Return only the translation, nothing else.", targetLang)},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(text)),
		},
	}

	msg, err := c.apiClient.Messages.New(ctx, params)
	if err != nil {
		return "", classifyError(err)
	}

	for _, block := range msg.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}

	return "", &StreamError{Code: ErrCodeAPI, Message: "no text in response"}
}

// classifyError maps SDK and transport errors to typed StreamError values.
// The error codes match the contract defined in spec/API.md:
//   - "network":    DNS failure, timeout, connection refused
//   - "auth":       invalid or expired API key (HTTP 401)
//   - "rate_limit": too many requests (HTTP 429)
//   - "cancelled":  user cancelled via CancelStream (context.Canceled)
//   - "api":        other Anthropic API errors (HTTP 4xx/5xx)
func classifyError(err error) *StreamError {
	// Context cancellation (user-initiated stop)
	if errors.Is(err, context.Canceled) {
		return &StreamError{
			Code:    ErrCodeCancelled,
			Message: "stream cancelled",
		}
	}

	// Context deadline exceeded
	if errors.Is(err, context.DeadlineExceeded) {
		return &StreamError{
			Code:    ErrCodeNetwork,
			Message: "request timed out",
		}
	}

	// Anthropic API errors (HTTP status-based)
	var apierr *anthropic.Error
	if errors.As(err, &apierr) {
		fmt.Fprintf(os.Stderr, "llm: API error (HTTP %d): %s\n", apierr.StatusCode, apierr.Error())
		switch apierr.StatusCode {
		case 401:
			return &StreamError{
				Code:    ErrCodeAuth,
				Message: "invalid or expired API key",
			}
		case 429:
			return &StreamError{
				Code:    ErrCodeRateLimit,
				Message: "rate limit exceeded",
			}
		default:
			msg := fmt.Sprintf("API error (HTTP %d)", apierr.StatusCode)
			if body := apierr.RawJSON(); body != "" {
				msg = fmt.Sprintf("API error (HTTP %d): %s", apierr.StatusCode, body)
			}
			return &StreamError{
				Code:    ErrCodeAPI,
				Message: msg,
			}
		}
	}

	// Transport-level errors (DNS, connection refused, etc.)
	// Log the full error for debugging but return a generic message
	// to avoid leaking internal hostnames, DNS details, or SDK internals.
	fmt.Fprintf(os.Stderr, "llm: network error: %v\n", err)
	return &StreamError{
		Code:    ErrCodeNetwork,
		Message: "connection failed",
	}
}
