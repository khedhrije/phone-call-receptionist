package model

// AudioCache represents a cached TTS audio file to avoid redundant synthesis.
type AudioCache struct {
	// ID is the unique identifier for the audio cache entry.
	ID string
	// Hash is the SHA-256 hash of the text content combined with the voice ID.
	Hash string
	// VoiceID is the ElevenLabs voice identifier used for synthesis.
	VoiceID string
	// FilePath is the storage path of the cached audio file.
	FilePath string
	// CreatedAt is the creation timestamp in RFC3339 format.
	CreatedAt string
}
