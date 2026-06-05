package document_templates

import "github.com/gin-gonic/gin"

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	templates := protected.Group("/document-templates")
	{
		templates.GET("", handler.ListTemplates)
		templates.POST("", handler.CreateTemplate)
		templates.POST("/structure", handler.CreateTemplateWithStructure)
		templates.GET("/code/:code/structure", handler.GetTemplateStructure)
		templates.GET("/:id", handler.GetTemplate)
		templates.PUT("/:id", handler.UpdateTemplate)
		templates.DELETE("/:id", handler.DeleteTemplate)
		templates.POST("/:id/publish", handler.PublishTemplate)
		templates.POST("/:id/archive", handler.ArchiveTemplate)
		templates.GET("/:id/structure", handler.GetTemplateStructure)
		templates.GET("/:id/sheets", handler.ListSheets)
		templates.GET("/:id/sheets/:sheetId/columns", handler.ListColumns)
	}
}
