package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"docker-go/domains"

	commonServices "github.com/wfu-work/nav-common-go-lib/services"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
)

type DeployTemplateService struct {
	commonServices.CrudService[domains.DeployTemplate]
}

type DeployTemplateSaveRequest struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Content     string          `json:"content"`
	Variables   json.RawMessage `json:"variables"`
	Description string          `json:"description"`
}

type DeployTemplateRenderRequest struct {
	Values map[string]any `json:"values"`
}

type DeployTemplateRenderResponse struct {
	TemplateGuid string         `json:"templateGuid"`
	Content      string         `json:"content"`
	Values       map[string]any `json:"values"`
}

type DeployTemplateCreateConfigRequest struct {
	Name        string         `json:"name"`
	ProjectName string         `json:"projectName"`
	Description string         `json:"description"`
	Message     string         `json:"message"`
	CreatedBy   string         `json:"createdBy"`
	Values      map[string]any `json:"values"`
}

var DeployTemplateServiceApp = new(DeployTemplateService)

func (s DeployTemplateService) List(params map[string]string) (interface{}, int64, error) {
	if params["content"] == "" {
		params["content"] = params["keyword"]
	}
	return s.CrudService.List(commonUtils.ToPageInfo(params), "name,type,description")
}

func (s DeployTemplateService) Get(guid string) (*domains.DeployTemplate, error) {
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return nil, errors.New("missing deploy template guid")
	}
	item, err := s.CrudService.GetByGuid(guid)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("deploy template not found")
	}
	return item, nil
}

func (s DeployTemplateService) Create(req DeployTemplateSaveRequest) (*domains.DeployTemplate, error) {
	item, err := normalizeDeployTemplate(req)
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

func (s DeployTemplateService) Update(guid string, req DeployTemplateSaveRequest) (*domains.DeployTemplate, error) {
	existing, err := s.Get(guid)
	if err != nil {
		return nil, err
	}
	item, err := normalizeDeployTemplate(req)
	if err != nil {
		return nil, err
	}
	db := s.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	if err = db.Model(&domains.DeployTemplate{}).Where("guid = ?", existing.Guid).Updates(map[string]any{
		"name":        item.Name,
		"type":        item.Type,
		"content":     item.Content,
		"variables":   item.Variables,
		"description": item.Description,
	}).Error; err != nil {
		return nil, err
	}
	return s.Get(existing.Guid)
}

func (s DeployTemplateService) Delete(guid string) error {
	item, err := s.Get(guid)
	if err != nil {
		return err
	}
	return s.CrudService.DeleteByGuid(item.Guid)
}

func (s DeployTemplateService) Render(guid string, req DeployTemplateRenderRequest) (*DeployTemplateRenderResponse, error) {
	item, err := s.Get(guid)
	if err != nil {
		return nil, err
	}
	content := renderTemplateContent(item.Content, req.Values)
	if err = validateDeployConfigContent(item.Type, content); err != nil {
		return nil, err
	}
	return &DeployTemplateRenderResponse{TemplateGuid: item.Guid, Content: content, Values: req.Values}, nil
}

func (s DeployTemplateService) CreateConfig(guid string, req DeployTemplateCreateConfigRequest) (*DeployConfigDetail, error) {
	rendered, err := s.Render(guid, DeployTemplateRenderRequest{Values: req.Values})
	if err != nil {
		return nil, err
	}
	template, err := s.Get(guid)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = template.Name
	}
	return DeployConfigServiceApp.Create(DeployConfigSaveRequest{
		Name:        name,
		Type:        template.Type,
		ProjectName: req.ProjectName,
		Description: req.Description,
		Content:     &rendered.Content,
		Message:     req.Message,
		CreatedBy:   req.CreatedBy,
	})
}

func normalizeDeployTemplate(req DeployTemplateSaveRequest) (domains.DeployTemplate, error) {
	item := domains.DeployTemplate{
		Name:        strings.TrimSpace(req.Name),
		Type:        normalizeDeployConfigType(req.Type),
		Content:     req.Content,
		Description: strings.TrimSpace(req.Description),
	}
	if item.Name == "" {
		return item, errors.New("missing deploy template name")
	}
	if strings.TrimSpace(item.Content) == "" {
		return item, errors.New("missing template content")
	}
	variables, err := normalizeJSONText(req.Variables)
	if err != nil {
		return item, err
	}
	item.Variables = variables
	return item, nil
}

func renderTemplateContent(content string, values map[string]any) string {
	for key, value := range values {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		text := fmt.Sprint(value)
		content = strings.ReplaceAll(content, "{{"+key+"}}", text)
		content = strings.ReplaceAll(content, "${"+key+"}", text)
	}
	return content
}

func normalizeJSONText(raw json.RawMessage) (string, error) {
	value := strings.TrimSpace(string(raw))
	if value == "" || value == "null" {
		return "{}", nil
	}
	if !json.Valid([]byte(value)) {
		return "", errors.New("invalid json")
	}
	return value, nil
}
