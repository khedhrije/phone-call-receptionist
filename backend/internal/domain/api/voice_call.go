package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"
)

// BookingState tracks the multi-turn appointment booking conversation state.
type BookingState struct {
	// Step is the current step in the booking flow.
	Step string `json:"step"`
	// Name is the caller's name collected during booking.
	Name string `json:"name"`
	// Email is the caller's email collected during booking.
	Email string `json:"email"`
	// ServiceType is the service type selected during booking.
	ServiceType string `json:"serviceType"`
	// Slot is the preferred time slot selected during booking.
	Slot string `json:"slot"`
}

// VoiceCallApi provides business operations for handling inbound voice calls.
type VoiceCallApi struct {
	callPort    port.InboundCall
	leadPort    port.Lead
	ragApi      *RAGApi
	apptApi     *AppointmentApi
	tts         port.TextToSpeech
	cache       port.Cache
	audioCachePort port.AudioCache
	fileStorage port.FileStorage
	broadcaster port.EventBroadcaster
	settingsPort port.SystemSettings
	logger      *zerolog.Logger
}

// NewVoiceCallApi creates a new VoiceCallApi with the given dependencies.
func NewVoiceCallApi(
	callPort port.InboundCall,
	leadPort port.Lead,
	ragApi *RAGApi,
	apptApi *AppointmentApi,
	tts port.TextToSpeech,
	cache port.Cache,
	audioCachePort port.AudioCache,
	fileStorage port.FileStorage,
	broadcaster port.EventBroadcaster,
	settingsPort port.SystemSettings,
	logger *zerolog.Logger,
) *VoiceCallApi {
	return &VoiceCallApi{
		callPort:       callPort,
		leadPort:       leadPort,
		ragApi:         ragApi,
		apptApi:        apptApi,
		tts:            tts,
		cache:          cache,
		audioCachePort: audioCachePort,
		fileStorage:    fileStorage,
		broadcaster:    broadcaster,
		settingsPort:   settingsPort,
		logger:         logger,
	}
}

// HandleInbound creates a new call record and returns the greeting text.
func (v *VoiceCallApi) HandleInbound(ctx context.Context, twilioSID string, callerPhone string) (string, error) {
	now := time.Now().Format(time.RFC3339)
	call := model.InboundCall{
		ID:            uuid.New().String(),
		TwilioCallSID: twilioSID,
		CallerPhone:   callerPhone,
		Status:        "in_progress",
		Transcript:    []model.TranscriptEntry{},
		RAGQueries:    []model.RAGQuery{},
		CreatedAt:     now,
	}

	if err := v.callPort.Create(ctx, call); err != nil {
		return "", fmt.Errorf("failed to create call record: %w", err)
	}

	greeting := "Thank you for calling our IT services. My name is Alex. How can I help you today?"

	call.Transcript = append(call.Transcript, model.TranscriptEntry{
		Speaker: "assistant",
		Text:    greeting,
		At:      now,
	})
	if err := v.callPort.Update(ctx, call); err != nil {
		v.logger.Error().Err(err).Str("callSid", twilioSID).Msg("Failed to update call transcript")
	}

	v.broadcaster.Broadcast(map[string]interface{}{
		"type":        "call_started",
		"callId":      call.ID,
		"callerPhone": callerPhone,
		"at":          now,
	})

	return greeting, nil
}

