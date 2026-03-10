// Package claude implements the LLM port using the Anthropic Claude Messages API.
package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/port"
)

const messagesEndpoint = "https://api.anthropic.com/v1/messages"

// Adapter implements port.LLM using the Anthropic Claude Messages API.
type Adapter struct {
	apiKey string
	client *http.Client
	logger *zerolog.Logger
}

// NewClaudeAdapter creates a new Claude LLM adapter.
func NewClaudeAdapter(apiKey string, logger *zerolog.Logger) port.LLM {
	return &Adapter{
		apiKey: apiKey,
		client: &http.Client{},
		logger: logger,
	}
}

// messagesRequest is the request body for the Claude Messages API.
type messagesRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []message `json:"messages"`
}

// message represents a single message in the conversation.
type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// messagesResponse is the response from the Claude Messages API.
type messagesResponse struct {
	Content []contentBlock `json:"content"`
	Usage   usage          `json:"usage"`
}

// contentBlock represents a content block in the response.
type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// usage contains token usage information.
type usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// errorResponse represents an API error response.
type errorResponse struct {
	Error errorDetail `json:"error"`
}

// errorDetail contains error details.
type errorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Generate produces a text response from the given system and user prompts using Claude.
func (a *Adapter) Generate(ctx context.Context, systemPrompt string, userPrompt string) (string, int, error) {
	reqBody := messagesRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages: []message{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal claude request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, messagesEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create claude request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to call claude: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read claude response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if json.Unmarshal(respBody, &errResp) == nil {
			return "", 0, fmt.Errorf("failed to generate with claude: %s: %s", errResp.Error.Type, errResp.Error.Message)
		}
		return "", 0, fmt.Errorf("failed to generate with claude: status %d", resp.StatusCode)
	}

	var msgResp messagesResponse
	if err := json.Unmarshal(respBody, &msgResp); err != nil {
		return "", 0, fmt.Errorf("failed to unmarshal claude response: %w", err)
	}

	if len(msgResp.Content) == 0 {
		return "", 0, fmt.Errorf("failed to generate with claude: empty response")
	}

	text := ""
	for _, block := range msgResp.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}

	tokens := msgResp.Usage.InputTokens + msgResp.Usage.OutputTokens

	a.logger.Debug().Int("tokens", tokens).Msg("Claude generation completed")
	return text, tokens, nil
}

// Provider returns the name of the LLM provider.
func (a *Adapter) Provider() string {
	return "claude"
}
