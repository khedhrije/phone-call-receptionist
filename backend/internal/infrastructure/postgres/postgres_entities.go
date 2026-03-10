package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"phone-call-receptionist/backend/internal/domain/model"
)

// UserDB is the database entity for the users table.
type UserDB struct {
	// ID is the primary key.
	ID string `db:"id"`
	// Email is the user's email address.
	Email string `db:"email"`
	// PasswordHash is the bcrypt-hashed password.
	PasswordHash string `db:"password_hash"`
	// Role is the user's permission level.
	Role string `db:"role"`
	// IsBlocked indicates whether the user account is blocked.
	IsBlocked bool `db:"is_blocked"`
	// CreatedAt is the record creation timestamp.
	CreatedAt time.Time `db:"created_at"`
	// UpdatedAt is the record last modification timestamp.
	UpdatedAt time.Time `db:"updated_at"`
}

// ToDomain converts a UserDB to a domain User model.
func (u UserDB) ToDomain() model.User {
	return model.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		IsBlocked:    u.IsBlocked,
		CreatedAt:    u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    u.UpdatedAt.Format(time.RFC3339),
	}
}

// FromDomain populates a UserDB from a domain User model.
func (u *UserDB) FromDomain(m model.User) {
	u.ID = m.ID
	u.Email = m.Email
	u.PasswordHash = m.PasswordHash
	u.Role = m.Role
	u.IsBlocked = m.IsBlocked
	if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
		u.CreatedAt = t
	} else {
		u.CreatedAt = time.Now().UTC()
	}
	if t, err := time.Parse(time.RFC3339, m.UpdatedAt); err == nil {
		u.UpdatedAt = t
	} else {
		u.UpdatedAt = time.Now().UTC()
	}
}

// KnowledgeDocumentDB is the database entity for the knowledge_documents table.
type KnowledgeDocumentDB struct {
	// ID is the primary key.
	ID string `db:"id"`
	// Filename is the original name of the uploaded file.
	Filename string `db:"filename"`
	// MimeType is the MIME type of the uploaded file.
	MimeType string `db:"mime_type"`
	// FilePath is the storage path of the uploaded file.
	FilePath string `db:"file_path"`
	// ChunkCount is the number of chunks the document was split into.
	ChunkCount int `db:"chunk_count"`
	// Status is the current indexing status.
	Status string `db:"status"`
	// IndexedAt is the timestamp when indexing completed (nullable).
	IndexedAt sql.NullTime `db:"indexed_at"`
	// CreatedAt is the record creation timestamp.
	CreatedAt time.Time `db:"created_at"`
	// UpdatedAt is the record last modification timestamp.
	UpdatedAt time.Time `db:"updated_at"`
}

// ToDomain converts a KnowledgeDocumentDB to a domain KnowledgeDocument model.
func (d KnowledgeDocumentDB) ToDomain() model.KnowledgeDocument {
	var indexedAt string
	if d.IndexedAt.Valid {
		indexedAt = d.IndexedAt.Time.Format(time.RFC3339)
	}
	return model.KnowledgeDocument{
		ID:         d.ID,
		Filename:   d.Filename,
		MimeType:   d.MimeType,
		FilePath:   d.FilePath,
		ChunkCount: d.ChunkCount,
		Status:     d.Status,
		IndexedAt:  indexedAt,
		CreatedAt:  d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  d.UpdatedAt.Format(time.RFC3339),
	}
}

// FromDomain populates a KnowledgeDocumentDB from a domain KnowledgeDocument model.
func (d *KnowledgeDocumentDB) FromDomain(m model.KnowledgeDocument) {
	d.ID = m.ID
	d.Filename = m.Filename
	d.MimeType = m.MimeType
	d.FilePath = m.FilePath
	d.ChunkCount = m.ChunkCount
	d.Status = m.Status
	if m.IndexedAt != "" {
		if t, err := time.Parse(time.RFC3339, m.IndexedAt); err == nil {
			d.IndexedAt = sql.NullTime{Time: t, Valid: true}
		}
	}
	if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
		d.CreatedAt = t
	} else {
		d.CreatedAt = time.Now().UTC()
	}
	if t, err := time.Parse(time.RFC3339, m.UpdatedAt); err == nil {
		d.UpdatedAt = t
	} else {
		d.UpdatedAt = time.Now().UTC()
	}
}

