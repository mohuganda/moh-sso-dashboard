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
		response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	userID := userIDValue.(string)

	sessions, err := h.sessionService.GetUserSessions(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toSessionResponses(sessions))
}

func (h *Handler) LogoutSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "session id is required")
		return
	}

	if err := h.sessionService.LogoutSession(sessionID); err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "request failed")
		return
	}

	c.Status(http.StatusNoContent)
}
