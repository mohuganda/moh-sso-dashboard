package routes

import (
	"github.com/gin-gonic/gin"

	announcementfeature "github.com/moh-sso-dashboard/internal/features/announcements"
)

func RegisterPublicAnnouncementRoutes(api *gin.RouterGroup, deps Dependencies) {
	announcementfeature.RegisterPublicRoutes(api, deps.Announcements)
}
