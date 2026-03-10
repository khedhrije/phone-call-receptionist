package model

// SystemSettings represents the global system configuration.
type SystemSettings struct {
	// DefaultLLMProvider is the name of the default LLM provider to use.
	DefaultLLMProvider string
	// DefaultVoiceID is the default ElevenLabs voice identifier for TTS.
	DefaultVoiceID string
	// TopK is the number of top results to retrieve in vector search.
	TopK int
	// MaxCallDurationSecs is the maximum allowed call duration in seconds.
	MaxCallDurationSecs int
	// UpdatedAt is the last modification timestamp in RFC3339 format.
	UpdatedAt string
}
