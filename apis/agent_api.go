package apis

import (
	agentHub "docker-go/agent"
	"docker-go/services"

	"github.com/gin-gonic/gin"
	commonDomains "github.com/wfu-work/nav-common-go-lib/domains"
	"github.com/wfu-work/nav-common-go-lib/response"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
)

type AgentApi struct{}

// List 查询 Agent 列表
// @Summary 查询 Agent 列表
// @Description 分页查询远程 Agent
// @Tags Agent模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键字"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /agents/list [get]
func (AgentApi) List(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.AgentServiceApp.List(params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}

// Get 查询 Agent 详情
// @Summary 查询 Agent 详情
// @Description 根据 GUID 查询 Agent 详情
// @Tags Agent模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "Agent GUID"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /agents/{guid} [get]
func (AgentApi) Get(c *gin.Context) {
	agent, err := services.AgentServiceApp.Get(c.Param("guid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(agent, c)
}

// Connect Agent WebSocket 连接
// @Summary Agent WebSocket 连接
// @Description Agent 使用 Token 建立 WebSocket 长连接
// @Tags Agent模块
// @Accept json
// @Produce json
// @Param agentGuid query string false "Agent GUID"
// @Param token query string false "Agent Token"
// @Success 101 {string} string "Switching Protocols"
// @Router /agent/ws [get]
func (AgentApi) Connect(c *gin.Context) {
	agentGuid := c.Query("agentGuid")
	if agentGuid == "" {
		agentGuid = c.GetHeader("X-Agent-Guid")
	}
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("X-Agent-Token")
	}
	agent, err := services.AgentServiceApp.Authenticate(agentGuid, token)
	if err != nil {
		response.NoAuth(err.Error(), c)
		return
	}
	if err = agentHub.DefaultHub.Handle(c, *agent, token); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
}

// Register Agent 注册
// @Summary Agent 注册
// @Description Agent 使用注册 Token 上报基础信息并绑定主机
// @Tags Agent模块
// @Accept json
// @Produce json
// @Param data body services.AgentRegisterRequest true "Agent 注册信息"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /agents/register [post]
func (AgentApi) Register(c *gin.Context) {
	var req services.AgentRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	agent, err := services.AgentServiceApp.Register(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(agent, c)
}

// Heartbeat Agent 心跳
// @Summary Agent 心跳
// @Description Agent 上报心跳、版本和 Docker 版本
// @Tags Agent模块
// @Accept json
// @Produce json
// @Param data body services.AgentHeartbeatRequest true "Agent 心跳信息"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /agents/heartbeat [post]
func (AgentApi) Heartbeat(c *gin.Context) {
	var req services.AgentHeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	agent, err := services.AgentServiceApp.Heartbeat(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(agent, c)
}
