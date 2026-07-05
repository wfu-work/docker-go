package apis

import (
	"docker-go/domains"
	"docker-go/services"

	"github.com/gin-gonic/gin"
	commonDomains "github.com/wfu-work/nav-common-go-lib/domains"
	"github.com/wfu-work/nav-common-go-lib/response"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
)

type HostApi struct{}

// List 查询主机列表
// @Summary 查询主机列表
// @Description 分页查询远程 Docker 主机
// @Tags 主机模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键字"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /hosts/list [get]
func (HostApi) List(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.HostServiceApp.List(params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}

// Get 查询主机详情
// @Summary 查询主机详情
// @Description 根据 GUID 查询远程 Docker 主机详情
// @Tags 主机模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Success 200 {object} response.Response{data=domains.Host,msg=string}
// @Router /hosts/{guid} [get]
func (HostApi) Get(c *gin.Context) {
	host, err := services.HostServiceApp.Get(c.Param("guid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(host, c)
}

// Create 创建主机
// @Summary 创建主机
// @Description 创建远程 Docker 主机记录
// @Tags 主机模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body domains.Host true "主机信息"
// @Success 200 {object} response.Response{data=domains.Host,msg=string}
// @Router /hosts [post]
func (HostApi) Create(c *gin.Context) {
	var host domains.Host
	if err := c.ShouldBindJSON(&host); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	saved, err := services.HostServiceApp.Create(host)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(saved, c)
}

// Update 更新主机
// @Summary 更新主机
// @Description 根据 GUID 更新远程 Docker 主机
// @Tags 主机模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Param data body domains.Host true "主机信息"
// @Success 200 {object} response.Response{data=domains.Host,msg=string}
// @Router /hosts/{guid} [put]
func (HostApi) Update(c *gin.Context) {
	var host domains.Host
	if err := c.ShouldBindJSON(&host); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	saved, err := services.HostServiceApp.Update(c.Param("guid"), host)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(saved, c)
}

// Delete 删除主机
// @Summary 删除主机
// @Description 根据 GUID 删除远程 Docker 主机
// @Tags 主机模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /hosts/{guid} [delete]
func (HostApi) Delete(c *gin.Context) {
	if err := services.HostServiceApp.Delete(c.Param("guid")); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(true, c)
}

// GenerateAgentToken 生成 Agent Token
// @Summary 生成 Agent Token
// @Description 为指定主机生成或刷新 Agent 注册 Token
// @Tags 主机模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Success 200 {object} response.Response{data=services.AgentTokenResponse,msg=string}
// @Router /hosts/{guid}/agent-token [post]
func (HostApi) GenerateAgentToken(c *gin.Context) {
	token, err := services.HostServiceApp.GenerateAgentToken(c.Param("guid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(token, c)
}
