package documents

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
	"github.com/moh-sso-dashboard/internal/ratelimit"
)

func RegisterProtectedRoutes(
	protected *gin.RouterGroup,
	handler *Handler,
	limiter *ratelimit.Limiter,
) {
	documents := protected.Group("/documents")

	readPermission := middleware.RequirePermission(
		authz.PermissionDocumentsRead,
	)

	writePermission := middleware.RequirePermission(
		authz.PermissionDocumentsWrite,
	)

	processPermission := middleware.RequirePermission(
		authz.PermissionDocumentsProcess,
	)

	writeLimit := ratelimit.MiddlewareForPolicy(
		limiter,
		ratelimit.DocumentsWritePolicy(),
	)

	processLimit := ratelimit.MiddlewareForPolicy(
		limiter,
		ratelimit.DocumentsProcessPolicy(),
	)

	{
		documents.GET("", readPermission, handler.ListDocuments)
		documents.GET("/stats", readPermission, handler.GetDocumentStats)
		documents.POST("", writePermission, writeLimit, handler.CreateDocument)
		documents.POST("/scan-structure", writePermission, handler.ScanStructure)
		documents.GET("/:id", readPermission, handler.GetDocument)
		documents.PUT("/:id", writePermission, handler.EditDocument)
		documents.DELETE("/:id", writePermission, handler.DeleteDocument)
		documents.GET("/:id/processes", readPermission, handler.ListDocumentProcesses)
		documents.POST("/:id/reprocess", processPermission, processLimit, handler.ReprocessDocument)
		documents.GET("/:id/data-preview", readPermission, handler.DataPreview)
		documents.GET("/:id/parse-structure", readPermission, handler.ParseStructure)
	}

	files := documents.Group("/files")
	{
		files.GET("/:id/view", readPermission, handler.ViewDocument)
		files.GET("/:id/download", readPermission, handler.DownloadDocument)
	}
}
