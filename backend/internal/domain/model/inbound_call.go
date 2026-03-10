package model

// InboundCall represents a phone call received by the system.
type InboundCall struct {
	// ID is the unique identifier for the call.
	ID string
	// TwilioCallSID is the Twilio-assigned call session identifier.
	TwilioCallSID string
	// CallerPhone is the phone number of the caller.
	CallerPhone string
	// Status is the current call status (ringing, in_progress, completed, failed).
	Status string
	// Transcript is the ordered list of transcript entries for this call.
	Transcript []TranscriptEntry
	// RAGQueries is the list of RAG queries executed during this call.
	RAGQueries []RAGQuery
	// DurationSeconds is the total call duration in seconds.
	DurationSeconds int
	// TwilioCostUSD is the Twilio usage cost in US dollars.
	TwilioCostUSD float64
	// LLMCostUSD is the LLM API usage cost in US dollars.
	LLMCostUSD float64
	// TotalCostUSD is the total cost of the call in US dollars.
	TotalCostUSD float64
	// CreatedAt is the creation timestamp in RFC3339 format.
	CreatedAt string
	// EndedAt is the call end timestamp in RFC3339 format.
	// Empty string means the call has not ended.
	EndedAt string
}

// TranscriptEntry represents a single speech entry in a call transcript.
type TranscriptEntry struct {
	// Speaker identifies who spoke (caller or assistant).
	Speaker string
	// Text is the transcribed speech content.
	Text string
	// At is the timestamp of the speech entry in RFC3339 format.
	At string
}

// RAGQuery represents a knowledge base query executed during a call.
type RAGQuery struct {
	// Query is the user's question sent to the RAG pipeline.
	Query string
	// Chunks contains the retrieved chunk IDs used for context.
	Chunks []string
	// Response is the generated answer from the LLM.
	Response string
	// Provider is the LLM provider that generated the response.
	Provider string
	// Tokens is the number of tokens consumed for this query.
	Tokens int
	// At is the timestamp when the query was executed in RFC3339 format.
	At string
}
