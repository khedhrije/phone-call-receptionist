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
	v.logger.Info().Str("callSid", twilioSID).Str("callerPhone", callerPhone).Msg("[VoiceCallApi] HandleInbound started")

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
		v.logger.Error().Err(err).Str("callSid", twilioSID).Msg("[VoiceCallApi] Failed to create call record")
		return "", fmt.Errorf("failed to create call record: %w", err)
	}
	v.logger.Info().Str("callId", call.ID).Str("callSid", twilioSID).Msg("[VoiceCallApi] Call record created")

	greeting := "Merci d'avoir appelé nos services informatiques. Je suis Alex. Comment puis-je vous aider aujourd'hui ?"

	call.Transcript = append(call.Transcript, model.TranscriptEntry{
		Speaker: "assistant",
		Text:    greeting,
		At:      now,
	})
	if err := v.callPort.Update(ctx, call); err != nil {
		v.logger.Error().Err(err).Str("callSid", twilioSID).Msg("[VoiceCallApi] Failed to update call transcript")
	}

	v.broadcaster.Broadcast(ctx, map[string]interface{}{
		"type":        "call_started",
		"callId":      call.ID,
		"callerPhone": callerPhone,
		"at":          now,
	})
	v.logger.Info().Str("callSid", twilioSID).Msg("[VoiceCallApi] Broadcast call_started event")

	return greeting, nil
}

