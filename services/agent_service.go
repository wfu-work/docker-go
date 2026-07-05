package services

import (
	"errors"
	"strings"
	"time"

	"docker-go/domains"

	commonServices "github.com/wfu-work/nav-common-go-lib/services"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
	"gorm.io/gorm"
)

type AgentService struct {
	commonServices.CrudService[domains.Agent]
}

type AgentRegisterRequest struct {
	Token         string `json:"token"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	DockerVersion string `json:"dockerVersion"`
	OS            string `json:"os"`
	Arch          string `json:"arch"`
}

type AgentHeartbeatRequest struct {
	AgentGuid     string `json:"agentGuid"`
	Token         string `json:"token"`
	Version       string `json:"version"`
	DockerVersion string `json:"dockerVersion"`
	Status        string `json:"status"`
}

var AgentServiceApp = new(AgentService)

func (s AgentService) List(params map[string]string) (interface{}, int64, error) {
	if params["content"] == "" {
		params["content"] = params["keyword"]
	}
	return s.CrudService.List(commonUtils.ToPageInfo(params), "host_guid,name,version,docker_version,os,arch,status")
}

func (s AgentService) Get(guid string) (*domains.Agent, error) {
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return nil, errors.New("missing agent guid")
	}
	agent, err := s.CrudService.GetByGuid(guid)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, errors.New("agent not found")
	}
	return agent, nil
}

func (s AgentService) Register(req AgentRegisterRequest) (*domains.Agent, error) {
	token := strings.TrimSpace(req.Token)
	if token == "" {
		return nil, errors.New("missing agent token")
	}
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}

	var agent domains.Agent
	err := db.Where("token_hash = ?", hashAgentToken(token)).First(&agent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("invalid agent token")
	}
	if err != nil {
		return nil, err
	}
	if agent.Status == domains.AgentStatusDisabled {
		return nil, errors.New("agent is disabled")
	}

	now := time.Now().UnixMilli()
	updates := map[string]any{
		"name":              strings.TrimSpace(req.Name),
		"version":           strings.TrimSpace(req.Version),
		"docker_version":    strings.TrimSpace(req.DockerVersion),
		"os":                strings.TrimSpace(req.OS),
		"arch":              strings.TrimSpace(req.Arch),
		"status":            domains.AgentStatusOnline,
		"registered_at":     now,
		"last_heartbeat_at": now,
		"update_time":       now,
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domains.Agent{}).Where("guid = ?", agent.Guid).Updates(updates).Error; err != nil {
			return err
		}
		return tx.Model(&domains.Host{}).Where("guid = ?", agent.HostGuid).Updates(map[string]any{
			"status":            domains.HostStatusOnline,
			"last_heartbeat_at": now,
			"update_time":       now,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.Get(agent.Guid)
}

func (s AgentService) Heartbeat(req AgentHeartbeatRequest) (*domains.Agent, error) {
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	agent, err := validateAgentCredential(db, req.AgentGuid, req.Token)
	if err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = domains.AgentStatusOnline
	}
	if status != domains.AgentStatusOnline && status != domains.AgentStatusOffline {
		return nil, errors.New("unsupported agent status")
	}
	updates := map[string]any{
		"status":            status,
		"last_heartbeat_at": now,
		"update_time":       now,
	}
	if strings.TrimSpace(req.Version) != "" {
		updates["version"] = strings.TrimSpace(req.Version)
	}
	if strings.TrimSpace(req.DockerVersion) != "" {
		updates["docker_version"] = strings.TrimSpace(req.DockerVersion)
	}

	hostStatus := domains.HostStatusOnline
	if status == domains.AgentStatusOffline {
		hostStatus = domains.HostStatusOffline
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domains.Agent{}).Where("guid = ?", agent.Guid).Updates(updates).Error; err != nil {
			return err
		}
		return tx.Model(&domains.Host{}).Where("guid = ?", agent.HostGuid).Updates(map[string]any{
			"status":            hostStatus,
			"last_heartbeat_at": now,
			"update_time":       now,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.Get(agent.Guid)
}

func (s AgentService) Authenticate(agentGuid string, token string) (*domains.Agent, error) {
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	return validateAgentCredential(db, agentGuid, token)
}

func (s AgentService) MarkConnected(agentGuid string) (*domains.Agent, error) {
	agent, err := s.Get(agentGuid)
	if err != nil {
		return nil, err
	}
	if agent.Status == domains.AgentStatusDisabled {
		return nil, errors.New("agent is disabled")
	}
	now := time.Now().UnixMilli()
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domains.Agent{}).Where("guid = ?", agent.Guid).Updates(map[string]any{
			"status":            domains.AgentStatusOnline,
			"last_heartbeat_at": now,
			"update_time":       now,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&domains.Host{}).Where("guid = ?", agent.HostGuid).Updates(map[string]any{
			"status":            domains.HostStatusOnline,
			"last_heartbeat_at": now,
			"update_time":       now,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.Get(agent.Guid)
}

func (s AgentService) MarkDisconnected(agentGuid string) (*domains.Agent, error) {
	agent, err := s.Get(agentGuid)
	if err != nil {
		return nil, err
	}
	if agent.Status == domains.AgentStatusDisabled {
		return agent, nil
	}
	now := time.Now().UnixMilli()
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domains.Agent{}).Where("guid = ?", agent.Guid).Updates(map[string]any{
			"status":      domains.AgentStatusOffline,
			"update_time": now,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&domains.Host{}).Where("guid = ?", agent.HostGuid).Updates(map[string]any{
			"status":      domains.HostStatusOffline,
			"update_time": now,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.Get(agent.Guid)
}

func validateAgentCredential(db *gorm.DB, agentGuid string, token string) (*domains.Agent, error) {
	agentGuid = strings.TrimSpace(agentGuid)
	token = strings.TrimSpace(token)
	if agentGuid == "" {
		return nil, errors.New("missing agent guid")
	}
	if token == "" {
		return nil, errors.New("missing agent token")
	}
	var agent domains.Agent
	err := db.Where("guid = ? AND token_hash = ?", agentGuid, hashAgentToken(token)).First(&agent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("invalid agent credential")
	}
	if err != nil {
		return nil, err
	}
	if agent.Status == domains.AgentStatusDisabled {
		return nil, errors.New("agent is disabled")
	}
	return &agent, nil
}
