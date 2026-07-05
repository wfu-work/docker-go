package domains

import commonDomains "github.com/wfu-work/nav-common-go-lib/domains"

type Host struct {
	commonDomains.BaseDataEntity
	Name            string `gorm:"column:name;size:100;index" json:"name"`
	Address         string `gorm:"column:address;size:255" json:"address"`
	Description     string `gorm:"column:description;size:500" json:"description"`
	Status          string `gorm:"column:status;size:30;index" json:"status"`
	LastHeartbeatAt int64  `gorm:"column:last_heartbeat_at;index" json:"lastHeartbeatAt"`
}

func (Host) TableName() string {
	return "nav_docker_hosts"
}
