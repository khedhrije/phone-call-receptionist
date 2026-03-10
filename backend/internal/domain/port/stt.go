package port

import "context"

// SpeechToText defines the interface for speech-to-text transcription.
// Implementations convert audio data into text.
type SpeechToText interface {
	// Transcribe converts the given audio data into text.
	Transcribe(ctx context.Context, audioData []byte) (string, error)
}
