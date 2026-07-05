package apis

import (
	"docker-go/services"

	"github.com/gin-gonic/gin"
	commonDomains "github.com/wfu-work/nav-common-go-lib/domains"
	"github.com/wfu-work/nav-common-go-lib/response"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
)

type RegistryCredentialApi struct{}

// List 查询镜像仓库凭据列表
// @Summary 查询镜像仓库凭据列表
// @Description 分页查询镜像仓库凭据
// @Tags 镜像仓库凭据模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键字"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /registry-credentials/list [get]
func (RegistryCredentialApi) List(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.RegistryCredentialServiceApp.List(params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}

// Get 查询镜像仓库凭据详情
// @Summary 查询镜像仓库凭据详情
// @Description 根据 GUID 查询镜像仓库凭据
// @Tags 镜像仓库凭据模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "凭据 GUID"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /registry-credentials/{guid} [get]
func (RegistryCredentialApi) Get(c *gin.Context) {
	item, err := services.RegistryCredentialServiceApp.Get(c.Param("guid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// Create 创建镜像仓库凭据
// @Summary 创建镜像仓库凭据
// @Description 创建镜像仓库认证凭据
// @Tags 镜像仓库凭据模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body services.RegistryCredentialSaveRequest true "凭据信息"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /registry-credentials [post]
func (RegistryCredentialApi) Create(c *gin.Context) {
	var req services.RegistryCredentialSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := services.RegistryCredentialServiceApp.Create(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// Update 更新镜像仓库凭据
// @Summary 更新镜像仓库凭据
// @Description 根据 GUID 更新镜像仓库认证凭据
// @Tags 镜像仓库凭据模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "凭据 GUID"
// @Param data body services.RegistryCredentialSaveRequest true "凭据信息"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /registry-credentials/{guid} [put]
func (RegistryCredentialApi) Update(c *gin.Context) {
	var req services.RegistryCredentialSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := services.RegistryCredentialServiceApp.Update(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// Delete 删除镜像仓库凭据
// @Summary 删除镜像仓库凭据
// @Description 根据 GUID 删除镜像仓库凭据
// @Tags 镜像仓库凭据模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "凭据 GUID"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /registry-credentials/{guid} [delete]
func (RegistryCredentialApi) Delete(c *gin.Context) {
	if err := services.RegistryCredentialServiceApp.Delete(c.Param("guid")); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(true, c)
}

type DeployTemplateApi struct{}

// List 查询部署模板列表
// @Summary 查询部署模板列表
// @Description 分页查询 Compose 部署模板
// @Tags 部署模板模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键字"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /deploy-templates/list [get]
func (DeployTemplateApi) List(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.DeployTemplateServiceApp.List(params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}

// Get 查询部署模板详情
// @Summary 查询部署模板详情
// @Description 根据 GUID 查询部署模板
// @Tags 部署模板模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "模板 GUID"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /deploy-templates/{guid} [get]
func (DeployTemplateApi) Get(c *gin.Context) {
	item, err := services.DeployTemplateServiceApp.Get(c.Param("guid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// Create 创建部署模板
// @Summary 创建部署模板
// @Description 创建 Compose 部署模板
// @Tags 部署模板模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body services.DeployTemplateSaveRequest true "模板信息"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /deploy-templates [post]
func (DeployTemplateApi) Create(c *gin.Context) {
	var req services.DeployTemplateSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := services.DeployTemplateServiceApp.Create(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// Update 更新部署模板
// @Summary 更新部署模板
// @Description 根据 GUID 更新 Compose 部署模板
// @Tags 部署模板模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "模板 GUID"
// @Param data body services.DeployTemplateSaveRequest true "模板信息"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /deploy-templates/{guid} [put]
func (DeployTemplateApi) Update(c *gin.Context) {
	var req services.DeployTemplateSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := services.DeployTemplateServiceApp.Update(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// Delete 删除部署模板
// @Summary 删除部署模板
// @Description 根据 GUID 删除部署模板
// @Tags 部署模板模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "模板 GUID"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /deploy-templates/{guid} [delete]
func (DeployTemplateApi) Delete(c *gin.Context) {
	if err := services.DeployTemplateServiceApp.Delete(c.Param("guid")); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(true, c)
}

// Render 渲染部署模板
// @Summary 渲染部署模板
// @Description 使用变量渲染部署模板并做基础校验
// @Tags 部署模板模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "模板 GUID"
// @Param data body services.DeployTemplateRenderRequest true "渲染变量"
// @Success 200 {object} response.Response{data=services.DeployTemplateRenderResponse,msg=string}
// @Router /deploy-templates/{guid}/render [post]
func (DeployTemplateApi) Render(c *gin.Context) {
	var req services.DeployTemplateRenderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := services.DeployTemplateServiceApp.Render(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// CreateConfig 根据模板创建部署配置
// @Summary 根据模板创建部署配置
// @Description 渲染模板并创建部署配置和初始版本
// @Tags 部署模板模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "模板 GUID"
// @Param data body services.DeployTemplateCreateConfigRequest true "创建参数"
// @Success 200 {object} response.Response{data=services.DeployConfigDetail,msg=string}
// @Router /deploy-templates/{guid}/create-config [post]
func (DeployTemplateApi) CreateConfig(c *gin.Context) {
	var req services.DeployTemplateCreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := services.DeployTemplateServiceApp.CreateConfig(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

type OperationApi struct{}

// ListPolicies 查询操作策略列表
// @Summary 查询操作策略列表
// @Description 分页查询远程操作策略
// @Tags 操作策略模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键字"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /operation-policies/list [get]
func (OperationApi) ListPolicies(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.OperationPolicyServiceApp.List(params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}

// GetPolicy 查询操作策略详情
// @Summary 查询操作策略详情
// @Description 根据 GUID 查询远程操作策略
// @Tags 操作策略模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "策略 GUID"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /operation-policies/{guid} [get]
func (OperationApi) GetPolicy(c *gin.Context) {
	item, err := services.OperationPolicyServiceApp.Get(c.Param("guid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// CreatePolicy 创建操作策略
// @Summary 创建操作策略
// @Description 创建远程操作审批策略
// @Tags 操作策略模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body services.OperationPolicySaveRequest true "策略信息"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /operation-policies [post]
func (OperationApi) CreatePolicy(c *gin.Context) {
	var req services.OperationPolicySaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := services.OperationPolicyServiceApp.Create(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// UpdatePolicy 更新操作策略
// @Summary 更新操作策略
// @Description 根据 GUID 更新远程操作审批策略
// @Tags 操作策略模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "策略 GUID"
// @Param data body services.OperationPolicySaveRequest true "策略信息"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /operation-policies/{guid} [put]
func (OperationApi) UpdatePolicy(c *gin.Context) {
	var req services.OperationPolicySaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := services.OperationPolicyServiceApp.Update(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// DeletePolicy 删除操作策略
// @Summary 删除操作策略
// @Description 根据 GUID 删除远程操作审批策略
// @Tags 操作策略模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "策略 GUID"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /operation-policies/{guid} [delete]
func (OperationApi) DeletePolicy(c *gin.Context) {
	if err := services.OperationPolicyServiceApp.Delete(c.Param("guid")); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(true, c)
}

// ListApprovals 查询操作审批列表
// @Summary 查询操作审批列表
// @Description 分页查询远程操作审批单
// @Tags 操作审批模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键字"
// @Param status query string false "审批状态"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /operation-approvals/list [get]
func (OperationApi) ListApprovals(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.OperationApprovalServiceApp.List(params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}

// GetApproval 查询操作审批详情
// @Summary 查询操作审批详情
// @Description 根据 GUID 查询远程操作审批单
// @Tags 操作审批模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "审批单 GUID"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /operation-approvals/{guid} [get]
func (OperationApi) GetApproval(c *gin.Context) {
	item, err := services.OperationApprovalServiceApp.Get(c.Param("guid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// CreateApproval 创建操作审批
// @Summary 创建操作审批
// @Description 创建远程操作审批单
// @Tags 操作审批模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body services.OperationApprovalCreateRequest true "审批单信息"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /operation-approvals [post]
func (OperationApi) CreateApproval(c *gin.Context) {
	var req services.OperationApprovalCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := services.OperationApprovalServiceApp.Create(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// Approve 通过操作审批
// @Summary 通过操作审批
// @Description 审批通过远程操作审批单
// @Tags 操作审批模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "审批单 GUID"
// @Param data body services.OperationApprovalReviewRequest false "审批意见"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /operation-approvals/{guid}/approve [post]
func (OperationApi) Approve(c *gin.Context) {
	var req services.OperationApprovalReviewRequest
	_ = c.ShouldBindJSON(&req)
	item, err := services.OperationApprovalServiceApp.Approve(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// Reject 拒绝操作审批
// @Summary 拒绝操作审批
// @Description 拒绝远程操作审批单
// @Tags 操作审批模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "审批单 GUID"
// @Param data body services.OperationApprovalReviewRequest false "审批意见"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /operation-approvals/{guid}/reject [post]
func (OperationApi) Reject(c *gin.Context) {
	var req services.OperationApprovalReviewRequest
	_ = c.ShouldBindJSON(&req)
	item, err := services.OperationApprovalServiceApp.Reject(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// Cancel 取消操作审批
// @Summary 取消操作审批
// @Description 取消远程操作审批单
// @Tags 操作审批模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "审批单 GUID"
// @Param data body services.OperationApprovalReviewRequest false "取消原因"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /operation-approvals/{guid}/cancel [post]
func (OperationApi) Cancel(c *gin.Context) {
	var req services.OperationApprovalReviewRequest
	_ = c.ShouldBindJSON(&req)
	item, err := services.OperationApprovalServiceApp.Cancel(c.Param("guid"), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

// ListAudits 查询操作审计列表
// @Summary 查询操作审计列表
// @Description 分页查询远程操作审计记录
// @Tags 操作审计模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键字"
// @Param status query string false "任务状态"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /operation-audits/list [get]
func (OperationApi) ListAudits(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.OperationAuditServiceApp.List(params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}

// GetAudit 查询操作审计详情
// @Summary 查询操作审计详情
// @Description 根据 GUID 查询远程操作审计记录
// @Tags 操作审计模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "审计 GUID"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /operation-audits/{guid} [get]
func (OperationApi) GetAudit(c *gin.Context) {
	item, err := services.OperationAuditServiceApp.Get(c.Param("guid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}
