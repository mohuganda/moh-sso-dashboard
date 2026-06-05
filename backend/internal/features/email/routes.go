package email

import "github.com/gin-gonic/gin"

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	email := protected.Group("/emails")
	{
		email.POST("/send", handler.Send)
		email.POST("/queue", handler.Queue)
		email.GET("", handler.List)
		email.GET("/status/:status", handler.ListByStatus)
		email.GET("/:id", handler.GetByID)
		email.POST("/:id/retry", handler.Retry)
		email.DELETE("/:id", handler.Delete)
	}
}