// InboundCallDB is the database entity for the inbound_calls table.
type InboundCallDB struct {
	// ID is the primary key.
	ID string `db:"id"`
	// TwilioCallSID is the Twilio-assigned call session identifier.
	TwilioCallSID string `db:"twilio_call_sid"`
	// CallerPhone is the phone number of the caller.
	CallerPhone string `db:"caller_phone"`
	// Status is the current call status.
	Status string `db:"status"`
	// Transcript is the JSONB transcript data.
	Transcript []byte `db:"transcript"`
	// RAGQueries is the JSONB RAG queries data.
	RAGQueries []byte `db:"rag_queries"`
	// DurationSeconds is the total call duration in seconds.
	DurationSeconds int `db:"duration_seconds"`
	// TwilioCostUSD is the Twilio usage cost in US dollars.
	TwilioCostUSD float64 `db:"twilio_cost_usd"`
	// LLMCostUSD is the LLM API usage cost in US dollars.
	LLMCostUSD float64 `db:"llm_cost_usd"`
	// TotalCostUSD is the total cost of the call in US dollars.
	TotalCostUSD float64 `db:"total_cost_usd"`
	// CreatedAt is the record creation timestamp.
	CreatedAt time.Time `db:"created_at"`
	// EndedAt is the call end timestamp (nullable).
	EndedAt sql.NullTime `db:"ended_at"`
}

// ToDomain converts an InboundCallDB to a domain InboundCall model.
func (c InboundCallDB) ToDomain() (model.InboundCall, error) {
	var transcript []model.TranscriptEntry
	if len(c.Transcript) > 0 {
		if err := json.Unmarshal(c.Transcript, &transcript); err != nil {
			return model.InboundCall{}, fmt.Errorf("failed to unmarshal transcript: %w", err)
		}
	}

	var ragQueries []model.RAGQuery
	if len(c.RAGQueries) > 0 {
		if err := json.Unmarshal(c.RAGQueries, &ragQueries); err != nil {
			return model.InboundCall{}, fmt.Errorf("failed to unmarshal rag_queries: %w", err)
		}
	}

	var endedAt string
	if c.EndedAt.Valid {
		endedAt = c.EndedAt.Time.Format(time.RFC3339)
	}

	return model.InboundCall{
		ID:              c.ID,
		TwilioCallSID:   c.TwilioCallSID,
		CallerPhone:     c.CallerPhone,
		Status:          c.Status,
		Transcript:      transcript,
		RAGQueries:      ragQueries,
		DurationSeconds: c.DurationSeconds,
		TwilioCostUSD:   c.TwilioCostUSD,
		LLMCostUSD:      c.LLMCostUSD,
		TotalCostUSD:    c.TotalCostUSD,
		CreatedAt:       c.CreatedAt.Format(time.RFC3339),
		EndedAt:         endedAt,
	}, nil
}

// FromDomain populates an InboundCallDB from a domain InboundCall model.
func (c *InboundCallDB) FromDomain(m model.InboundCall) error {
	c.ID = m.ID
	c.TwilioCallSID = m.TwilioCallSID
	c.CallerPhone = m.CallerPhone
	c.Status = m.Status
	c.DurationSeconds = m.DurationSeconds
	c.TwilioCostUSD = m.TwilioCostUSD
	c.LLMCostUSD = m.LLMCostUSD
	c.TotalCostUSD = m.TotalCostUSD

	if len(m.Transcript) > 0 {
		data, err := json.Marshal(m.Transcript)
		if err != nil {
			return fmt.Errorf("failed to marshal transcript: %w", err)
		}
		c.Transcript = data
	} else {
		c.Transcript = []byte("[]")
	}

	if len(m.RAGQueries) > 0 {
		data, err := json.Marshal(m.RAGQueries)
		if err != nil {
			return fmt.Errorf("failed to marshal rag_queries: %w", err)
		}
		c.RAGQueries = data
	} else {
		c.RAGQueries = []byte("[]")
	}

	if m.EndedAt != "" {
		if t, err := time.Parse(time.RFC3339, m.EndedAt); err == nil {
			c.EndedAt = sql.NullTime{Time: t, Valid: true}
		}
	}
	if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
		c.CreatedAt = t
	} else {
		c.CreatedAt = time.Now().UTC()
	}

	return nil
}

