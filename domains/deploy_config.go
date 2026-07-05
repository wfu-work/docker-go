package domains

import commonDomains "github.com/wfu-work/nav-common-go-lib/domains"

type DeployConfig struct {
	commonDomains.BaseDataEntity
	Name               string `gorm:"column:name;size:120;index" json:"name"`
	Type               string `gorm:"column:type;size:30;index" json:"type"`
	ProjectName        string `gorm:"column:project_name;size:120;index" json:"projectName"`
	CurrentVersionGuid string `gorm:"column:current_version_guid;size:50;index" json:"currentVersionGuid"`
	Description        string `gorm:"column:description;size:1000" json:"description"`
}

func (DeployConfig) TableName() string {
	return "nav_docker_deploy_configs"
}

type DeployConfigVersion struct {
	commonDomains.BaseDataEntity
	ConfigGuid string `gorm:"column:config_guid;size:50;uniqueIndex:idx_deploy_config_version_no;index" json:"configGuid"`
	VersionNo  int    `gorm:"column:version_no;uniqueIndex:idx_deploy_config_version_no" json:"versionNo"`
	Content    string `gorm:"column:content;type:longtext" json:"content"`
	Checksum   string `gorm:"column:checksum;size:64;index" json:"checksum"`
	Message    string `gorm:"column:message;size:500" json:"message"`
	CreatedBy  string `gorm:"column:created_by;size:50" json:"createdBy"`
}

func (DeployConfigVersion) TableName() string {
	return "nav_docker_deploy_config_versions"
}

type DeployRelease struct {
	commonDomains.BaseDataEntity
	ConfigGuid   string `gorm:"column:config_guid;size:50;index" json:"configGuid"`
	VersionGuid  string `gorm:"column:version_guid;size:50;index" json:"versionGuid"`
	HostGuid     string `gorm:"column:host_guid;size:50;index" json:"hostGuid"`
	AgentGuid    string `gorm:"column:agent_guid;size:50;index" json:"agentGuid"`
	TaskGuid     string `gorm:"column:task_guid;size:50;uniqueIndex" json:"taskGuid"`
	Action       string `gorm:"column:action;size:30;index" json:"action"`
	ProjectName  string `gorm:"column:project_name;size:120;index" json:"projectName"`
	Status       string `gorm:"column:status;size:30;index" json:"status"`
	Result       string `gorm:"column:result;type:longtext" json:"result"`
	ErrorMessage string `gorm:"column:error_message;size:1000" json:"errorMessage"`
	StartedAt    int64  `gorm:"column:started_at" json:"startedAt"`
	FinishedAt   int64  `gorm:"column:finished_at" json:"finishedAt"`
}

func (DeployRelease) TableName() string {
	return "nav_docker_deploy_releases"
}

type ComposeTaskPayload struct {
	ConfigGuid    string   `json:"configGuid"`
	VersionGuid   string   `json:"versionGuid"`
	ProjectName   string   `json:"projectName"`
	Content       string   `json:"content"`
	Action        string   `json:"action"`
	Pull          bool     `json:"pull"`
	RemoveOrphans bool     `json:"removeOrphans"`
	Services      []string `json:"services,omitempty"`
	Profiles      []string `json:"profiles,omitempty"`
}

type ComposeCommandStep struct {
	Name       string   `json:"name"`
	Command    []string `json:"command"`
	Stdout     string   `json:"stdout"`
	Stderr     string   `json:"stderr"`
	ExitCode   int      `json:"exitCode"`
	StartedAt  int64    `json:"startedAt"`
	FinishedAt int64    `json:"finishedAt"`
}

type ComposeTaskResult struct {
	ConfigGuid  string               `json:"configGuid"`
	VersionGuid string               `json:"versionGuid"`
	ProjectName string               `json:"projectName"`
	Action      string               `json:"action"`
	Workdir     string               `json:"workdir"`
	Steps       []ComposeCommandStep `json:"steps"`
}
