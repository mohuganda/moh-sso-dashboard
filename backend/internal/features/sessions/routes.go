package sessions

import "github.com/gin-gonic/gin"

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	sessions := protected.Group("/sessions")
	{
		sessions.GET("", handler.GetUserSessions)
		sessions.DELETE("/:id", handler.LogoutSession)
	}
}
