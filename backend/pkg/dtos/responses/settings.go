package responses

// SystemSettingsResponse contains the system settings returned by the API.
type SystemSettingsResponse struct {
	// DefaultLLMProvider is the name of the default LLM provider.
	DefaultLLMProvider string `json:"defaultLlmProvider"`
	// DefaultVoiceID is the default ElevenLabs voice identifier.
	DefaultVoiceID string `json:"defaultVoiceId"`
	// TopK is the number of top results in vector search.
	TopK int `json:"topK"`
	// MaxCallDurationSecs is the maximum call duration in seconds.
	MaxCallDurationSecs int `json:"maxCallDurationSecs"`
	// UpdatedAt is the last modification timestamp in RFC3339 format.
	UpdatedAt string `json:"updatedAt"`
}
