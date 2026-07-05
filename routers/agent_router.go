package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/wfu-work/nav-common-go-lib/middlewares"
)

type AgentRouter struct{}

func (r *AgentRouter) InitAgentRouter(privateGroup *gin.RouterGroup, publicGroup *gin.RouterGroup) {
	publicGroup.GET("agent/ws", agentApi.Connect)

	publicLogger := publicGroup.Group("agents").Use(middlewares.ApiLogger())
	{
		publicLogger.POST("register", agentApi.Register)
		publicLogger.POST("heartbeat", agentApi.Heartbeat)
	}

	group := privateGroup.Group("agents")
	{
		group.GET("list", agentApi.List)
		group.GET(":guid", agentApi.Get)
	}
}
