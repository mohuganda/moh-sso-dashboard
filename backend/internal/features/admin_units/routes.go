package admin_units

import "github.com/gin-gonic/gin"

func RegisterVisualiserRoutes(visualiser *gin.RouterGroup, handler *Handler) {
	visualiser.GET("/adminunits/orgunits", handler.GetOrgUnits)
	visualiser.GET("/adminunits/facilities", handler.GetFacilities)
	visualiser.GET("/adminunits/district", handler.GetDistricts)
	visualiser.POST("/adminunits/subcounties", handler.GetSubCounties)
	visualiser.POST("/adminunits/localgovt", handler.GetLocalGovt)
	visualiser.POST("/adminunits/districts", handler.GetDistrictsByRegion)
	visualiser.GET("/adminunits/region", handler.GetRegions)
	visualiser.GET("/adminunits/national", handler.GetNational)
	visualiser.GET("/adminunits/hierarchy", handler.GetHierarchy)
}
