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

type OperationApprovalService struct {
	commonServices.CrudService[domains.OperationApproval]
}

type OperationAuditService struct {
	commonServices.CrudService[domains.OperationAudit]
}

type OperationPolicyService struct {
	commonServices.CrudService[domains.OperationPolicy]
}

type OperationApprovalCreateRequest struct {
	HostGuid       string          `json:"hostGuid"`
	Action         string          `json:"action"`
	ResourceType   string          `json:"resourceType"`
	ResourceGuid   string          `json:"resourceGuid"`
	RequestPayload json.RawMessage `json:"requestPayload"`
	Reason         string          `json:"reason"`
	CreatedBy      string          `json:"createdBy"`
}

type OperationApprovalReviewRequest struct {
	ReviewedBy string `json:"reviewedBy"`
	Reason     string `json:"reason"`
}

type OperationPolicySaveRequest struct {
	Name             string `json:"name"`
	HostGuid         string `json:"hostGuid"`
	Action           string `json:"action"`
	ResourceType     string `json:"resourceType"`
	ApprovalRequired bool   `json:"approvalRequired"`
	Enabled          *bool  `json:"enabled"`
	Description      string `json:"description"`
}

var (
	OperationApprovalServiceApp = new(OperationApprovalService)
	OperationPolicyServiceApp   = new(OperationPolicyService)
	OperationAuditServiceApp    = new(OperationAuditService)
)

func (s OperationPolicyService) List(params map[string]string) (interface{}, int64, error) {
	if params["content"] == "" {
		params["content"] = params["keyword"]
	}
	return s.CrudService.List(commonUtils.ToPageInfo(params), "name,host_guid,action,resource_type,description")
}

func (s OperationPolicyService) Get(guid string) (*domains.OperationPolicy, error) {
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return nil, errors.New("missing operation policy guid")
	}
	item, err := s.CrudService.GetByGuid(guid)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("operation policy not found")
	}
	return item, nil
}

func (s OperationPolicyService) Create(req OperationPolicySaveRequest) (*domains.OperationPolicy, error) {
	item, err := normalizeOperationPolicy(req, nil)
	if err != nil {
		return nil, err
	}
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	if err = db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s OperationPolicyService) Update(guid string, req OperationPolicySaveRequest) (*domains.OperationPolicy, error) {
	existing, err := s.Get(guid)
	if err != nil {
		return nil, err
	}
	item, err := normalizeOperationPolicy(req, existing)
	if err != nil {
		return nil, err
	}
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	if err = db.Model(&domains.OperationPolicy{}).Where("guid = ?", existing.Guid).Updates(map[string]any{
		"name":              item.Name,
		"host_guid":         item.HostGuid,
		"action":            item.Action,
		"resource_type":     item.ResourceType,
		"approval_required": item.ApprovalRequired,
		"enabled":           item.Enabled,
		"description":       item.Description,
		"update_time":       time.Now().UnixMilli(),
	}).Error; err != nil {
		return nil, err
	}
	return s.Get(existing.Guid)
}

func (s OperationPolicyService) Delete(guid string) error {
	item, err := s.Get(guid)
	if err != nil {
		return err
	}
	return s.CrudService.DeleteByGuid(item.Guid)
}

func (s OperationPolicyService) RequiresApproval(hostGuid string, action string, resourceType string) (bool, error) {
	db := s.DB()
	if db == nil {
		return false, errors.New("database is not initialized")
	}
	var policies []domains.OperationPolicy
	err := db.Where(
		"enabled = ? AND approval_required = ? AND action = ? AND (host_guid = '' OR host_guid = ?) AND (resource_type = '' OR resource_type = ?)",
		true,
		true,
		strings.TrimSpace(action),
		strings.TrimSpace(hostGuid),
		strings.TrimSpace(resourceType),
	).Order("host_guid desc, resource_type desc, id desc").Limit(1).Find(&policies).Error
	if err != nil {
		return false, err
	}
	return len(policies) > 0, nil
}

