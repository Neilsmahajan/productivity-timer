package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
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
	component := templates.IndexPage()
	if err := component.Render(context.Background(), c.Writer); err != nil {
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

	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		_, err = fmt.Fprintln(c.Writer, err)
		if err != nil {
			return
		}
	}

	component := templates.UserPage(gothUser)
	if err = component.Render(context.Background(), c.Writer); err != nil {
		return
	}
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

	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request); err == nil {
		component := templates.UserPage(gothUser)
		if err = component.Render(context.Background(), c.Writer); err != nil {
			return
		}
	} else {
		gothic.BeginAuthHandler(c.Writer, c.Request)
	}
}
