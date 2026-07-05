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

type HostService struct {
	commonServices.CrudService[domains.Host]
}

type AgentTokenResponse struct {
	HostGuid  string `json:"hostGuid"`
	AgentGuid string `json:"agentGuid"`
	Token     string `json:"token"`
}

var HostServiceApp = new(HostService)

func (s HostService) List(params map[string]string) (interface{}, int64, error) {
	if params["content"] == "" {
		params["content"] = params["keyword"]
	}
	return s.CrudService.List(commonUtils.ToPageInfo(params), "name,address,description,status")
}

func (s HostService) Get(guid string) (*domains.Host, error) {
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return nil, errors.New("missing host guid")
	}
	host, err := s.CrudService.GetByGuid(guid)
	if err != nil {
		return nil, err
	}
	if host == nil {
		return nil, errors.New("host not found")
	}
	return host, nil
}

func (s HostService) Create(host domains.Host) (*domains.Host, error) {
	host.Name = strings.TrimSpace(host.Name)
	host.Address = strings.TrimSpace(host.Address)
	host.Description = strings.TrimSpace(host.Description)
	host.Status = normalizeHostStatus(host.Status)
	if host.Name == "" {
		return nil, errors.New("missing host name")
	}
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	if err := db.Create(&host).Error; err != nil {
		return nil, err
	}
	return &host, nil
}

func (s HostService) Update(guid string, host domains.Host) (*domains.Host, error) {
	existing, err := s.Get(guid)
	if err != nil {
		return nil, err
	}
	host.Name = strings.TrimSpace(host.Name)
	host.Address = strings.TrimSpace(host.Address)
	host.Description = strings.TrimSpace(host.Description)
	if host.Name == "" {
		return nil, errors.New("missing host name")
	}
	status := strings.TrimSpace(host.Status)
	if status == "" {
		status = existing.Status
	} else {
		status = normalizeHostStatus(status)
	}
	now := time.Now().UnixMilli()
	updates := map[string]any{
		"name":        host.Name,
		"address":     host.Address,
		"description": host.Description,
		"status":      status,
		"update_time": now,
	}
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	if err := db.Model(&domains.Host{}).Where("guid = ?", existing.Guid).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.Get(existing.Guid)
}

func (s HostService) Delete(guid string) error {
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return errors.New("missing host guid")
	}
	return s.CrudService.DeleteByGuid(guid)
}

func (s HostService) GenerateAgentToken(hostGuid string) (*AgentTokenResponse, error) {
	host, err := s.Get(hostGuid)
	if err != nil {
		return nil, err
	}
	token, err := newAgentToken()
	if err != nil {
		return nil, err
	}
	tokenHash := hashAgentToken(token)
	now := time.Now().UnixMilli()
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}

	var agent domains.Agent
	err = db.Where("host_guid = ?", host.Guid).First(&agent).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		agent = domains.Agent{
			HostGuid:  host.Guid,
			TokenHash: tokenHash,
			Status:    domains.AgentStatusPending,
		}
		if err = db.Create(&agent).Error; err != nil {
			return nil, err
		}
	case err != nil:
		return nil, err
	default:
		updates := map[string]any{
			"token_hash":  tokenHash,
			"status":      domains.AgentStatusPending,
			"update_time": now,
		}
		if err = db.Model(&domains.Agent{}).Where("guid = ?", agent.Guid).Updates(updates).Error; err != nil {
			return nil, err
		}
		agent.TokenHash = tokenHash
		agent.Status = domains.AgentStatusPending
	}

	return &AgentTokenResponse{
		HostGuid:  host.Guid,
		AgentGuid: agent.Guid,
		Token:     token,
	}, nil
}

func normalizeHostStatus(status string) string {
	switch strings.TrimSpace(status) {
	case domains.HostStatusOnline:
		return domains.HostStatusOnline
	case domains.HostStatusDisabled:
		return domains.HostStatusDisabled
	default:
		return domains.HostStatusOffline
	}
}
