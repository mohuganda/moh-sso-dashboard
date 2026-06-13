package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	adminunitsfeature "github.com/moh-sso-dashboard/internal/features/admin_units"
	visualiserfeature "github.com/moh-sso-dashboard/internal/features/visualiser"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func registerVisualiserRoutes(protected *gin.RouterGroup, deps Dependencies) {
	visualiser := protected.Group("/visualizer")
	visualiser.Use(middleware.RequireAnyPermission(
		authz.PermissionReportBrowserRead,
		authz.PermissionDocumentsRead,
		authz.PermissionDataQualityRead,
		authz.PermissionSurveillanceRead,
	))
	adminunitsfeature.RegisterVisualiserRoutes(visualiser, deps.AdminUnits)
	visualiserfeature.RegisterRoutes(visualiser, deps.Visualiser)
}
