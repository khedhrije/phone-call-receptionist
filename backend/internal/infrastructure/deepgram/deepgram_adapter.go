// Package deepgram implements the SpeechToText port using the Deepgram API.
package deepgram

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

const listenEndpoint = "https://api.deepgram.com/v1/listen?model=nova-2&smart_format=true"

// Adapter implements port.SpeechToText using the Deepgram API.
type Adapter struct {
	apiKey string
	client *http.Client
	logger *zerolog.Logger
}

// NewDeepgramAdapter creates a new Deepgram speech-to-text adapter.
func NewDeepgramAdapter(apiKey string, logger *zerolog.Logger) port.SpeechToText {
	return &Adapter{
		apiKey: apiKey,
		client: &http.Client{},
		logger: logger,
	}
}

// transcriptionResponse is the response from the Deepgram listen endpoint.
type transcriptionResponse struct {
	Results results `json:"results"`
}

// results contains the transcription results.
type results struct {
	Channels []channel `json:"channels"`
}

// channel contains alternatives for a single audio channel.
type channel struct {
	Alternatives []alternative `json:"alternatives"`
}

// alternative represents a single transcription result.
type alternative struct {
	Transcript string  `json:"transcript"`
	Confidence float64 `json:"confidence"`
}

// Transcribe converts the given audio data into text using Deepgram.
func (a *Adapter) Transcribe(ctx context.Context, audioData []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, listenEndpoint, bytes.NewReader(audioData))
	if err != nil {
		return "", fmt.Errorf("failed to create deepgram request: %w", err)
	}
	req.Header.Set("Content-Type", "audio/wav")
	req.Header.Set("Authorization", "Token "+a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call deepgram: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read deepgram response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to transcribe with deepgram: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var txResp transcriptionResponse
	if err := json.Unmarshal(respBody, &txResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal deepgram response: %w", err)
	}

	if len(txResp.Results.Channels) == 0 || len(txResp.Results.Channels[0].Alternatives) == 0 {
		return "", fmt.Errorf("failed to transcribe with deepgram: empty response")
	}

	transcript := txResp.Results.Channels[0].Alternatives[0].Transcript

	a.logger.Debug().
		Int("audioBytes", len(audioData)).
		Float64("confidence", txResp.Results.Channels[0].Alternatives[0].Confidence).
		Msg("Deepgram transcription completed")

	return transcript, nil
}
