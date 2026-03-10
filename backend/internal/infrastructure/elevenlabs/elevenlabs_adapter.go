// Package elevenlabs implements the TextToSpeech port using the ElevenLabs API.
package elevenlabs

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

const baseURL = "https://api.elevenlabs.io/v1/text-to-speech"

// Adapter implements port.TextToSpeech using the ElevenLabs API.
type Adapter struct {
	apiKey string
	client *http.Client
	logger *zerolog.Logger
}

// NewElevenLabsAdapter creates a new ElevenLabs text-to-speech adapter.
func NewElevenLabsAdapter(apiKey string, logger *zerolog.Logger) port.TextToSpeech {
	return &Adapter{
		apiKey: apiKey,
		client: &http.Client{},
		logger: logger,
	}
}

// synthesizeRequest is the request body for the ElevenLabs TTS endpoint.
type synthesizeRequest struct {
	Text    string `json:"text"`
	ModelID string `json:"model_id"`
}

// Synthesize converts the given text to audio using the specified voice.
// Returns the raw MP3 audio bytes.
func (a *Adapter) Synthesize(ctx context.Context, text string, voiceID string) ([]byte, error) {
	reqBody := synthesizeRequest{
		Text:    text,
		ModelID: "eleven_multilingual_v2",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal elevenlabs request: %w", err)
	}

	url := fmt.Sprintf("%s/%s", baseURL, voiceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create elevenlabs request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", a.apiKey)
	req.Header.Set("Accept", "audio/mpeg")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call elevenlabs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to synthesize with elevenlabs: status %d, body: %s", resp.StatusCode, string(errBody))
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read elevenlabs audio response: %w", err)
	}

	a.logger.Debug().
		Int("audioBytes", len(audioData)).
		Str("voiceID", voiceID).
		Msg("ElevenLabs synthesis completed")

	return audioData, nil
}
