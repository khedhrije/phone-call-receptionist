package api

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/port"
	"phone-call-receptionist/backend/pkg/dtos/responses"
)

// DashboardApi provides business operations for dashboard analytics.
type DashboardApi struct {
	callPort port.InboundCall
	apptPort port.Appointment
	leadPort port.Lead
	logger   *zerolog.Logger
}

// NewDashboardApi creates a new DashboardApi with the given dependencies.
func NewDashboardApi(
	callPort port.InboundCall,
	apptPort port.Appointment,
	leadPort port.Lead,
	logger *zerolog.Logger,
) *DashboardApi {
	return &DashboardApi{
		callPort: callPort,
		apptPort: apptPort,
		leadPort: leadPort,
		logger:   logger,
	}
}

// Stats retrieves the dashboard overview statistics.
func (d *DashboardApi) Stats(ctx context.Context) (responses.DashboardStatsResponse, error) {
	d.logger.Info().Msg("[DashboardApi] Stats started")

	calls, totalCalls, err := d.callPort.List(ctx, port.CallFilters{Limit: 1})
	if err != nil {
		d.logger.Error().Err(err).Msg("[DashboardApi] Failed to count calls")
		return responses.DashboardStatsResponse{}, fmt.Errorf("failed to count calls: %w", err)
	}
	d.logger.Debug().Int("totalCalls", totalCalls).Msg("[DashboardApi] Calls counted")

	_, totalAppts, err := d.apptPort.List(ctx, port.AppointmentFilters{Limit: 1})
	if err != nil {
		d.logger.Error().Err(err).Msg("[DashboardApi] Failed to count appointments")
		return responses.DashboardStatsResponse{}, fmt.Errorf("failed to count appointments: %w", err)
	}
	d.logger.Debug().Int("totalAppts", totalAppts).Msg("[DashboardApi] Appointments counted")

	_, totalLeads, err := d.leadPort.List(ctx, port.LeadFilters{Limit: 1})
	if err != nil {
		d.logger.Error().Err(err).Msg("[DashboardApi] Failed to count leads")
		return responses.DashboardStatsResponse{}, fmt.Errorf("failed to count leads: %w", err)
	}
	d.logger.Debug().Int("totalLeads", totalLeads).Msg("[DashboardApi] Leads counted")

	var totalCost float64
	allCalls, _, _ := d.callPort.List(ctx, port.CallFilters{Limit: 10000})
	for _, c := range allCalls {
		totalCost += c.TotalCostUSD
	}

	var avgCost float64
	if totalCalls > 0 {
		avgCost = totalCost / float64(totalCalls)
	}

	_ = calls

	d.logger.Info().Int("totalCalls", totalCalls).Int("totalAppts", totalAppts).Int("totalLeads", totalLeads).Float64("totalCost", totalCost).Msg("[DashboardApi] Stats completed")
	return responses.DashboardStatsResponse{
		TotalCalls:        totalCalls,
		TotalAppointments: totalAppts,
		TotalLeads:        totalLeads,
		AvgCostPerCall:    avgCost,
		TotalCostUSD:      totalCost,
	}, nil
}

// CostAnalytics retrieves cost breakdown data for the given time range.
func (d *DashboardApi) CostAnalytics(ctx context.Context, from string, to string) (responses.CostAnalyticsResponse, error) {
	d.logger.Info().Str("from", from).Str("to", to).Msg("[DashboardApi] CostAnalytics started")

	calls, _, err := d.callPort.List(ctx, port.CallFilters{From: from, To: to, Limit: 10000})
	if err != nil {
		d.logger.Error().Err(err).Msg("[DashboardApi] Failed to list calls for cost analytics")
		return responses.CostAnalyticsResponse{}, fmt.Errorf("failed to list calls: %w", err)
	}

	dayMap := make(map[string]*responses.DayCostResponse)
	var totalTwilio, totalLLM float64

	for _, call := range calls {
		date := call.CreatedAt[:10]
		day, ok := dayMap[date]
		if !ok {
			day = &responses.DayCostResponse{Date: date}
			dayMap[date] = day
		}
		day.TwilioCost += call.TwilioCostUSD
		day.LLMCost += call.LLMCostUSD
		day.TotalCost += call.TotalCostUSD
		day.CallCount++
		totalTwilio += call.TwilioCostUSD
		totalLLM += call.LLMCostUSD
	}

	var days []responses.DayCostResponse
	for _, day := range dayMap {
		days = append(days, *day)
	}

	d.logger.Info().Int("days", len(days)).Float64("totalCost", totalTwilio+totalLLM).Msg("[DashboardApi] CostAnalytics completed")
	return responses.CostAnalyticsResponse{
		Days:            days,
		TotalTwilioCost: totalTwilio,
		TotalLLMCost:    totalLLM,
		TotalCost:       totalTwilio + totalLLM,
	}, nil
}

// CallVolume retrieves call volume data for the given time range.
func (d *DashboardApi) CallVolume(ctx context.Context, from string, to string) (responses.CallVolumeResponse, error) {
	d.logger.Info().Str("from", from).Str("to", to).Msg("[DashboardApi] CallVolume started")

	calls, _, err := d.callPort.List(ctx, port.CallFilters{From: from, To: to, Limit: 10000})
	if err != nil {
		d.logger.Error().Err(err).Msg("[DashboardApi] Failed to list calls for call volume")
		return responses.CallVolumeResponse{}, fmt.Errorf("failed to list calls: %w", err)
	}

	dayMap := make(map[string]*responses.DayVolumeResponse)
	totalCalls := 0

	for _, call := range calls {
		date := call.CreatedAt[:10]
		day, ok := dayMap[date]
		if !ok {
			day = &responses.DayVolumeResponse{Date: date}
			dayMap[date] = day
		}
		day.CallCount++
		day.AvgDuration += call.DurationSeconds
		totalCalls++
	}

	var days []responses.DayVolumeResponse
	for _, day := range dayMap {
		if day.CallCount > 0 {
			day.AvgDuration = day.AvgDuration / day.CallCount
		}
		days = append(days, *day)
	}

	d.logger.Info().Int("days", len(days)).Int("totalCalls", totalCalls).Msg("[DashboardApi] CallVolume completed")
	return responses.CallVolumeResponse{
		Days:       days,
		TotalCalls: totalCalls,
	}, nil
}
