package requests

// UpdateSettingsRequest contains the fields for updating system settings.
type UpdateSettingsRequest struct {
	// DefaultLLMProvider is the name of the default LLM provider.
	DefaultLLMProvider string `json:"defaultLlmProvider" binding:"omitempty"`
	// DefaultVoiceID is the default ElevenLabs voice identifier.
	DefaultVoiceID string `json:"defaultVoiceId" binding:"omitempty"`
	// TopK is the number of top results in vector search.
	TopK int `json:"topK" binding:"omitempty,min=1,max=20"`
	// MaxCallDurationSecs is the maximum call duration in seconds.
	MaxCallDurationSecs int `json:"maxCallDurationSecs" binding:"omitempty,min=60,max=1800"`
}
