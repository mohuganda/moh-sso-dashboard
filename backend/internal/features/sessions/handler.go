package sessions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/http/response"
)

type Handler struct {
	sessionService Service
}

func NewHandler(
	sessionService Service,
) *Handler {
	return &Handler{
		sessionService: sessionService,
	}
}

func (h *Handler) GetUserSessions(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	userID := userIDValue.(string)

	sessions, err := h.sessionService.GetUserSessions(userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	response.OK(c, http.StatusOK, sessions)
}

func (h *Handler) LogoutSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "session id is required",
		})
		return
	}

	if err := h.sessionService.LogoutSession(sessionID); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}
