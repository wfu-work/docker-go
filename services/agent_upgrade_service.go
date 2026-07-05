package services

import (
	"encoding/json"
	"errors"
	"strings"

	"docker-go/domains"

	commonServices "github.com/wfu-work/nav-common-go-lib/services"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
)

type AgentUpgradePackageService struct {
	commonServices.CrudService[domains.AgentUpgradePackage]
}

type AgentUpgradePackageSaveRequest struct {
	Version      string `json:"version"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	DownloadURL  string `json:"downloadUrl"`
	SHA256       string `json:"sha256"`
	Signature    string `json:"signature"`
	Status       string `json:"status"`
	ReleaseNotes string `json:"releaseNotes"`
}

type AgentUpgradeRequest struct {
	PackageGuid         string `json:"packageGuid"`
	Force               bool   `json:"force"`
	RestartDelaySeconds int    `json:"restartDelaySeconds"`
	ApprovalGuid        string `json:"approvalGuid"`
	Operator            string `json:"operator"`
}

var AgentUpgradePackageServiceApp = new(AgentUpgradePackageService)

func (s AgentUpgradePackageService) List(params map[string]string) (interface{}, int64, error) {
	if params["content"] == "" {
		params["content"] = params["keyword"]
	}
	return s.CrudService.List(commonUtils.ToPageInfo(params), "version,os,arch,download_url,sha256,status,release_notes")
}

func (s AgentUpgradePackageService) Get(guid string) (*domains.AgentUpgradePackage, error) {
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return nil, errors.New("missing upgrade package guid")
	}
	item, err := s.CrudService.GetByGuid(guid)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("agent upgrade package not found")
	}
	return item, nil
}

func (s AgentUpgradePackageService) Create(req AgentUpgradePackageSaveRequest) (*domains.AgentUpgradePackage, error) {
	item, err := normalizeAgentUpgradePackage(req)
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

func (s AgentUpgradePackageService) Update(guid string, req AgentUpgradePackageSaveRequest) (*domains.AgentUpgradePackage, error) {
	existing, err := s.Get(guid)
	if err != nil {
		return nil, err
	}
	item, err := normalizeAgentUpgradePackage(req)
	if err != nil {
		return nil, err
	}
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	if err = db.Model(&domains.AgentUpgradePackage{}).Where("guid = ?", existing.Guid).Updates(map[string]any{
		"version":       item.Version,
		"os":            item.OS,
		"arch":          item.Arch,
		"download_url":  item.DownloadURL,
		"sha256":        item.SHA256,
		"signature":     item.Signature,
		"status":        item.Status,
		"release_notes": item.ReleaseNotes,
	}).Error; err != nil {
		return nil, err
	}
	return s.Get(existing.Guid)
}

func (s AgentUpgradePackageService) Delete(guid string) error {
	item, err := s.Get(guid)
	if err != nil {
		return err
	}
	return s.CrudService.DeleteByGuid(item.Guid)
}

func (s AgentUpgradePackageService) CreateUpgradeTask(agentGuid string, req AgentUpgradeRequest) (*domains.Task, error) {
	agent, err := AgentServiceApp.Get(agentGuid)
	if err != nil {
		return nil, err
	}
	pkg, err := s.Get(req.PackageGuid)
	if err != nil {
		return nil, err
	}
	if pkg.Status != domains.AgentUpgradePackageStatusEnabled {
		return nil, errors.New("agent upgrade package is disabled")
	}
	if pkg.OS != "" && agent.OS != "" && pkg.OS != agent.OS {
		return nil, errors.New("agent upgrade package os does not match agent")
	}
	if pkg.Arch != "" && agent.Arch != "" && pkg.Arch != agent.Arch {
		return nil, errors.New("agent upgrade package arch does not match agent")
	}
	if req.RestartDelaySeconds <= 0 {
		req.RestartDelaySeconds = 3
	}
	payload := domains.AgentUpgradePayload{
		AgentGuid:           agent.Guid,
		PackageGuid:         pkg.Guid,
		Version:             pkg.Version,
		DownloadURL:         pkg.DownloadURL,
		SHA256:              pkg.SHA256,
		Signature:           pkg.Signature,
		Force:               req.Force,
		RestartDelaySeconds: req.RestartDelaySeconds,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return TaskServiceApp.Create(TaskCreateRequest{
		HostGuid:       agent.HostGuid,
		AgentGuid:      agent.Guid,
		ApprovalGuid:   strings.TrimSpace(req.ApprovalGuid),
		Type:           domains.TaskTypeAgentUpgrade,
		Payload:        raw,
		TimeoutSeconds: 600,
		ResourceType:   "agent",
		ResourceGuid:   agent.Guid,
		Operator:       strings.TrimSpace(req.Operator),
	})
}

func normalizeAgentUpgradePackage(req AgentUpgradePackageSaveRequest) (domains.AgentUpgradePackage, error) {
	item := domains.AgentUpgradePackage{
		Version:      strings.TrimSpace(req.Version),
		OS:           strings.TrimSpace(req.OS),
		Arch:         strings.TrimSpace(req.Arch),
		DownloadURL:  strings.TrimSpace(req.DownloadURL),
		SHA256:       strings.ToLower(strings.TrimSpace(req.SHA256)),
		Signature:    strings.TrimSpace(req.Signature),
		Status:       normalizeAgentUpgradePackageStatus(req.Status),
		ReleaseNotes: strings.TrimSpace(req.ReleaseNotes),
	}
	if item.Version == "" {
		return item, errors.New("missing upgrade package version")
	}
	if item.DownloadURL == "" {
		return item, errors.New("missing upgrade package download url")
	}
	if len(item.SHA256) != 64 {
		return item, errors.New("sha256 must be 64 hex characters")
	}
	for _, r := range item.SHA256 {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') {
			continue
		}
		return item, errors.New("sha256 must be hex encoded")
	}
	return item, nil
}

func normalizeAgentUpgradePackageStatus(status string) string {
	switch strings.TrimSpace(status) {
	case domains.AgentUpgradePackageStatusDisabled:
		return domains.AgentUpgradePackageStatusDisabled
	default:
		return domains.AgentUpgradePackageStatusEnabled
	}
}
