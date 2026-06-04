package routes

import "github.com/gin-gonic/gin"

func registerVisualiserRoutes(protected *gin.RouterGroup, deps Dependencies) {
	visualiser := protected.Group("/visualizer")
	{
		visualiser.GET("/adminunits/orgunits", deps.AdminUnits.GetOrgUnits)
		visualiser.GET("/adminunits/facilities", deps.AdminUnits.GetFacilities)
		visualiser.GET("/adminunits/district", deps.AdminUnits.GetDistricts)
		visualiser.POST("/adminunits/subcounties", deps.AdminUnits.GetSubCounties)
		visualiser.POST("/adminunits/localgovt", deps.AdminUnits.GetLocalGovt)
		visualiser.POST("/adminunits/districts", deps.AdminUnits.GetDistrictsByRegion)
		visualiser.GET("/adminunits/region", deps.AdminUnits.GetRegions)
		visualiser.GET("/adminunits/national", deps.AdminUnits.GetNational)
		visualiser.GET("/adminunits/hierarchy", deps.AdminUnits.GetHierarchy)

		visualiser.GET("/datasets", deps.Visualiser.GetDatasets)
		visualiser.POST("/dataelements", deps.Visualiser.GetDataElements)
		visualiser.POST("/datavalues", deps.Visualiser.GetDataValues)
		visualiser.GET("/themes", deps.Visualiser.GetThemes)
		visualiser.POST("/dataelements/theme", deps.Visualiser.GetDataElementsByTheme)
		visualiser.GET("/hiv/summary", deps.Visualiser.GetHIVSummary)
		visualiser.GET("/hiv/tested", deps.Visualiser.GetHIVTested)
		visualiser.GET("/hiv/regimen", deps.Visualiser.GetHIVRegimen)
	}
}
