package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/wfu-work/nav-common-go-lib/middlewares"
)

type TaskRouter struct{}

func (r *TaskRouter) InitTaskRouter(privateGroup *gin.RouterGroup, publicGroup *gin.RouterGroup) {
	publicLogger := publicGroup.Group("tasks").Use(middlewares.ApiLogger())
	{
		publicLogger.POST(":guid/events", taskApi.AppendEvent)
	}

	groupLogger := privateGroup.Group("tasks").Use(middlewares.ApiLogger())
	group := privateGroup.Group("tasks")
	{
		group.GET("list", taskApi.List)
		group.GET(":guid", taskApi.Get)
		group.GET(":guid/events", taskApi.ListEvents)
	}
	{
		groupLogger.POST("", taskApi.Create)
		groupLogger.POST(":guid/cancel", taskApi.Cancel)
	}
}
