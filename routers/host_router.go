package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/wfu-work/nav-common-go-lib/middlewares"
)

type HostRouter struct{}

func (r *HostRouter) InitHostRouter(privateGroup *gin.RouterGroup) {
	groupLogger := privateGroup.Group("hosts").Use(middlewares.ApiLogger())
	group := privateGroup.Group("hosts")
	{
		group.GET("list", hostApi.List)
		group.GET(":guid", hostApi.Get)
	}
	{
		groupLogger.POST("", hostApi.Create)
		groupLogger.PUT(":guid", hostApi.Update)
		groupLogger.DELETE(":guid", hostApi.Delete)
		groupLogger.POST(":guid/agent-token", hostApi.GenerateAgentToken)
	}
}
