package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neilsmahajan/productivity-timer/internal/models"
	"github.com/neilsmahajan/productivity-timer/web/templates"
	"go.mongodb.org/mongo-driver/mongo"
)

// startTimerHandler godoc
// @Summary Start a timer session
// @Description Starts a new timer session or resumes an existing stopped session for the specified tag
// @Tags timer
// @Accept x-www-form-urlencoded
// @Produce html
// @Param tag formData string true "Tag name for the timer session"
// @Success 200 {string} string "HTML component for running timer"
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/timer/start [post]
func (s *Server) startTimerHandler(c *gin.Context) {
	gothUser, tag, err := s.getGothUserAndTag(c)
	if err != nil || gothUser == nil || tag == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		return
	}

	// Abandon any orphaned running timers for this user+tag (e.g., from closed tabs)
	if err = s.db.AbandonRunningTimers(c.Request.Context(), gothUser.UserID, tag); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}

	currentTime := time.Now()
	timerSession, err := s.db.FindTimerSession(c.Request.Context(), gothUser.UserID, tag, models.StatusStopped)
	if errors.Is(err, mongo.ErrNoDocuments) {
		timerSession = models.NewTimerSession(gothUser.UserID, tag)

		if err = s.db.CreateTimerSession(c.Request.Context(), timerSession); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
			return
		}

		userTagStats, err2 := s.db.FindUserTagStats(c.Request.Context(), gothUser.UserID, tag)
		if errors.Is(err2, mongo.ErrNoDocuments) {
			userTagStats = models.NewUserTagStats(gothUser.UserID, tag)
			if err2 = s.db.CreateUserTagStats(c.Request.Context(), userTagStats); err2 != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
				return
			}
		} else if err2 != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
			return
		} else {
			userTagStats.SessionCount++
			userTagStats.LastUpdated = currentTime
			if err2 = s.db.UpdateUserTagStats(c.Request.Context(), userTagStats); err2 != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
				return
			}
		}
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	} else {
		timerSession.Status = models.StatusRunning
		timerSession.LastUpdated = currentTime

		if err = s.db.UpdateTimerSession(c.Request.Context(), timerSession); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}

	component := templates.TimerRunning(timerSession, timerSession.Duration)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err = component.Render(c.Request.Context(), c.Writer); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}
}

// stopTimerHandler godoc
// @Summary Stop a running timer
// @Description Stops the currently running timer session and updates the elapsed time
// @Tags timer
// @Accept x-www-form-urlencoded
// @Produce html
// @Param tag formData string true "Tag name for the timer session"
// @Success 200 {string} string "HTML component for stopped timer"
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/timer/stop [post]
func (s *Server) stopTimerHandler(c *gin.Context) {
	gothUser, tag, err := s.getGothUserAndTag(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		return
	}

	timerSession, err := s.db.FindTimerSession(c.Request.Context(), gothUser.UserID, tag, models.StatusRunning)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if timerSession == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		return
	}

	currentTime := time.Now()
	elapsedTime := int64(currentTime.Sub(timerSession.LastUpdated).Seconds())

	timerSession.Duration += elapsedTime
	timerSession.Status = models.StatusStopped
	timerSession.LastUpdated = currentTime
	if err = s.db.UpdateTimerSession(c.Request.Context(), timerSession); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}

	userTagStats, err := s.db.FindUserTagStats(c.Request.Context(), gothUser.UserID, tag)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}

	userTagStats.LastUpdated = currentTime
	userTagStats.TotalDuration += elapsedTime

	if err = s.db.UpdateUserTagStats(c.Request.Context(), userTagStats); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}

	component := templates.TimerStopped(timerSession, timerSession.Duration)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err = component.Render(c.Request.Context(), c.Writer); err != nil {
		return
	}
}

// resetTimerHandler godoc
// @Summary Reset and complete a timer session
// @Description Marks the current timer session as completed and returns to idle state
// @Tags timer
// @Accept x-www-form-urlencoded
// @Produce html
// @Param tag formData string true "Tag name for the timer session"
// @Success 200 {string} string "HTML component for idle timer"
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/timer/reset [post]
func (s *Server) resetTimerHandler(c *gin.Context) {
	gothUser, tag, err := s.getGothUserAndTag(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		return
	}

	timerSession, err := s.db.FindTimerSession(c.Request.Context(), gothUser.UserID, tag, models.StatusStopped)
	if err != nil || timerSession == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}

	currentTime := time.Now()
	timerSession.Status = models.StatusCompleted
	timerSession.LastUpdated = currentTime
	timerSession.EndTime = &currentTime
	if err = s.db.UpdateTimerSession(c.Request.Context(), timerSession); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}

	allUserTagStats, err := s.db.FindAllUserTagStats(c.Request.Context(), gothUser.UserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}
	tags := make([]string, 0, len(allUserTagStats))
	for _, tagStats := range allUserTagStats {
		tags = append(tags, tagStats.Tag)
	}

	component := templates.TimerIdle(tags)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err = component.Render(c.Request.Context(), c.Writer); err != nil {
		return
	}
}
