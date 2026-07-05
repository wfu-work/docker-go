package routers

import "github.com/gin-gonic/gin"

type SystemMonitorRouter struct{}

func (r *SystemMonitorRouter) InitSystemMonitorRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("system/monitor")
	{
		group.GET("runtime", systemMonitorApi.Runtime)
	}
}
