package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/wfu-work/nav-common-go-lib/middlewares"
)

type DockerRouter struct{}

func (r *DockerRouter) InitDockerRouter(privateGroup *gin.RouterGroup) {
	containers := privateGroup.Group("hosts/:guid/containers")
	containerLogger := privateGroup.Group("hosts/:guid/containers").Use(middlewares.ApiLogger())
	{
		containers.GET("", dockerApi.ListContainers)
		containerLogger.POST("sync", dockerApi.SyncContainers)
		containerLogger.POST(":containerId/start", dockerApi.StartContainer)
		containerLogger.POST(":containerId/stop", dockerApi.StopContainer)
		containerLogger.POST(":containerId/restart", dockerApi.RestartContainer)
		containerLogger.DELETE(":containerId", dockerApi.RemoveContainer)
	}

	images := privateGroup.Group("hosts/:guid/images")
	imageLogger := privateGroup.Group("hosts/:guid/images").Use(middlewares.ApiLogger())
	{
		images.GET("", dockerApi.ListImages)
		imageLogger.POST("sync", dockerApi.SyncImages)
		imageLogger.POST("pull", dockerApi.PullImage)
	}
}
