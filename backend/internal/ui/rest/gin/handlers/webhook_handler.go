package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/api"
	"phone-call-receptionist/backend/internal/infrastructure/twiml"
)

// WebhookHandler handles Twilio webhook HTTP requests.
type WebhookHandler struct {
	voiceCallApi *api.VoiceCallApi
	logger       *zerolog.Logger
}

// NewWebhookHandler creates a new WebhookHandler with the given dependencies.
func NewWebhookHandler(voiceCallApi *api.VoiceCallApi, logger *zerolog.Logger) *WebhookHandler {
	return &WebhookHandler{voiceCallApi: voiceCallApi, logger: logger}
}

// HandleVoice godoc
//
//	@Summary		Twilio voice webhook
//	@Description	Handles inbound call from Twilio
//	@Tags			Webhooks
//	@Accept			x-www-form-urlencoded
//	@Produce		xml
//	@Success		200	{string}	string	"TwiML XML response"
//	@Router			/webhooks/twilio/voice [post]
func (h *WebhookHandler) HandleVoice(c *gin.Context) {
	callSID := c.PostForm("CallSid")
	callerPhone := c.PostForm("From")

	h.logger.Info().Str("callSid", callSID).Str("callerPhone", callerPhone).Msg("[WebhookHandler] HandleVoice inbound call received")

	greeting, err := h.voiceCallApi.HandleInbound(c.Request.Context(), callSID, callerPhone)
	if err != nil {
		h.logger.Error().Err(err).Str("callSid", callSID).Str("callerPhone", callerPhone).Msg("[WebhookHandler] HandleVoice failed to handle inbound call")
		greeting = "Merci de votre appel. Nous rencontrons des difficultés techniques. Veuillez réessayer plus tard."
	}

	h.logger.Info().Str("callSid", callSID).Msg("[WebhookHandler] HandleVoice responding with TwiML")

	xmlStr, _ := twiml.Build([]twiml.Node{
		twiml.Gather{
			Input:         "speech",
			Timeout:       5,
			SpeechTimeout: "auto",
			Language:      "fr-FR",
			Action:        "/api/webhooks/twilio/gather",
			Method:        "POST",
			Children: []twiml.Node{
				twiml.Say{Text: greeting, Voice: "Polly.Lea", Language: "fr-FR"},
			},
		},
	})

	c.Data(http.StatusOK, "application/xml", []byte(xmlStr))
}

// HandleGather godoc
//
//	@Summary		Twilio gather webhook
//	@Description	Handles speech input from caller
//	@Tags			Webhooks
//	@Accept			x-www-form-urlencoded
//	@Produce		xml
//	@Success		200	{string}	string	"TwiML XML response"
//	@Router			/webhooks/twilio/gather [post]
func (h *WebhookHandler) HandleGather(c *gin.Context) {
	callSID := c.PostForm("CallSid")
	speechResult := c.PostForm("SpeechResult")

	h.logger.Info().Str("callSid", callSID).Str("speechResult", speechResult).Msg("[WebhookHandler] HandleGather speech input received")

	if speechResult == "" {
		h.logger.Info().Str("callSid", callSID).Msg("[WebhookHandler] HandleGather empty speech result, prompting retry")
		xmlStr, _ := twiml.Build([]twiml.Node{
			twiml.Gather{
				Input:         "speech",
				Timeout:       5,
				SpeechTimeout: "auto",
				Language:      "fr-FR",
				Action:        "/api/webhooks/twilio/gather",
				Method:        "POST",
				Children: []twiml.Node{
					twiml.Say{Text: "Je n'ai pas compris. Pourriez-vous répéter s'il vous plaît ?", Voice: "Polly.Lea", Language: "fr-FR"},
				},
			},
		})
		c.Data(http.StatusOK, "application/xml", []byte(xmlStr))
		return
	}

	response, err := h.voiceCallApi.ProcessSpeech(c.Request.Context(), callSID, speechResult)
	if err != nil {
		h.logger.Error().Err(err).Str("callSid", callSID).Msg("[WebhookHandler] HandleGather failed to process speech")
		response = "Je suis désolé, j'ai un problème technique. Pourriez-vous répéter s'il vous plaît ?"
	}

	h.logger.Info().Str("callSid", callSID).Msg("[WebhookHandler] HandleGather responding with TwiML")

	xmlStr, _ := twiml.Build([]twiml.Node{
		twiml.Gather{
			Input:         "speech",
			Timeout:       5,
			SpeechTimeout: "auto",
			Language:      "fr-FR",
			Action:        "/api/webhooks/twilio/gather",
			Method:        "POST",
			Children: []twiml.Node{
				twiml.Say{Text: response, Voice: "Polly.Lea", Language: "fr-FR"},
			},
		},
		twiml.Say{Text: "Merci de votre appel. Au revoir !", Voice: "Polly.Lea", Language: "fr-FR"},
		twiml.Hangup{},
	})

	c.Data(http.StatusOK, "application/xml", []byte(xmlStr))
}

// HandleStatus godoc
//
//	@Summary		Twilio status webhook
//	@Description	Handles call status updates from Twilio
//	@Tags			Webhooks
//	@Accept			x-www-form-urlencoded
//	@Produce		xml
//	@Success		200	{string}	string	"OK"
//	@Router			/webhooks/twilio/status [post]
func (h *WebhookHandler) HandleStatus(c *gin.Context) {
	callSID := c.PostForm("CallSid")
	callStatus := c.PostForm("CallStatus")
	durationStr := c.PostForm("CallDuration")

	h.logger.Info().
		Str("callSid", callSID).
		Str("callStatus", callStatus).
		Str("duration", durationStr).
		Msg("[WebhookHandler] HandleStatus call status update received")

	if callStatus == "completed" || callStatus == "failed" || callStatus == "no-answer" {
		duration, _ := strconv.Atoi(durationStr)
		h.logger.Info().Str("callSid", callSID).Str("callStatus", callStatus).Int("duration", duration).Msg("[WebhookHandler] HandleStatus ending call")
		if err := h.voiceCallApi.EndCall(c.Request.Context(), callSID, duration); err != nil {
			h.logger.Error().Err(err).Str("callSid", callSID).Msg("[WebhookHandler] HandleStatus failed to end call")
		}
	}

	h.logger.Info().Str("callSid", callSID).Msg("[WebhookHandler] HandleStatus succeeded")

	c.String(http.StatusOK, "OK")
}

// HandleRecording godoc
//
//	@Summary		Twilio recording webhook
//	@Description	Handles recording callbacks from Twilio
//	@Tags			Webhooks
//	@Accept			x-www-form-urlencoded
//	@Produce		json
//	@Success		200	{string}	string	"OK"
//	@Router			/webhooks/twilio/recording [post]
func (h *WebhookHandler) HandleRecording(c *gin.Context) {
	recordingURL := c.PostForm("RecordingUrl")
	callSID := c.PostForm("CallSid")
	h.logger.Info().Str("callSid", callSID).Str("recordingUrl", recordingURL).Msg("[WebhookHandler] HandleRecording recording received")
	_ = fmt.Sprintf("recording for %s", callSID)

	h.logger.Info().Str("callSid", callSID).Msg("[WebhookHandler] HandleRecording succeeded")

	c.String(http.StatusOK, "OK")
}