// ProcessSpeech handles caller speech input, detects intent, and generates a response.
func (v *VoiceCallApi) ProcessSpeech(ctx context.Context, twilioSID string, transcript string) (string, error) {
	v.logger.Info().Str("callSid", twilioSID).Str("transcript", transcript).Msg("[VoiceCallApi] ProcessSpeech started")

	call, err := v.callPort.FindByTwilioSID(ctx, twilioSID)
	if err != nil {
		v.logger.Error().Err(err).Str("callSid", twilioSID).Msg("[VoiceCallApi] Failed to find call by TwilioSID")
		return "", fmt.Errorf("failed to find call: %w", err)
	}
	v.logger.Debug().Str("callId", call.ID).Str("callSid", twilioSID).Msg("[VoiceCallApi] Found call record")

	now := time.Now().Format(time.RFC3339)
	call.Transcript = append(call.Transcript, model.TranscriptEntry{
		Speaker: "caller",
		Text:    transcript,
		At:      now,
	})

	intent := v.detectIntent(transcript)
	v.logger.Info().Str("callSid", twilioSID).Str("intent", intent).Msg("[VoiceCallApi] Intent detected")

	var response string

	switch intent {
	case "book", "schedule":
		v.logger.Info().Str("callSid", twilioSID).Msg("[VoiceCallApi] Handling booking intent")
		response, err = v.handleBookingIntent(ctx, call, transcript)
	case "reschedule":
		v.logger.Info().Str("callSid", twilioSID).Msg("[VoiceCallApi] Handling reschedule intent")
		response = "Je serais ravi de vous aider à reprogrammer votre rendez-vous. Pourriez-vous me donner votre nom ou numéro de téléphone pour que je puisse retrouver votre rendez-vous ?"
	case "cancel":
		v.logger.Info().Str("callSid", twilioSID).Msg("[VoiceCallApi] Handling cancel intent")
		response = "Je peux vous aider à annuler votre rendez-vous. Pourriez-vous me donner votre nom ou numéro de téléphone pour que je puisse le retrouver ?"
	default:
		v.logger.Info().Str("callSid", twilioSID).Msg("[VoiceCallApi] Handling question intent via RAG")
		response, err = v.handleQuestionIntent(ctx, &call, transcript)
	}

	if err != nil {
		v.logger.Error().Err(err).Str("callSid", twilioSID).Str("intent", intent).Msg("[VoiceCallApi] Failed to process intent")
		response = "Je suis désolé, j'ai du mal à traiter votre demande. Pourriez-vous répéter s'il vous plaît ?"
	}

	v.logger.Info().Str("callSid", twilioSID).Str("response", response).Msg("[VoiceCallApi] Generated response")

	call.Transcript = append(call.Transcript, model.TranscriptEntry{
		Speaker: "assistant",
		Text:    response,
		At:      time.Now().Format(time.RFC3339),
	})

	if err := v.callPort.Update(ctx, call); err != nil {
		v.logger.Error().Err(err).Str("callSid", twilioSID).Msg("[VoiceCallApi] Failed to update call")
	}

	v.broadcaster.Broadcast(ctx, map[string]interface{}{
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

	bookKeywords := []string{"book", "schedule", "appointment", "make an appointment", "set up a meeting", "rendez-vous", "prendre rendez-vous", "réserver", "planifier"}
	for _, kw := range bookKeywords {
		if strings.Contains(lower, kw) {
			return "book"
		}
	}

	rescheduleKeywords := []string{"reschedule", "change my appointment", "move my appointment", "reprogrammer", "déplacer", "changer mon rendez-vous", "modifier mon rendez-vous"}
	for _, kw := range rescheduleKeywords {
		if strings.Contains(lower, kw) {
			return "reschedule"
		}
	}

	cancelKeywords := []string{"cancel", "cancel my appointment", "annuler", "annuler mon rendez-vous"}
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
		v.logger.Debug().Str("callSid", call.TwilioCallSID).Str("step", state.Step).Msg("[VoiceCallApi] Loaded booking state from cache")
	} else {
		v.logger.Debug().Str("callSid", call.TwilioCallSID).Msg("[VoiceCallApi] No existing booking state, starting new flow")
	}

	if state.Step == "" {
		state.Step = "ask_name"
		v.saveBookingState(ctx, stateKey, state)
		v.logger.Info().Str("callSid", call.TwilioCallSID).Str("step", "ask_name").Msg("[VoiceCallApi] Booking flow: asking for name")
		return "Je serais ravi de vous aider à prendre rendez-vous. Puis-je avoir votre nom complet s'il vous plaît ?", nil
	}

	switch state.Step {
	case "ask_name":
		state.Name = transcript
		state.Step = "ask_email"
		v.saveBookingState(ctx, stateKey, state)
		v.logger.Info().Str("callSid", call.TwilioCallSID).Str("name", state.Name).Str("step", "ask_email").Msg("[VoiceCallApi] Booking flow: name collected, asking for email")
		return fmt.Sprintf("Merci, %s. Quelle est votre adresse email ?", state.Name), nil

	case "ask_email":
		state.Email = transcript
		state.Step = "ask_service"
		v.saveBookingState(ctx, stateKey, state)
		v.logger.Info().Str("callSid", call.TwilioCallSID).Str("email", state.Email).Str("step", "ask_service").Msg("[VoiceCallApi] Booking flow: email collected, asking for service")
		return "Quel type de service recherchez-vous ? Nous proposons l'installation réseau, le conseil en cybersécurité, la migration cloud, le support informatique et le développement logiciel.", nil

	case "ask_service":
		state.ServiceType = transcript
		state.Step = "ask_time"
		v.saveBookingState(ctx, stateKey, state)
		v.logger.Info().Str("callSid", call.TwilioCallSID).Str("service", state.ServiceType).Str("step", "ask_time").Msg("[VoiceCallApi] Booking flow: service collected, asking for time")
		return "Quand souhaitez-vous planifier votre rendez-vous ? Veuillez indiquer la date et l'heure souhaitées.", nil

	case "ask_time":
		state.Slot = transcript
		state.Step = "confirm"
		v.saveBookingState(ctx, stateKey, state)
		v.logger.Info().Str("callSid", call.TwilioCallSID).Str("slot", state.Slot).Str("step", "confirm").Msg("[VoiceCallApi] Booking flow: time collected, asking for confirmation")
		return fmt.Sprintf("Pour confirmer — %s, %s le %s. Est-ce correct ?", state.Name, state.ServiceType, state.Slot), nil

	case "confirm":
		lower := strings.ToLower(transcript)
		if strings.Contains(lower, "yes") || strings.Contains(lower, "correct") || strings.Contains(lower, "that's right") || strings.Contains(lower, "oui") || strings.Contains(lower, "c'est correct") || strings.Contains(lower, "exactement") {
			v.cache.Delete(ctx, stateKey)
			v.logger.Info().Str("callSid", call.TwilioCallSID).Str("name", state.Name).Str("service", state.ServiceType).Str("slot", state.Slot).Msg("[VoiceCallApi] Booking confirmed")
			return fmt.Sprintf("Votre rendez-vous a été réservé. Vous recevrez un SMS de confirmation sous peu. Y a-t-il autre chose que je puisse faire pour vous ?"), nil
		}
		state.Step = "ask_name"
		v.saveBookingState(ctx, stateKey, state)
		v.logger.Info().Str("callSid", call.TwilioCallSID).Msg("[VoiceCallApi] Booking not confirmed, restarting flow")
		return "Pas de problème, recommençons. Puis-je avoir votre nom complet s'il vous plaît ?", nil
	}

	return "Je suis désolé, pourriez-vous répéter s'il vous plaît ?", nil
}

func (v *VoiceCallApi) saveBookingState(ctx context.Context, key string, state BookingState) {
	data, _ := json.Marshal(state)
	v.cache.Create(ctx, key, data, 15*time.Minute)
}

func (v *VoiceCallApi) handleQuestionIntent(ctx context.Context, call *model.InboundCall, transcript string) (string, error) {
	v.logger.Debug().Str("callSid", call.TwilioCallSID).Msg("[VoiceCallApi] Loading system settings for RAG query")

	settings, err := v.settingsPort.Find(ctx)
	if err != nil {
		v.logger.Warn().Err(err).Msg("[VoiceCallApi] Failed to load settings, using defaults")
		settings = model.SystemSettings{TopK: 5}
	}

	v.logger.Info().Str("callSid", call.TwilioCallSID).Str("query", transcript).Int("topK", settings.TopK).Msg("[VoiceCallApi] Querying knowledge base")

	answer, chunks, provider, tokens, err := v.ragApi.QueryKnowledgeBase(ctx, transcript, settings.TopK)
	if err != nil {
		return "", fmt.Errorf("failed to query knowledge base: %w", err)
	}

	v.logger.Info().Str("callSid", call.TwilioCallSID).Str("provider", provider).Int("tokens", tokens).Int("chunks", len(chunks)).Msg("[VoiceCallApi] RAG query completed")

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
	v.logger.Info().Str("callSid", twilioSID).Int("duration", durationSeconds).Msg("[VoiceCallApi] EndCall started")

	call, err := v.callPort.FindByTwilioSID(ctx, twilioSID)
	if err != nil {
		v.logger.Error().Err(err).Str("callSid", twilioSID).Msg("[VoiceCallApi] Failed to find call for ending")
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

	v.logger.Info().
		Str("callId", call.ID).
		Str("callSid", twilioSID).
		Int("duration", durationSeconds).
		Int("totalTokens", totalTokens).
		Float64("twilioCost", call.TwilioCostUSD).
		Float64("llmCost", call.LLMCostUSD).
		Float64("totalCost", call.TotalCostUSD).
		Msg("[VoiceCallApi] Call costs calculated")

	if err := v.callPort.Update(ctx, call); err != nil {
		v.logger.Error().Err(err).Str("callSid", twilioSID).Msg("[VoiceCallApi] Failed to update call on end")
		return fmt.Errorf("failed to update call: %w", err)
	}

	go v.createOrUpdateLead(context.Background(), call)

	v.broadcaster.Broadcast(ctx, map[string]interface{}{
		"type":     "call_ended",
		"callId":   call.ID,
		"duration": durationSeconds,
		"cost":     call.TotalCostUSD,
		"at":       now,
	})

	v.logger.Info().Str("callSid", twilioSID).Int("duration", durationSeconds).Msg("[VoiceCallApi] Call ended successfully")
	return nil
}

func (v *VoiceCallApi) createOrUpdateLead(ctx context.Context, call model.InboundCall) {
	v.logger.Debug().Str("callId", call.ID).Str("phone", call.CallerPhone).Msg("[VoiceCallApi] Creating/updating lead from call")

	existing, err := v.leadPort.FindByPhone(ctx, call.CallerPhone)
	if err == nil {
		existing.CallID = call.ID
		existing.UpdatedAt = time.Now().Format(time.RFC3339)
		if err := v.leadPort.Update(ctx, existing); err != nil {
			v.logger.Error().Err(err).Str("phone", call.CallerPhone).Msg("[VoiceCallApi] Failed to update lead")
		} else {
			v.logger.Info().Str("leadId", existing.ID).Str("phone", call.CallerPhone).Msg("[VoiceCallApi] Lead updated")
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
		v.logger.Error().Err(err).Str("phone", call.CallerPhone).Msg("[VoiceCallApi] Failed to create lead")
	} else {
		v.logger.Info().Str("leadId", lead.ID).Str("phone", call.CallerPhone).Msg("[VoiceCallApi] Lead created")
	}
}

// CallHistory retrieves a paginated list of inbound calls.
func (v *VoiceCallApi) CallHistory(ctx context.Context, filters port.CallFilters) ([]model.InboundCall, int, error) {
	v.logger.Debug().Int("limit", filters.Limit).Int("offset", filters.Offset).Str("status", filters.Status).Msg("[VoiceCallApi] Listing call history")
	calls, total, err := v.callPort.List(ctx, filters)
	if err != nil {
		v.logger.Error().Err(err).Msg("[VoiceCallApi] Failed to list calls")
		return nil, 0, fmt.Errorf("failed to list calls: %w", err)
	}
	v.logger.Debug().Int("total", total).Int("returned", len(calls)).Msg("[VoiceCallApi] Call history retrieved")
	return calls, total, nil
}

// CallDetail retrieves the full details of a specific call.
func (v *VoiceCallApi) CallDetail(ctx context.Context, id string) (model.InboundCall, error) {
	v.logger.Debug().Str("callId", id).Msg("[VoiceCallApi] Fetching call detail")
	call, err := v.callPort.FindByID(ctx, id)
	if err != nil {
		v.logger.Error().Err(err).Str("callId", id).Msg("[VoiceCallApi] Failed to find call")
		return model.InboundCall{}, fmt.Errorf("failed to find call: %w", err)
	}
	v.logger.Debug().Str("callId", id).Str("status", call.Status).Msg("[VoiceCallApi] Call detail retrieved")
	return call, nil
}
