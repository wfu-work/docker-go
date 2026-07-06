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
