package server

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
)

func (s *Server) getGothUserAndTag(c *gin.Context) (*goth.User, string, error) {
	gothUser, err := s.auth.GetUserFromSession(c.Request)
	if err != nil {
		return nil, "", err
	}

	tag := c.Param("tag")
	if tag == "" {
		return nil, "", errors.New("tag required")
	}

	return gothUser, tag, nil
}
