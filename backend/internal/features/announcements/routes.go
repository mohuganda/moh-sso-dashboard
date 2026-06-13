package announcements

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func RegisterPublicRoutes(api *gin.RouterGroup, handler *Handler) {
	announcements := api.Group("/announcements")
	{
		announcements.GET("/public", handler.ListPublicAnnouncements)
		announcements.GET("/active", handler.ListActivePublishedAnnouncements)
	}
}

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	announcements := protected.Group("/announcements")
	announcements.Use(middleware.RequirePermission(authz.PermissionAnnouncementsRead))
	{
		announcements.GET("/me", handler.ListMyAnnouncements)
		announcements.GET("/user", handler.ListAnnouncementsForUser)
		announcements.GET("/role/:role_name", handler.ListAnnouncementsForRole)
		announcements.GET("/client/:client_id", handler.ListAnnouncementsForClient)
	}
}

func RegisterAdminRoutes(admin *gin.RouterGroup, handler *Handler) {
	announcements := admin.Group("/announcements")
	{
		announcements.GET("", middleware.RequirePermission(authz.PermissionAnnouncementsRead), handler.ListAnnouncementsAdmin)
		announcements.GET("/stats", middleware.RequirePermission(authz.PermissionAnnouncementsRead), handler.GetAnnouncementStats)
		announcements.GET("/:id", middleware.RequirePermission(authz.PermissionAnnouncementsRead), handler.GetAnnouncementByID)
		announcements.POST("", middleware.RequirePermission(authz.PermissionAnnouncementsWrite), handler.CreateAnnouncement)
		announcements.PUT("/:id", middleware.RequirePermission(authz.PermissionAnnouncementsWrite), handler.UpdateAnnouncement)
		announcements.DELETE("/:id", middleware.RequirePermission(authz.PermissionAnnouncementsWrite), handler.DeleteAnnouncement)
		announcements.POST("/:id/restore", middleware.RequirePermission(authz.PermissionAnnouncementsWrite), handler.RestoreAnnouncement)
		announcements.POST("/:id/publish", middleware.RequirePermission(authz.PermissionAnnouncementsPublish), handler.PublishAnnouncementNow)
		announcements.POST("/:id/draft", middleware.RequirePermission(authz.PermissionAnnouncementsPublish), handler.MoveAnnouncementToDraft)
		announcements.POST("/:id/schedule", middleware.RequirePermission(authz.PermissionAnnouncementsPublish), handler.ScheduleAnnouncement)
		announcements.POST("/:id/archive", middleware.RequirePermission(authz.PermissionAnnouncementsPublish), handler.ArchiveAnnouncement)
		announcements.PATCH("/:id/pin", middleware.RequirePermission(authz.PermissionAnnouncementsWrite), handler.SetAnnouncementPinned)
		announcements.PATCH("/:id/priority", middleware.RequirePermission(authz.PermissionAnnouncementsWrite), handler.SetAnnouncementPriority)
	}
}
