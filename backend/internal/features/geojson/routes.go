package geojson

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	geojson := protected.Group("/geojson")
	{
		geojson.GET("/:name", middleware.RequireAnyPermission(
			authz.PermissionSurveillanceRead,
			authz.PermissionDataQualityRead,
			authz.PermissionReportBrowserRead,
		), handler.GetGeoJSON)
	}
}
