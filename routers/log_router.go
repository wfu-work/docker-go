package routers

import "github.com/gin-gonic/gin"

type LogRouter struct{}

func (r *LogRouter) InitLogRouter(privateGroup *gin.RouterGroup) {
	privateGroup.GET("hosts/:guid/containers/:containerId/logs", logApi.RecentContainerLogs)
	privateGroup.GET("ws/hosts/:guid/containers/:containerId/logs", logApi.StreamContainerLogs)
}
