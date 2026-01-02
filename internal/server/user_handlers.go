package server

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"

	"github.com/neilsmahajan/productivity-timer/internal/models"
)

// callbackHandler godoc
// @Summary OAuth callback handler
// @Description Handles the OAuth callback from the provider and creates/updates user session
// @Tags auth
// @Param provider path string true "OAuth provider (e.g., google, github)"
// @Success 307 {string} string "Redirect to home page"
// @Router /auth/{provider}/callback [get]
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

// logoutHandler godoc
// @Summary Logout user
// @Description Clears user session and logs out from OAuth provider
// @Tags auth
// @Param provider path string true "OAuth provider (e.g., google, github)"
// @Success 307 {string} string "Redirect to home page"
// @Router /logout/{provider} [get]
func (s *Server) logoutHandler(c *gin.Context) {
	// Clear Gothic session (OAuth state)
	if err := gothic.Logout(c.Writer, c.Request); err != nil {
		log.Printf("Error clearing gothic session: %v", err)
	}

	// Clear our custom user session
	if err := s.auth.ClearUserSession(c.Writer, c.Request); err != nil {
		log.Printf("Error clearing user session: %v", err)
	}

	c.Redirect(http.StatusTemporaryRedirect, "/")
}

// authHandler godoc
// @Summary Initiate OAuth authentication
// @Description Begins the OAuth authentication flow with the specified provider
// @Tags auth
// @Param provider path string true "OAuth provider (e.g., google, github)"
// @Success 307 {string} string "Redirect to OAuth provider"
// @Router /auth/{provider} [get]
func (s *Server) authHandler(c *gin.Context) {
	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	// Check if user is already authenticated
	gothUser, err := s.auth.GetUserFromSession(c.Request)
	if err == nil && gothUser != nil {
		// Check if user exists in database
		user, dbErr := s.db.GetUserByID(c.Request.Context(), gothUser.UserID)
		if dbErr == nil && user != nil {
			// User exists in both session and database, redirect to index
			c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}
		// User in session but not in database, clear session and re-authenticate
		log.Printf("User in session but not in database, clearing session")
		if clearErr := s.auth.ClearUserSession(c.Writer, c.Request); clearErr != nil {
			log.Printf("Error clearing session: %v", clearErr)
		}
	}

	// Not authenticated, begin OAuth flow
	gothic.BeginAuthHandler(c.Writer, c.Request)
}