// AppointmentDB is the database entity for the appointments table.
type AppointmentDB struct {
	// ID is the primary key.
	ID string `db:"id"`
	// CallID is the identifier of the related inbound call (nullable).
	CallID sql.NullString `db:"call_id"`
	// CallerPhone is the phone number of the person who booked.
	CallerPhone string `db:"caller_phone"`
	// CallerName is the full name of the person who booked.
	CallerName string `db:"caller_name"`
	// CallerEmail is the email address of the person who booked.
	CallerEmail string `db:"caller_email"`
	// ServiceType is the type of IT service requested.
	ServiceType string `db:"service_type"`
	// ScheduledAt is the appointment date and time.
	ScheduledAt time.Time `db:"scheduled_at"`
	// DurationMins is the appointment duration in minutes.
	DurationMins int `db:"duration_mins"`
	// Status is the current appointment status.
	Status string `db:"status"`
	// GoogleEventID is the Google Calendar event identifier (nullable).
	GoogleEventID sql.NullString `db:"google_event_id"`
	// SMSSentAt is the timestamp when the SMS confirmation was sent (nullable).
	SMSSentAt sql.NullTime `db:"sms_sent_at"`
	// Notes contains additional notes about the appointment (nullable).
	Notes sql.NullString `db:"notes"`
	// CreatedAt is the record creation timestamp.
	CreatedAt time.Time `db:"created_at"`
	// UpdatedAt is the record last modification timestamp.
	UpdatedAt time.Time `db:"updated_at"`
}

// ToDomain converts an AppointmentDB to a domain Appointment model.
func (a AppointmentDB) ToDomain() model.Appointment {
	var callID, googleEventID, smsSentAt, notes string
	if a.CallID.Valid {
		callID = a.CallID.String
	}
	if a.GoogleEventID.Valid {
		googleEventID = a.GoogleEventID.String
	}
	if a.SMSSentAt.Valid {
		smsSentAt = a.SMSSentAt.Time.Format(time.RFC3339)
	}
	if a.Notes.Valid {
		notes = a.Notes.String
	}
	return model.Appointment{
		ID:            a.ID,
		CallID:        callID,
		CallerPhone:   a.CallerPhone,
		CallerName:    a.CallerName,
		CallerEmail:   a.CallerEmail,
		ServiceType:   a.ServiceType,
		ScheduledAt:   a.ScheduledAt.Format(time.RFC3339),
		DurationMins:  a.DurationMins,
		Status:        a.Status,
		GoogleEventID: googleEventID,
		SMSSentAt:     smsSentAt,
		Notes:         notes,
		CreatedAt:     a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     a.UpdatedAt.Format(time.RFC3339),
	}
}