func normalizeOperationPolicy(req OperationPolicySaveRequest, existing *domains.OperationPolicy) (domains.OperationPolicy, error) {
	item := domains.OperationPolicy{}
	if existing != nil {
		item = *existing
	}
	item.Name = strings.TrimSpace(req.Name)
	item.HostGuid = strings.TrimSpace(req.HostGuid)
	item.Action = strings.TrimSpace(req.Action)
	item.ResourceType = strings.TrimSpace(req.ResourceType)
	item.ApprovalRequired = req.ApprovalRequired
	item.Description = strings.TrimSpace(req.Description)
	if req.Enabled == nil {
		item.Enabled = true
	} else {
		item.Enabled = *req.Enabled
	}
	if item.Name == "" {
		return item, errors.New("missing policy name")
	}
	if item.Action == "" {
		return item, errors.New("missing policy action")
	}
	return item, nil
}

func (s OperationApprovalService) List(params map[string]string) (interface{}, int64, error) {
	pageInfo := commonUtils.ToPageInfo(params)
	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.Size <= 0 {
		pageInfo.Size = 20
	}
	if pageInfo.Size > 100 {
		pageInfo.Size = 100
	}
	db := s.DB()
	if db == nil {
		return nil, 0, errors.New("database is not initialized")
	}
	query := db.Model(&domains.OperationApproval{})
	for _, field := range []string{"host_guid", "action", "resource_type", "resource_guid", "status", "task_guid"} {
		if value := strings.TrimSpace(params[snakeToLowerCamel(field)]); value != "" {
			query = query.Where(field+" = ?", value)
		}
	}
	if keyword := strings.TrimSpace(params["keyword"]); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("action LIKE ? OR resource_guid LIKE ? OR reason LIKE ?", like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domains.OperationApproval
	err := query.Order("id desc").Limit(pageInfo.Size).Offset(pageInfo.Size * (pageInfo.Page - 1)).Find(&items).Error
	return items, total, err
}

func (s OperationApprovalService) Get(guid string) (*domains.OperationApproval, error) {
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return nil, errors.New("missing approval guid")
	}
	item, err := s.CrudService.GetByGuid(guid)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("operation approval not found")
	}
	return item, nil
}

func (s OperationApprovalService) Create(req OperationApprovalCreateRequest) (*domains.OperationApproval, error) {
	req.HostGuid = strings.TrimSpace(req.HostGuid)
	req.Action = strings.TrimSpace(req.Action)
	req.ResourceType = strings.TrimSpace(req.ResourceType)
	req.ResourceGuid = strings.TrimSpace(req.ResourceGuid)
	if req.HostGuid == "" {
		return nil, errors.New("missing host guid")
	}
	if req.Action == "" {
		return nil, errors.New("missing action")
	}
	payload, err := normalizeRawJSON(req.RequestPayload)
	if err != nil {
		return nil, err
	}
	payload = redactJSONText(payload)
	item := domains.OperationApproval{
		HostGuid:       req.HostGuid,
		Action:         req.Action,
		ResourceType:   req.ResourceType,
		ResourceGuid:   req.ResourceGuid,
		Status:         domains.OperationApprovalStatusPending,
		RequestPayload: payload,
		Reason:         strings.TrimSpace(req.Reason),
		CreatedBy:      strings.TrimSpace(req.CreatedBy),
	}
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	if err = db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s OperationApprovalService) Approve(guid string, req OperationApprovalReviewRequest) (*domains.OperationApproval, error) {
	return s.review(guid, domains.OperationApprovalStatusApproved, req)
}

func (s OperationApprovalService) Reject(guid string, req OperationApprovalReviewRequest) (*domains.OperationApproval, error) {
	return s.review(guid, domains.OperationApprovalStatusRejected, req)
}

