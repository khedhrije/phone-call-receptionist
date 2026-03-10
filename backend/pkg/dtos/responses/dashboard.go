package responses

// DashboardStatsResponse contains the dashboard overview statistics.
type DashboardStatsResponse struct {
	// TotalCalls is the total number of inbound calls.
	TotalCalls int `json:"totalCalls"`
	// TotalAppointments is the total number of appointments booked.
	TotalAppointments int `json:"totalAppointments"`
	// TotalLeads is the total number of leads captured.
	TotalLeads int `json:"totalLeads"`
	// AvgCostPerCall is the average total cost per call in US dollars.
	AvgCostPerCall float64 `json:"avgCostPerCall"`
	// TotalCostUSD is the total cost across all calls in US dollars.
	TotalCostUSD float64 `json:"totalCostUsd"`
	// CallsToday is the number of calls received today.
	CallsToday int `json:"callsToday"`
}

// CostAnalyticsResponse contains cost breakdown data for a time range.
type CostAnalyticsResponse struct {
	// Days is the daily cost breakdown.
	Days []DayCostResponse `json:"days"`
	// TotalTwilioCost is the total Twilio cost in the period.
	TotalTwilioCost float64 `json:"totalTwilioCost"`
	// TotalLLMCost is the total LLM cost in the period.
	TotalLLMCost float64 `json:"totalLlmCost"`
	// TotalCost is the combined total cost in the period.
	TotalCost float64 `json:"totalCost"`
}

// DayCostResponse contains cost data for a single day.
type DayCostResponse struct {
	// Date is the day in YYYY-MM-DD format.
	Date string `json:"date"`
	// TwilioCost is the Twilio cost for this day.
	TwilioCost float64 `json:"twilioCost"`
	// LLMCost is the LLM cost for this day.
	LLMCost float64 `json:"llmCost"`
	// TotalCost is the total cost for this day.
	TotalCost float64 `json:"totalCost"`
	// CallCount is the number of calls on this day.
	CallCount int `json:"callCount"`
}

// CallVolumeResponse contains call volume data for a time range.
type CallVolumeResponse struct {
	// Days is the daily call volume.
	Days []DayVolumeResponse `json:"days"`
	// TotalCalls is the total number of calls in the period.
	TotalCalls int `json:"totalCalls"`
}

// DayVolumeResponse contains call volume data for a single day.
type DayVolumeResponse struct {
	// Date is the day in YYYY-MM-DD format.
	Date string `json:"date"`
	// CallCount is the number of calls on this day.
	CallCount int `json:"callCount"`
	// AvgDuration is the average call duration in seconds.
	AvgDuration int `json:"avgDuration"`
}