// FromDomain populates an AppointmentDB from a domain Appointment model.
func (a *AppointmentDB) FromDomain(m model.Appointment) {
	a.ID = m.ID
	if m.CallID != "" {
		a.CallID = sql.NullString{String: m.CallID, Valid: true}
	}
	a.CallerPhone = m.CallerPhone
	a.CallerName = m.CallerName
	a.CallerEmail = m.CallerEmail
	a.ServiceType = m.ServiceType
	if t, err := time.Parse(time.RFC3339, m.ScheduledAt); err == nil {
		a.ScheduledAt = t
	}
	a.DurationMins = m.DurationMins
	a.Status = m.Status
	if m.GoogleEventID != "" {
		a.GoogleEventID = sql.NullString{String: m.GoogleEventID, Valid: true}
	}
	if m.SMSSentAt != "" {
		if t, err := time.Parse(time.RFC3339, m.SMSSentAt); err == nil {
			a.SMSSentAt = sql.NullTime{Time: t, Valid: true}
		}
	}
	if m.Notes != "" {
		a.Notes = sql.NullString{String: m.Notes, Valid: true}
	}
	if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
		a.CreatedAt = t
	} else {
		a.CreatedAt = time.Now().UTC()
	}
	if t, err := time.Parse(time.RFC3339, m.UpdatedAt); err == nil {
		a.UpdatedAt = t
	} else {
		a.UpdatedAt = time.Now().UTC()
	}
}

// LeadDB is the database entity for the leads table.
type LeadDB struct {
	// ID is the primary key.
	ID string `db:"id"`
	// CallID is the identifier of the related inbound call (nullable).
	CallID sql.NullString `db:"call_id"`
	// Phone is the lead's phone number.
	Phone string `db:"phone"`
	// Name is the lead's full name.
	Name string `db:"name"`
	// Email is the lead's email address.
	Email string `db:"email"`
	// Status is the current lead status.
	Status string `db:"status"`
	// Notes contains additional notes about the lead (nullable).
	Notes sql.NullString `db:"notes"`
	// CreatedAt is the record creation timestamp.
	CreatedAt time.Time `db:"created_at"`
	// UpdatedAt is the record last modification timestamp.
	UpdatedAt time.Time `db:"updated_at"`
}

// ToDomain converts a LeadDB to a domain Lead model.
func (l LeadDB) ToDomain() model.Lead {
	var callID, notes string
	if l.CallID.Valid {
		callID = l.CallID.String
	}
	if l.Notes.Valid {
		notes = l.Notes.String
	}
	return model.Lead{
		ID:        l.ID,
		CallID:    callID,
		Phone:     l.Phone,
		Name:      l.Name,
		Email:     l.Email,
		Status:    l.Status,
		Notes:     notes,
		CreatedAt: l.CreatedAt.Format(time.RFC3339),
		UpdatedAt: l.UpdatedAt.Format(time.RFC3339),
	}
}

// FromDomain populates a LeadDB from a domain Lead model.
func (l *LeadDB) FromDomain(m model.Lead) {
	l.ID = m.ID
	if m.CallID != "" {
		l.CallID = sql.NullString{String: m.CallID, Valid: true}
	}
	l.Phone = m.Phone
	l.Name = m.Name
	l.Email = m.Email
	l.Status = m.Status
	if m.Notes != "" {
		l.Notes = sql.NullString{String: m.Notes, Valid: true}
	}
	if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
		l.CreatedAt = t
	} else {
		l.CreatedAt = time.Now().UTC()
	}
	if t, err := time.Parse(time.RFC3339, m.UpdatedAt); err == nil {
		l.UpdatedAt = t
	} else {
		l.UpdatedAt = time.Now().UTC()
	}
}

// SMSLogDB is the database entity for the sms_logs table.
type SMSLogDB struct {
	// ID is the primary key.
	ID string `db:"id"`
	// CallID is the identifier of the related inbound call (nullable).
	CallID sql.NullString `db:"call_id"`
	// ToPhone is the recipient phone number.
	ToPhone string `db:"to_phone"`
	// Message is the SMS message body.
	Message string `db:"message"`
	// TwilioSID is the Twilio message session identifier.
	TwilioSID string `db:"twilio_sid"`
	// Status is the current SMS delivery status.
	Status string `db:"status"`
	// CreatedAt is the record creation timestamp.
	CreatedAt time.Time `db:"created_at"`
}

// ToDomain converts an SMSLogDB to a domain SMSLog model.
func (s SMSLogDB) ToDomain() model.SMSLog {
	var callID string
	if s.CallID.Valid {
		callID = s.CallID.String
	}
	return model.SMSLog{
		ID:        s.ID,
		CallID:    callID,
		ToPhone:   s.ToPhone,
		Message:   s.Message,
		TwilioSID: s.TwilioSID,
		Status:    s.Status,
		CreatedAt: s.CreatedAt.Format(time.RFC3339),
	}
}

