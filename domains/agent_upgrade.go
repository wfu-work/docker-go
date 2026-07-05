package domains

import commonDomains "github.com/wfu-work/nav-common-go-lib/domains"

type AgentUpgradePackage struct {
	commonDomains.BaseDataEntity
	Version      string `gorm:"column:version;size:80;index" json:"version"`
	OS           string `gorm:"column:os;size:50;index" json:"os"`
	Arch         string `gorm:"column:arch;size:50;index" json:"arch"`
	DownloadURL  string `gorm:"column:download_url;size:1000" json:"downloadUrl"`
	SHA256       string `gorm:"column:sha256;size:64;index" json:"sha256"`
	Signature    string `gorm:"column:signature;type:text" json:"signature"`
	Status       string `gorm:"column:status;size:30;index" json:"status"`
	ReleaseNotes string `gorm:"column:release_notes;type:text" json:"releaseNotes"`
}

func (AgentUpgradePackage) TableName() string {
	return "nav_docker_agent_upgrade_packages"
}

type AgentUpgradePayload struct {
	AgentGuid           string `json:"agentGuid"`
	PackageGuid         string `json:"packageGuid"`
	Version             string `json:"version"`
	DownloadURL         string `json:"downloadUrl"`
	SHA256              string `json:"sha256"`
	Signature           string `json:"signature"`
	Force               bool   `json:"force"`
	RestartDelaySeconds int    `json:"restartDelaySeconds"`
}

type AgentUpgradeResult struct {
	AgentGuid           string `json:"agentGuid"`
	PackageGuid         string `json:"packageGuid"`
	PreviousVersion     string `json:"previousVersion"`
	TargetVersion       string `json:"targetVersion"`
	ExecutablePath      string `json:"executablePath"`
	BackupPath          string `json:"backupPath"`
	DownloadedBytes     int64  `json:"downloadedBytes"`
	SHA256              string `json:"sha256"`
	SignatureVerified   bool   `json:"signatureVerified"`
	RestartScheduled    bool   `json:"restartScheduled"`
	RestartDelaySeconds int    `json:"restartDelaySeconds"`
}
