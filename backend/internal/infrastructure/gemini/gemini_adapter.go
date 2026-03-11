// Package gemini implements the LLM and Embedding ports using the Google Gemini REST API.
package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/port"
)

const (
	generateEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
	embedEndpoint    = "https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-001:embedContent"
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

// AsSpeechToText returns the adapter as a port.SpeechToText interface.
func (a *Adapter) AsSpeechToText() port.SpeechToText {
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

// part represents a content part (text or inline audio).
type part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inlineData,omitempty"`
}

// inlineData represents inline binary data (audio, image, etc).
type inlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
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
	a.logger.Debug().Int("systemPromptLen", len(systemPrompt)).Int("userPromptLen", len(userPrompt)).Msg("[GeminiAdapter] generating response")

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
		a.logger.Error().Err(err).Msg("[GeminiAdapter] failed to marshal request")
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
		a.logger.Error().Err(err).Msg("[GeminiAdapter] failed to call API")
		return "", 0, fmt.Errorf("failed to call gemini: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		a.logger.Error().Err(err).Msg("[GeminiAdapter] failed to read response body")
		return "", 0, fmt.Errorf("failed to read gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		a.logger.Error().Int("statusCode", resp.StatusCode).Msg("[GeminiAdapter] API returned non-OK status")
		return "", 0, fmt.Errorf("failed to generate with gemini: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var genResp generateResponse
	if err := json.Unmarshal(respBody, &genResp); err != nil {
		a.logger.Error().Err(err).Msg("[GeminiAdapter] failed to unmarshal response")
		return "", 0, fmt.Errorf("failed to unmarshal gemini response: %w", err)
	}

	if len(genResp.Candidates) == 0 || len(genResp.Candidates[0].Content.Parts) == 0 {
		a.logger.Warn().Msg("[GeminiAdapter] received empty response")
		return "", 0, fmt.Errorf("failed to generate with gemini: empty response")
	}

	text := genResp.Candidates[0].Content.Parts[0].Text
	tokens := 0
	if genResp.UsageMetadata != nil {
		tokens = genResp.UsageMetadata.TotalTokenCount
	}

	a.logger.Debug().Int("tokens", tokens).Int("responseLen", len(text)).Msg("[GeminiAdapter] generation completed")
	return text, tokens, nil
}

// Provider returns the name of the LLM provider.
func (a *Adapter) Provider() string {
	return "gemini"
}

// Embed converts the given text into a vector embedding using Gemini gemini-embedding-001.
func (a *Adapter) Embed(ctx context.Context, text string) ([]float32, error) {
	a.logger.Debug().Int("textLen", len(text)).Msg("[GeminiAdapter] generating embedding")

	reqBody := embedRequest{
		Model: "models/gemini-embedding-001",
		Content: embedContent{
			Parts: []part{{Text: text}},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		a.logger.Error().Err(err).Msg("[GeminiAdapter] failed to marshal embed request")
		return nil, fmt.Errorf("failed to marshal gemini embed request: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", embedEndpoint, a.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		a.logger.Error().Err(err).Msg("[GeminiAdapter] failed to create embed request")
		return nil, fmt.Errorf("failed to create gemini embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		a.logger.Error().Err(err).Msg("[GeminiAdapter] failed to call embed API")
		return nil, fmt.Errorf("failed to call gemini embed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		a.logger.Error().Err(err).Msg("[GeminiAdapter] failed to read embed response body")
		return nil, fmt.Errorf("failed to read gemini embed response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		a.logger.Error().Int("statusCode", resp.StatusCode).Msg("[GeminiAdapter] embed API returned non-OK status")
		return nil, fmt.Errorf("failed to embed with gemini: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var embedResp embedResponse
	if err := json.Unmarshal(respBody, &embedResp); err != nil {
		a.logger.Error().Err(err).Msg("[GeminiAdapter] failed to unmarshal embed response")
		return nil, fmt.Errorf("failed to unmarshal gemini embed response: %w", err)
	}

	if len(embedResp.Embedding.Values) == 0 {
		a.logger.Warn().Msg("[GeminiAdapter] received empty embedding")
		return nil, fmt.Errorf("failed to embed with gemini: empty embedding")
	}

	a.logger.Debug().Int("dimensions", len(embedResp.Embedding.Values)).Msg("[GeminiAdapter] embedding completed")
	return embedResp.Embedding.Values, nil
}

// Transcribe converts audio data into text using Gemini's multimodal capabilities.
func (a *Adapter) Transcribe(ctx context.Context, audioData []byte) (string, error) {
	a.logger.Debug().Int("audioBytes", len(audioData)).Msg("[GeminiAdapter] transcribing audio")
	reqBody := generateRequest{
		SystemInstruction: &systemInstruction{
			Parts: []part{{Text: "You are a transcription assistant. Transcribe the audio exactly as spoken. Return only the transcription text, nothing else."}},
		},
		Contents: []content{
			{
				Role: "user",
				Parts: []part{
					{InlineData: &inlineData{
						MimeType: "audio/wav",
						Data:     base64.StdEncoding.EncodeToString(audioData),
					}},
					{Text: "Transcribe this audio."},
				},
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		a.logger.Error().Err(err).Msg("[GeminiAdapter] failed to marshal transcribe request")
		return "", fmt.Errorf("failed to marshal gemini transcribe request: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", generateEndpoint, a.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		a.logger.Error().Err(err).Msg("[GeminiAdapter] failed to create transcribe request")
		return "", fmt.Errorf("failed to create gemini transcribe request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		a.logger.Error().Err(err).Msg("[GeminiAdapter] failed to call transcribe API")
		return "", fmt.Errorf("failed to call gemini transcribe: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		a.logger.Error().Err(err).Msg("[GeminiAdapter] failed to read transcribe response body")
		return "", fmt.Errorf("failed to read gemini transcribe response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		a.logger.Error().Int("statusCode", resp.StatusCode).Msg("[GeminiAdapter] transcribe API returned non-OK status")
		return "", fmt.Errorf("failed to transcribe with gemini: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var genResp generateResponse
	if err := json.Unmarshal(respBody, &genResp); err != nil {
		a.logger.Error().Err(err).Msg("[GeminiAdapter] failed to unmarshal transcribe response")
		return "", fmt.Errorf("failed to unmarshal gemini transcribe response: %w", err)
	}

	if len(genResp.Candidates) == 0 || len(genResp.Candidates[0].Content.Parts) == 0 {
		a.logger.Warn().Msg("[GeminiAdapter] received empty transcription response")
		return "", fmt.Errorf("failed to transcribe with gemini: empty response")
	}

	transcript := genResp.Candidates[0].Content.Parts[0].Text
	a.logger.Debug().Int("audioBytes", len(audioData)).Int("transcriptLen", len(transcript)).Msg("[GeminiAdapter] transcription completed")
	return transcript, nil
}
