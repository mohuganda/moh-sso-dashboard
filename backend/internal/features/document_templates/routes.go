package document_templates

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	templates := protected.Group("/document-templates")
	{
		templates.GET("", middleware.RequirePermission(authz.PermissionDocumentTemplatesRead), handler.ListTemplates)
		templates.POST("", middleware.RequirePermission(authz.PermissionDocumentTemplatesWrite), handler.CreateTemplate)
		templates.POST("/structure", middleware.RequirePermission(authz.PermissionDocumentTemplatesWrite), handler.CreateTemplateWithStructure)
		templates.GET("/code/:code/structure", middleware.RequirePermission(authz.PermissionDocumentTemplatesRead), handler.GetTemplateStructure)
		templates.GET("/code/:code/has-data", middleware.RequirePermission(authz.PermissionDocumentTemplatesRead), handler.HasData)
		templates.GET("/:id", middleware.RequirePermission(authz.PermissionDocumentTemplatesRead), handler.GetTemplate)
		templates.PUT("/:id", middleware.RequirePermission(authz.PermissionDocumentTemplatesWrite), handler.UpdateTemplate)
		templates.DELETE("/:id", middleware.RequirePermission(authz.PermissionDocumentTemplatesWrite), handler.DeleteTemplate)
		templates.POST("/:id/publish", middleware.RequirePermission(authz.PermissionDocumentTemplatesPublish), handler.PublishTemplate)
		templates.POST("/:id/archive", middleware.RequirePermission(authz.PermissionDocumentTemplatesPublish), handler.ArchiveTemplate)
		templates.GET("/:id/structure", middleware.RequirePermission(authz.PermissionDocumentTemplatesRead), handler.GetTemplateStructure)
		templates.PUT("/:id/structure", middleware.RequirePermission(authz.PermissionDocumentTemplatesWrite), handler.ReplaceStructure)
		templates.GET("/:id/sheets", middleware.RequirePermission(authz.PermissionDocumentTemplatesRead), handler.ListSheets)
		templates.GET("/:id/sheets/:sheetId/columns", middleware.RequirePermission(authz.PermissionDocumentTemplatesRead), handler.ListColumns)
	}
}
