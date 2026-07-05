package apis

import "github.com/gin-gonic/gin"

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	SettingApi
	HostApi
	AgentApi
	AgentUpgradeApi
	TaskApi
	DockerApi
	LogApi
	MetricApi
	DeployConfigApi
	RegistryCredentialApi
	DeployTemplateApi
	OperationApi
	SystemMonitorApi
}

func queryParams(c *gin.Context) map[string]string {
	params := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params
}
