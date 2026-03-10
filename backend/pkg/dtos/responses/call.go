package responses

// CallResponse contains the inbound call data returned by the API.
type CallResponse struct {
	// ID is the call's unique identifier.
	ID string `json:"id"`
	// TwilioCallSID is the Twilio-assigned call session identifier.
	TwilioCallSID string `json:"twilioCallSid"`
	// CallerPhone is the caller's phone number.
	CallerPhone string `json:"callerPhone"`
	// Status is the current call status.
	Status string `json:"status"`
	// Transcript is the call transcript entries.
	Transcript []TranscriptEntryResponse `json:"transcript"`
	// RAGQueries is the RAG queries executed during the call.
	RAGQueries []RAGQueryResponse `json:"ragQueries"`
	// DurationSeconds is the call duration in seconds.
	DurationSeconds int `json:"durationSeconds"`
	// TwilioCostUSD is the Twilio cost in US dollars.
	TwilioCostUSD float64 `json:"twilioCostUsd"`
	// LLMCostUSD is the LLM cost in US dollars.
	LLMCostUSD float64 `json:"llmCostUsd"`
	// TotalCostUSD is the total cost in US dollars.
	TotalCostUSD float64 `json:"totalCostUsd"`
	// CreatedAt is the call start timestamp in RFC3339 format.
	CreatedAt string `json:"createdAt"`
	// EndedAt is the call end timestamp in RFC3339 format.
	EndedAt string `json:"endedAt"`
}

// TranscriptEntryResponse contains a single transcript entry.
type TranscriptEntryResponse struct {
	// Speaker identifies who spoke (caller or assistant).
	Speaker string `json:"speaker"`
	// Text is the spoken text.
	Text string `json:"text"`
	// At is the speech timestamp in RFC3339 format.
	At string `json:"at"`
}

// RAGQueryResponse contains the details of a RAG query.
type RAGQueryResponse struct {
	// Query is the user's question.
	Query string `json:"query"`
	// Chunks is the list of retrieved chunk IDs.
	Chunks []string `json:"chunks"`
	// Response is the generated answer.
	Response string `json:"response"`
	// Provider is the LLM provider used.
	Provider string `json:"provider"`
	// Tokens is the number of tokens consumed.
	Tokens int `json:"tokens"`
	// At is the query timestamp in RFC3339 format.
	At string `json:"at"`
}
