package server

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	"github.com/neilsmahajan/productivity-timer/internal/models"
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
		component := templates.IndexPage()
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
		component := templates.IndexPage()
		if err := component.Render(context.Background(), c.Writer); err != nil {
			log.Printf("Error rendering index page: %v", err)
			c.String(http.StatusInternalServerError, "Error rendering page")
			return
		}
		return
	}

	// Show user page with user information
	component := templates.UserPage(*gothUser)
	if err := component.Render(context.Background(), c.Writer); err != nil {
		log.Printf("Error rendering user page: %v", err)
		c.String(http.StatusInternalServerError, "Error rendering page")
		return
	}
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) callbackHandler(c *gin.Context) {
	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	// Complete the OAuth authentication flow
	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		log.Printf("Error completing auth: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	// Convert goth.User to our User model and save to database
	user := models.FromGothUser(gothUser)
	_, err = s.db.FindOrCreateUser(c.Request.Context(), user)
	if err != nil {
		log.Printf("Error saving user to database: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	// Store user in our custom session
	err = s.auth.StoreUserInSession(c.Writer, c.Request, &gothUser)
	if err != nil {
		log.Printf("Error storing user in session: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	// Redirect to index page (will show user info)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (s *Server) logoutHandler(c *gin.Context) {
	if err := gothic.Logout(c.Writer, c.Request); err != nil {
		return
	}

	c.Header("Location", "/")
	c.Status(http.StatusTemporaryRedirect)
}

func (s *Server) authHandler(c *gin.Context) {
	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	// Check if user is already authenticated
	gothUser, err := s.auth.GetUserFromSession(c.Request)
	if err == nil && gothUser != nil {
		// Already authenticated, redirect to index
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	// Not authenticated, begin OAuth flow
	gothic.BeginAuthHandler(c.Writer, c.Request)
}
