// Package gemini implements the LLM and Embedding ports using the Google Gemini REST API.
package gemini

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

const (
	generateEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
	embedEndpoint    = "https://generativelanguage.googleapis.com/v1beta/models/text-embedding-004:embedContent"
)

// Adapter implements both port.LLM and port.Embedding using the Gemini REST API.
type Adapter struct {
	apiKey string
	client *http.Client
	logger *zerolog.Logger
}

// NewGeminiAdapter creates a new Gemini adapter for LLM and embedding operations.
func NewGeminiAdapter(apiKey string, logger *zerolog.Logger) *Adapter {
	return &Adapter{
		apiKey: apiKey,
		client: &http.Client{},
		logger: logger,
	}
}

// AsLLM returns the adapter as a port.LLM interface.
func (a *Adapter) AsLLM() port.LLM {
	return a
}

// AsEmbedding returns the adapter as a port.Embedding interface.
func (a *Adapter) AsEmbedding() port.Embedding {
	return a
}

// generateRequest is the request body for the Gemini generateContent endpoint.
type generateRequest struct {
	SystemInstruction *systemInstruction `json:"systemInstruction,omitempty"`
	Contents          []content          `json:"contents"`
}

// systemInstruction wraps the system prompt.
type systemInstruction struct {
	Parts []part `json:"parts"`
}

// content represents a message in the conversation.
type content struct {
	Role  string `json:"role"`
	Parts []part `json:"parts"`
}

// part represents a content part.
type part struct {
	Text string `json:"text"`
}

// generateResponse is the response from the Gemini generateContent endpoint.
type generateResponse struct {
	Candidates []candidate `json:"candidates"`
	UsageMetadata *usageMetadata `json:"usageMetadata"`
}

// candidate is a single generated response candidate.
type candidate struct {
	Content content `json:"content"`
}

// usageMetadata contains token usage information.
type usageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// embedRequest is the request body for the Gemini embedContent endpoint.
type embedRequest struct {
	Model   string       `json:"model"`
	Content embedContent `json:"content"`
}

// embedContent wraps the text to embed.
type embedContent struct {
	Parts []part `json:"parts"`
}

// embedResponse is the response from the Gemini embedContent endpoint.
type embedResponse struct {
	Embedding embeddingValues `json:"embedding"`
}

// embeddingValues contains the embedding vector.
type embeddingValues struct {
	Values []float32 `json:"values"`
}

// Generate produces a text response from the given system and user prompts using Gemini.
func (a *Adapter) Generate(ctx context.Context, systemPrompt string, userPrompt string) (string, int, error) {
	reqBody := generateRequest{
		Contents: []content{
			{
				Role:  "user",
				Parts: []part{{Text: userPrompt}},
			},
		},
	}

	if systemPrompt != "" {
		reqBody.SystemInstruction = &systemInstruction{
			Parts: []part{{Text: systemPrompt}},
		}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal gemini request: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", generateEndpoint, a.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to call gemini: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("failed to generate with gemini: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var genResp generateResponse
	if err := json.Unmarshal(respBody, &genResp); err != nil {
		return "", 0, fmt.Errorf("failed to unmarshal gemini response: %w", err)
	}

	if len(genResp.Candidates) == 0 || len(genResp.Candidates[0].Content.Parts) == 0 {
		return "", 0, fmt.Errorf("failed to generate with gemini: empty response")
	}

	text := genResp.Candidates[0].Content.Parts[0].Text
	tokens := 0
	if genResp.UsageMetadata != nil {
		tokens = genResp.UsageMetadata.TotalTokenCount
	}

	a.logger.Debug().Int("tokens", tokens).Msg("Gemini generation completed")
	return text, tokens, nil
}

// Provider returns the name of the LLM provider.
func (a *Adapter) Provider() string {
	return "gemini"
}

// Embed converts the given text into a vector embedding using Gemini text-embedding-004.
func (a *Adapter) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody := embedRequest{
		Model: "models/text-embedding-004",
		Content: embedContent{
			Parts: []part{{Text: text}},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gemini embed request: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", embedEndpoint, a.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call gemini embed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read gemini embed response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to embed with gemini: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var embedResp embedResponse
	if err := json.Unmarshal(respBody, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal gemini embed response: %w", err)
	}

	if len(embedResp.Embedding.Values) == 0 {
		return nil, fmt.Errorf("failed to embed with gemini: empty embedding")
	}

	a.logger.Debug().Int("dimensions", len(embedResp.Embedding.Values)).Msg("Gemini embedding completed")
	return embedResp.Embedding.Values, nil
}
