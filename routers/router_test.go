package routers

import (
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
