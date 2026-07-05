package apis

import (
	agentHub "docker-go/agent"
	"docker-go/services"

	"github.com/gin-gonic/gin"
	commonDomains "github.com/wfu-work/nav-common-go-lib/domains"
	"github.com/wfu-work/nav-common-go-lib/response"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
)

type DeployConfigApi struct{}

// ListDeployConfigs 查询部署配置列表
// @Summary 查询部署配置列表
// @Description 分页查询 YAML/Compose 部署配置
// @Tags 部署配置模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键字"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /deploy-configs/list [get]
func (DeployConfigApi) ListDeployConfigs(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.DeployConfigServiceApp.List(params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}

// GetDeployConfig 查询部署配置详情
// @Summary 查询部署配置详情
// @Description 查询部署配置及当前版本内容
// @Tags 部署配置模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "部署配置 GUID"
// @Success 200 {object} response.Response{data=services.DeployConfigDetail,msg=string}
// @Router /deploy-configs/{guid} [get]
func (DeployConfigApi) GetDeployConfig(c *gin.Context) {
	item, err := services.DeployConfigServiceApp.Get(c.Param("guid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// CreateDeployConfig 创建部署配置
// @Summary 创建部署配置
// @Description 创建部署配置并生成初始版本
// @Tags 部署配置模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body services.DeployConfigSaveRequest true "部署配置信息"
// @Success 200 {object} response.Response{data=services.DeployConfigDetail,msg=string}
// @Router /deploy-configs [post]
func (DeployConfigApi) CreateDeployConfig(c *gin.Context) {
	var req services.DeployConfigSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := services.DeployConfigServiceApp.Create(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// UpdateDeployConfig 更新部署配置
// @Summary 更新部署配置
// @Description 更新部署配置，内容变化时自动生成新版本
// @Tags 部署配置模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "部署配置 GUID"
// @Param data body services.DeployConfigSaveRequest true "部署配置信息"
// @Success 200 {object} response.Response{data=services.DeployConfigDetail,msg=string}
// @Router /deploy-configs/{guid} [put]
func (DeployConfigApi) UpdateDeployConfig(c *gin.Context) {
	var req services.DeployConfigSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := services.DeployConfigServiceApp.Update(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// DeleteDeployConfig 删除部署配置
// @Summary 删除部署配置
// @Description 根据 GUID 删除部署配置
// @Tags 部署配置模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "部署配置 GUID"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /deploy-configs/{guid} [delete]
func (DeployConfigApi) DeleteDeployConfig(c *gin.Context) {
	if err := services.DeployConfigServiceApp.Delete(c.Param("guid")); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(true, c)
}

// ListDeployConfigVersions 查询部署配置版本
// @Summary 查询部署配置版本
// @Description 分页查询指定部署配置的版本记录
// @Tags 部署配置模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "部署配置 GUID"
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /deploy-configs/{guid}/versions [get]
func (DeployConfigApi) ListDeployConfigVersions(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.DeployConfigServiceApp.ListVersions(c.Param("guid"), params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}

// GetDeployConfigVersion 查询部署配置版本详情
// @Summary 查询部署配置版本详情
// @Description 查询指定部署配置的某个版本内容
// @Tags 部署配置模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "部署配置 GUID"
// @Param versionGuid path string true "版本 GUID"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /deploy-configs/{guid}/versions/{versionGuid} [get]
func (DeployConfigApi) GetDeployConfigVersion(c *gin.Context) {
	item, err := services.DeployConfigServiceApp.GetVersion(c.Param("guid"), c.Param("versionGuid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// CreateDeployConfigVersion 创建部署配置版本
// @Summary 创建部署配置版本
// @Description 为指定部署配置创建新版本并设置为当前版本
// @Tags 部署配置模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "部署配置 GUID"
// @Param data body services.DeployConfigVersionCreateRequest true "版本内容"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /deploy-configs/{guid}/versions [post]
func (DeployConfigApi) CreateDeployConfigVersion(c *gin.Context) {
	var req services.DeployConfigVersionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := services.DeployConfigServiceApp.CreateVersion(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// ValidateDeployConfig 校验部署配置
// @Summary 校验部署配置
// @Description 创建任务让 Agent 执行 docker compose config 校验
// @Tags 部署配置模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "部署配置 GUID"
// @Param data body services.DeployConfigTaskRequest true "校验参数"
// @Success 200 {object} response.Response{data=services.DeployConfigActionResponse,msg=string}
// @Router /deploy-configs/{guid}/validate [post]
func (DeployConfigApi) ValidateDeployConfig(c *gin.Context) {
	var req services.DeployConfigTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := services.DeployConfigServiceApp.Validate(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	dispatchDeployTask(result)
	response.Ok(result, c)
}

// DeployConfig 发布部署配置
// @Summary 发布部署配置
// @Description 创建 Compose 发布任务并记录发布历史
// @Tags 部署配置模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "部署配置 GUID"
// @Param data body services.DeployConfigTaskRequest true "发布参数"
// @Success 200 {object} response.Response{data=services.DeployConfigActionResponse,msg=string}
// @Router /deploy-configs/{guid}/deploy [post]
func (DeployConfigApi) DeployConfig(c *gin.Context) {
	var req services.DeployConfigTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := services.DeployConfigServiceApp.Deploy(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	dispatchDeployTask(result)
	response.Ok(result, c)
}

// RollbackDeployConfig 回滚部署配置
// @Summary 回滚部署配置
// @Description 使用历史版本创建发布任务并在成功后切换当前版本
// @Tags 部署配置模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "部署配置 GUID"
// @Param data body services.DeployConfigTaskRequest true "回滚参数"
// @Success 200 {object} response.Response{data=services.DeployConfigActionResponse,msg=string}
// @Router /deploy-configs/{guid}/rollback [post]
func (DeployConfigApi) RollbackDeployConfig(c *gin.Context) {
	var req services.DeployConfigTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := services.DeployConfigServiceApp.Rollback(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	dispatchDeployTask(result)
	response.Ok(result, c)
}

// ListDeployReleases 查询发布记录
// @Summary 查询发布记录
// @Description 分页查询 Compose 发布记录
// @Tags 部署发布模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键字"
// @Param hostGuid query string false "主机 GUID"
// @Param configGuid query string false "部署配置 GUID"
// @Param status query string false "发布状态"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /deploy-releases/list [get]
func (DeployConfigApi) ListDeployReleases(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.DeployConfigServiceApp.ListReleases(params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}

// GetDeployRelease 查询发布记录详情
// @Summary 查询发布记录详情
// @Description 根据 GUID 查询 Compose 发布记录
// @Tags 部署发布模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "发布记录 GUID"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /deploy-releases/{guid} [get]
func (DeployConfigApi) GetDeployRelease(c *gin.Context) {
	item, err := services.DeployConfigServiceApp.GetRelease(c.Param("guid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

func dispatchDeployTask(result *services.DeployConfigActionResponse) {
	if result == nil || result.Task == nil {
		return
	}
	_ = agentHub.DefaultHub.DispatchTask(result.Task)
	current, err := services.TaskServiceApp.Get(result.Task.Guid)
	if err == nil {
		result.Task = current
	}
	if result.Release == nil {
		return
	}
	_ = services.DeployConfigServiceApp.SyncReleaseFromTask(*result.Task)
	release, err := services.DeployConfigServiceApp.GetRelease(result.Release.Guid)
	if err == nil {
		result.Release = release
	}
}
