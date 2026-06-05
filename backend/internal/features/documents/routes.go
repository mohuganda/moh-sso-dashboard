package documents

import "github.com/gin-gonic/gin"

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	documents := protected.Group("/documents")
	{
		documents.GET("", handler.ListDocuments)
		documents.POST("", handler.CreateDocument)
		documents.GET("/:id", handler.GetDocument)
		documents.PUT("/:id", handler.EditDocument)
		documents.DELETE("/:id", handler.DeleteDocument)
		documents.GET("/files/:id/view", handler.ViewDocument)
		documents.GET("/files/:id/download", handler.DownloadDocument)
		documents.GET("/:id/processes", handler.ListDocumentProcesses)
		documents.POST("/:id/reprocess", handler.ReprocessDocument)
	}
}
