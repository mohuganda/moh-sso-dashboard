package routes

import (
	"github.com/gin-gonic/gin"

	adminunitsfeature "github.com/moh-sso-dashboard/internal/features/admin_units"
	visualiserfeature "github.com/moh-sso-dashboard/internal/features/visualiser"
)

func registerVisualiserRoutes(protected *gin.RouterGroup, deps Dependencies) {
	visualiser := protected.Group("/visualizer")
	adminunitsfeature.RegisterVisualiserRoutes(visualiser, deps.AdminUnits)
	visualiserfeature.RegisterRoutes(visualiser, deps.Visualiser)
}