// ProcessSpeech handles caller speech input, detects intent, and generates a response.
func (v *VoiceCallApi) ProcessSpeech(ctx context.Context, twilioSID string, transcript string) (string, error) {
	call, err := v.callPort.FindByTwilioSID(ctx, twilioSID)
	if err != nil {
		return "", fmt.Errorf("failed to find call: %w", err)
	}

	now := time.Now().Format(time.RFC3339)
	call.Transcript = append(call.Transcript, model.TranscriptEntry{
		Speaker: "caller",
		Text:    transcript,
		At:      now,
	})

	intent := v.detectIntent(transcript)
	var response string

	switch intent {
	case "book", "schedule":
		response, err = v.handleBookingIntent(ctx, call, transcript)
	case "reschedule":
		response = "I'd be happy to help you reschedule. Could you please provide me with your name or phone number so I can look up your appointment?"
	case "cancel":
		response = "I can help you cancel your appointment. Could you please provide me with your name or phone number so I can look up your appointment?"
	default:
		response, err = v.handleQuestionIntent(ctx, call, transcript)
	}

	if err != nil {
		v.logger.Error().Err(err).Str("callSid", twilioSID).Msg("Failed to process speech")
		response = "I'm sorry, I'm having trouble processing your request. Could you please repeat that?"
	}

	call.Transcript = append(call.Transcript, model.TranscriptEntry{
		Speaker: "assistant",
		Text:    response,
		At:      time.Now().Format(time.RFC3339),
	})

	if err := v.callPort.Update(ctx, call); err != nil {
		v.logger.Error().Err(err).Str("callSid", twilioSID).Msg("Failed to update call")
	}

	v.broadcaster.Broadcast(map[string]interface{}{
		"type":       "call_speech",
		"callId":     call.ID,
		"transcript": transcript,
		"response":   response,
		"at":         now,
	})

	return response, nil
}

func (v *VoiceCallApi) detectIntent(text string) string {
	lower := strings.ToLower(text)

	bookKeywords := []string{"book", "schedule", "appointment", "make an appointment", "set up a meeting"}
	for _, kw := range bookKeywords {
		if strings.Contains(lower, kw) {
			return "book"
		}
	}

	rescheduleKeywords := []string{"reschedule", "change my appointment", "move my appointment"}
	for _, kw := range rescheduleKeywords {
		if strings.Contains(lower, kw) {
			return "reschedule"
		}
	}

	cancelKeywords := []string{"cancel", "cancel my appointment"}
	for _, kw := range cancelKeywords {
		if strings.Contains(lower, kw) {
			return "cancel"
		}
	}

	return "question"
}

func (v *VoiceCallApi) handleBookingIntent(ctx context.Context, call model.InboundCall, transcript string) (string, error) {
	stateKey := fmt.Sprintf("call:%s:state", call.TwilioCallSID)
	var state BookingState

	stateData, err := v.cache.Find(ctx, stateKey)
	if err == nil {
		json.Unmarshal(stateData, &state)
	}

	if state.Step == "" {
		state.Step = "ask_name"
		v.saveBookingState(ctx, stateKey, state)
		return "I'd be happy to help you book an appointment. May I have your full name, please?", nil
	}

	switch state.Step {
	case "ask_name":
		state.Name = transcript
		state.Step = "ask_email"
		v.saveBookingState(ctx, stateKey, state)
		return fmt.Sprintf("Thank you, %s. And what's your email address?", state.Name), nil

	case "ask_email":
		state.Email = transcript
		state.Step = "ask_service"
		v.saveBookingState(ctx, stateKey, state)
		return "What type of service are you looking for? We offer network setup, cybersecurity consulting, cloud migration, IT support, and software development.", nil

	case "ask_service":
		state.ServiceType = transcript
		state.Step = "ask_time"
		v.saveBookingState(ctx, stateKey, state)
		return "When would you like to schedule your appointment? Please provide your preferred date and time.", nil

	case "ask_time":
		state.Slot = transcript
		state.Step = "confirm"
		v.saveBookingState(ctx, stateKey, state)
		return fmt.Sprintf("Just to confirm — %s, %s on %s. Is that correct?", state.Name, state.ServiceType, state.Slot), nil

	case "confirm":
		lower := strings.ToLower(transcript)
		if strings.Contains(lower, "yes") || strings.Contains(lower, "correct") || strings.Contains(lower, "that's right") {
			v.cache.Delete(ctx, stateKey)
			return fmt.Sprintf("Your appointment has been booked. You'll receive an SMS confirmation at your phone number shortly. Is there anything else I can help you with?"), nil
		}
		state.Step = "ask_name"
		v.saveBookingState(ctx, stateKey, state)
		return "No problem, let's start over. May I have your full name, please?", nil
	}

	return "I'm sorry, could you please repeat that?", nil
}

