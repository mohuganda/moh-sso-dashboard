package apidocgen

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	scheduler "github.com/moh-sso-dashboard/internal/adapters/modules/report_scheduler"
)

func TestSchedulerContractMatchesMountedSurfaces(t *testing.T) {
	operations := map[string]object{}
	if err := loadSchedulerContract(backendRoot(t), object{}, operations); err != nil {
		t.Fatal(err)
	}
	engine := gin.New()
	handler := scheduler.NewHandler(scheduler.NewService(nil, nil, nil, nil, nil, nil))
	scheduler.RegisterPublicRoutes(engine.Group("/public"), handler, nil)
	scheduler.RegisterProtectedRoutes(engine.Group("/protected"), handler)
	for _, route := range engine.Routes() {
		surface, path, _ := strings.Cut(strings.TrimPrefix(route.Path, "/"), "/")
		public := surface == "public"
		key := route.Method + " /api/v1/" + OpenAPIPath(path)
		op, found := operations[key]
		if !found {
			t.Fatalf("mounted scheduler operation missing from OpenAPI: %s", key)
		}
		security, anonymous := op["security"].([]any)
		if public != (anonymous && len(security) == 0) {
			t.Errorf("scheduler operation %s security differs from its mounted surface", key)
		}
		delete(operations, key)
	}
	if len(operations) != 0 {
		t.Errorf("OpenAPI describes %d unmounted scheduler operations", len(operations))
	}
}
