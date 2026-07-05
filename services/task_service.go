package services

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"docker-go/domains"

	commonServices "github.com/wfu-work/nav-common-go-lib/services"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
	"gorm.io/gorm"
)

type TaskService struct {
	commonServices.CrudService[domains.Task]
}

type TaskEventService struct {
	commonServices.CrudService[domains.TaskEvent]
}

type TaskCreateRequest struct {
	HostGuid       string          `json:"hostGuid"`
	AgentGuid      string          `json:"agentGuid"`
	ApprovalGuid   string          `json:"approvalGuid"`
	Type           string          `json:"type"`
	Payload        json.RawMessage `json:"payload"`
	TimeoutSeconds int             `json:"timeoutSeconds"`
	ResourceType   string          `json:"resourceType"`
	ResourceGuid   string          `json:"resourceGuid"`
	Operator       string          `json:"operator"`
	ClientIP       string          `json:"clientIp"`
}

type TaskEventCreateRequest struct {
	AgentGuid    string          `json:"agentGuid"`
	Token        string          `json:"token"`
	Status       string          `json:"status"`
	Message      string          `json:"message"`
	Data         json.RawMessage `json:"data"`
	Result       json.RawMessage `json:"result"`
	ErrorMessage string          `json:"errorMessage"`
}

var (
	TaskServiceApp      = new(TaskService)
	TaskEventServiceApp = new(TaskEventService)
)

func (s TaskService) List(params map[string]string) (interface{}, int64, error) {
	if params["content"] == "" {
		params["content"] = params["keyword"]
	}
	return s.CrudService.List(commonUtils.ToPageInfo(params), "host_guid,agent_guid,type,status,error_message")
}

func (s TaskService) Get(guid string) (*domains.Task, error) {
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return nil, errors.New("missing task guid")
	}
	task, err := s.CrudService.GetByGuid(guid)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	return task, nil
}

func (s TaskService) Create(req TaskCreateRequest) (*domains.Task, error) {
	req.HostGuid = strings.TrimSpace(req.HostGuid)
	req.AgentGuid = strings.TrimSpace(req.AgentGuid)
	req.ApprovalGuid = strings.TrimSpace(req.ApprovalGuid)
	req.Type = strings.TrimSpace(req.Type)
	if req.HostGuid == "" {
		return nil, errors.New("missing host guid")
	}
	if req.Type == "" {
		return nil, errors.New("missing task type")
	}
	if req.TimeoutSeconds <= 0 {
		req.TimeoutSeconds = 600
	}
	payload, err := normalizeRawJSON(req.Payload)
	if err != nil {
		return nil, err
	}
	req.Payload = json.RawMessage(payload)
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	var host domains.Host
	err = db.Where("guid = ?", req.HostGuid).First(&host).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("host not found")
	}
	if err != nil {
		return nil, err
	}
	if req.AgentGuid != "" {
		var agent domains.Agent
		err = db.Where("guid = ?", req.AgentGuid).First(&agent).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("agent not found")
		}
		if err != nil {
			return nil, err
		}
		if agent.HostGuid != req.HostGuid {
			return nil, errors.New("agent host does not match task host")
		}
	}
	resourceType, resourceGuid := inferTaskResource(req.Type, payload)
	if req.ResourceType != "" {
		resourceType = strings.TrimSpace(req.ResourceType)
	}
	if req.ResourceGuid != "" {
		resourceGuid = strings.TrimSpace(req.ResourceGuid)
	}
	if req.ApprovalGuid == "" {
		requiresApproval, err := OperationPolicyServiceApp.RequiresApproval(req.HostGuid, req.Type, resourceType)
		if err != nil {
			return nil, err
		}
		if requiresApproval {
			return nil, errors.New("operation requires approval")
		}
	}
	req.ResourceType = resourceType
	req.ResourceGuid = resourceGuid

	task := domains.Task{
		HostGuid:       req.HostGuid,
		AgentGuid:      req.AgentGuid,
		ApprovalGuid:   req.ApprovalGuid,
		Type:           req.Type,
		Status:         domains.TaskStatusPending,
		Payload:        payload,
		TimeoutSeconds: req.TimeoutSeconds,
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&task).Error; err != nil {
			return err
		}
		if err := OperationApprovalServiceApp.UseApproval(tx, req.ApprovalGuid, req, task); err != nil {
			return err
		}
		if err := OperationAuditServiceApp.CreateAuditInTx(tx, task, req); err != nil {
			return err
		}
		event := domains.TaskEvent{
			TaskGuid: task.Guid,
			Status:   domains.TaskStatusPending,
			Message:  "task created",
			Data:     "{}",
		}
		return tx.Create(&event).Error
	})
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s TaskService) Cancel(guid string) (*domains.Task, error) {
	task, err := s.Get(guid)
	if err != nil {
		return nil, err
	}
	if isFinalTaskStatus(task.Status) {
		return task, nil
	}
	now := time.Now().UnixMilli()
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domains.Task{}).Where("guid = ?", task.Guid).Updates(map[string]any{
			"status":      domains.TaskStatusCancelled,
			"finished_at": now,
			"update_time": now,
		}).Error; err != nil {
			return err
		}
		return tx.Create(&domains.TaskEvent{
			TaskGuid: task.Guid,
			Status:   domains.TaskStatusCancelled,
			Message:  "task cancelled",
			Data:     "{}",
		}).Error
	})
	if err != nil {
		return nil, err
	}
	updated, err := s.Get(task.Guid)
	if err != nil {
		return nil, err
	}
	if syncErr := DeployConfigServiceApp.SyncReleaseFromTask(*updated); syncErr != nil {
		return nil, syncErr
	}
	if syncErr := OperationAuditServiceApp.SyncFromTask(*updated); syncErr != nil {
		return nil, syncErr
	}
	return updated, nil
}

