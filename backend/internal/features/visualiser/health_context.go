package visualiser

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	healthcontext "github.com/moh-sso-dashboard/internal/features/health_context"
)

var errHealthContextAliasMissing = errors.New("health context has no DWH organisation-unit alias")

var visualiserOrgUnitAliasNamespaces = map[string]struct{}{
	"dwh-org-unit":            {},
	"surveillance-region":     {},
	"surveillance-district":   {},
	"surveillance-sub-county": {},
	"surveillance-facility":   {},
}

func (h *Handler) scopeDataValuesRequest(
	c *gin.Context,
	request DataValuesRequest,
) (DataValuesRequest, error) {
	rawID, exists := c.Get("health_context_id")
	if !exists {
		return request, nil
	}
	contextID, ok := rawID.(uuid.UUID)
	if !ok || contextID == uuid.Nil {
		return request, errHealthContextAliasMissing
	}

	node, err := h.healthContexts.GetNode(c.Request.Context(), contextID)
	if err != nil {
		return request, err
	}
	rawMode, _ := c.Get("health_context_scope_mode")
	mode, _ := rawMode.(healthcontext.ScopeMode)
	includeDescendants := mode == healthcontext.ScopeNodeAndDescendants

	// Uganda with descendant scope already represents the complete DWH domain.
	if node.Code == "UG" && includeDescendants {
		return request, nil
	}

	aliases, err := h.healthContexts.ListAliasesInScope(
		c.Request.Context(),
		contextID,
		includeDescendants,
	)
	if err != nil {
		return request, err
	}
	allowed := make(map[string]struct{})
	for _, alias := range aliases {
		if _, supported := visualiserOrgUnitAliasNamespaces[alias.Namespace]; supported {
			externalID := strings.TrimSpace(alias.ExternalID)
			if externalID != "" {
				allowed[externalID] = struct{}{}
			}
		}
	}
	if len(allowed) == 0 {
		return request, errHealthContextAliasMissing
	}

	if len(request.OU) == 0 {
		request.OU = make([]string, 0, len(allowed))
		for externalID := range allowed {
			request.OU = append(request.OU, externalID)
		}
		return request, nil
	}

	scoped := make([]string, 0, len(request.OU))
	for _, externalID := range request.OU {
		if _, ok := allowed[strings.TrimSpace(externalID)]; ok {
			scoped = append(scoped, externalID)
		}
	}
	if len(scoped) == 0 {
		return request, errHealthContextAliasMissing
	}
	request.OU = scoped
	return request, nil
}