// FromDomain populates an SMSLogDB from a domain SMSLog model.
func (s *SMSLogDB) FromDomain(m model.SMSLog) {
	s.ID = m.ID
	if m.CallID != "" {
		s.CallID = sql.NullString{String: m.CallID, Valid: true}
	}
	s.ToPhone = m.ToPhone
	s.Message = m.Message
	s.TwilioSID = m.TwilioSID
	s.Status = m.Status
	if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
		s.CreatedAt = t
	} else {
		s.CreatedAt = time.Now().UTC()
	}
}

// AudioCacheDB is the database entity for the audio_cache table.
type AudioCacheDB struct {
	// ID is the primary key.
	ID string `db:"id"`
	// Hash is the SHA-256 hash of the text content combined with the voice ID.
	Hash string `db:"hash"`
	// VoiceID is the ElevenLabs voice identifier used for synthesis.
	VoiceID string `db:"voice_id"`
	// FilePath is the storage path of the cached audio file.
	FilePath string `db:"file_path"`
	// CreatedAt is the record creation timestamp.
	CreatedAt time.Time `db:"created_at"`
}

// ToDomain converts an AudioCacheDB to a domain AudioCache model.
func (a AudioCacheDB) ToDomain() model.AudioCache {
	return model.AudioCache{
		ID:        a.ID,
		Hash:      a.Hash,
		VoiceID:   a.VoiceID,
		FilePath:  a.FilePath,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
	}
}

// FromDomain populates an AudioCacheDB from a domain AudioCache model.
func (a *AudioCacheDB) FromDomain(m model.AudioCache) {
	a.ID = m.ID
	a.Hash = m.Hash
	a.VoiceID = m.VoiceID
	a.FilePath = m.FilePath
	if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
		a.CreatedAt = t
	} else {
		a.CreatedAt = time.Now().UTC()
	}
}

// SystemSettingsDB is the database entity for the system_settings table.
type SystemSettingsDB struct {
	// ID is the primary key (singleton row).
	ID int `db:"id"`
	// DefaultLLMProvider is the name of the default LLM provider.
	DefaultLLMProvider string `db:"default_llm_provider"`
	// DefaultVoiceID is the default ElevenLabs voice identifier for TTS.
	DefaultVoiceID string `db:"default_voice_id"`
	// TopK is the number of top results to retrieve in vector search.
	TopK int `db:"top_k"`
	// MaxCallDurationSecs is the maximum allowed call duration in seconds.
	MaxCallDurationSecs int `db:"max_call_duration_secs"`
	// UpdatedAt is the record last modification timestamp.
	UpdatedAt time.Time `db:"updated_at"`
}

// ToDomain converts a SystemSettingsDB to a domain SystemSettings model.
func (s SystemSettingsDB) ToDomain() model.SystemSettings {
	return model.SystemSettings{
		DefaultLLMProvider:  s.DefaultLLMProvider,
		DefaultVoiceID:      s.DefaultVoiceID,
		TopK:                s.TopK,
		MaxCallDurationSecs: s.MaxCallDurationSecs,
		UpdatedAt:           s.UpdatedAt.Format(time.RFC3339),
	}
}

// FromDomain populates a SystemSettingsDB from a domain SystemSettings model.
func (s *SystemSettingsDB) FromDomain(m model.SystemSettings) {
	s.ID = 1
	s.DefaultLLMProvider = m.DefaultLLMProvider
	s.DefaultVoiceID = m.DefaultVoiceID
	s.TopK = m.TopK
	s.MaxCallDurationSecs = m.MaxCallDurationSecs
	if t, err := time.Parse(time.RFC3339, m.UpdatedAt); err == nil {
		s.UpdatedAt = t
	} else {
		s.UpdatedAt = time.Now().UTC()
	}
}
