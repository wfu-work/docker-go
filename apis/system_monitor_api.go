package apis

import (
	"docker-go/services"

	"github.com/gin-gonic/gin"
	"github.com/wfu-work/nav-common-go-lib/response"
)

type SystemMonitorApi struct{}

// Runtime 查询服务运行态
// @Summary 查询服务运行态
// @Description 查询 Docker Gateway 服务进程、Go Runtime、CPU、内存和磁盘采样信息
// @Tags 系统监控模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=services.SystemMonitorInfo,msg=string}
// @Router /system/monitor/runtime [get]
func (SystemMonitorApi) Runtime(c *gin.Context) {
	response.Ok(services.SystemMonitorServiceApp.Runtime(), c)
}
