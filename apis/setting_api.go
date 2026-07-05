package apis

import (
	"docker-go/domains"
	"docker-go/services"

	"github.com/gin-gonic/gin"
	commonDomains "github.com/wfu-work/nav-common-go-lib/domains"
	"github.com/wfu-work/nav-common-go-lib/response"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
)

type SettingApi struct{}

// List 查询设置列表
// @Summary 查询设置列表
// @Description 分页查询运行时设置
// @Tags 设置模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键字"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /settings/list [get]
func (SettingApi) List(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.SettingServiceApp.List(params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}

// Save 保存设置
// @Summary 保存设置
// @Description 创建或更新运行时设置
// @Tags 设置模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body domains.Setting true "设置信息"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /settings [post]
func (SettingApi) Save(c *gin.Context) {
	var setting domains.Setting
	if err := c.ShouldBindJSON(&setting); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := services.SettingServiceApp.Save(setting); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(true, c)
}

// Delete 删除设置
// @Summary 删除设置
// @Description 根据 GUID 删除运行时设置
// @Tags 设置模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "GUID"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /settings/{guid} [delete]
func (SettingApi) Delete(c *gin.Context) {
	if err := services.SettingServiceApp.Delete(c.Param("guid")); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(true, c)
}