func (s OperationApprovalService) Cancel(guid string, req OperationApprovalReviewRequest) (*domains.OperationApproval, error) {
	return s.review(guid, domains.OperationApprovalStatusCancelled, req)
}

func (s OperationApprovalService) UseApproval(tx *gorm.DB, approvalGuid string, req TaskCreateRequest, task domains.Task) error {
	approvalGuid = strings.TrimSpace(approvalGuid)
	if approvalGuid == "" {
		return nil
	}
	var approval domains.OperationApproval
	if err := tx.Where("guid = ?", approvalGuid).First(&approval).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("operation approval not found")
		}
		return err
	}
	if approval.Status != domains.OperationApprovalStatusApproved {
		return errors.New("operation approval is not approved")
	}
	if approval.TaskGuid != "" {
		return errors.New("operation approval is already used")
	}
	if approval.HostGuid != req.HostGuid {
		return errors.New("operation approval host does not match task")
	}
	if approval.Action != req.Type {
		return errors.New("operation approval action does not match task")
	}
	resourceType, resourceGuid := inferTaskResource(req.Type, string(req.Payload))
	if approval.ResourceType != "" && approval.ResourceType != resourceType {
		return errors.New("operation approval resource type does not match task")
	}
	if approval.ResourceGuid != "" && approval.ResourceGuid != resourceGuid {
		return errors.New("operation approval resource does not match task")
	}
	now := time.Now().UnixMilli()
	return tx.Model(&domains.OperationApproval{}).Where("guid = ?", approval.Guid).Updates(map[string]any{
		"status":      domains.OperationApprovalStatusUsed,
		"task_guid":   task.Guid,
		"update_time": now,
	}).Error
}

func (s OperationApprovalService) review(guid string, status string, req OperationApprovalReviewRequest) (*domains.OperationApproval, error) {
	item, err := s.Get(guid)
	if err != nil {
		return nil, err
	}
	if item.Status != domains.OperationApprovalStatusPending && item.Status != domains.OperationApprovalStatusApproved {
		return nil, errors.New("operation approval cannot be reviewed")
	}
	now := time.Now().UnixMilli()
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	if err = db.Model(&domains.OperationApproval{}).Where("guid = ?", item.Guid).Updates(map[string]any{
		"status":      status,
		"reason":      strings.TrimSpace(req.Reason),
		"reviewed_by": strings.TrimSpace(req.ReviewedBy),
		"reviewed_at": now,
		"update_time": now,
	}).Error; err != nil {
		return nil, err
	}
	return s.Get(item.Guid)
}

func (s OperationAuditService) List(params map[string]string) (interface{}, int64, error) {
	pageInfo := commonUtils.ToPageInfo(params)
	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.Size <= 0 {
		pageInfo.Size = 20
	}
	if pageInfo.Size > 100 {
		pageInfo.Size = 100
	}
	db := s.DB()
	if db == nil {
		return nil, 0, errors.New("database is not initialized")
	}
	query := db.Model(&domains.OperationAudit{})
	for _, field := range []string{"task_guid", "approval_guid", "host_guid", "agent_guid", "action", "resource_type", "resource_guid", "status"} {
		if value := strings.TrimSpace(params[snakeToLowerCamel(field)]); value != "" {
			query = query.Where(field+" = ?", value)
		}
	}
	if keyword := strings.TrimSpace(params["keyword"]); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("action LIKE ? OR resource_guid LIKE ? OR error_message LIKE ?", like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domains.OperationAudit
	err := query.Order("id desc").Limit(pageInfo.Size).Offset(pageInfo.Size * (pageInfo.Page - 1)).Find(&items).Error
	return items, total, err
}

func (s OperationAuditService) Get(guid string) (*domains.OperationAudit, error) {
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return nil, errors.New("missing audit guid")
	}
	item, err := s.CrudService.GetByGuid(guid)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("operation audit not found")
	}
	return item, nil
}

