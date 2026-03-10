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
	existing, err := l.leadPort.FindByPhone(ctx, phone)
	if err == nil {
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
			return model.Lead{}, fmt.Errorf("failed to update lead: %w", err)
		}
		return existing, nil
	}

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
		return model.Lead{}, fmt.Errorf("failed to create lead: %w", err)
	}

	l.logger.Info().Str("phone", phone).Msg("Lead created")
	return lead, nil
}

// UpdateStatus changes the status and notes of a lead.
func (l *LeadApi) UpdateStatus(ctx context.Context, id string, status string, notes string) (model.Lead, error) {
	lead, err := l.leadPort.FindByID(ctx, id)
	if err != nil {
		return model.Lead{}, fmt.Errorf("failed to find lead: %w", err)
	}

	lead.Status = status
	lead.Notes = notes
	lead.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := l.leadPort.Update(ctx, lead); err != nil {
		return model.Lead{}, fmt.Errorf("failed to update lead: %w", err)
	}

	l.logger.Info().Str("leadId", id).Str("status", status).Msg("Lead status updated")
	return lead, nil
}

// List retrieves a paginated list of leads.
func (l *LeadApi) List(ctx context.Context, filters port.LeadFilters) ([]model.Lead, int, error) {
	leads, total, err := l.leadPort.List(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list leads: %w", err)
	}
	return leads, total, nil
}

// FindByID retrieves a specific lead by its ID.
func (l *LeadApi) FindByID(ctx context.Context, id string) (model.Lead, error) {
	lead, err := l.leadPort.FindByID(ctx, id)
	if err != nil {
		return model.Lead{}, fmt.Errorf("failed to find lead: %w", err)
	}
	return lead, nil
}
