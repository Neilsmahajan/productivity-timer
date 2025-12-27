package server

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/neilsmahajan/productivity-timer/web/templates"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*", "http://localhost:3000"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	r.GET("/", s.indexHandler)

	r.GET("/health", s.healthHandler)

	r.GET("/auth/:provider/callback", s.callbackHandler)

	r.GET("/logout/:provider", s.logoutHandler)

	r.GET("/auth/:provider", s.authHandler)

	return r
}

func (s *Server) indexHandler(c *gin.Context) {
	// Try to get user from session
	gothUser, err := s.auth.GetUserFromSession(c.Request)
	if err != nil || gothUser == nil {
		// No user logged in, show index page with login button
		component := templates.LoginPage()
		if err = component.Render(context.Background(), c.Writer); err != nil {
			log.Printf("Error rendering index page: %v", err)
			c.String(http.StatusInternalServerError, "Error rendering page")
			return
		}
		return
	}

	// User is logged in, get from database
	user, err := s.db.GetUserByID(c.Request.Context(), gothUser.UserID)
	if err != nil {
		log.Printf("Error getting user from database: %v", err)
		c.String(http.StatusInternalServerError, "Error getting user")
		return
	}

	if user == nil {
		// User not in database, show login page
		component := templates.LoginPage()
		if err = component.Render(context.Background(), c.Writer); err != nil {
			log.Printf("Error rendering index page: %v", err)
			c.String(http.StatusInternalServerError, "Error rendering page")
			return
		}
		return
	}

	// Show user page with user information
	// TODO: Get active timer session from database once implemented
	component := templates.TimerPage(*gothUser, nil)
	if err := component.Render(context.Background(), c.Writer); err != nil {
		log.Printf("Error rendering user page: %v", err)
		c.String(http.StatusInternalServerError, "Error rendering page")
		return
	}
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
