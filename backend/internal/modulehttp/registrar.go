// Package modulehttp integrates standard-library modules with the trusted host.
// It contains no application implementation or dependency on a published module.
package modulehttp

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/http/apierror"
	"github.com/moh-sso-dashboard/internal/http/response"
)

type accessKey struct{}
type access struct {
	actor         authz.Context
	valid         bool
	healthContext map[string]string
}

// AuthorizeAny reads only the verified context supplied by the registrar after
// the platform authentication middleware. Request headers never create claims.
func AuthorizeAny(r *http.Request, permissions ...string) error {
	current, ok := r.Context().Value(accessKey{}).(access)
	if !ok || !current.valid || current.actor.UserID == "" {
		return apierror.New(http.StatusUnauthorized, "UNAUTHORIZED", http.StatusText(http.StatusUnauthorized))
	}
	for _, permission := range permissions {
		if current.actor.HasPermission(authz.Permission(permission)) {
			return nil
		}
	}
	return apierror.New(http.StatusForbidden, "FORBIDDEN", http.StatusText(http.StatusForbidden))
}

// Actor returns only the identity verified by the platform middleware.
func Actor(r *http.Request) (authz.Context, bool) {
	current, ok := r.Context().Value(accessKey{}).(access)
	return current.actor, ok && current.valid && current.actor.UserID != ""
}

// RequireAdmin grants access only from verified platform identity.
func RequireAdmin(r *http.Request) error {
	actor, valid := Actor(r)
	if !valid {
		return apierror.New(http.StatusUnauthorized, "UNAUTHORIZED", http.StatusText(http.StatusUnauthorized))
	}
	if !actor.IsAdmin {
		return apierror.New(http.StatusForbidden, "FORBIDDEN", http.StatusText(http.StatusForbidden))
	}
	return nil
}
func RequireSystem(r *http.Request, system string) error {
	actor, valid := Actor(r)
	if !valid {
		return apierror.New(http.StatusUnauthorized, "UNAUTHORIZED", http.StatusText(http.StatusUnauthorized))
	}
	if !actor.HasSystem(system) {
		return apierror.New(http.StatusForbidden, "FORBIDDEN", http.StatusText(http.StatusForbidden))
	}
	return nil
}

// HealthContext returns the district and facility populated by host authentication.
func HealthContext(r *http.Request) map[string]string {
	current, _ := r.Context().Value(accessKey{}).(access)
	return current.healthContext
}

func RequestID(r *http.Request) string { return response.MetaFromRequest(r).RequestID }

type HTTPModule interface{ http.Handler }
type Registrar struct{ group *gin.RouterGroup }

func NewRegistrar(group *gin.RouterGroup) *Registrar { return &Registrar{group: group} }
func Register(group *gin.RouterGroup, module HTTPModule) {
	if group == nil {
		return
	}
	registrar := NewRegistrar(group)
	switch described := module.(type) {
	case interface{ Patterns() []string }:
		for _, pattern := range described.Patterns() {
			registrar.Handle(pattern, module)
		}
	case interface{ Pattern() string }:
		registrar.Handle(described.Pattern(), module)
	default:
		panic("module does not declare HTTP patterns")
	}
}

var parameter = regexp.MustCompile(`\{([A-Za-z_][A-Za-z0-9_]*)\}`)

func (r *Registrar) Handle(pattern string, handler http.Handler) {
	fields := strings.Fields(pattern)
	if len(fields) != 2 || !strings.HasPrefix(fields[1], "/") {
		panic(fmt.Sprintf("invalid module HTTP pattern %q", pattern))
	}
	route := parameter.ReplaceAllString(fields[1], ":$1")
	if strings.ContainsAny(route, "{}") {
		panic("unsupported module wildcard pattern")
	}
	base := strings.TrimSuffix(r.group.BasePath(), "/")
	mounted := handler
	if base != "" {
		mounted = http.StripPrefix(base, handler)
	}
	r.group.Handle(fields[0], route, func(c *gin.Context) {
		actor, valid := authz.FromGin(c)
		rawHealth, _ := c.Get("health_context")
		health, _ := rawHealth.(map[string]string)
		request := c.Request.WithContext(context.WithValue(c.Request.Context(), accessKey{}, access{actor: actor, valid: valid, healthContext: health}))
		mounted.ServeHTTP(c.Writer, request)
	})
}