func (s OperationAuditService) CreateAuditInTx(tx *gorm.DB, task domains.Task, req TaskCreateRequest) error {
	resourceType, resourceGuid := inferTaskResource(task.Type, task.Payload)
	if req.ResourceType != "" {
		resourceType = strings.TrimSpace(req.ResourceType)
	}
	if req.ResourceGuid != "" {
		resourceGuid = strings.TrimSpace(req.ResourceGuid)
	}
	audit := domains.OperationAudit{
		TaskGuid:       task.Guid,
		ApprovalGuid:   task.ApprovalGuid,
		HostGuid:       task.HostGuid,
		AgentGuid:      task.AgentGuid,
		Action:         task.Type,
		ResourceType:   resourceType,
		ResourceGuid:   resourceGuid,
		Status:         task.Status,
		RequestPayload: redactJSONText(task.Payload),
		Operator:       strings.TrimSpace(req.Operator),
		ClientIP:       strings.TrimSpace(req.ClientIP),
		StartedAt:      task.StartedAt,
		FinishedAt:     task.FinishedAt,
	}
	return tx.Create(&audit).Error
}

func (s OperationAuditService) SyncFromTask(task domains.Task) error {
	db := s.DB()
	if db == nil {
		return errors.New("database is not initialized")
	}
	updates := map[string]any{
		"agent_guid":    task.AgentGuid,
		"approval_guid": task.ApprovalGuid,
		"status":        task.Status,
		"started_at":    task.StartedAt,
		"finished_at":   task.FinishedAt,
		"update_time":   time.Now().UnixMilli(),
	}
	if strings.TrimSpace(task.Result) != "" {
		updates["result"] = task.Result
	}
	if strings.TrimSpace(task.ErrorMessage) != "" {
		updates["error_message"] = task.ErrorMessage
	}
	return db.Model(&domains.OperationAudit{}).Where("task_guid = ?", task.Guid).Updates(updates).Error
}

func inferTaskResource(taskType string, payload string) (string, string) {
	var raw map[string]any
	_ = json.Unmarshal([]byte(payload), &raw)
	stringField := func(key string) string {
		if value, ok := raw[key].(string); ok {
			return strings.TrimSpace(value)
		}
		return ""
	}
	switch taskType {
	case domains.TaskTypeAgentUpgrade:
		return "agent", stringField("agentGuid")
	case domains.TaskTypeDockerContainerStart,
		domains.TaskTypeDockerContainerStop,
		domains.TaskTypeDockerContainerRestart,
		domains.TaskTypeDockerContainerRemove,
		domains.TaskTypeDockerContainerLogs,
		domains.TaskTypeDockerContainerStream:
		return "container", stringField("containerId")
	case domains.TaskTypeDockerImagePull:
		return "image", stringField("image")
	case domains.TaskTypeDockerConfigValidate,
		domains.TaskTypeDockerConfigDeploy,
		domains.TaskTypeDockerComposeUp,
		domains.TaskTypeDockerComposeDown,
		domains.TaskTypeDockerComposeRestart,
		domains.TaskTypeDockerComposePull:
		return "deploy_config", stringField("configGuid")
	default:
		return "", ""
	}
}

func redactJSONText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "null" || !json.Valid([]byte(value)) {
		return value
	}
	var raw any
	if err := json.Unmarshal([]byte(value), &raw); err != nil {
		return value
	}
	redactJSONValue(raw)
	out, err := json.Marshal(raw)
	if err != nil {
		return value
	}
	return string(out)
}

func redactJSONValue(value any) {
	switch item := value.(type) {
	case map[string]any:
		for key, child := range item {
			if isSensitiveJSONKey(key) {
				item[key] = "***"
				continue
			}
			redactJSONValue(child)
		}
	case []any:
		for _, child := range item {
			redactJSONValue(child)
		}
	}
}

func isSensitiveJSONKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "password", "passwordsecret", "registryauth", "token", "secret":
		return true
	default:
		return false
	}
}
