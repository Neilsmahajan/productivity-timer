package server

import (
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

	r.POST("/timer/start", s.startTimerHandler)

	r.POST("/timer/stop", s.stopTimerHandler)

	r.POST("/timer/reset", s.resetTimerHandler)

	return r
}

func (s *Server) indexHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Try to get user from session
	gothUser, err := s.auth.GetUserFromSession(c.Request)
	if err != nil || gothUser == nil {
		// No user logged in, show login page
		component := templates.LoginPage()
		if err = component.Render(ctx, c.Writer); err != nil {
			log.Printf("Error rendering login page: %v", err)
			c.String(http.StatusInternalServerError, "Error rendering page")
		}
		return
	}

	// User is logged in, get from database
	user, err := s.db.GetUserByID(ctx, gothUser.UserID)
	if err != nil {
		log.Printf("Error getting user from database: %v", err)
		c.String(http.StatusInternalServerError, "Error getting user")
		return
	}

	if user == nil {
		// User not in database, show login page
		component := templates.LoginPage()
		if err = component.Render(ctx, c.Writer); err != nil {
			log.Printf("Error rendering login page: %v", err)
			c.String(http.StatusInternalServerError, "Error rendering page")
		}
		return
	}

	userTagStats, err := s.db.FindAllUserTagStats(ctx, gothUser.UserID)
	if err != nil {
		log.Printf("Error getting user tag stats: %v", err)
	}
	var tags []string
	for _, tagStats := range userTagStats {
		tags = append(tags, tagStats.Tag)
	}

	component := templates.IndexPage(gothUser, nil, tags)
	if err = component.Render(ctx, c.Writer); err != nil {
		log.Printf("Error rendering index page: %v", err)
		c.String(http.StatusInternalServerError, "Error rendering page")
	}
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
