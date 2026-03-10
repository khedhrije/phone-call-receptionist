package port

import "context"

// LLM defines the interface for large language model interactions.
// Implementations may use different providers (Gemini, Claude, OpenAI, etc.).
type LLM interface {
	// Generate produces a text response from the given system and user prompts.
	// Returns the generated text, token count, and any error.
	Generate(ctx context.Context, systemPrompt string, userPrompt string) (string, int, error)
	// Provider returns the name of the LLM provider (e.g., "gemini", "claude", "openai").
	Provider() string
}