func (s TaskService) MarkDispatched(taskGuid string, agentGuid string) (*domains.Task, error) {
	task, err := s.Get(taskGuid)
	if err != nil {
		return nil, err
	}
	if task.Status != domains.TaskStatusPending {
		return task, nil
	}
	if strings.TrimSpace(agentGuid) == "" {
		return nil, errors.New("missing agent guid")
	}
	now := time.Now().UnixMilli()
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domains.Task{}).Where("guid = ?", task.Guid).Updates(map[string]any{
			"agent_guid":  agentGuid,
			"status":      domains.TaskStatusDispatched,
			"update_time": now,
		}).Error; err != nil {
			return err
		}
		return tx.Create(&domains.TaskEvent{
			TaskGuid: task.Guid,
			Status:   domains.TaskStatusDispatched,
			Message:  "task dispatched",
			Data:     "{}",
		}).Error
	})
	if err != nil {
		return nil, err
	}
	updated, err := s.Get(task.Guid)
	if err != nil {
		return nil, err
	}
	if syncErr := DeployConfigServiceApp.SyncReleaseFromTask(*updated); syncErr != nil {
		return nil, syncErr
	}
	if syncErr := OperationAuditServiceApp.SyncFromTask(*updated); syncErr != nil {
		return nil, syncErr
	}
	return updated, nil
}

func (s TaskService) MarkTimeout(taskGuid string) (*domains.Task, error) {
	task, err := s.Get(taskGuid)
	if err != nil {
		return nil, err
	}
	if isFinalTaskStatus(task.Status) || task.Status == domains.TaskStatusPending {
		return task, nil
	}
	now := time.Now().UnixMilli()
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domains.Task{}).Where("guid = ?", task.Guid).Updates(map[string]any{
			"status":        domains.TaskStatusTimeout,
			"error_message": "task timeout",
			"finished_at":   now,
			"update_time":   now,
		}).Error; err != nil {
			return err
		}
		return tx.Create(&domains.TaskEvent{
			TaskGuid: task.Guid,
			Status:   domains.TaskStatusTimeout,
			Message:  "task timeout",
			Data:     "{}",
		}).Error
	})
	if err != nil {
		return nil, err
	}
	updated, err := s.Get(task.Guid)
	if err != nil {
		return nil, err
	}
	if syncErr := DeployConfigServiceApp.SyncReleaseFromTask(*updated); syncErr != nil {
		return nil, syncErr
	}
	if syncErr := OperationAuditServiceApp.SyncFromTask(*updated); syncErr != nil {
		return nil, syncErr
	}
	return updated, nil
}

func (s TaskService) ListPendingForAgent(agent domains.Agent) ([]domains.Task, error) {
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	var tasks []domains.Task
	err := db.Where(
		"host_guid = ? AND status = ? AND (agent_guid = '' OR agent_guid = ?)",
		agent.HostGuid,
		domains.TaskStatusPending,
		agent.Guid,
	).Order("id asc").Limit(100).Find(&tasks).Error
	return tasks, err
}

