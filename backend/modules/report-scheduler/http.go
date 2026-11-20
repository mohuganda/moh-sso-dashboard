package reportscheduler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

const (
	RequiredSystem         = "data-statistics"
	PermissionRead         = "report_scheduler:read"
	PermissionCreate       = "report_scheduler:create"
	PermissionUpdate       = "report_scheduler:update"
	PermissionDelete       = "report_scheduler:delete"
	PermissionExecute      = "report_scheduler:execute"
	PermissionHistory      = "report_scheduler:history"
	PermissionManage       = "report_scheduler:manage"
	PermissionMetrics      = "metrics:read"
	DeliveryDownloadPolicy = "report-delivery-download"
	DeliveryDownloadLimit  = 60
)

// HTTPActor represents a verified identity and its resolved access.
type HTTPActor struct {
	UserID        string
	HealthContext HealthContext
	HasSystem     func(string) bool
	HasPermission func(string) bool
}

type HTTPOptions struct {
	// Actor must obtain its identity from trusted authentication middleware.
	Actor     func(*http.Request) (HTTPActor, bool)
	RequestID func(*http.Request) string
	// LimitPublicDownload wraps the anonymous capability route in a per-IP limiter.
	LimitPublicDownload func(policy string, maxPerMinute int, next http.Handler) http.Handler
}

type HTTPModule struct {
	protected         *http.ServeMux
	public            *http.ServeMux
	protectedPatterns []string
	publicPatterns    []string
}

func NewHTTPModule(service *Service, options HTTPOptions) (*HTTPModule, error) {
	if service == nil || options.Actor == nil || options.LimitPublicDownload == nil {
		return nil, errors.New("scheduler HTTP requires a service, actor resolver, and public download limiter")
	}
	h := &httpHandler{service: service}
	m := &HTTPModule{protected: http.NewServeMux(), public: http.NewServeMux()}
	for _, route := range h.endpoints() {
		pattern := route.pattern()
		serve := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := &httpRequest{writer: w, request: r}
			if options.RequestID != nil {
				c.meta.RequestID = options.RequestID(r)
			}
			if !route.public {
				actor, ok := options.Actor(r)
				if !ok || strings.TrimSpace(actor.UserID) == "" {
					c.fail(http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
					return
				}
				if actor.HasSystem == nil || !actor.HasSystem(RequiredSystem) {
					c.fail(http.StatusForbidden, "FORBIDDEN", "system access required: "+RequiredSystem)
					return
				}
				if actor.HasPermission == nil {
					actor.HasPermission = func(string) bool { return false }
				}
				for _, anyOf := range route.required {
					allowed := false
					for _, permission := range anyOf {
						allowed = allowed || actor.HasPermission(permission)
					}
					if !allowed {
						message := "permission required: " + anyOf[0]
						if len(anyOf) > 1 {
							message = "one of these permissions is required: " + strings.Join(anyOf, ", ")
						}
						c.fail(http.StatusForbidden, "FORBIDDEN", message)
						return
					}
				}
				actor.HealthContext.District = strings.TrimSpace(actor.HealthContext.District)
				actor.HealthContext.Facility = strings.TrimSpace(actor.HealthContext.Facility)
				c.actor = actor
			}
			route.serve(c)
		})
		if route.public {
			m.publicPatterns = append(m.publicPatterns, pattern)
			m.public.Handle(pattern, options.LimitPublicDownload(DeliveryDownloadPolicy, DeliveryDownloadLimit, serve))
		} else {
			m.protectedPatterns = append(m.protectedPatterns, pattern)
			m.protected.Handle(pattern, serve)
		}
	}
	return m, nil
}

// Patterns lists the module's public and protected HTTP operations.
func (m *HTTPModule) Patterns() []string {
	return append(m.ProtectedPatterns(), m.publicPatterns...)
}

// ServeHTTP dispatches public capability routes to the rate-limited handler.
// Every other route is served by the protected handler.
func (m *HTTPModule) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, pattern := m.public.Handler(r); pattern != "" {
		m.public.ServeHTTP(w, r)
		return
	}
	m.protected.ServeHTTP(w, r)
}

func (m *HTTPModule) ProtectedHandler() http.Handler { return m.protected }
func (m *HTTPModule) PublicHandler() http.Handler    { return m.public }
func (m *HTTPModule) ProtectedPatterns() []string {
	return append([]string(nil), m.protectedPatterns...)
}
func (m *HTTPModule) PublicPatterns() []string { return append([]string(nil), m.publicPatterns...) }

type httpHandler struct{ service *Service }

type responseMeta struct {
	RequestID string `json:"requestId,omitempty"`
}

type httpRequest struct {
	writer  http.ResponseWriter
	request *http.Request
	actor   HTTPActor
	meta    responseMeta
}

func (c *httpRequest) decode(value any) error {
	return json.NewDecoder(c.request.Body).Decode(value)
}

func (c *httpRequest) ok(status int, data any) {
	c.write(status, struct {
		Success bool         `json:"success"`
		Data    any          `json:"data"`
		Meta    responseMeta `json:"meta"`
	}{true, data, c.meta})
}

func (c *httpRequest) fail(status int, code, message string) {
	c.write(status, struct {
		Success bool              `json:"success"`
		Error   map[string]string `json:"error"`
		Meta    responseMeta      `json:"meta"`
	}{false, map[string]string{"code": code, "message": message}, c.meta})
}

func (c *httpRequest) write(status int, body any) {
	c.writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.writer.WriteHeader(status)
	_ = json.NewEncoder(c.writer).Encode(body)
}
