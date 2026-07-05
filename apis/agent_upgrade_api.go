package apis

import (
	agentHub "docker-go/agent"
	"docker-go/services"

	"github.com/gin-gonic/gin"
	commonDomains "github.com/wfu-work/nav-common-go-lib/domains"
	"github.com/wfu-work/nav-common-go-lib/response"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
)

type AgentUpgradeApi struct{}

// ListPackages 查询 Agent 升级包列表
// @Summary 查询 Agent 升级包列表
// @Description 分页查询 Agent 自动升级包元数据
// @Tags Agent升级模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键字"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /agent-upgrade-packages/list [get]
func (AgentUpgradeApi) ListPackages(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.AgentUpgradePackageServiceApp.List(params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}

// GetPackage 查询 Agent 升级包详情
// @Summary 查询 Agent 升级包详情
// @Description 根据 GUID 查询 Agent 自动升级包
// @Tags Agent升级模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "升级包 GUID"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /agent-upgrade-packages/{guid} [get]
func (AgentUpgradeApi) GetPackage(c *gin.Context) {
	item, err := services.AgentUpgradePackageServiceApp.Get(c.Param("guid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// CreatePackage 创建 Agent 升级包
// @Summary 创建 Agent 升级包
// @Description 创建 Agent 自动升级包元数据
// @Tags Agent升级模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body services.AgentUpgradePackageSaveRequest true "升级包信息"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /agent-upgrade-packages [post]
func (AgentUpgradeApi) CreatePackage(c *gin.Context) {
	var req services.AgentUpgradePackageSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := services.AgentUpgradePackageServiceApp.Create(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// UpdatePackage 更新 Agent 升级包
// @Summary 更新 Agent 升级包
// @Description 根据 GUID 更新 Agent 自动升级包元数据
// @Tags Agent升级模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "升级包 GUID"
// @Param data body services.AgentUpgradePackageSaveRequest true "升级包信息"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /agent-upgrade-packages/{guid} [put]
func (AgentUpgradeApi) UpdatePackage(c *gin.Context) {
	var req services.AgentUpgradePackageSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := services.AgentUpgradePackageServiceApp.Update(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// DeletePackage 删除 Agent 升级包
// @Summary 删除 Agent 升级包
// @Description 根据 GUID 删除 Agent 自动升级包元数据
// @Tags Agent升级模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "升级包 GUID"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /agent-upgrade-packages/{guid} [delete]
func (AgentUpgradeApi) DeletePackage(c *gin.Context) {
	if err := services.AgentUpgradePackageServiceApp.Delete(c.Param("guid")); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(true, c)
}

// UpgradeAgent 下发 Agent 升级任务
// @Summary 下发 Agent 升级任务
// @Description 给指定 Agent 创建并下发自动升级任务
// @Tags Agent升级模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "Agent GUID"
// @Param data body services.AgentUpgradeRequest true "升级任务参数"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /agents/{guid}/upgrade [post]
func (AgentUpgradeApi) UpgradeAgent(c *gin.Context) {
	var req services.AgentUpgradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	task, err := services.AgentUpgradePackageServiceApp.CreateUpgradeTask(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	_ = agentHub.DefaultHub.DispatchTask(task)
	current, err := services.TaskServiceApp.Get(task.Guid)
	if err == nil {
		task = current
	}
	response.Ok(task, c)
}
