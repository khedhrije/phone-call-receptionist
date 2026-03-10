// Package deepseek implements the LLM port using the DeepSeek OpenAI-compatible API.
package deepseek

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

const endpoint = "https://api.deepseek.com/v1/chat/completions"

// Adapter implements port.LLM using the DeepSeek API.
type Adapter struct {
	apiKey string
	client *http.Client
	logger *zerolog.Logger
}

// NewDeepSeekAdapter creates a new DeepSeek LLM adapter.
func NewDeepSeekAdapter(apiKey string, logger *zerolog.Logger) port.LLM {
	return &Adapter{
		apiKey: apiKey,
		client: &http.Client{},
		logger: logger,
	}
}

// chatRequest is the OpenAI-compatible request body.
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

// chatMessage represents a message in the conversation.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse is the OpenAI-compatible response body.
type chatResponse struct {
	Choices []chatChoice `json:"choices"`
	Usage   chatUsage    `json:"usage"`
}

// chatChoice represents a single response choice.
type chatChoice struct {
	Message chatMessage `json:"message"`
}

// chatUsage contains token usage information.
type chatUsage struct {
	TotalTokens int `json:"total_tokens"`
}

// Generate produces a text response from the given system and user prompts using DeepSeek.
func (a *Adapter) Generate(ctx context.Context, systemPrompt string, userPrompt string) (string, int, error) {
	reqBody := chatRequest{
		Model: "deepseek-chat",
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal deepseek request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create deepseek request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to call deepseek: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read deepseek response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("failed to generate with deepseek: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", 0, fmt.Errorf("failed to unmarshal deepseek response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", 0, fmt.Errorf("failed to generate with deepseek: empty response")
	}

	text := chatResp.Choices[0].Message.Content
	tokens := chatResp.Usage.TotalTokens

	a.logger.Debug().Int("tokens", tokens).Msg("DeepSeek generation completed")
	return text, tokens, nil
}

// Provider returns the name of the LLM provider.
func (a *Adapter) Provider() string {
	return "deepseek"
}