func (v *VoiceCallApi) saveBookingState(ctx context.Context, key string, state BookingState) {
	data, _ := json.Marshal(state)
	v.cache.Store(ctx, key, data, 15*time.Minute)
}

func (v *VoiceCallApi) handleQuestionIntent(ctx context.Context, call model.InboundCall, transcript string) (string, error) {
	settings, err := v.settingsPort.Find(ctx)
	if err != nil {
		v.logger.Warn().Err(err).Msg("Failed to load settings, using defaults")
		settings = model.SystemSettings{TopK: 5}
	}

	answer, chunks, provider, tokens, err := v.ragApi.QueryKnowledgeBase(ctx, transcript, settings.TopK)
	if err != nil {
		return "", fmt.Errorf("failed to query knowledge base: %w", err)
	}

	var chunkIDs []string
	for _, c := range chunks {
		chunkIDs = append(chunkIDs, c.ID)
	}

	call.RAGQueries = append(call.RAGQueries, model.RAGQuery{
		Query:    transcript,
		Chunks:   chunkIDs,
		Response: answer,
		Provider: provider,
		Tokens:   tokens,
		At:       time.Now().Format(time.RFC3339),
	})

	return answer, nil
}

// EndCall finalizes the call record with duration and costs, and creates/updates a lead.
func (v *VoiceCallApi) EndCall(ctx context.Context, twilioSID string, durationSeconds int) error {
	call, err := v.callPort.FindByTwilioSID(ctx, twilioSID)
	if err != nil {
		return fmt.Errorf("failed to find call: %w", err)
	}

	now := time.Now().Format(time.RFC3339)
	call.Status = "completed"
	call.DurationSeconds = durationSeconds
	call.EndedAt = now
	call.TwilioCostUSD = float64(durationSeconds) * 0.013 / 60.0

	var totalTokens int
	for _, q := range call.RAGQueries {
		totalTokens += q.Tokens
	}
	call.LLMCostUSD = float64(totalTokens) * 0.00001
	call.TotalCostUSD = call.TwilioCostUSD + call.LLMCostUSD

	if err := v.callPort.Update(ctx, call); err != nil {
		return fmt.Errorf("failed to update call: %w", err)
	}

	go v.createOrUpdateLead(context.Background(), call)

	v.broadcaster.Broadcast(map[string]interface{}{
		"type":     "call_ended",
		"callId":   call.ID,
		"duration": durationSeconds,
		"cost":     call.TotalCostUSD,
		"at":       now,
	})

	v.logger.Info().Str("callSid", twilioSID).Int("duration", durationSeconds).Msg("Call ended")
	return nil
}

func (v *VoiceCallApi) createOrUpdateLead(ctx context.Context, call model.InboundCall) {
	existing, err := v.leadPort.FindByPhone(ctx, call.CallerPhone)
	if err == nil {
		existing.CallID = call.ID
		existing.UpdatedAt = time.Now().Format(time.RFC3339)
		if err := v.leadPort.Update(ctx, existing); err != nil {
			v.logger.Error().Err(err).Str("phone", call.CallerPhone).Msg("Failed to update lead")
		}
		return
	}

	now := time.Now().Format(time.RFC3339)
	lead := model.Lead{
		ID:        uuid.New().String(),
		CallID:    call.ID,
		Phone:     call.CallerPhone,
		Status:    "new",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := v.leadPort.Create(ctx, lead); err != nil {
		v.logger.Error().Err(err).Str("phone", call.CallerPhone).Msg("Failed to create lead")
	}
}

// CallHistory retrieves a paginated list of inbound calls.
func (v *VoiceCallApi) CallHistory(ctx context.Context, filters port.CallFilters) ([]model.InboundCall, int, error) {
	calls, total, err := v.callPort.List(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list calls: %w", err)
	}
	return calls, total, nil
}

// CallDetail retrieves the full details of a specific call.
func (v *VoiceCallApi) CallDetail(ctx context.Context, id string) (model.InboundCall, error) {
	call, err := v.callPort.FindByID(ctx, id)
	if err != nil {
		return model.InboundCall{}, fmt.Errorf("failed to find call: %w", err)
	}
	return call, nil
}
