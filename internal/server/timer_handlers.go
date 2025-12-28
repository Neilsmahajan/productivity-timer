package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neilsmahajan/productivity-timer/internal/models"
	"github.com/neilsmahajan/productivity-timer/web/templates"
)

func (s *Server) startTimerHandler(c *gin.Context) {
	gothUser, err := s.auth.GetUserFromSession(c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"success": false,
		})
	}

	tag := c.PostForm("tag")
	timerSession := models.NewTimerSession(gothUser.UserID, tag)

	// upsert timer session
	if err = s.db.UpsertTimerSession(context.Background(), timerSession); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
	}

	component := templates.TimerRunning(timerSession, 0)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err = component.Render(context.Background(), c.Writer); err != nil {
		return
	}
}

func (s *Server) getCurrentTimerHandler(c *gin.Context) {
	gothUser, err := s.auth.GetUserFromSession(c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"success": false,
		})
	}

	tag := c.Query("tag")
	timerSession, err := s.db.GetTimerSessionByID(context.Background(), gothUser.UserID, tag)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
	}
	if timerSession == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
	}

	component := templates.TimerRunning(timerSession, timerSession.Duration)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err = component.Render(context.Background(), c.Writer); err != nil {
		return
	}
}
