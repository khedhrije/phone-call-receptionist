package port

import "context"

// VoiceCaller defines the interface for voice call and SMS operations.
// Implementations handle outbound SMS and TwiML generation.
type VoiceCaller interface {
	// SendSMS sends an SMS message to the given phone number.
	// Returns the Twilio message SID and any error.
	SendSMS(ctx context.Context, to string, message string) (string, error)
	// ValidateSignature verifies that a webhook request came from Twilio.
	ValidateSignature(url string, params map[string]string, signature string) bool
}
