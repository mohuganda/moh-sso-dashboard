package announcements

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
	"github.com/moh-sso-dashboard/internal/ratelimit"
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
		announcements.GET("/:id/attachments/:attachmentId/download", handler.DownloadAnnouncementAttachment)
		announcements.GET("/me", handler.ListMyAnnouncements)
		announcements.GET("/user", handler.ListAnnouncementsForUser)
		announcements.GET("/role/:role_name", handler.ListAnnouncementsForRole)
		announcements.GET("/client/:client_id", handler.ListAnnouncementsForClient)
	}
}

func RegisterAdminRoutes(admin *gin.RouterGroup, handler *Handler, limiter *ratelimit.Limiter) {
	announcements := admin.Group("/announcements")
	writeLimit := ratelimit.MiddlewareForPolicy(limiter, ratelimit.AnnouncementsWritePolicy())
	publishLimit := ratelimit.MiddlewareForPolicy(limiter, ratelimit.AnnouncementsPublishPolicy())
	{
		announcements.GET("", middleware.RequirePermission(authz.PermissionAnnouncementsRead), handler.ListAnnouncementsAdmin)
		announcements.GET("/stats", middleware.RequirePermission(authz.PermissionAnnouncementsRead), handler.GetAnnouncementStats)
		announcements.GET("/:id/attachments", middleware.RequirePermission(authz.PermissionAnnouncementsRead), handler.ListAnnouncementAttachments)
		announcements.GET("/:id/attachments/:attachmentId/download", middleware.RequirePermission(authz.PermissionAnnouncementsRead), handler.DownloadAnnouncementAttachment)
		announcements.POST("/:id/attachments", middleware.RequirePermission(authz.PermissionAnnouncementsWrite), writeLimit, handler.UploadAnnouncementAttachment)
		announcements.PATCH("/:id/attachments/:attachmentId", middleware.RequirePermission(authz.PermissionAnnouncementsWrite), writeLimit, handler.UpdateAnnouncementAttachment)
		announcements.DELETE("/:id/attachments/:attachmentId", middleware.RequirePermission(authz.PermissionAnnouncementsWrite), writeLimit, handler.DeleteAnnouncementAttachment)
		announcements.GET("/:id", middleware.RequirePermission(authz.PermissionAnnouncementsRead), handler.GetAnnouncementByID)
		announcements.POST("", middleware.RequirePermission(authz.PermissionAnnouncementsWrite), writeLimit, handler.CreateAnnouncement)
		announcements.PUT("/:id", middleware.RequirePermission(authz.PermissionAnnouncementsWrite), writeLimit, handler.UpdateAnnouncement)
		announcements.DELETE("/:id", middleware.RequirePermission(authz.PermissionAnnouncementsWrite), writeLimit, handler.DeleteAnnouncement)
		announcements.POST("/:id/restore", middleware.RequirePermission(authz.PermissionAnnouncementsWrite), writeLimit, handler.RestoreAnnouncement)
		announcements.POST("/:id/publish", middleware.RequirePermission(authz.PermissionAnnouncementsPublish), publishLimit, handler.PublishAnnouncementNow)
		announcements.POST("/:id/draft", middleware.RequirePermission(authz.PermissionAnnouncementsPublish), publishLimit, handler.MoveAnnouncementToDraft)
		announcements.POST("/:id/schedule", middleware.RequirePermission(authz.PermissionAnnouncementsPublish), publishLimit, handler.ScheduleAnnouncement)
		announcements.POST("/:id/archive", middleware.RequirePermission(authz.PermissionAnnouncementsPublish), publishLimit, handler.ArchiveAnnouncement)
		announcements.PATCH("/:id/pin", middleware.RequirePermission(authz.PermissionAnnouncementsWrite), writeLimit, handler.SetAnnouncementPinned)
		announcements.PATCH("/:id/priority", middleware.RequirePermission(authz.PermissionAnnouncementsWrite), writeLimit, handler.SetAnnouncementPriority)
	}
}
