package services

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"docker-go/domains"

	commonServices "github.com/wfu-work/nav-common-go-lib/services"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
)

type RegistryCredentialService struct {
	commonServices.CrudService[domains.RegistryCredential]
}

type RegistryCredentialSaveRequest struct {
	Name        string `json:"name"`
	Registry    string `json:"registry"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

var RegistryCredentialServiceApp = new(RegistryCredentialService)

func (s RegistryCredentialService) List(params map[string]string) (interface{}, int64, error) {
	if params["content"] == "" {
		params["content"] = params["keyword"]
	}
	return s.CrudService.List(commonUtils.ToPageInfo(params), "name,registry,username,email,status,description")
}

func (s RegistryCredentialService) Get(guid string) (*domains.RegistryCredential, error) {
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return nil, errors.New("missing registry credential guid")
	}
	item, err := s.CrudService.GetByGuid(guid)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("registry credential not found")
	}
	return item, nil
}

func (s RegistryCredentialService) Create(req RegistryCredentialSaveRequest) (*domains.RegistryCredential, error) {
	item, err := normalizeRegistryCredential(req, nil)
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

func (s RegistryCredentialService) Update(guid string, req RegistryCredentialSaveRequest) (*domains.RegistryCredential, error) {
	existing, err := s.Get(guid)
	if err != nil {
		return nil, err
	}
	item, err := normalizeRegistryCredential(req, existing)
	if err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	err = db.Model(&domains.RegistryCredential{}).Where("guid = ?", existing.Guid).Updates(map[string]any{
		"name":            item.Name,
		"registry":        item.Registry,
		"username":        item.Username,
		"password_secret": item.PasswordSecret,
		"email":           item.Email,
		"status":          item.Status,
		"description":     item.Description,
		"update_time":     now,
	}).Error
	if err != nil {
		return nil, err
	}
	return s.Get(existing.Guid)
}

func (s RegistryCredentialService) Delete(guid string) error {
	item, err := s.Get(guid)
	if err != nil {
		return err
	}
	return s.CrudService.DeleteByGuid(item.Guid)
}

func (s RegistryCredentialService) BuildRegistryAuth(guid string) (string, error) {
	item, err := s.Get(guid)
	if err != nil {
		return "", err
	}
	if item.Status != domains.RegistryCredentialStatusEnabled {
		return "", errors.New("registry credential is disabled")
	}
	password, err := decodeCredentialPassword(item.PasswordSecret)
	if err != nil {
		return "", err
	}
	authConfig := map[string]string{
		"username":      item.Username,
		"password":      password,
		"email":         item.Email,
		"serveraddress": item.Registry,
	}
	raw, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	db := s.DB()
	if db != nil {
		_ = db.Model(&domains.RegistryCredential{}).Where("guid = ?", item.Guid).Updates(map[string]any{
			"last_used_at": time.Now().UnixMilli(),
			"update_time":  time.Now().UnixMilli(),
		}).Error
	}
	return base64.URLEncoding.EncodeToString(raw), nil
}

func normalizeRegistryCredential(req RegistryCredentialSaveRequest, existing *domains.RegistryCredential) (domains.RegistryCredential, error) {
	item := domains.RegistryCredential{}
	if existing != nil {
		item = *existing
	}
	item.Name = strings.TrimSpace(req.Name)
	item.Registry = strings.TrimRight(strings.TrimSpace(req.Registry), "/")
	item.Username = strings.TrimSpace(req.Username)
	item.Email = strings.TrimSpace(req.Email)
	item.Status = normalizeRegistryCredentialStatus(req.Status)
	item.Description = strings.TrimSpace(req.Description)
	if item.Name == "" {
		return item, errors.New("missing credential name")
	}
	if item.Registry == "" {
		return item, errors.New("missing registry")
	}
	if item.Username == "" {
		return item, errors.New("missing username")
	}
	if req.Password != "" {
		item.PasswordSecret = encodeCredentialPassword(req.Password)
	}
	if item.PasswordSecret == "" {
		return item, errors.New("missing password")
	}
	return item, nil
}

func normalizeRegistryCredentialStatus(status string) string {
	switch strings.TrimSpace(status) {
	case domains.RegistryCredentialStatusDisabled:
		return domains.RegistryCredentialStatusDisabled
	default:
		return domains.RegistryCredentialStatusEnabled
	}
}

func encodeCredentialPassword(password string) string {
	return base64.StdEncoding.EncodeToString([]byte(password))
}

func decodeCredentialPassword(secret string) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", errors.New("missing password secret")
	}
	raw, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", errors.New("invalid password secret")
	}
	return string(raw), nil
}
