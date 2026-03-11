package api

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"
)

// LeadApi provides business operations for lead management.
type LeadApi struct {
	leadPort port.Lead
	logger   *zerolog.Logger
}

// NewLeadApi creates a new LeadApi with the given dependencies.
func NewLeadApi(leadPort port.Lead, logger *zerolog.Logger) *LeadApi {
	return &LeadApi{leadPort: leadPort, logger: logger}
}

// CreateOrUpdate upserts a lead by phone number.
func (l *LeadApi) CreateOrUpdate(ctx context.Context, phone string, name string, email string, callID string) (model.Lead, error) {
	l.logger.Info().Str("phone", phone).Str("name", name).Str("callID", callID).Msg("[LeadApi] CreateOrUpdate started")

	existing, err := l.leadPort.FindByPhone(ctx, phone)
	if err == nil {
		l.logger.Info().Str("leadId", existing.ID).Str("phone", phone).Msg("[LeadApi] Existing lead found, updating")
		if name != "" {
			existing.Name = name
		}
		if email != "" {
			existing.Email = email
		}
		if callID != "" {
			existing.CallID = callID
		}
		existing.UpdatedAt = time.Now().Format(time.RFC3339)

		if err := l.leadPort.Update(ctx, existing); err != nil {
			l.logger.Error().Err(err).Str("leadId", existing.ID).Msg("[LeadApi] Failed to update existing lead")
			return model.Lead{}, fmt.Errorf("failed to update lead: %w", err)
		}
		l.logger.Info().Str("leadId", existing.ID).Msg("[LeadApi] Lead updated successfully")
		return existing, nil
	}

	l.logger.Debug().Str("phone", phone).Msg("[LeadApi] No existing lead found, creating new")
	now := time.Now().Format(time.RFC3339)
	lead := model.Lead{
		ID:        fmt.Sprintf("%s", time.Now().UnixNano()),
		CallID:    callID,
		Phone:     phone,
		Name:      name,
		Email:     email,
		Status:    "new",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := l.leadPort.Create(ctx, lead); err != nil {
		l.logger.Error().Err(err).Str("phone", phone).Msg("[LeadApi] Failed to create lead")
		return model.Lead{}, fmt.Errorf("failed to create lead: %w", err)
	}

	l.logger.Info().Str("leadId", lead.ID).Str("phone", phone).Msg("[LeadApi] Lead created successfully")
	return lead, nil
}

// UpdateStatus changes the status and notes of a lead.
func (l *LeadApi) UpdateStatus(ctx context.Context, id string, status string, notes string) (model.Lead, error) {
	l.logger.Info().Str("leadId", id).Str("status", status).Msg("[LeadApi] UpdateStatus started")

	lead, err := l.leadPort.FindByID(ctx, id)
	if err != nil {
		l.logger.Error().Err(err).Str("leadId", id).Msg("[LeadApi] Failed to find lead")
		return model.Lead{}, fmt.Errorf("failed to find lead: %w", err)
	}

	lead.Status = status
	lead.Notes = notes
	lead.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := l.leadPort.Update(ctx, lead); err != nil {
		l.logger.Error().Err(err).Str("leadId", id).Msg("[LeadApi] Failed to update lead status")
		return model.Lead{}, fmt.Errorf("failed to update lead: %w", err)
	}

	l.logger.Info().Str("leadId", id).Str("status", status).Msg("[LeadApi] Lead status updated successfully")
	return lead, nil
}

// List retrieves a paginated list of leads.
func (l *LeadApi) List(ctx context.Context, filters port.LeadFilters) ([]model.Lead, int, error) {
	l.logger.Debug().Int("limit", filters.Limit).Int("offset", filters.Offset).Msg("[LeadApi] Listing leads")
	leads, total, err := l.leadPort.List(ctx, filters)
	if err != nil {
		l.logger.Error().Err(err).Msg("[LeadApi] Failed to list leads")
		return nil, 0, fmt.Errorf("failed to list leads: %w", err)
	}
	l.logger.Debug().Int("total", total).Int("returned", len(leads)).Msg("[LeadApi] Leads listed")
	return leads, total, nil
}

// FindByID retrieves a specific lead by its ID.
func (l *LeadApi) FindByID(ctx context.Context, id string) (model.Lead, error) {
	l.logger.Debug().Str("leadId", id).Msg("[LeadApi] Finding lead by ID")
	lead, err := l.leadPort.FindByID(ctx, id)
	if err != nil {
		l.logger.Error().Err(err).Str("leadId", id).Msg("[LeadApi] Failed to find lead")
		return model.Lead{}, fmt.Errorf("failed to find lead: %w", err)
	}
	return lead, nil
}
