package apis

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	agentHub "docker-go/agent"
	"docker-go/domains"
	"docker-go/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/wfu-work/nav-common-go-lib/response"
)

type LogApi struct{}

type dockerLogsPayload struct {
	ContainerID string `json:"containerId"`
	Tail        string `json:"tail"`
	Since       string `json:"since"`
	Until       string `json:"until"`
	Timestamps  bool   `json:"timestamps"`
	Follow      bool   `json:"follow"`
}

var frontendLogUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// RecentContainerLogs 查询容器最近日志
// @Summary 查询容器最近日志
// @Description 创建任务读取指定容器最近 N 行日志
// @Tags 日志模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Param containerId path string true "容器 ID"
// @Param tail query string false "日志行数"
// @Param since query string false "开始时间"
// @Param until query string false "结束时间"
// @Param timestamps query bool false "是否包含时间戳"
// @Success 200 {object} response.Response{data=domains.Task,msg=string}
// @Router /hosts/{guid}/containers/{containerId}/logs [get]
func (LogApi) RecentContainerLogs(c *gin.Context) {
	payload := logPayloadFromQuery(c, false)
	task, ok := createDockerTask(c, dockerHostGuid(c), domains.TaskTypeDockerContainerLogs, payload, 120)
	if !ok {
		return
	}
	response.Ok(task, c)
}

// StreamContainerLogs 实时容器日志
// @Summary 实时容器日志
// @Description 建立 WebSocket 并转发 Agent 回传的容器实时日志事件
// @Tags 日志模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "主机 GUID"
// @Param containerId path string true "容器 ID"
// @Param tail query string false "初始日志行数"
// @Param timeoutSeconds query int false "订阅超时时间"
// @Success 101 {string} string "Switching Protocols"
// @Router /ws/hosts/{guid}/containers/{containerId}/logs [get]
func (LogApi) StreamContainerLogs(c *gin.Context) {
	payload := logPayloadFromQuery(c, true)
	raw, err := json.Marshal(payload)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	task, err := services.TaskServiceApp.Create(services.TaskCreateRequest{
		HostGuid:       dockerHostGuid(c),
		Type:           domains.TaskTypeDockerContainerStream,
		Payload:        raw,
		TimeoutSeconds: logStreamTimeout(c),
	})
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	events, cancel := agentHub.DefaultHub.SubscribeTaskEvents(task.Guid)
	defer cancel()

	conn, err := frontendLogUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	_ = conn.WriteJSON(gin.H{"type": "task.created", "task": task})
	dispatchErr := agentHub.DefaultHub.DispatchTask(task)
	if dispatchErr != nil {
		_ = conn.WriteJSON(gin.H{"type": "error", "errorMessage": dispatchErr.Error(), "taskGuid": task.Guid})
		return
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	ping := time.NewTicker(30 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-done:
			return
		case <-ping.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case event := <-events:
			if err := conn.WriteJSON(event); err != nil {
				return
			}
			if isFinalLogStatus(event.Status) {
				return
			}
		}
	}
}

func logPayloadFromQuery(c *gin.Context, follow bool) dockerLogsPayload {
	tail := c.DefaultQuery("tail", "200")
	return dockerLogsPayload{
		ContainerID: c.Param("containerId"),
		Tail:        tail,
		Since:       c.Query("since"),
		Until:       c.Query("until"),
		Timestamps:  c.Query("timestamps") == "true",
		Follow:      follow,
	}
}

func logStreamTimeout(c *gin.Context) int {
	timeout, err := strconv.Atoi(c.DefaultQuery("timeoutSeconds", "3600"))
	if err != nil || timeout <= 0 {
		return 3600
	}
	return timeout
}

func isFinalLogStatus(status string) bool {
	switch status {
	case domains.TaskStatusSuccess, domains.TaskStatusFailed, domains.TaskStatusTimeout, domains.TaskStatusCancelled:
		return true
	default:
		return false
	}
}
