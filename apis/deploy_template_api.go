package apis

import (
	"docker-go/services"

	"github.com/gin-gonic/gin"
	commonDomains "github.com/wfu-work/nav-common-go-lib/domains"
	"github.com/wfu-work/nav-common-go-lib/response"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
)

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
