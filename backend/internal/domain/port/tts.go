package port

import "context"

// TextToSpeech defines the interface for text-to-speech synthesis.
// Implementations convert text into audio data.
type TextToSpeech interface {
	// Synthesize converts the given text to audio using the specified voice.
	// Returns the raw audio bytes.
	Synthesize(ctx context.Context, text string, voiceID string) ([]byte, error)
}
