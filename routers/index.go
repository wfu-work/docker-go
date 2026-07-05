package routers

import (
	"docker-go/apis"

	"github.com/gin-gonic/gin"
)

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	SettingRouter
	HostRouter
	AgentRouter
	AgentUpgradeRouter
	TaskRouter
	DockerRouter
	LogRouter
	MetricRouter
	DeployConfigRouter
	RegistryCredentialRouter
	DeployTemplateRouter
	OperationRouter
	SystemMonitorRouter
}

var (
	settingApi            = apis.ApiGroupApp.SettingApi
	hostApi               = apis.ApiGroupApp.HostApi
	agentApi              = apis.ApiGroupApp.AgentApi
	agentUpgradeApi       = apis.ApiGroupApp.AgentUpgradeApi
	taskApi               = apis.ApiGroupApp.TaskApi
	dockerApi             = apis.ApiGroupApp.DockerApi
	logApi                = apis.ApiGroupApp.LogApi
	metricApi             = apis.ApiGroupApp.MetricApi
	deployConfigApi       = apis.ApiGroupApp.DeployConfigApi
	registryCredentialApi = apis.ApiGroupApp.RegistryCredentialApi
	deployTemplateApi     = apis.ApiGroupApp.DeployTemplateApi
	operationApi          = apis.ApiGroupApp.OperationApi
	systemMonitorApi      = apis.ApiGroupApp.SystemMonitorApi
)

func (r *RouterGroup) InitRouters(publicGroup *gin.RouterGroup, privateGroup *gin.RouterGroup) {
	r.InitSettingRouter(privateGroup, publicGroup)
	r.InitHostRouter(privateGroup)
	r.InitAgentRouter(privateGroup, publicGroup)
	r.InitAgentUpgradeRouter(privateGroup)
	r.InitTaskRouter(privateGroup, publicGroup)
	r.InitDockerRouter(privateGroup)
	r.InitLogRouter(privateGroup)
	r.InitMetricRouter(privateGroup)
	r.InitDeployConfigRouter(privateGroup)
	r.InitRegistryCredentialRouter(privateGroup)
	r.InitDeployTemplateRouter(privateGroup)
	r.InitOperationRouter(privateGroup)
	r.InitSystemMonitorRouter(privateGroup)
}
