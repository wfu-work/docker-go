package apis

import (
	agentHub "docker-go/agent"
	"docker-go/domains"
	"docker-go/services"

	"github.com/gin-gonic/gin"
	"github.com/wfu-work/nav-common-go-lib/response"
)

func createAndDispatchTask(c *gin.Context, req services.TaskCreateRequest) (*domains.Task, bool) {
	if req.ClientIP == "" {
		req.ClientIP = c.ClientIP()
	}
	task, err := services.TaskServiceApp.Create(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return nil, false
	}
	_ = agentHub.DefaultHub.DispatchTask(task)
	current, err := services.TaskServiceApp.Get(task.Guid)
	if err == nil {
		task = current
	}
	return task, true
}
