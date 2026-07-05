package apis

import (
	"encoding/json"
	"strings"

	"docker-go/domains"
	"docker-go/services"

	"github.com/gin-gonic/gin"
	commonDomains "github.com/wfu-work/nav-common-go-lib/domains"
	"github.com/wfu-work/nav-common-go-lib/response"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
)

type DockerApi struct{}

type dockerContainerPayload struct {
	ContainerID   string `json:"containerId"`
	ApprovalGuid  string `json:"approvalGuid,omitempty"`
	Force         bool   `json:"force,omitempty"`
	RemoveVolumes bool   `json:"removeVolumes,omitempty"`
	Timeout       int    `json:"timeout,omitempty"`
	Operator      string `json:"operator,omitempty"`
}

type dockerImagePullRequest struct {
	Image                  string `json:"image"`
	Platform               string `json:"platform"`
	RegistryAuth           string `json:"registryAuth"`
	RegistryCredentialGuid string `json:"registryCredentialGuid"`
	ApprovalGuid           string `json:"approvalGuid"`
	Operator               string `json:"operator"`
}

// ListContainers 查询容器列表
// @Summary 查询容器列表
// @Description 查询指定主机已同步的 Docker 容器列表
// @Tags Docker资源模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键字"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /hosts/{guid}/containers [get]
func (DockerApi) ListContainers(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.DockerResourceServiceApp.ListContainers(dockerHostGuid(c), params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}

// SyncContainers 同步容器列表
// @Summary 同步容器列表
// @Description 创建任务让 Agent 同步指定主机 Docker 容器列表
// @Tags Docker资源模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Success 200 {object} response.Response{data=domains.Task,msg=string}
// @Router /hosts/{guid}/containers/sync [post]
func (DockerApi) SyncContainers(c *gin.Context) {
	task, ok := createDockerTask(c, dockerHostGuid(c), domains.TaskTypeDockerContainerList, map[string]any{}, 120)
	if !ok {
		return
	}
	response.Ok(task, c)
}

// StartContainer 启动容器
// @Summary 启动容器
// @Description 创建任务启动指定容器
// @Tags Docker资源模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Param containerId path string true "容器 ID"
// @Success 200 {object} response.Response{data=domains.Task,msg=string}
// @Router /hosts/{guid}/containers/{containerId}/start [post]
func (api DockerApi) StartContainer(c *gin.Context) {
	api.createContainerTask(c, domains.TaskTypeDockerContainerStart, dockerContainerPayload{
		ContainerID: c.Param("containerId"),
	})
}

// StopContainer 停止容器
// @Summary 停止容器
// @Description 创建任务停止指定容器
// @Tags Docker资源模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Param containerId path string true "容器 ID"
// @Param data body object false "停止参数"
// @Success 200 {object} response.Response{data=domains.Task,msg=string}
// @Router /hosts/{guid}/containers/{containerId}/stop [post]
func (api DockerApi) StopContainer(c *gin.Context) {
	var payload dockerContainerPayload
	_ = c.ShouldBindJSON(&payload)
	payload.ContainerID = c.Param("containerId")
	api.createContainerTask(c, domains.TaskTypeDockerContainerStop, payload)
}

// RestartContainer 重启容器
// @Summary 重启容器
// @Description 创建任务重启指定容器
// @Tags Docker资源模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Param containerId path string true "容器 ID"
// @Param data body object false "重启参数"
// @Success 200 {object} response.Response{data=domains.Task,msg=string}
// @Router /hosts/{guid}/containers/{containerId}/restart [post]
func (api DockerApi) RestartContainer(c *gin.Context) {
	var payload dockerContainerPayload
	_ = c.ShouldBindJSON(&payload)
	payload.ContainerID = c.Param("containerId")
	api.createContainerTask(c, domains.TaskTypeDockerContainerRestart, payload)
}

// RemoveContainer 删除容器
// @Summary 删除容器
// @Description 创建任务删除指定容器
// @Tags Docker资源模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Param containerId path string true "容器 ID"
// @Param data body object false "删除参数"
// @Success 200 {object} response.Response{data=domains.Task,msg=string}
// @Router /hosts/{guid}/containers/{containerId} [delete]
func (api DockerApi) RemoveContainer(c *gin.Context) {
	var payload dockerContainerPayload
	_ = c.ShouldBindJSON(&payload)
	payload.ContainerID = c.Param("containerId")
	api.createContainerTask(c, domains.TaskTypeDockerContainerRemove, payload)
}

// ListImages 查询镜像列表
// @Summary 查询镜像列表
// @Description 查询指定主机已同步的 Docker 镜像列表
// @Tags Docker资源模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param keyword query string false "搜索关键字"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /hosts/{guid}/images [get]
func (DockerApi) ListImages(c *gin.Context) {
	params := queryParams(c)
	items, total, err := services.DockerResourceServiceApp.ListImages(dockerHostGuid(c), params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(commonDomains.PageResult{Data: items, Total: total, Page: commonUtils.Str2Int(params["page"]), Size: commonUtils.Str2Int(params["size"])}, c)
}

// SyncImages 同步镜像列表
// @Summary 同步镜像列表
// @Description 创建任务让 Agent 同步指定主机 Docker 镜像列表
// @Tags Docker资源模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Success 200 {object} response.Response{data=domains.Task,msg=string}
// @Router /hosts/{guid}/images/sync [post]
func (DockerApi) SyncImages(c *gin.Context) {
	task, ok := createDockerTask(c, dockerHostGuid(c), domains.TaskTypeDockerImageList, map[string]any{}, 120)
	if !ok {
		return
	}
	response.Ok(task, c)
}

// PullImage 拉取镜像
// @Summary 拉取镜像
// @Description 创建任务在指定主机拉取 Docker 镜像
// @Tags Docker资源模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Param data body object true "镜像拉取参数"
// @Success 200 {object} response.Response{data=domains.Task,msg=string}
// @Router /hosts/{guid}/images/pull [post]
func (DockerApi) PullImage(c *gin.Context) {
	var req dockerImagePullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	req.Image = strings.TrimSpace(req.Image)
	if req.Image == "" {
		response.FailWithMessage("missing image", c)
		return
	}
	if strings.TrimSpace(req.RegistryAuth) == "" && strings.TrimSpace(req.RegistryCredentialGuid) != "" {
		registryAuth, err := services.RegistryCredentialServiceApp.BuildRegistryAuth(req.RegistryCredentialGuid)
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
		req.RegistryAuth = registryAuth
	}
	task, ok := createDockerTaskWithMeta(c, dockerHostGuid(c), domains.TaskTypeDockerImagePull, req, 1800, req.ApprovalGuid, "image", req.Image, req.Operator)
	if !ok {
		return
	}
	response.Ok(task, c)
}

func (DockerApi) createContainerTask(c *gin.Context, taskType string, payload dockerContainerPayload) {
	payload.ContainerID = strings.TrimSpace(payload.ContainerID)
	if payload.ContainerID == "" {
		response.FailWithMessage("missing container id", c)
		return
	}
	timeout := 120
	if taskType == domains.TaskTypeDockerContainerRemove {
		timeout = 300
	}
	task, ok := createDockerTaskWithMeta(c, dockerHostGuid(c), taskType, payload, timeout, payload.ApprovalGuid, "container", payload.ContainerID, payload.Operator)
	if !ok {
		return
	}
	response.Ok(task, c)
}

func createDockerTask(c *gin.Context, hostGuid string, taskType string, payload any, timeoutSeconds int) (*domains.Task, bool) {
	return createDockerTaskWithMeta(c, hostGuid, taskType, payload, timeoutSeconds, "", "", "", "")
}

func createDockerTaskWithMeta(c *gin.Context, hostGuid string, taskType string, payload any, timeoutSeconds int, approvalGuid string, resourceType string, resourceGuid string, operator string) (*domains.Task, bool) {
	raw, err := json.Marshal(payload)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return nil, false
	}
	return createAndDispatchTask(c, services.TaskCreateRequest{
		HostGuid:       hostGuid,
		ApprovalGuid:   approvalGuid,
		Type:           taskType,
		Payload:        raw,
		TimeoutSeconds: timeoutSeconds,
		ResourceType:   resourceType,
		ResourceGuid:   resourceGuid,
		Operator:       operator,
	})
}

func dockerHostGuid(c *gin.Context) string {
	if guid := c.Param("hostGuid"); guid != "" {
		return guid
	}
	return c.Param("guid")
}
