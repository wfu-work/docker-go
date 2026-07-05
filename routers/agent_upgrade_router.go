package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/wfu-work/nav-common-go-lib/middlewares"
)

type AgentUpgradeRouter struct{}

func (r *AgentUpgradeRouter) InitAgentUpgradeRouter(privateGroup *gin.RouterGroup) {
	packages := privateGroup.Group("agent-upgrade-packages")
	packageLogger := privateGroup.Group("agent-upgrade-packages").Use(middlewares.ApiLogger())
	{
		packages.GET("list", agentUpgradeApi.ListPackages)
		packages.GET(":guid", agentUpgradeApi.GetPackage)

		packageLogger.POST("", agentUpgradeApi.CreatePackage)
		packageLogger.PUT(":guid", agentUpgradeApi.UpdatePackage)
		packageLogger.DELETE(":guid", agentUpgradeApi.DeletePackage)
	}

	agentLogger := privateGroup.Group("agents").Use(middlewares.ApiLogger())
	{
		agentLogger.POST(":guid/upgrade", agentUpgradeApi.UpgradeAgent)
	}
}
