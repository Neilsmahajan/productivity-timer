package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neilsmahajan/productivity-timer/internal/models"
	"github.com/neilsmahajan/productivity-timer/web/templates"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Server) startTimerHandler(c *gin.Context) {
	gothUser, tag, err := s.getGothUserAndTag(c)
	if err != nil || gothUser == nil || tag == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		return
	}
	fmt.Printf("start timer handler, gothUser: %s, gothTag: %s", gothUser, tag)

	// Find or create tag
	if _, err = s.db.FindOrCreateTagStats(c, gothUser.UserID, tag); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}

	timerSession, err := s.db.FindTimerSession(c.Request.Context(), gothUser.UserID, tag)
	if errors.Is(err, mongo.ErrNoDocuments) {
		fmt.Println("find timer session not found")
		timerSession = models.NewTimerSession(gothUser.UserID, tag)

		// upsert timer session
		if err = s.db.CreateTimerSession(context.Background(), timerSession); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
			return
		}
		component := templates.TimerRunning(timerSession, 0)
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err = component.Render(context.Background(), c.Writer); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
			return
		}
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}
	fmt.Println("found timerSession:", timerSession)

	timerSession.Status = "running"
	timerSession.LastUpdated = time.Now()

	if err = s.db.UpdateTimerSession(context.Background(), timerSession); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}

	component := templates.TimerRunning(timerSession, timerSession.Duration)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err = component.Render(context.Background(), c.Writer); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}
}

func (s *Server) getCurrentTimerHandler(c *gin.Context) {
	gothUser, err := s.auth.GetUserFromSession(c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		return
	}

	tag := c.Query("tag")
	if tag == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		return
	}

	timerSession, err := s.db.FindTimerSession(context.Background(), gothUser.UserID, tag)
	if err != nil || timerSession == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
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
		return
	}

	timerSession, err := s.db.FindTimerSession(context.Background(), gothUser.UserID, tag)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if timerSession == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		return
	}

	elapsedTime := int64(time.Now().Sub(timerSession.LastUpdated).Seconds())

	timerSession.Duration += elapsedTime
	timerSession.Status = "stopped"
	timerSession.LastUpdated = time.Now()
	if err = s.db.UpdateTimerSession(context.Background(), timerSession); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}

	userTagStats, err := s.db.FindTagStats(context.Background(), gothUser.UserID, tag)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}

	userTagStats.LastUpdated = time.Now()
	userTagStats.TotalDuration += elapsedTime
	userTagStats.SessionCount++

	if err = s.db.UpdateTagStats(context.Background(), userTagStats); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
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
		return
	}
}
