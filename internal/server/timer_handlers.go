package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neilsmahajan/productivity-timer/internal/models"
	"github.com/neilsmahajan/productivity-timer/web/templates"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Server) startTimerHandler(c *gin.Context) {
	gothUser, tag, err := s.getGothUserAndTag(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
	}

	timerSession, err := s.db.GetTimerSession(c.Request.Context(), gothUser.UserID, tag)
	if errors.Is(err, mongo.ErrNoDocuments) {
		timerSession = models.NewTimerSession(gothUser.UserID, tag)

		// upsert timer session
		if err = s.db.CreateTimerSession(context.Background(), timerSession); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		}
		component := templates.TimerRunning(timerSession, 0)
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err = component.Render(context.Background(), c.Writer); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		}
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
	}

	timerSession.Status = "running"
	timerSession.LastUpdated = time.Now()

	if err = s.db.UpdateTimerSession(context.Background(), timerSession); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
	}

	component := templates.TimerRunning(timerSession, timerSession.Duration)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err = component.Render(context.Background(), c.Writer); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
	}
}

func (s *Server) getCurrentTimerHandler(c *gin.Context) {
	gothUser, tag, err := s.getGothUserAndTag(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
	}

	timerSession, err := s.db.GetTimerSession(context.Background(), gothUser.UserID, tag)
	if err != nil || timerSession == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
	}

	elapsedTime := int64(time.Now().Sub(timerSession.LastUpdated).Seconds())

	component := templates.TimerRunning(timerSession, timerSession.Duration+elapsedTime)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err = component.Render(context.Background(), c.Writer); err != nil {
		return
	}
}

func (s *Server) stopTimerHandler(c *gin.Context) {
	gothUser, tag, err := s.getGothUserAndTag(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
	}

	timerSession, err := s.db.GetTimerSession(context.Background(), gothUser.UserID, tag)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
	}
	if timerSession == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
	}

	elapsedTime := int64(time.Now().Sub(timerSession.LastUpdated).Seconds())

	timerSession.Duration += elapsedTime
	timerSession.Status = "stopped"
	timerSession.LastUpdated = time.Now()
	if err = s.db.UpdateTimerSession(context.Background(), timerSession); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
	}

	userTagStats, err := s.db.FindOrCreateTagStats(context.Background(), gothUser.UserID, tag)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
	}

	userTagStats.LastUpdated = time.Now()
	userTagStats.TotalDuration += elapsedTime

	if err = s.db.UpdateTagStats(context.Background(), userTagStats, elapsedTime); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
	}

	component := templates.TimerStopped(timerSession, timerSession.Duration)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err = component.Render(context.Background(), c.Writer); err != nil {
		return
	}
}

func (s *Server) resetTimerHandler(c *gin.Context) {
	_, _, err := s.getGothUserAndTag(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
	}

	return
}
