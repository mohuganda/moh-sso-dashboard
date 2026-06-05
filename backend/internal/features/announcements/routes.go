package announcements

import "github.com/gin-gonic/gin"

func RegisterPublicRoutes(api *gin.RouterGroup, handler *Handler) {
	announcements := api.Group("/announcements")
	{
		announcements.GET("/public", handler.ListPublicAnnouncements)
		announcements.GET("/me", handler.ListMyAnnouncements)
		announcements.GET("/active", handler.ListActivePublishedAnnouncements)
		announcements.GET("/user", handler.ListAnnouncementsForUser)
		announcements.GET("/role/:role_name", handler.ListAnnouncementsForRole)
		announcements.GET("/client/:client_id", handler.ListAnnouncementsForClient)
	}
}

func RegisterAdminRoutes(admin *gin.RouterGroup, handler *Handler) {
	announcements := admin.Group("/announcements")
	{
		announcements.GET("", handler.ListAnnouncementsAdmin)
		announcements.GET("/stats", handler.GetAnnouncementStats)
		announcements.GET("/:id", handler.GetAnnouncementByID)
		announcements.POST("", handler.CreateAnnouncement)
		announcements.PUT("/:id", handler.UpdateAnnouncement)
		announcements.DELETE("/:id", handler.DeleteAnnouncement)
		announcements.POST("/:id/restore", handler.RestoreAnnouncement)
		announcements.POST("/:id/publish", handler.PublishAnnouncementNow)
		announcements.POST("/:id/draft", handler.MoveAnnouncementToDraft)
		announcements.POST("/:id/schedule", handler.ScheduleAnnouncement)
		announcements.POST("/:id/archive", handler.ArchiveAnnouncement)
		announcements.PATCH("/:id/pin", handler.SetAnnouncementPinned)
		announcements.PATCH("/:id/priority", handler.SetAnnouncementPriority)
	}
}
