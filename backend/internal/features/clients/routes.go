package clients

import "github.com/gin-gonic/gin"

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	clients := protected.Group("/clients")
	{
		clients.GET("", handler.ListClients)
		clients.GET("/:id", handler.GetClient)
		clients.POST("", handler.CreateClient)
		clients.PATCH("/:id/toggle", handler.ToggleClientEnabled)
		clients.DELETE("/:id", handler.DeleteClient)
		clients.GET("/:id/roles", handler.ListClientRoles)
	}
}

func RegisterAdminRoutes(admin *gin.RouterGroup, handler *Handler) {
	admin.POST("/clients/:id/roles", handler.CreateClientRole)
	admin.DELETE("/clients/:id/roles/:role", handler.DeleteClientRole)
}
