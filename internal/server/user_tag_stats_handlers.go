package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neilsmahajan/productivity-timer/web/templates"
)

func (s *Server) statsPageHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Try to get user from session
	gothUser, err := s.auth.GetUserFromSession(c.Request)
	if err != nil || gothUser == nil {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	component := templates.StatsPage()
	if err = component.Render(ctx, c.Writer); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

// statsSummaryHandler godoc
// @Summary Get stats summary
// @Description Returns aggregated statistics for the authenticated user within a date range
// @Tags stats
// @Produce html
// @Param start query string false "Start datetime (format: 2006-01-02T15:04)"
// @Param end query string false "End datetime (format: 2006-01-02T15:04)"
// @Success 200 {string} string "HTML component with stats summary"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/stats/summary [get]
func (s *Server) statsSummaryHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Try to get user from session
	gothUser, err := s.auth.GetUserFromSession(c.Request)
	if err != nil || gothUser == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	startDate, endDate, err := parseStatsQueryParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	statsSummary, err := s.db.GetStatsSummary(ctx, gothUser.UserID, startDate, endDate)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	component := templates.StatsSummary(statsSummary)
	if err = component.Render(ctx, c.Writer); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

// parseStatsQueryParams extracts and validates start/end dates from query params
func parseStatsQueryParams(c *gin.Context) (time.Time, time.Time, error) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	// Default to today if no params provided
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())

	if startStr == "" && endStr == "" {
		return startOfDay, endOfDay, nil
	}

	// Parse datetime-local format: "2006-01-02T15:04"
	const layout = "2006-01-02T15:04"

	startDate, err := time.ParseInLocation(layout, startStr, now.Location())
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endDate, err := time.ParseInLocation(layout, endStr, now.Location())
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return startDate, endDate, nil
}

// tagSessionsHandler godoc
// @Summary Get sessions for a specific tag
// @Description Returns all timer sessions for a specific tag within a date range
// @Tags stats
// @Produce html
// @Param tag path string true "Tag name"
// @Param start query string false "Start datetime (format: 2006-01-02T15:04)"
// @Param end query string false "End datetime (format: 2006-01-02T15:04)"
// @Success 200 {string} string "HTML component with tag sessions"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/stats/tag/{tag}/sessions [get]
func (s *Server) tagSessionsHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Try to get user from session
	gothUser, err := s.auth.GetUserFromSession(c.Request)
	if err != nil || gothUser == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	tag := c.Param("tag")
	if tag == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	startDate, endDate, err := parseStatsQueryParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	sessions, err := s.db.GetTagSessions(ctx, gothUser.UserID, tag, startDate, endDate)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	component := templates.TagSessions(tag, sessions)
	if err = component.Render(ctx, c.Writer); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

// deleteTagHandler godoc
// @Summary Delete a tag and all its sessions
// @Description Deletes all timer sessions and statistics for a specific tag
// @Tags stats
// @Param tag path string true "Tag name to delete"
// @Success 200 {string} string "Empty response on successful deletion"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/stats/tag/{tag} [delete]
func (s *Server) deleteTagHandler(c *gin.Context) {
	ctx := c.Request.Context()

	gothUser, err := s.auth.GetUserFromSession(c.Request)
	if err != nil || gothUser == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	tag := c.Param("tag")
	if tag == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err = s.db.DeleteUserTagStats(ctx, gothUser.UserID, tag); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if err = s.db.DeleteTimerSession(ctx, gothUser.UserID, tag); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Return empty response - HTMX will remove the deleted row
	c.Status(http.StatusOK)
}
