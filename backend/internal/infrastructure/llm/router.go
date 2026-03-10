// Package llm provides an LLM router that tries multiple providers in order.
package llm

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/port"
)

// Router implements port.LLM by trying multiple LLM providers in order.
// It returns the first successful response, falling back to the next provider on failure.
type Router struct {
	providers []port.LLM
	logger    *zerolog.Logger
}

// NewRouter creates a new LLM router with the given ordered list of providers.
func NewRouter(providers []port.LLM, logger *zerolog.Logger) port.LLM {
	return &Router{
		providers: providers,
		logger:    logger,
	}
}

// Generate tries each LLM provider in order and returns the first successful response.
// If all providers fail, it returns the last error encountered.
func (r *Router) Generate(ctx context.Context, systemPrompt string, userPrompt string) (string, int, error) {
	var lastErr error

	for _, p := range r.providers {
		text, tokens, err := p.Generate(ctx, systemPrompt, userPrompt)
		if err != nil {
			r.logger.Warn().
				Err(err).
				Str("provider", p.Provider()).
				Msg("LLM provider failed, trying next")
			lastErr = err
			continue
		}

		r.logger.Debug().
			Str("provider", p.Provider()).
			Int("tokens", tokens).
			Msg("LLM generation succeeded")

		return text, tokens, nil
	}

	if lastErr != nil {
		return "", 0, fmt.Errorf("failed to generate with all providers: %w", lastErr)
	}

	return "", 0, fmt.Errorf("failed to generate: no providers configured")
}

// Provider returns the name of the LLM router.
func (r *Router) Provider() string {
	return "router"
}
