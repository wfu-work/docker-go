package domains

import commonDomains "github.com/wfu-work/nav-common-go-lib/domains"

type Task struct {
	commonDomains.BaseDataEntity
	HostGuid       string `gorm:"column:host_guid;size:50;index" json:"hostGuid"`
	AgentGuid      string `gorm:"column:agent_guid;size:50;index" json:"agentGuid"`
	ApprovalGuid   string `gorm:"column:approval_guid;size:50;index" json:"approvalGuid"`
	Type           string `gorm:"column:type;size:100;index" json:"type"`
	Status         string `gorm:"column:status;size:30;index" json:"status"`
	Payload        string `gorm:"column:payload;type:text" json:"payload"`
	Result         string `gorm:"column:result;type:text" json:"result"`
	ErrorMessage   string `gorm:"column:error_message;size:1000" json:"errorMessage"`
	TimeoutSeconds int    `gorm:"column:timeout_seconds" json:"timeoutSeconds"`
	StartedAt      int64  `gorm:"column:started_at" json:"startedAt"`
	FinishedAt     int64  `gorm:"column:finished_at" json:"finishedAt"`
}

func (Task) TableName() string {
	return "nav_docker_tasks"
}

type TaskEvent struct {
	commonDomains.BaseDataEntity
	TaskGuid string `gorm:"column:task_guid;size:50;index" json:"taskGuid"`
	Status   string `gorm:"column:status;size:30;index" json:"status"`
	Message  string `gorm:"column:message;size:1000" json:"message"`
	Data     string `gorm:"column:data;type:text" json:"data"`
}

func (TaskEvent) TableName() string {
	return "nav_docker_task_events"
}
