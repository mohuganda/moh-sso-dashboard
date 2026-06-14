package users

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
	"github.com/moh-sso-dashboard/internal/ratelimit"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	users := protected.Group("/users")
	{
		users.GET("", middleware.RequirePermission(authz.PermissionUsersRead), handler.ListUsers)
		users.GET("/:id", middleware.RequirePermission(authz.PermissionUsersRead), handler.GetUser)
	}
}

func RegisterAdminRoutes(admin *gin.RouterGroup, handler *Handler, limiter *ratelimit.Limiter) {
	users := admin.Group("/users")
	emailActionLimit := ratelimit.MiddlewareForPolicy(limiter, ratelimit.UsersEmailActionPolicy())
	passwordActionLimit := ratelimit.MiddlewareForPolicy(limiter, ratelimit.UsersPasswordActionPolicy())
	{
		users.GET("", middleware.RequirePermission(authz.PermissionUsersRead), handler.ListUsers)
		users.GET("/:id", middleware.RequirePermission(authz.PermissionUsersRead), handler.GetUser)
		users.POST("", middleware.RequirePermission(authz.PermissionUsersWrite), handler.CreateUser)
		users.PUT("/:id", middleware.RequirePermission(authz.PermissionUsersWrite), handler.UpdateUser)
		users.DELETE("/:id", middleware.RequirePermission(authz.PermissionUsersWrite), handler.DeleteUser)
		users.PATCH("/:id/enabled", middleware.RequirePermission(authz.PermissionUsersWrite), handler.SetUserEnabled)
		users.PATCH("/:id/toggle", middleware.RequirePermission(authz.PermissionUsersWrite), handler.SetUserEnabled)
		users.POST("/:id/onboarding-email", middleware.RequirePermission(authz.PermissionUsersWrite), emailActionLimit, handler.SendUserOnboardingEmail)
		users.POST("/:id/verification-email", middleware.RequirePermission(authz.PermissionUsersWrite), emailActionLimit, handler.SendUserVerificationEmail)
		users.POST("/:id/password-reset", middleware.RequirePermission(authz.PermissionUsersWrite), passwordActionLimit, handler.SendUserPasswordResetEmail)
		users.POST("/:id/reset-password", middleware.RequirePermission(authz.PermissionUsersWrite), passwordActionLimit, handler.ResetUserPassword)
		users.GET("/:id/client-roles", middleware.RequirePermission(authz.PermissionUsersRead), handler.GetUserClientRoles)
		users.GET("/:id/clients/:clientID/roles", middleware.RequirePermission(authz.PermissionUsersRead), handler.GetUserClientRolesForClient)
		users.POST("/:id/clients/:clientID/roles", middleware.RequirePermission(authz.PermissionUsersRolesWrite), handler.AddUserClientRoles)
		users.DELETE("/:id/clients/:clientID/roles", middleware.RequirePermission(authz.PermissionUsersRolesWrite), handler.RemoveUserClientRoles)
		users.PUT("/:id/client-roles", middleware.RequirePermission(authz.PermissionUsersRolesWrite), handler.UpdateUserClientRoles)
	}
}
