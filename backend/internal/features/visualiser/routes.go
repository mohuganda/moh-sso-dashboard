package visualiser

import "github.com/gin-gonic/gin"

func RegisterRoutes(visualiser *gin.RouterGroup, handler *Handler) {
	visualiser.GET("/datasets", handler.GetDatasets)
	visualiser.POST("/dataelements", handler.GetDataElements)
	visualiser.POST("/datavalues", handler.GetDataValues)
	visualiser.GET("/themes", handler.GetThemes)
	visualiser.POST("/dataelements/theme", handler.GetDataElementsByTheme)
	visualiser.GET("/hiv/summary", handler.GetHIVSummary)
	visualiser.GET("/hiv/tested", handler.GetHIVTested)
	visualiser.GET("/hiv/regimen", handler.GetHIVRegimen)
}
