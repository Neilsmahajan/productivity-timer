package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neilsmahajan/productivity-timer/web/templates"
)

func (s *Server) statsPageHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Try to get user from session
	gothUser, err := s.auth.GetUserFromSession(c.Request)
	if err != nil || gothUser == nil {
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	allUserTagStats, err := s.db.FindAllUserTagStats(ctx, gothUser.UserID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	component := templates.StatsPage(allUserTagStats)
	if err = component.Render(ctx, c.Writer); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}
