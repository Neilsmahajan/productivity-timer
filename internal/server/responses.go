package server

import "github.com/neilsmahajan/productivity-timer/internal/models"

// ErrorResponse represents an API error response
// @Description Error response returned when an API request fails
type ErrorResponse struct {
	Error   string `json:"error" example:"unauthorized"`
	Message string `json:"message,omitempty" example:"User not authenticated"`
}

// HealthResponse represents the health check response
// @Description Health check response with database status
type HealthResponse struct {
	Status  string `json:"status" example:"up"`
	Message string `json:"message,omitempty" example:"It's healthy"`
}

// TimerResponse represents a timer session response
// @Description Timer session state returned after timer operations
type TimerResponse struct {
	Session  *models.TimerSession `json:"session"`
	Duration int64                `json:"duration" example:"3600"`
	Status   string               `json:"status" example:"running"`
}

// StatsQueryParams represents the query parameters for stats endpoints
// @Description Query parameters for filtering stats by date range
type StatsQueryParams struct {
	Start string `form:"start" example:"2026-01-01T00:00"`
	End   string `form:"end" example:"2026-01-31T23:59"`
}

// TagSessionsResponse represents the response for tag sessions endpoint
// @Description List of timer sessions for a specific tag
type TagSessionsResponse struct {
	Tag      string                `json:"tag" example:"coding"`
	Sessions []models.TimerSession `json:"sessions"`
}

// TagListResponse represents a list of tags
// @Description List of tags for a user
type TagListResponse struct {
	Tags []string `json:"tags" example:"coding,reading,exercise"`
}
