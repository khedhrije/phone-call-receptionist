// Package openai implements the LLM port using the OpenAI API via the go-openai client.
package openai

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	goopenai "github.com/sashabaranov/go-openai"

	"phone-call-receptionist/backend/internal/domain/port"
)

// Adapter implements port.LLM using the OpenAI API.
type Adapter struct {
	client *goopenai.Client
	logger *zerolog.Logger
}

// NewOpenAIAdapter creates a new OpenAI LLM adapter.
func NewOpenAIAdapter(apiKey string, logger *zerolog.Logger) port.LLM {
	client := goopenai.NewClient(apiKey)
	return &Adapter{
		client: client,
		logger: logger,
	}
}

// Generate produces a text response from the given system and user prompts using OpenAI GPT-4o.
func (a *Adapter) Generate(ctx context.Context, systemPrompt string, userPrompt string) (string, int, error) {
	messages := []goopenai.ChatCompletionMessage{
		{
			Role:    goopenai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    goopenai.ChatMessageRoleUser,
			Content: userPrompt,
		},
	}

	resp, err := a.client.CreateChatCompletion(ctx, goopenai.ChatCompletionRequest{
		Model:    goopenai.GPT4o,
		Messages: messages,
	})
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate with openai: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", 0, fmt.Errorf("failed to generate with openai: empty response")
	}

	text := resp.Choices[0].Message.Content
	tokens := resp.Usage.TotalTokens

	a.logger.Debug().Int("tokens", tokens).Msg("OpenAI generation completed")
	return text, tokens, nil
}

// Provider returns the name of the LLM provider.
func (a *Adapter) Provider() string {
	return "openai"
}
