package domains

import commonDomains "github.com/wfu-work/nav-common-go-lib/domains"

type Agent struct {
	commonDomains.BaseDataEntity
	HostGuid        string `gorm:"column:host_guid;size:50;index" json:"hostGuid"`
	Name            string `gorm:"column:name;size:100" json:"name"`
	TokenHash       string `gorm:"column:token_hash;size:128;uniqueIndex" json:"-"`
	Version         string `gorm:"column:version;size:50" json:"version"`
	DockerVersion   string `gorm:"column:docker_version;size:100" json:"dockerVersion"`
	OS              string `gorm:"column:os;size:50" json:"os"`
	Arch            string `gorm:"column:arch;size:50" json:"arch"`
	Status          string `gorm:"column:status;size:30;index" json:"status"`
	RegisteredAt    int64  `gorm:"column:registered_at" json:"registeredAt"`
	LastHeartbeatAt int64  `gorm:"column:last_heartbeat_at;index" json:"lastHeartbeatAt"`
}

func (Agent) TableName() string {
	return "nav_docker_agents"
}
