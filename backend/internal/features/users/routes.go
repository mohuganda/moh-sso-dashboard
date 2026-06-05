package users

import "github.com/gin-gonic/gin"

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	users := protected.Group("/users")
	{
		users.GET("", handler.ListUsers)
		users.GET("/:id", handler.GetUser)
	}
}

func RegisterAdminRoutes(admin *gin.RouterGroup, handler *Handler) {
	users := admin.Group("/users")
	{
		users.GET("", handler.ListUsers)
		users.GET("/:id", handler.GetUser)
		users.POST("", handler.CreateUser)
		users.PUT("/:id", handler.UpdateUser)
		users.DELETE("/:id", handler.DeleteUser)
		users.PATCH("/:id/enabled", handler.SetUserEnabled)
		users.PATCH("/:id/toggle", handler.SetUserEnabled)
		users.POST("/:id/onboarding-email", handler.SendUserOnboardingEmail)
		users.POST("/:id/verification-email", handler.SendUserVerificationEmail)
		users.POST("/:id/password-reset", handler.SendUserPasswordResetEmail)
		users.POST("/:id/reset-password", handler.ResetUserPassword)
		users.GET("/:id/client-roles", handler.GetUserClientRoles)
		users.GET("/:id/clients/:clientID/roles", handler.GetUserClientRolesForClient)
		users.POST("/:id/clients/:clientID/roles", handler.AddUserClientRoles)
		users.DELETE("/:id/clients/:clientID/roles", handler.RemoveUserClientRoles)
		users.PUT("/:id/client-roles", handler.UpdateUserClientRoles)
	}
}
