package server

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/neilsmahajan/productivity-timer/web/templates"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/neilsmahajan/productivity-timer/docs"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*", "http://localhost:3000"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	// Static files
	r.StaticFile("/favicon.ico", "./favicon.ico")

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Page routes (HTML responses)
	r.GET("/", s.indexHandler)
	r.GET("/stats", s.statsPageHandler)

	// Health check
	r.GET("/health", s.healthHandler)

	// Auth routes (at root level for OAuth compatibility)
	// These are browser redirect flows, not REST API endpoints
	r.GET("/auth/:provider", s.authHandler)
	r.GET("/auth/:provider/callback", s.callbackHandler)
	r.GET("/logout/:provider", s.logoutHandler)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Timer routes
		timer := v1.Group("/timer")
		{
			timer.POST("/start", s.startTimerHandler)
			timer.POST("/stop", s.stopTimerHandler)
			timer.POST("/reset", s.resetTimerHandler)
		}

		// Stats routes
		stats := v1.Group("/stats")
		{
			stats.GET("/summary", s.statsSummaryHandler)
			stats.GET("/tag/:tag/sessions", s.tagSessionsHandler)
			stats.DELETE("/tag/:tag", s.deleteTagHandler)
		}
	}

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
	tags := make([]string, 0, len(userTagStats))
	for _, tagStats := range userTagStats {
		tags = append(tags, tagStats.Tag)
	}

	component := templates.IndexPage(gothUser, nil, tags)
	if err = component.Render(ctx, c.Writer); err != nil {
		log.Printf("Error rendering index page: %v", err)
		c.String(http.StatusInternalServerError, "Error rendering page")
	}
}

// healthHandler godoc
// @Summary Health check endpoint
// @Description Returns the health status of the API and database
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health [get]
func (s *Server) healthHandler(c *gin.Context) {
	health := s.db.Health()
	if health["status"] == "unhealthy" {
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}
	c.JSON(http.StatusOK, health)
}
