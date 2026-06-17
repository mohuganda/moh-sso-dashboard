package storage_locations

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	storageLocations := protected.Group("/storage-locations")
	{
		storageLocations.POST("", middleware.RequirePermission(authz.PermissionStorageLocationsWrite), handler.Create)
		storageLocations.GET("", middleware.RequirePermission(authz.PermissionStorageLocationsRead), handler.ListActive)
		storageLocations.GET("/:id", middleware.RequirePermission(authz.PermissionStorageLocationsRead), handler.GetByID)
		storageLocations.PUT("/:id", middleware.RequirePermission(authz.PermissionStorageLocationsWrite), handler.Update)
		storageLocations.DELETE("/:id", middleware.RequirePermission(authz.PermissionStorageLocationsWrite), handler.Delete)
	}
}
