package routes

import "github.com/gin-gonic/gin"

func registerDocumentRoutes(protected *gin.RouterGroup, deps Dependencies) {
	documents := protected.Group("/documents")
	{
		documents.GET("", deps.Documents.ListDocuments)
		documents.POST("", deps.Documents.CreateDocument)
		documents.GET("/:id", deps.Documents.GetDocument)
		documents.PUT("/:id", deps.Documents.EditDocument)
		documents.DELETE("/:id", deps.Documents.DeleteDocument)
		documents.GET("/files/:id/view", deps.Documents.ViewDocument)
		documents.GET("/files/:id/download", deps.Documents.DownloadDocument)
		documents.GET("/:id/processes", deps.Documents.ListDocumentProcesses)
		documents.POST("/:id/reprocess", deps.Documents.ReprocessDocument)
	}

	templates := protected.Group("/document-templates")
	{
		templates.GET("", deps.DocumentTemplates.ListTemplates)
		templates.POST("", deps.DocumentTemplates.CreateTemplate)
		templates.POST("/structure", deps.DocumentTemplates.CreateTemplateWithStructure)
		templates.GET("/code/:code/structure", deps.DocumentTemplates.GetTemplateStructure)
		templates.GET("/:id", deps.DocumentTemplates.GetTemplate)
		templates.PUT("/:id", deps.DocumentTemplates.UpdateTemplate)
		templates.DELETE("/:id", deps.DocumentTemplates.DeleteTemplate)
		templates.POST("/:id/publish", deps.DocumentTemplates.PublishTemplate)
		templates.POST("/:id/archive", deps.DocumentTemplates.ArchiveTemplate)
		templates.GET("/:id/structure", deps.DocumentTemplates.GetTemplateStructure)
		templates.GET("/:id/sheets", deps.DocumentTemplates.ListSheets)
		templates.GET("/:id/sheets/:sheetId/columns", deps.DocumentTemplates.ListColumns)
	}
}
