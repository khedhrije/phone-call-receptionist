// Package twilio implements the VoiceCaller port using the Twilio API.
package twilio

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/port"
)

const messagesEndpoint = "https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json"

// Adapter implements port.VoiceCaller using the Twilio API.
type Adapter struct {
	accountSID  string
	authToken   string
	phoneNumber string
	client      *http.Client
	logger      *zerolog.Logger
}

// NewTwilioAdapter creates a new Twilio voice caller adapter.
func NewTwilioAdapter(accountSID string, authToken string, phoneNumber string, logger *zerolog.Logger) port.VoiceCaller {
	return &Adapter{
		accountSID:  accountSID,
		authToken:   authToken,
		phoneNumber: phoneNumber,
		client:      &http.Client{},
		logger:      logger,
	}
}

// smsResponse represents the Twilio Messages API response.
type smsResponse struct {
	SID          string `json:"sid"`
	ErrorCode    *int   `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

// SendSMS sends an SMS message to the given phone number using Twilio.
// Returns the Twilio message SID and any error.
func (a *Adapter) SendSMS(ctx context.Context, to string, message string) (string, error) {
	a.logger.Debug().Str("to", to).Int("messageLen", len(message)).Msg("[TwilioAdapter] sending SMS")

	endpoint := fmt.Sprintf(messagesEndpoint, a.accountSID)

	data := url.Values{}
	data.Set("To", to)
	data.Set("From", a.phoneNumber)
	data.Set("Body", message)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		a.logger.Error().Err(err).Str("to", to).Msg("[TwilioAdapter] failed to create request")
		return "", fmt.Errorf("failed to create twilio request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(a.accountSID, a.authToken)

	resp, err := a.client.Do(req)
	if err != nil {
		a.logger.Error().Err(err).Str("to", to).Msg("[TwilioAdapter] failed to call API")
		return "", fmt.Errorf("failed to call twilio: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		a.logger.Error().Err(err).Str("to", to).Msg("[TwilioAdapter] failed to read response body")
		return "", fmt.Errorf("failed to read twilio response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		a.logger.Error().Int("statusCode", resp.StatusCode).Str("to", to).Msg("[TwilioAdapter] API returned non-created status")
		return "", fmt.Errorf("failed to send sms via twilio: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var smsResp smsResponse
	if err := json.Unmarshal(respBody, &smsResp); err != nil {
		a.logger.Error().Err(err).Str("to", to).Msg("[TwilioAdapter] failed to unmarshal response")
		return "", fmt.Errorf("failed to unmarshal twilio response: %w", err)
	}

	if smsResp.ErrorCode != nil {
		a.logger.Error().Int("errorCode", *smsResp.ErrorCode).Str("errorMessage", smsResp.ErrorMessage).Str("to", to).Msg("[TwilioAdapter] API returned error")
		return "", fmt.Errorf("failed to send sms via twilio: error %d: %s", *smsResp.ErrorCode, smsResp.ErrorMessage)
	}

	a.logger.Info().
		Str("sid", smsResp.SID).
		Str("to", to).
		Msg("[TwilioAdapter] SMS sent")

	return smsResp.SID, nil
}

// ValidateSignature verifies that a webhook request came from Twilio by
// validating the X-Twilio-Signature using HMAC-SHA1.
func (a *Adapter) ValidateSignature(requestURL string, params map[string]string, signature string) bool {
	a.logger.Debug().Str("url", requestURL).Int("params", len(params)).Msg("[TwilioAdapter] validating signature")
	// Sort parameter keys
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build the data string: URL + sorted key-value pairs
	var builder strings.Builder
	builder.WriteString(requestURL)
	for _, k := range keys {
		builder.WriteString(k)
		builder.WriteString(params[k])
	}

	// Compute HMAC-SHA1
	mac := hmac.New(sha1.New, []byte(a.authToken))
	mac.Write([]byte(builder.String()))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	valid := hmac.Equal([]byte(expected), []byte(signature))
	if !valid {
		a.logger.Warn().Str("url", requestURL).Msg("[TwilioAdapter] signature validation failed")
	} else {
		a.logger.Debug().Str("url", requestURL).Msg("[TwilioAdapter] signature validated")
	}
	return valid
}
