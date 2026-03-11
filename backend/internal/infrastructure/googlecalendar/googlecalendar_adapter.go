// Package googlecalendar implements the Calendar port using the Google Calendar API.
package googlecalendar

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"
)

// Adapter implements port.Calendar using the Google Calendar API with service account auth.
type Adapter struct {
	service    *calendar.Service
	calendarID string
	logger     *zerolog.Logger
}

// NewGoogleCalendarAdapter creates a new Google Calendar adapter using service account credentials.
func NewGoogleCalendarAdapter(credentialsJSON string, calendarID string, logger *zerolog.Logger) (port.Calendar, error) {
	ctx := context.Background()
	srv, err := calendar.NewService(ctx, option.WithCredentialsJSON([]byte(credentialsJSON)))
	if err != nil {
		logger.Error().Err(err).Msg("[GoogleCalendarAdapter] failed to create service")
		return nil, fmt.Errorf("failed to create google calendar service: %w", err)
	}

	logger.Info().Str("calendarID", calendarID).Msg("[GoogleCalendarAdapter] connected")

	return &Adapter{
		service:    srv,
		calendarID: calendarID,
		logger:     logger,
	}, nil
}

// CheckAvailability returns available time slots within the given time range
// by querying the Google Calendar FreeBusy API.
func (a *Adapter) CheckAvailability(ctx context.Context, from string, to string) ([]model.TimeSlot, error) {
	a.logger.Debug().Str("from", from).Str("to", to).Msg("[GoogleCalendarAdapter] checking availability")
	freeBusyReq := &calendar.FreeBusyRequest{
		TimeMin: from,
		TimeMax: to,
		Items: []*calendar.FreeBusyRequestItem{
			{Id: a.calendarID},
		},
	}

	resp, err := a.service.Freebusy.Query(freeBusyReq).Context(ctx).Do()
	if err != nil {
		a.logger.Error().Err(err).Str("from", from).Str("to", to).Msg("[GoogleCalendarAdapter] failed to query freebusy")
		return nil, fmt.Errorf("failed to query freebusy: %w", err)
	}

	calBusy, ok := resp.Calendars[a.calendarID]
	if !ok {
		// No busy times means entirely free
		return []model.TimeSlot{{Start: from, End: to}}, nil
	}

	fromTime, err := time.Parse(time.RFC3339, from)
	if err != nil {
		a.logger.Error().Err(err).Str("from", from).Msg("[GoogleCalendarAdapter] failed to parse from time")
		return nil, fmt.Errorf("failed to parse from time: %w", err)
	}
	toTime, err := time.Parse(time.RFC3339, to)
	if err != nil {
		a.logger.Error().Err(err).Str("to", to).Msg("[GoogleCalendarAdapter] failed to parse to time")
		return nil, fmt.Errorf("failed to parse to time: %w", err)
	}

	slots := buildFreeSlots(fromTime, toTime, calBusy.Busy)

	a.logger.Debug().
		Int("freeSlots", len(slots)).
		Str("from", from).
		Str("to", to).
		Msg("[GoogleCalendarAdapter] availability checked")

	return slots, nil
}

// buildFreeSlots computes free time slots from busy periods within a range.
func buildFreeSlots(from time.Time, to time.Time, busy []*calendar.TimePeriod) []model.TimeSlot {
	var slots []model.TimeSlot
	cursor := from

	for _, period := range busy {
		busyStart, err := time.Parse(time.RFC3339, period.Start)
		if err != nil {
			continue
		}
		busyEnd, err := time.Parse(time.RFC3339, period.End)
		if err != nil {
			continue
		}

		if cursor.Before(busyStart) {
			slots = append(slots, model.TimeSlot{
				Start: cursor.Format(time.RFC3339),
				End:   busyStart.Format(time.RFC3339),
			})
		}
		if busyEnd.After(cursor) {
			cursor = busyEnd
		}
	}

	if cursor.Before(to) {
		slots = append(slots, model.TimeSlot{
			Start: cursor.Format(time.RFC3339),
			End:   to.Format(time.RFC3339),
		})
	}

	return slots
}

// CreateEvent creates a calendar event for the given appointment.
// Returns the external event ID and any error.
func (a *Adapter) CreateEvent(ctx context.Context, appt model.Appointment) (string, error) {
	a.logger.Debug().Str("appointmentID", appt.ID).Str("scheduledAt", appt.ScheduledAt).Msg("[GoogleCalendarAdapter] creating event")

	event := appointmentToEvent(appt)

	created, err := a.service.Events.Insert(a.calendarID, event).Context(ctx).Do()
	if err != nil {
		a.logger.Error().Err(err).Str("appointmentID", appt.ID).Msg("[GoogleCalendarAdapter] failed to create event")
		return "", fmt.Errorf("failed to create calendar event: %w", err)
	}

	a.logger.Info().
		Str("eventID", created.Id).
		Str("appointmentID", appt.ID).
		Msg("[GoogleCalendarAdapter] event created")

	return created.Id, nil
}

// UpdateEvent modifies an existing calendar event.
func (a *Adapter) UpdateEvent(ctx context.Context, eventID string, appt model.Appointment) error {
	a.logger.Debug().Str("eventID", eventID).Str("appointmentID", appt.ID).Msg("[GoogleCalendarAdapter] updating event")

	event := appointmentToEvent(appt)

	_, err := a.service.Events.Update(a.calendarID, eventID, event).Context(ctx).Do()
	if err != nil {
		a.logger.Error().Err(err).Str("eventID", eventID).Str("appointmentID", appt.ID).Msg("[GoogleCalendarAdapter] failed to update event")
		return fmt.Errorf("failed to update calendar event: %w", err)
	}

	a.logger.Info().
		Str("eventID", eventID).
		Str("appointmentID", appt.ID).
		Msg("[GoogleCalendarAdapter] event updated")

	return nil
}

// DeleteEvent removes a calendar event by its external event ID.
func (a *Adapter) DeleteEvent(ctx context.Context, eventID string) error {
	a.logger.Debug().Str("eventID", eventID).Msg("[GoogleCalendarAdapter] deleting event")

	err := a.service.Events.Delete(a.calendarID, eventID).Context(ctx).Do()
	if err != nil {
		a.logger.Error().Err(err).Str("eventID", eventID).Msg("[GoogleCalendarAdapter] failed to delete event")
		return fmt.Errorf("failed to delete calendar event: %w", err)
	}

	a.logger.Info().Str("eventID", eventID).Msg("[GoogleCalendarAdapter] event deleted")
	return nil
}

// appointmentToEvent converts a domain appointment to a Google Calendar event.
func appointmentToEvent(appt model.Appointment) *calendar.Event {
	startTime, _ := time.Parse(time.RFC3339, appt.ScheduledAt)
	endTime := startTime.Add(time.Duration(appt.DurationMins) * time.Minute)

	summary := fmt.Sprintf("%s - %s", appt.ServiceType, appt.CallerName)
	description := fmt.Sprintf(
		"Phone: %s\nEmail: %s\nService: %s\nNotes: %s",
		appt.CallerPhone, appt.CallerEmail, appt.ServiceType, appt.Notes,
	)

	return &calendar.Event{
		Summary:     summary,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: appt.ScheduledAt,
			TimeZone: "UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		Attendees: []*calendar.EventAttendee{
			{Email: appt.CallerEmail},
		},
	}
}
