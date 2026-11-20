package reportscheduler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPModuleAccess(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, body string
		authenticated, system    bool
		permissions              []string
		status                   int
		want                     string
	}{
		{"anonymous", "GET", "", "", false, false, nil, 401, "UNAUTHORIZED"},
		{"missing system", "GET", "", "", true, false, []string{PermissionRead}, 403, "system access required"},
		{"missing permission", "GET", "", "", true, true, nil, 403, PermissionRead},
		{"module status preserves scope and request id", "GET", "", "", true, true, []string{PermissionRead}, 200, `"district":"Kampala"`},
		{"metrics requires manage too", "GET", "/metrics", "", true, true, []string{PermissionMetrics}, 403, PermissionManage},
		{"recipient options accepts update", "GET", "/recipients/options?kind=invalid", "", true, true, []string{PermissionUpdate}, 400, "INVALID_KIND"},
		{"preview requires recipients", "POST", "/recipients/preview", "{}", true, true, []string{PermissionCreate}, 400, "INVALID_BODY"},
		{"create requires schedule fields", "POST", "/schedules", "{}", true, true, []string{PermissionCreate}, 400, "INVALID_BODY"},
		{"raw generation is unavailable", "POST", "/reports/example/generate", "", true, true, []string{PermissionExecute}, 404, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			module, err := NewHTTPModule(NewService(nil, nil, nil, nil, nil, nil), HTTPOptions{
				Actor: func(*http.Request) (HTTPActor, bool) {
					return HTTPActor{
						UserID:        "verified-user",
						HealthContext: HealthContext{District: " Kampala "},
						HasSystem:     func(system string) bool { return tc.system && system == RequiredSystem },
						HasPermission: func(permission string) bool {
							for _, granted := range tc.permissions {
								if granted == permission {
									return true
								}
							}
							return false
						},
					}, tc.authenticated
				},
				RequestID:           func(*http.Request) string { return "test-request" },
				LimitPublicDownload: func(_ string, _ int, next http.Handler) http.Handler { return next },
			})
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/report-scheduler"+tc.path, strings.NewReader(tc.body))
			req.Header.Set("X-User-ID", "spoofed")
			module.ProtectedHandler().ServeHTTP(recorder, req)
			if recorder.Code != tc.status || !strings.Contains(recorder.Body.String(), tc.want) {
				t.Fatalf("got status %d and body %s", recorder.Code, recorder.Body)
			}
			if tc.status != 404 && !strings.Contains(recorder.Body.String(), `"requestId":"test-request"`) {
				t.Fatalf("request ID missing from %s", recorder.Body)
			}
		})
	}
}

func TestHTTPPublicDeliveryUsesLimiter(t *testing.T) {
	module, err := NewHTTPModule(NewService(nil, nil, nil, nil, nil, nil), HTTPOptions{
		Actor: func(*http.Request) (HTTPActor, bool) {
			t.Fatal("delivery links should not resolve a portal actor")
			return HTTPActor{}, false
		},
		LimitPublicDownload: func(policy string, limit int, _ http.Handler) http.Handler {
			if policy != "report-delivery-download" || limit != 60 {
				t.Fatalf("unexpected policy %q with limit %d", policy, limit)
			}
			return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
			})
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/report-scheduler/public/deliveries/example/download", nil)
	module.PublicHandler().ServeHTTP(recorder, req)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("got %d", recorder.Code)
	}
	recorder = httptest.NewRecorder()
	module.ProtectedHandler().ServeHTTP(recorder, req)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("protected surface returned %d", recorder.Code)
	}
}
