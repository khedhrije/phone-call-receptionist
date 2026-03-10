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

	greeting, err := h.voiceCallApi.HandleInbound(c.Request.Context(), callSID, callerPhone)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to handle inbound call")
		greeting = "Thank you for calling. We are experiencing technical difficulties. Please try again later."
	}

	xmlStr, _ := twiml.Build([]twiml.Node{
		twiml.Gather{
			Input:         "speech",
			Timeout:       5,
			SpeechTimeout: "auto",
			Action:        "/api/webhooks/twilio/gather",
			Method:        "POST",
			Children: []twiml.Node{
				twiml.Say{Text: greeting, Voice: "Polly.Joanna"},
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

	if speechResult == "" {
		xmlStr, _ := twiml.Build([]twiml.Node{
			twiml.Gather{
				Input:         "speech",
				Timeout:       5,
				SpeechTimeout: "auto",
				Action:        "/api/webhooks/twilio/gather",
				Method:        "POST",
				Children: []twiml.Node{
					twiml.Say{Text: "I didn't catch that. Could you please repeat?", Voice: "Polly.Joanna"},
				},
			},
		})
		c.Data(http.StatusOK, "application/xml", []byte(xmlStr))
		return
	}

	response, err := h.voiceCallApi.ProcessSpeech(c.Request.Context(), callSID, speechResult)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to process speech")
		response = "I'm sorry, I'm having trouble right now. Could you please repeat that?"
	}

	xmlStr, _ := twiml.Build([]twiml.Node{
		twiml.Gather{
			Input:         "speech",
			Timeout:       5,
			SpeechTimeout: "auto",
			Action:        "/api/webhooks/twilio/gather",
			Method:        "POST",
			Children: []twiml.Node{
				twiml.Say{Text: response, Voice: "Polly.Joanna"},
			},
		},
		twiml.Say{Text: "Thank you for calling. Goodbye!", Voice: "Polly.Joanna"},
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

	if callStatus == "completed" || callStatus == "failed" || callStatus == "no-answer" {
		duration, _ := strconv.Atoi(durationStr)
		if err := h.voiceCallApi.EndCall(c.Request.Context(), callSID, duration); err != nil {
			h.logger.Error().Err(err).Str("callSid", callSID).Msg("Failed to end call")
		}
	}

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
	h.logger.Info().Str("callSid", callSID).Str("recordingUrl", recordingURL).Msg("Recording received")
	_ = fmt.Sprintf("recording for %s", callSID)
	c.String(http.StatusOK, "OK")
}
