package routes

import "github.com/gin-gonic/gin"

func RegisterPublicAnnouncementRoutes(api *gin.RouterGroup, deps Dependencies) {
	announcements := api.Group("/announcements")
	{
		announcements.GET("/public", deps.Announcements.ListPublicAnnouncements)
		announcements.GET("/me", deps.Announcements.ListMyAnnouncements)
		announcements.GET("/active", deps.Announcements.ListActivePublishedAnnouncements)
		announcements.GET("/user", deps.Announcements.ListAnnouncementsForUser)
		announcements.GET("/role/:role_name", deps.Announcements.ListAnnouncementsForRole)
		announcements.GET("/client/:client_id", deps.Announcements.ListAnnouncementsForClient)
	}
}
