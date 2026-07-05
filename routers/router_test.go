package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInitRouters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	publicGroup := engine.Group("/api")
	privateGroup := engine.Group("/api")

	RouterGroupApp.InitRouters(publicGroup, privateGroup)
}

func TestSystemMonitorRuntimeRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	publicGroup := engine.Group("/api")
	privateGroup := engine.Group("/api")

	RouterGroupApp.InitRouters(publicGroup, privateGroup)

	req := httptest.NewRequest(http.MethodGet, "/api/system/monitor/runtime", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}
