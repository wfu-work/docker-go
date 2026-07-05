package apis

import (
	"docker-go/domains"
	"docker-go/services"

	"github.com/gin-gonic/gin"
	"github.com/wfu-work/nav-common-go-lib/response"
)

type MetricApi struct{}

// SyncMetrics 同步主机指标
// @Summary 同步主机指标
// @Description 创建任务让 Agent 采集主机和运行中容器指标
// @Tags 监控模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Success 200 {object} response.Response{data=domains.Task,msg=string}
// @Router /hosts/{guid}/metrics/sync [post]
func (MetricApi) SyncMetrics(c *gin.Context) {
	task, ok := createDockerTask(c, dockerHostGuid(c), domains.TaskTypeDockerMetricsSnapshot, map[string]any{}, 120)
	if !ok {
		return
	}
	response.Ok(task, c)
}

// LatestHostMetric 查询最新主机指标
// @Summary 查询最新主机指标
// @Description 查询指定主机最近一次采集的主机指标
// @Tags 监控模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Success 200 {object} response.Response{data=domains.HostMetric,msg=string}
// @Router /hosts/{guid}/metrics [get]
func (MetricApi) LatestHostMetric(c *gin.Context) {
	metric, err := services.MetricServiceApp.LatestHostMetric(dockerHostGuid(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(metric, c)
}

// LatestContainerMetrics 查询最新容器指标
// @Summary 查询最新容器指标
// @Description 查询指定主机每个容器最近一次采集的指标
// @Tags 监控模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Success 200 {object} response.Response{data=[]domains.ContainerMetric,msg=string}
// @Router /hosts/{guid}/metrics/containers [get]
func (MetricApi) LatestContainerMetrics(c *gin.Context) {
	metrics, err := services.MetricServiceApp.LatestContainerMetrics(dockerHostGuid(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(metrics, c)
}

// Overview 查询主机概览
// @Summary 查询主机概览
// @Description 查询主机信息、最新指标、容器和镜像统计
// @Tags 监控模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Success 200 {object} response.Response{data=services.HostOverview,msg=string}
// @Router /hosts/{guid}/overview [get]
func (MetricApi) Overview(c *gin.Context) {
	overview, err := services.MetricServiceApp.Overview(dockerHostGuid(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(overview, c)
}
