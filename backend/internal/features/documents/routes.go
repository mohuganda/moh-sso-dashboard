package documents

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	documents := protected.Group("/documents")
	{
		documents.GET("", middleware.RequirePermission(authz.PermissionDocumentsRead), handler.ListDocuments)
		documents.POST("", middleware.RequirePermission(authz.PermissionDocumentsWrite), handler.CreateDocument)
		documents.GET("/:id", middleware.RequirePermission(authz.PermissionDocumentsRead), handler.GetDocument)
		documents.PUT("/:id", middleware.RequirePermission(authz.PermissionDocumentsWrite), handler.EditDocument)
		documents.DELETE("/:id", middleware.RequirePermission(authz.PermissionDocumentsWrite), handler.DeleteDocument)
		documents.GET("/files/:id/view", middleware.RequirePermission(authz.PermissionDocumentsRead), handler.ViewDocument)
		documents.GET("/files/:id/download", middleware.RequirePermission(authz.PermissionDocumentsRead), handler.DownloadDocument)
		documents.GET("/:id/processes", middleware.RequirePermission(authz.PermissionDocumentsRead), handler.ListDocumentProcesses)
		documents.POST("/:id/reprocess", middleware.RequirePermission(authz.PermissionDocumentsProcess), handler.ReprocessDocument)
	}
}
