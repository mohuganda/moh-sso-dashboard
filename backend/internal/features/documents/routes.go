package documents

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
	"github.com/moh-sso-dashboard/internal/ratelimit"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler, limiter *ratelimit.Limiter) {
	documents := protected.Group("/documents")
	writeLimit := ratelimit.MiddlewareForPolicy(limiter, ratelimit.DocumentsWritePolicy())
	processLimit := ratelimit.MiddlewareForPolicy(limiter, ratelimit.DocumentsProcessPolicy())
	{
		documents.GET("", middleware.RequirePermission(authz.PermissionDocumentsRead), handler.ListDocuments)
		documents.GET("/stats", middleware.RequirePermission(authz.PermissionDocumentsRead), handler.GetDocumentStats)
		documents.POST("", middleware.RequirePermission(authz.PermissionDocumentsWrite), writeLimit, handler.CreateDocument)
		documents.POST("/scan-structure", middleware.RequirePermission(authz.PermissionDocumentsWrite), handler.ScanStructure)
		documents.GET("/:id", middleware.RequirePermission(authz.PermissionDocumentsRead), handler.GetDocument)
		documents.PUT("/:id", middleware.RequirePermission(authz.PermissionDocumentsWrite), handler.EditDocument)
		documents.DELETE("/:id", middleware.RequirePermission(authz.PermissionDocumentsWrite), handler.DeleteDocument)
		documents.GET("/files/:id/view", middleware.RequirePermission(authz.PermissionDocumentsRead), handler.ViewDocument)
		documents.GET("/files/:id/download", middleware.RequirePermission(authz.PermissionDocumentsRead), handler.DownloadDocument)
		documents.GET("/:id/processes", middleware.RequirePermission(authz.PermissionDocumentsRead), handler.ListDocumentProcesses)
		documents.POST("/:id/reprocess", middleware.RequirePermission(authz.PermissionDocumentsProcess), processLimit, handler.ReprocessDocument)
		documents.GET("/:id/data-preview", middleware.RequirePermission(authz.PermissionDocumentsRead), handler.DataPreview)
		documents.GET("/:id/parse-structure", middleware.RequirePermission(authz.PermissionDocumentsRead), handler.ParseStructure)
	}
}
