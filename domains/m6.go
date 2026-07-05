package domains

import commonDomains "github.com/wfu-work/nav-common-go-lib/domains"

type RegistryCredential struct {
	commonDomains.BaseDataEntity
	Name           string `gorm:"column:name;size:120;index" json:"name"`
	Registry       string `gorm:"column:registry;size:255;index" json:"registry"`
	Username       string `gorm:"column:username;size:255" json:"username"`
	PasswordSecret string `gorm:"column:password_secret;type:text" json:"-"`
	Email          string `gorm:"column:email;size:255" json:"email"`
	Status         string `gorm:"column:status;size:30;index" json:"status"`
	Description    string `gorm:"column:description;size:1000" json:"description"`
	LastUsedAt     int64  `gorm:"column:last_used_at" json:"lastUsedAt"`
}

func (RegistryCredential) TableName() string {
	return "nav_docker_registry_credentials"
}

type DeployTemplate struct {
	commonDomains.BaseDataEntity
	Name        string `gorm:"column:name;size:120;index" json:"name"`
	Type        string `gorm:"column:type;size:30;index" json:"type"`
	Content     string `gorm:"column:content;type:longtext" json:"content"`
	Variables   string `gorm:"column:variables;type:text" json:"variables"`
	Description string `gorm:"column:description;size:1000" json:"description"`
}

func (DeployTemplate) TableName() string {
	return "nav_docker_deploy_templates"
}

type OperationApproval struct {
	commonDomains.BaseDataEntity
	HostGuid       string `gorm:"column:host_guid;size:50;index" json:"hostGuid"`
	Action         string `gorm:"column:action;size:100;index" json:"action"`
	ResourceType   string `gorm:"column:resource_type;size:50;index" json:"resourceType"`
	ResourceGuid   string `gorm:"column:resource_guid;size:100;index" json:"resourceGuid"`
	TaskGuid       string `gorm:"column:task_guid;size:50;index" json:"taskGuid"`
	Status         string `gorm:"column:status;size:30;index" json:"status"`
	RequestPayload string `gorm:"column:request_payload;type:longtext" json:"requestPayload"`
	Reason         string `gorm:"column:reason;size:1000" json:"reason"`
	CreatedBy      string `gorm:"column:created_by;size:50" json:"createdBy"`
	ReviewedBy     string `gorm:"column:reviewed_by;size:50" json:"reviewedBy"`
	ReviewedAt     int64  `gorm:"column:reviewed_at" json:"reviewedAt"`
}

func (OperationApproval) TableName() string {
	return "nav_docker_operation_approvals"
}

type OperationPolicy struct {
	commonDomains.BaseDataEntity
	Name             string `gorm:"column:name;size:120;index" json:"name"`
	HostGuid         string `gorm:"column:host_guid;size:50;index" json:"hostGuid"`
	Action           string `gorm:"column:action;size:100;index" json:"action"`
	ResourceType     string `gorm:"column:resource_type;size:50;index" json:"resourceType"`
	ApprovalRequired bool   `gorm:"column:approval_required" json:"approvalRequired"`
	Enabled          bool   `gorm:"column:enabled;index" json:"enabled"`
	Description      string `gorm:"column:description;size:1000" json:"description"`
}

func (OperationPolicy) TableName() string {
	return "nav_docker_operation_policies"
}

type OperationAudit struct {
	commonDomains.BaseDataEntity
	TaskGuid       string `gorm:"column:task_guid;size:50;uniqueIndex" json:"taskGuid"`
	ApprovalGuid   string `gorm:"column:approval_guid;size:50;index" json:"approvalGuid"`
	HostGuid       string `gorm:"column:host_guid;size:50;index" json:"hostGuid"`
	AgentGuid      string `gorm:"column:agent_guid;size:50;index" json:"agentGuid"`
	Action         string `gorm:"column:action;size:100;index" json:"action"`
	ResourceType   string `gorm:"column:resource_type;size:50;index" json:"resourceType"`
	ResourceGuid   string `gorm:"column:resource_guid;size:100;index" json:"resourceGuid"`
	Status         string `gorm:"column:status;size:30;index" json:"status"`
	RequestPayload string `gorm:"column:request_payload;type:longtext" json:"requestPayload"`
	Result         string `gorm:"column:result;type:longtext" json:"result"`
	ErrorMessage   string `gorm:"column:error_message;size:1000" json:"errorMessage"`
	Operator       string `gorm:"column:operator;size:50" json:"operator"`
	ClientIP       string `gorm:"column:client_ip;size:100" json:"clientIp"`
	StartedAt      int64  `gorm:"column:started_at" json:"startedAt"`
	FinishedAt     int64  `gorm:"column:finished_at" json:"finishedAt"`
}

func (OperationAudit) TableName() string {
	return "nav_docker_operation_audits"
}