func (s TaskService) AppendEvent(taskGuid string, req TaskEventCreateRequest) (*domains.TaskEvent, error) {
	task, err := s.Get(taskGuid)
	if err != nil {
		return nil, err
	}
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	agent, err := validateAgentCredential(db, req.AgentGuid, req.Token)
	if err != nil {
		return nil, err
	}
	if task.AgentGuid != "" && task.AgentGuid != agent.Guid {
		return nil, errors.New("agent does not match task")
	}
	if task.HostGuid != "" && task.HostGuid != agent.HostGuid {
		return nil, errors.New("agent host does not match task")
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = task.Status
	}
	if isFinalTaskStatus(task.Status) && task.Status != status {
		return nil, errors.New("task already finished")
	}
	if !isAllowedTaskStatus(status) {
		return nil, errors.New("unsupported task status")
	}
	data, err := normalizeRawJSON(req.Data)
	if err != nil {
		return nil, err
	}
	result, err := normalizeOptionalRawJSON(req.Result)
	if err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()
	event := domains.TaskEvent{
		TaskGuid: task.Guid,
		Status:   status,
		Message:  strings.TrimSpace(req.Message),
		Data:     data,
	}
	updates := map[string]any{
		"status":      status,
		"update_time": now,
	}
	if status == domains.TaskStatusRunning && task.StartedAt == 0 {
		updates["started_at"] = now
	}
	if result != "" {
		updates["result"] = result
	}
	if strings.TrimSpace(req.ErrorMessage) != "" {
		updates["error_message"] = strings.TrimSpace(req.ErrorMessage)
	}
	if isFinalTaskStatus(status) {
		updates["finished_at"] = now
	}
	if task.AgentGuid == "" {
		updates["agent_guid"] = agent.Guid
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		return tx.Model(&domains.Task{}).Where("guid = ?", task.Guid).Updates(updates).Error
	})
	if err != nil {
		return nil, err
	}
	updated, getErr := s.Get(task.Guid)
	if getErr != nil {
		return nil, getErr
	}
	if syncErr := DeployConfigServiceApp.SyncReleaseFromTask(*updated); syncErr != nil {
		return nil, syncErr
	}
	if syncErr := OperationAuditServiceApp.SyncFromTask(*updated); syncErr != nil {
		return nil, syncErr
	}
	if status == domains.TaskStatusSuccess && result != "" {
		if syncErr := DockerResourceServiceApp.SyncFromTaskResult(*updated, result); syncErr != nil {
			return nil, syncErr
		}
		if syncErr := MetricServiceApp.SyncFromTaskResult(*updated, result); syncErr != nil {
			return nil, syncErr
		}
	}
	return &event, nil
}

func (s TaskService) ListEvents(taskGuid string, params map[string]string) (interface{}, int64, error) {
	task, err := s.Get(taskGuid)
	if err != nil {
		return nil, 0, err
	}
	pageInfo := commonUtils.ToPageInfo(params)
	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.Size <= 0 {
		pageInfo.Size = 50
	}
	if pageInfo.Size > 200 {
		pageInfo.Size = 200
	}
	db := s.DB()
	if db == nil {
		return nil, 0, errors.New("database is not initialized")
	}
	var events []domains.TaskEvent
	var total int64
	query := db.Model(&domains.TaskEvent{}).Where("task_guid = ?", task.Guid)
	if err = query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = query.Order("id asc").Limit(pageInfo.Size).Offset(pageInfo.Size * (pageInfo.Page - 1)).Find(&events).Error
	return events, total, err
}

func normalizeRawJSON(raw json.RawMessage) (string, error) {
	value := strings.TrimSpace(string(raw))
	if value == "" || value == "null" {
		return "{}", nil
	}
	if !json.Valid([]byte(value)) {
		return "", errors.New("invalid json payload")
	}
	return value, nil
}

func normalizeOptionalRawJSON(raw json.RawMessage) (string, error) {
	value := strings.TrimSpace(string(raw))
	if value == "" || value == "null" {
		return "", nil
	}
	if !json.Valid([]byte(value)) {
		return "", errors.New("invalid json result")
	}
	return value, nil
}

func isAllowedTaskStatus(status string) bool {
	switch status {
	case domains.TaskStatusPending,
		domains.TaskStatusDispatched,
		domains.TaskStatusRunning,
		domains.TaskStatusSuccess,
		domains.TaskStatusFailed,
		domains.TaskStatusTimeout,
		domains.TaskStatusCancelled:
		return true
	default:
		return false
	}
}

func isFinalTaskStatus(status string) bool {
	switch status {
	case domains.TaskStatusSuccess, domains.TaskStatusFailed, domains.TaskStatusTimeout, domains.TaskStatusCancelled:
		return true
	default:
		return false
	}
}
