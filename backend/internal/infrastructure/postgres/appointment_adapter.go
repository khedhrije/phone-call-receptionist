package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"

	"github.com/rs/zerolog"
)

// appointmentAdapter implements port.Appointment using PostgreSQL.
type appointmentAdapter struct {
	client *Client
	logger *zerolog.Logger
}

// NewAppointmentAdapter creates a new PostgreSQL appointment adapter.
func NewAppointmentAdapter(client *Client, logger *zerolog.Logger) port.Appointment {
	return &appointmentAdapter{client: client, logger: logger}
}

// Create persists a new appointment to PostgreSQL.
func (a *appointmentAdapter) Create(ctx context.Context, appt model.Appointment) error {
	a.logger.Debug().Str("id", appt.ID).Str("callerPhone", appt.CallerPhone).Msg("[PostgresAppointment] creating appointment")

	var db AppointmentDB
	db.FromDomain(appt)

	query := `INSERT INTO appointments (id, call_id, caller_phone, caller_name, caller_email, service_type,
	           scheduled_at, duration_mins, status, google_event_id, sms_sent_at, notes, created_at, updated_at)
	           VALUES (:id, :call_id, :caller_phone, :caller_name, :caller_email, :service_type,
	           :scheduled_at, :duration_mins, :status, :google_event_id, :sms_sent_at, :notes, :created_at, :updated_at)`

	_, err := a.client.DB.NamedExecContext(ctx, query, db)
	if err != nil {
		a.logger.Error().Err(err).Str("id", appt.ID).Msg("[PostgresAppointment] failed to create appointment")
		return fmt.Errorf("failed to create appointment: %w", err)
	}

	a.logger.Debug().Str("id", appt.ID).Msg("[PostgresAppointment] appointment created")
	return nil
}

// FindByID retrieves an appointment by its unique identifier from PostgreSQL.
func (a *appointmentAdapter) FindByID(ctx context.Context, id string) (model.Appointment, error) {
	a.logger.Debug().Str("id", id).Msg("[PostgresAppointment] finding appointment by ID")

	var db AppointmentDB
	query := `SELECT id, call_id, caller_phone, caller_name, caller_email, service_type,
	           scheduled_at, duration_mins, status, google_event_id, sms_sent_at, notes, created_at, updated_at
	           FROM appointments WHERE id = $1`

	if err := a.client.DB.GetContext(ctx, &db, query, id); err != nil {
		a.logger.Error().Err(err).Str("id", id).Msg("[PostgresAppointment] failed to find appointment by ID")
		return model.Appointment{}, fmt.Errorf("failed to find appointment by id: %w", err)
	}

	a.logger.Debug().Str("id", id).Msg("[PostgresAppointment] appointment found")
	return db.ToDomain(), nil
}

// FindByPhone retrieves all appointments for a given phone number from PostgreSQL.
func (a *appointmentAdapter) FindByPhone(ctx context.Context, phone string) ([]model.Appointment, error) {
	a.logger.Debug().Str("phone", phone).Msg("[PostgresAppointment] finding appointments by phone")

	query := `SELECT id, call_id, caller_phone, caller_name, caller_email, service_type,
	           scheduled_at, duration_mins, status, google_event_id, sms_sent_at, notes, created_at, updated_at
	           FROM appointments WHERE caller_phone = $1 ORDER BY scheduled_at DESC`

	var rows []AppointmentDB
	if err := a.client.DB.SelectContext(ctx, &rows, query, phone); err != nil {
		a.logger.Error().Err(err).Str("phone", phone).Msg("[PostgresAppointment] failed to find appointments by phone")
		return nil, fmt.Errorf("failed to find appointments by phone: %w", err)
	}

	appts := make([]model.Appointment, len(rows))
	for i, row := range rows {
		appts[i] = row.ToDomain()
	}

	a.logger.Debug().Str("phone", phone).Int("count", len(appts)).Msg("[PostgresAppointment] appointments found by phone")
	return appts, nil
}

