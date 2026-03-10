package model

// SMSLog represents a record of an SMS message sent by the system.
type SMSLog struct {
	// ID is the unique identifier for the SMS log entry.
	ID string
	// CallID is the identifier of the related inbound call.
	// Empty string means the SMS was not associated with a call.
	CallID string
	// ToPhone is the recipient phone number.
	ToPhone string
	// Message is the SMS message body.
	Message string
	// TwilioSID is the Twilio message session identifier.
	TwilioSID string
	// Status is the current SMS delivery status (queued, sent, delivered, failed).
	Status string
	// CreatedAt is the creation timestamp in RFC3339 format.
	CreatedAt string
}
