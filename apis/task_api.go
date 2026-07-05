package apis

import (
	"docker-go/services"

	"github.com/gin-gonic/gin"
	commonDomains "github.com/wfu-work/nav-common-go-lib/domains"
	"github.com/wfu-work/nav-common-go-lib/response"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
)

type TaskApi struct{}

// List 查询任务列表
// @Summary 查询任务列表
// @Description 分页查询远程操作任务
// @Tags 任务模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键字"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /tasks/list [get]
func (TaskApi) List(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.TaskServiceApp.List(params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}

// Get 查询任务详情
// @Summary 查询任务详情
// @Description 根据 GUID 查询远程操作任务
// @Tags 任务模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "任务 GUID"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /tasks/{guid} [get]
func (TaskApi) Get(c *gin.Context) {
	task, err := services.TaskServiceApp.Get(c.Param("guid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(task, c)
}

// Create 创建任务
// @Summary 创建任务
// @Description 创建远程操作任务并尝试下发到在线 Agent
// @Tags 任务模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body services.TaskCreateRequest true "任务信息"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /tasks [post]
func (TaskApi) Create(c *gin.Context) {
	var req services.TaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	task, ok := createAndDispatchTask(c, req)
	if !ok {
		return
	}
	response.Ok(task, c)
}

// Cancel 取消任务
// @Summary 取消任务
// @Description 取消尚未完成的远程操作任务
// @Tags 任务模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "任务 GUID"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /tasks/{guid}/cancel [post]
func (TaskApi) Cancel(c *gin.Context) {
	task, err := services.TaskServiceApp.Cancel(c.Param("guid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(task, c)
}

// AppendEvent 追加任务事件
// @Summary 追加任务事件
// @Description Agent 回传任务执行状态和结果
// @Tags 任务模块
// @Accept json
// @Produce json
// @Param guid path string true "任务 GUID"
// @Param data body services.TaskEventCreateRequest true "任务事件"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /tasks/{guid}/events [post]
func (TaskApi) AppendEvent(c *gin.Context) {
	var req services.TaskEventCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	event, err := services.TaskServiceApp.AppendEvent(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(event, c)
}

// ListEvents 查询任务事件
// @Summary 查询任务事件
// @Description 分页查询指定任务的事件记录
// @Tags 任务模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "任务 GUID"
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /tasks/{guid}/events [get]
func (TaskApi) ListEvents(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.TaskServiceApp.ListEvents(c.Param("guid"), params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}