// List retrieves appointments matching the given filters from PostgreSQL.
func (a *appointmentAdapter) List(ctx context.Context, filters port.AppointmentFilters) ([]model.Appointment, int, error) {
	a.logger.Debug().Int("limit", filters.Limit).Int("offset", filters.Offset).Msg("[PostgresAppointment] listing appointments")
	var conditions []string
	args := make(map[string]interface{})

	if filters.Status != "" {
		conditions = append(conditions, "status = :status")
		args["status"] = filters.Status
	}
	if filters.CallerPhone != "" {
		conditions = append(conditions, "caller_phone = :caller_phone")
		args["caller_phone"] = filters.CallerPhone
	}
	if filters.From != "" {
		if t, err := time.Parse(time.RFC3339, filters.From); err == nil {
			conditions = append(conditions, "scheduled_at >= :from_time")
			args["from_time"] = t
		}
	}
	if filters.To != "" {
		if t, err := time.Parse(time.RFC3339, filters.To); err == nil {
			conditions = append(conditions, "scheduled_at <= :to_time")
			args["to_time"] = t
		}
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total matching records.
	countQuery := "SELECT COUNT(*) FROM appointments" + where
	countStmt, countArgs, err := a.client.DB.BindNamed(countQuery, args)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to bind count query: %w", err)
	}

	var total int
	if err := a.client.DB.GetContext(ctx, &total, countStmt, countArgs...); err != nil {
		a.logger.Error().Err(err).Msg("[PostgresAppointment] failed to count appointments")
		return nil, 0, fmt.Errorf("failed to count appointments: %w", err)
	}

	// Fetch paginated results.
	selectQuery := `SELECT id, call_id, caller_phone, caller_name, caller_email, service_type,
	                 scheduled_at, duration_mins, status, google_event_id, sms_sent_at, notes, created_at, updated_at
	                 FROM appointments` + where + ` ORDER BY scheduled_at DESC LIMIT :limit OFFSET :offset`

	args["limit"] = filters.Limit
	args["offset"] = filters.Offset

	selectStmt, selectArgs, err := a.client.DB.BindNamed(selectQuery, args)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to bind select query: %w", err)
	}

	var rows []AppointmentDB
	if err := a.client.DB.SelectContext(ctx, &rows, selectStmt, selectArgs...); err != nil {
		a.logger.Error().Err(err).Msg("[PostgresAppointment] failed to list appointments")
		return nil, 0, fmt.Errorf("failed to list appointments: %w", err)
	}

	appts := make([]model.Appointment, len(rows))
	for i, row := range rows {
		appts[i] = row.ToDomain()
	}

	a.logger.Debug().Int("count", len(appts)).Int("total", total).Msg("[PostgresAppointment] appointments listed")
	return appts, total, nil
}

// Update modifies an existing appointment's data in PostgreSQL.
func (a *appointmentAdapter) Update(ctx context.Context, appt model.Appointment) error {
	a.logger.Debug().Str("id", appt.ID).Str("status", appt.Status).Msg("[PostgresAppointment] updating appointment")

	var db AppointmentDB
	db.FromDomain(appt)

	query := `UPDATE appointments SET call_id = :call_id, caller_phone = :caller_phone,
	           caller_name = :caller_name, caller_email = :caller_email, service_type = :service_type,
	           scheduled_at = :scheduled_at, duration_mins = :duration_mins, status = :status,
	           google_event_id = :google_event_id, sms_sent_at = :sms_sent_at, notes = :notes,
	           updated_at = :updated_at WHERE id = :id`

	_, err := a.client.DB.NamedExecContext(ctx, query, db)
	if err != nil {
		a.logger.Error().Err(err).Str("id", appt.ID).Msg("[PostgresAppointment] failed to update appointment")
		return fmt.Errorf("failed to update appointment: %w", err)
	}

	a.logger.Debug().Str("id", appt.ID).Msg("[PostgresAppointment] appointment updated")
	return nil
}

// Delete removes an appointment from PostgreSQL.
func (a *appointmentAdapter) Delete(ctx context.Context, id string) error {
	a.logger.Debug().Str("id", id).Msg("[PostgresAppointment] deleting appointment")

	query := `DELETE FROM appointments WHERE id = $1`

	_, err := a.client.DB.ExecContext(ctx, query, id)
	if err != nil {
		a.logger.Error().Err(err).Str("id", id).Msg("[PostgresAppointment] failed to delete appointment")
		return fmt.Errorf("failed to delete appointment: %w", err)
	}

	a.logger.Debug().Str("id", id).Msg("[PostgresAppointment] appointment deleted")
	return nil
}
