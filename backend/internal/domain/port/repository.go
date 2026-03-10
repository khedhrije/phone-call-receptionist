// Package port defines interfaces for external dependencies.
// These interfaces represent the boundaries of the application
// and are implemented by infrastructure adapters.
package port

import (
	"context"

	"phone-call-receptionist/backend/internal/domain/model"
)

// UserFilters contains filter criteria for listing users.
type UserFilters struct {
	// Role filters users by their role.
	Role string
	// IsBlocked filters users by their blocked status.
	IsBlocked *bool
	// Limit is the maximum number of results to return.
	Limit int
	// Offset is the number of results to skip for pagination.
	Offset int
}

// CallFilters contains filter criteria for listing inbound calls.
type CallFilters struct {
	// Status filters calls by their status.
	Status string
	// CallerPhone filters calls by the caller's phone number.
	CallerPhone string
	// From filters calls created after this RFC3339 timestamp.
	From string
	// To filters calls created before this RFC3339 timestamp.
	To string
	// Limit is the maximum number of results to return.
	Limit int
	// Offset is the number of results to skip for pagination.
	Offset int
}

// AppointmentFilters contains filter criteria for listing appointments.
type AppointmentFilters struct {
	// Status filters appointments by their status.
	Status string
	// CallerPhone filters appointments by the caller's phone number.
	CallerPhone string
	// From filters appointments scheduled after this RFC3339 timestamp.
	From string
	// To filters appointments scheduled before this RFC3339 timestamp.
	To string
	// Limit is the maximum number of results to return.
	Limit int
	// Offset is the number of results to skip for pagination.
	Offset int
}

// LeadFilters contains filter criteria for listing leads.
type LeadFilters struct {
	// Status filters leads by their status.
	Status string
	// Limit is the maximum number of results to return.
	Limit int
	// Offset is the number of results to skip for pagination.
	Offset int
}

// User defines the repository interface for user persistence.
type User interface {
	// Create persists a new user to the data store.
	Create(ctx context.Context, user model.User) error
	// FindByID retrieves a user by their unique identifier.
	FindByID(ctx context.Context, id string) (model.User, error)
	// FindByEmail retrieves a user by their email address.
	FindByEmail(ctx context.Context, email string) (model.User, error)
	// List retrieves users matching the given filters.
	List(ctx context.Context, filters UserFilters) ([]model.User, int, error)
	// Update modifies an existing user's data.
	Update(ctx context.Context, user model.User) error
	// Delete removes a user from the data store.
	Delete(ctx context.Context, id string) error
}

// KnowledgeDocument defines the repository interface for knowledge document persistence.
type KnowledgeDocument interface {
	// Create persists a new knowledge document to the data store.
	Create(ctx context.Context, doc model.KnowledgeDocument) error
	// FindByID retrieves a knowledge document by its unique identifier.
	FindByID(ctx context.Context, id string) (model.KnowledgeDocument, error)
	// List retrieves all knowledge documents.
	List(ctx context.Context) ([]model.KnowledgeDocument, error)
	// Update modifies an existing knowledge document's data.
	Update(ctx context.Context, doc model.KnowledgeDocument) error
	// Delete removes a knowledge document from the data store.
	Delete(ctx context.Context, id string) error
}

// InboundCall defines the repository interface for inbound call persistence.
type InboundCall interface {
	// Create persists a new inbound call record to the data store.
	Create(ctx context.Context, call model.InboundCall) error
	// FindByID retrieves an inbound call by its unique identifier.
	FindByID(ctx context.Context, id string) (model.InboundCall, error)
	// FindByTwilioSID retrieves an inbound call by its Twilio call SID.
	FindByTwilioSID(ctx context.Context, sid string) (model.InboundCall, error)
	// List retrieves inbound calls matching the given filters.
	List(ctx context.Context, filters CallFilters) ([]model.InboundCall, int, error)
	// Update modifies an existing inbound call's data.
	Update(ctx context.Context, call model.InboundCall) error
}

// Appointment defines the repository interface for appointment persistence.
type Appointment interface {
	// Create persists a new appointment to the data store.
	Create(ctx context.Context, appt model.Appointment) error
	// FindByID retrieves an appointment by its unique identifier.
	FindByID(ctx context.Context, id string) (model.Appointment, error)
	// FindByPhone retrieves all appointments for a given phone number.
	FindByPhone(ctx context.Context, phone string) ([]model.Appointment, error)
	// List retrieves appointments matching the given filters.
	List(ctx context.Context, filters AppointmentFilters) ([]model.Appointment, int, error)
	// Update modifies an existing appointment's data.
	Update(ctx context.Context, appt model.Appointment) error
	// Delete removes an appointment from the data store.
	Delete(ctx context.Context, id string) error
}

// Lead defines the repository interface for lead persistence.
type Lead interface {
	// Create persists a new lead to the data store.
	Create(ctx context.Context, lead model.Lead) error
	// FindByID retrieves a lead by its unique identifier.
	FindByID(ctx context.Context, id string) (model.Lead, error)
	// FindByPhone retrieves a lead by their phone number.
	FindByPhone(ctx context.Context, phone string) (model.Lead, error)
	// List retrieves leads matching the given filters.
	List(ctx context.Context, filters LeadFilters) ([]model.Lead, int, error)
	// Update modifies an existing lead's data.
	Update(ctx context.Context, lead model.Lead) error
}

// SMSLog defines the repository interface for SMS log persistence.
type SMSLog interface {
	// Create persists a new SMS log entry to the data store.
	Create(ctx context.Context, log model.SMSLog) error
	// FindByID retrieves an SMS log entry by its unique identifier.
	FindByID(ctx context.Context, id string) (model.SMSLog, error)
	// ListByCallID retrieves all SMS log entries for a given call.
	ListByCallID(ctx context.Context, callID string) ([]model.SMSLog, error)
}

// AudioCache defines the repository interface for audio cache persistence.
type AudioCache interface {
	// Create persists a new audio cache entry to the data store.
	Create(ctx context.Context, cache model.AudioCache) error
	// FindByHash retrieves an audio cache entry by its content hash.
	FindByHash(ctx context.Context, hash string) (model.AudioCache, error)
	// Delete removes an audio cache entry from the data store.
	Delete(ctx context.Context, id string) error
}

// SystemSettings defines the repository interface for system settings persistence.
type SystemSettings interface {
	// Find retrieves the current system settings.
	Find(ctx context.Context) (model.SystemSettings, error)
	// Update modifies the system settings.
	Update(ctx context.Context, settings model.SystemSettings) error
}
