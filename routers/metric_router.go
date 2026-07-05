package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/wfu-work/nav-common-go-lib/middlewares"
)

type MetricRouter struct{}

func (r *MetricRouter) InitMetricRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("hosts/:guid")
	logger := privateGroup.Group("hosts/:guid").Use(middlewares.ApiLogger())
	{
		group.GET("metrics", metricApi.LatestHostMetric)
		group.GET("metrics/containers", metricApi.LatestContainerMetrics)
		group.GET("overview", metricApi.Overview)
		logger.POST("metrics/sync", metricApi.SyncMetrics)
	}
}
