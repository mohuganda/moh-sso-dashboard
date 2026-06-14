package routes

import (
	"github.com/gin-gonic/gin"

	documenttemplatesfeature "github.com/moh-sso-dashboard/internal/features/document_templates"
	documentsfeature "github.com/moh-sso-dashboard/internal/features/documents"
)

func registerDocumentRoutes(protected *gin.RouterGroup, deps Dependencies) {
	documentsfeature.RegisterProtectedRoutes(protected, deps.Documents, deps.Limiter)
	documenttemplatesfeature.RegisterProtectedRoutes(protected, deps.DocumentTemplates)
}
