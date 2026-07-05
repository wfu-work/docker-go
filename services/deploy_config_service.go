package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"docker-go/domains"

	commonServices "github.com/wfu-work/nav-common-go-lib/services"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
	"go.yaml.in/yaml/v3"
	"gorm.io/gorm"
)

const maxDeployConfigContentBytes = 2 * 1024 * 1024

type DeployConfigCrudService struct {
	commonServices.CrudService[domains.DeployConfig]
}

type DeployConfigVersionService struct {
	commonServices.CrudService[domains.DeployConfigVersion]
}

type DeployReleaseService struct {
	commonServices.CrudService[domains.DeployRelease]
}

type DeployConfigService struct {
	ConfigService  DeployConfigCrudService
	VersionService DeployConfigVersionService
	ReleaseService DeployReleaseService
}

type DeployConfigSaveRequest struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	ProjectName string  `json:"projectName"`
	Description string  `json:"description"`
	Content     *string `json:"content"`
	Message     string  `json:"message"`
	CreatedBy   string  `json:"createdBy"`
}

type DeployConfigVersionCreateRequest struct {
	Content   string `json:"content"`
	Message   string `json:"message"`
	CreatedBy string `json:"createdBy"`
}

type DeployConfigTaskRequest struct {
	HostGuid       string   `json:"hostGuid"`
	ApprovalGuid   string   `json:"approvalGuid"`
	VersionGuid    string   `json:"versionGuid"`
	Content        *string  `json:"content"`
	ProjectName    string   `json:"projectName"`
	Action         string   `json:"action"`
	Pull           bool     `json:"pull"`
	RemoveOrphans  bool     `json:"removeOrphans"`
	Services       []string `json:"services"`
	Profiles       []string `json:"profiles"`
	TimeoutSeconds int      `json:"timeoutSeconds"`
	Operator       string   `json:"operator"`
}

type DeployConfigDetail struct {
	Config         *domains.DeployConfig        `json:"config"`
	CurrentVersion *domains.DeployConfigVersion `json:"currentVersion"`
}

type DeployConfigActionResponse struct {
	Task    *domains.Task                `json:"task"`
	Release *domains.DeployRelease       `json:"release,omitempty"`
	Version *domains.DeployConfigVersion `json:"version,omitempty"`
}

var DeployConfigServiceApp = new(DeployConfigService)

func (s DeployConfigService) List(params map[string]string) (interface{}, int64, error) {
	if params["content"] == "" {
		params["content"] = params["keyword"]
	}
	return s.ConfigService.CrudService.List(commonUtils.ToPageInfo(params), "name,type,project_name,description")
}

func (s DeployConfigService) Get(guid string) (*DeployConfigDetail, error) {
	config, err := s.getConfig(guid)
	if err != nil {
		return nil, err
	}
	version, err := s.currentVersion(*config)
	if err != nil {
		return nil, err
	}
	return &DeployConfigDetail{Config: config, CurrentVersion: version}, nil
}

func (s DeployConfigService) Create(req DeployConfigSaveRequest) (*DeployConfigDetail, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Type = normalizeDeployConfigType(req.Type)
	req.ProjectName = normalizeComposeProjectName(req.ProjectName, req.Name, "")
	req.Description = strings.TrimSpace(req.Description)
	req.Message = strings.TrimSpace(req.Message)
	req.CreatedBy = strings.TrimSpace(req.CreatedBy)
	if req.Name == "" {
		return nil, errors.New("missing deploy config name")
	}
	if req.Content == nil {
		return nil, errors.New("missing yaml content")
	}
	content := *req.Content
	if err := validateDeployConfigContent(req.Type, content); err != nil {
		return nil, err
	}
	db := s.ConfigService.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	config := domains.DeployConfig{
		Name:        req.Name,
		Type:        req.Type,
		ProjectName: req.ProjectName,
		Description: req.Description,
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&config).Error; err != nil {
			return err
		}
		version, err := s.createVersionInTx(tx, config.Guid, content, req.Message, req.CreatedBy)
		if err != nil {
			return err
		}
		config.CurrentVersionGuid = version.Guid
		return tx.Model(&domains.DeployConfig{}).Where("guid = ?", config.Guid).Updates(map[string]any{
			"current_version_guid": version.Guid,
			"update_time":          time.Now().UnixMilli(),
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.Get(config.Guid)
}

func (s DeployConfigService) Update(guid string, req DeployConfigSaveRequest) (*DeployConfigDetail, error) {
	config, err := s.getConfig(guid)
	if err != nil {
		return nil, err
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Type = normalizeDeployConfigType(req.Type)
	req.ProjectName = normalizeComposeProjectName(req.ProjectName, req.Name, config.Guid)
	req.Description = strings.TrimSpace(req.Description)
	req.Message = strings.TrimSpace(req.Message)
	req.CreatedBy = strings.TrimSpace(req.CreatedBy)
	if req.Name == "" {
		return nil, errors.New("missing deploy config name")
	}
	if req.Content != nil {
		if err = validateDeployConfigContent(req.Type, *req.Content); err != nil {
			return nil, err
		}
	}
	db := s.ConfigService.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()
		updates := map[string]any{
			"name":         req.Name,
			"type":         req.Type,
			"project_name": req.ProjectName,
			"description":  req.Description,
			"update_time":  now,
		}
		if req.Content != nil {
			shouldCreateVersion := true
			current, currentErr := s.currentVersionWithDB(tx, *config)
			if currentErr != nil {
				return currentErr
			}
			if current != nil && current.Checksum == contentChecksum(*req.Content) {
				shouldCreateVersion = false
			}
			if shouldCreateVersion {
				version, createErr := s.createVersionInTx(tx, config.Guid, *req.Content, req.Message, req.CreatedBy)
				if createErr != nil {
					return createErr
				}
				updates["current_version_guid"] = version.Guid
			}
		}
		return tx.Model(&domains.DeployConfig{}).Where("guid = ?", config.Guid).Updates(updates).Error
	})
	if err != nil {
		return nil, err
	}
	return s.Get(config.Guid)
}

func (s DeployConfigService) Delete(guid string) error {
	config, err := s.getConfig(guid)
	if err != nil {
		return err
	}
	return s.ConfigService.CrudService.DeleteByGuid(config.Guid)
}

func (s DeployConfigService) ListVersions(configGuid string, params map[string]string) (interface{}, int64, error) {
	config, err := s.getConfig(configGuid)
	if err != nil {
		return nil, 0, err
	}
	pageInfo := commonUtils.ToPageInfo(params)
	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.Size <= 0 {
		pageInfo.Size = 20
	}
	if pageInfo.Size > 100 {
		pageInfo.Size = 100
	}
	db := s.VersionService.DB()
	if db == nil {
		return nil, 0, errors.New("database is not initialized")
	}
	var total int64
	query := db.Model(&domains.DeployConfigVersion{}).Where("config_guid = ?", config.Guid)
	if err = query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var versions []domains.DeployConfigVersion
	err = query.Order("version_no desc, id desc").Limit(pageInfo.Size).Offset(pageInfo.Size * (pageInfo.Page - 1)).Find(&versions).Error
	return versions, total, err
}

func (s DeployConfigService) GetVersion(configGuid string, versionGuid string) (*domains.DeployConfigVersion, error) {
	config, err := s.getConfig(configGuid)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(versionGuid) == "" {
		version, err := s.currentVersion(*config)
		if err != nil {
			return nil, err
		}
		if version == nil {
			return nil, errors.New("deploy config version not found")
		}
		return version, nil
	}
	return s.getVersion(config.Guid, versionGuid)
}

func (s DeployConfigService) CreateVersion(configGuid string, req DeployConfigVersionCreateRequest) (*domains.DeployConfigVersion, error) {
	config, err := s.getConfig(configGuid)
	if err != nil {
		return nil, err
	}
	if err = validateDeployConfigContent(config.Type, req.Content); err != nil {
		return nil, err
	}
	db := s.VersionService.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	var version domains.DeployConfigVersion
	err = db.Transaction(func(tx *gorm.DB) error {
		var createErr error
		version, createErr = s.createVersionInTx(tx, config.Guid, req.Content, strings.TrimSpace(req.Message), strings.TrimSpace(req.CreatedBy))
		if createErr != nil {
			return createErr
		}
		return tx.Model(&domains.DeployConfig{}).Where("guid = ?", config.Guid).Updates(map[string]any{
			"current_version_guid": version.Guid,
			"update_time":          time.Now().UnixMilli(),
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (s DeployConfigService) Validate(configGuid string, req DeployConfigTaskRequest) (*DeployConfigActionResponse, error) {
	config, err := s.getConfig(configGuid)
	if err != nil {
		return nil, err
	}
	req.HostGuid = strings.TrimSpace(req.HostGuid)
	if req.HostGuid == "" {
		return nil, errors.New("missing host guid")
	}
	version, content, err := s.resolveTaskContent(*config, req.VersionGuid, req.Content)
	if err != nil {
		return nil, err
	}
	payload := domains.ComposeTaskPayload{
		ConfigGuid:  config.Guid,
		ProjectName: normalizeComposeProjectName(req.ProjectName, config.ProjectName, config.Guid),
		Content:     content,
		Action:      "validate",
		Services:    cleanStringSlice(req.Services),
		Profiles:    cleanStringSlice(req.Profiles),
	}
	if version != nil {
		payload.VersionGuid = version.Guid
	}
	task, err := s.createTask(req.HostGuid, req.ApprovalGuid, domains.TaskTypeDockerConfigValidate, payload, req.TimeoutSeconds, 120, strings.TrimSpace(req.Operator))
	if err != nil {
		return nil, err
	}
	return &DeployConfigActionResponse{Task: task, Version: version}, nil
}

func (s DeployConfigService) Deploy(configGuid string, req DeployConfigTaskRequest) (*DeployConfigActionResponse, error) {
	return s.deploy(configGuid, req, "", "")
}

func (s DeployConfigService) Rollback(configGuid string, req DeployConfigTaskRequest) (*DeployConfigActionResponse, error) {
	req.Action = domains.DeployReleaseActionUp
	if strings.TrimSpace(req.VersionGuid) == "" {
		return nil, errors.New("missing rollback version guid")
	}
	return s.deploy(configGuid, req, domains.DeployReleaseActionRollback, req.VersionGuid)
}

func (s DeployConfigService) ListReleases(params map[string]string) (interface{}, int64, error) {
	pageInfo := commonUtils.ToPageInfo(params)
	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.Size <= 0 {
		pageInfo.Size = 20
	}
	if pageInfo.Size > 100 {
		pageInfo.Size = 100
	}
	db := s.ReleaseService.DB()
	if db == nil {
		return nil, 0, errors.New("database is not initialized")
	}
	query := db.Model(&domains.DeployRelease{})
	for _, field := range []string{"config_guid", "version_guid", "host_guid", "task_guid", "action", "status", "project_name"} {
		paramName := snakeToLowerCamel(field)
		if value := strings.TrimSpace(params[paramName]); value != "" {
			query = query.Where(field+" = ?", value)
		}
	}
	if keyword := strings.TrimSpace(params["keyword"]); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("project_name LIKE ? OR error_message LIKE ?", like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var releases []domains.DeployRelease
	err := query.Order("id desc").Limit(pageInfo.Size).Offset(pageInfo.Size * (pageInfo.Page - 1)).Find(&releases).Error
	return releases, total, err
}

func (s DeployConfigService) GetRelease(guid string) (*domains.DeployRelease, error) {
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return nil, errors.New("missing release guid")
	}
	release, err := s.ReleaseService.CrudService.GetByGuid(guid)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return nil, errors.New("deploy release not found")
	}
	return release, nil
}

func (s DeployConfigService) SyncReleaseFromTask(task domains.Task) error {
	if !isDeployTaskType(task.Type) {
		return nil
	}
	db := s.ReleaseService.DB()
	if db == nil {
		return errors.New("database is not initialized")
	}
	var release domains.DeployRelease
	err := db.Where("task_guid = ?", task.Guid).First(&release).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	now := time.Now().UnixMilli()
	updates := map[string]any{
		"agent_guid":  task.AgentGuid,
		"status":      task.Status,
		"started_at":  task.StartedAt,
		"finished_at": task.FinishedAt,
		"update_time": now,
	}
	if strings.TrimSpace(task.Result) != "" {
		updates["result"] = task.Result
	}
	if strings.TrimSpace(task.ErrorMessage) != "" {
		updates["error_message"] = task.ErrorMessage
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domains.DeployRelease{}).Where("guid = ?", release.Guid).Updates(updates).Error; err != nil {
			return err
		}
		if task.Status == domains.TaskStatusSuccess && shouldPromoteReleaseVersion(release.Action) && release.VersionGuid != "" {
			return tx.Model(&domains.DeployConfig{}).Where("guid = ?", release.ConfigGuid).Updates(map[string]any{
				"current_version_guid": release.VersionGuid,
				"update_time":          now,
			}).Error
		}
		return nil
	})
}

func (s DeployConfigService) deploy(configGuid string, req DeployConfigTaskRequest, releaseAction string, forcedVersionGuid string) (*DeployConfigActionResponse, error) {
	config, err := s.getConfig(configGuid)
	if err != nil {
		return nil, err
	}
	req.HostGuid = strings.TrimSpace(req.HostGuid)
	if req.HostGuid == "" {
		return nil, errors.New("missing host guid")
	}
	versionGuid := strings.TrimSpace(req.VersionGuid)
	if forcedVersionGuid != "" {
		versionGuid = strings.TrimSpace(forcedVersionGuid)
	}
	version, content, err := s.resolveTaskContent(*config, versionGuid, nil)
	if err != nil {
		return nil, err
	}
	action := normalizeComposeAction(req.Action)
	if releaseAction == domains.DeployReleaseActionRollback {
		action = domains.DeployReleaseActionUp
	}
	projectName := normalizeComposeProjectName(req.ProjectName, config.ProjectName, config.Guid)
	payload := domains.ComposeTaskPayload{
		ConfigGuid:    config.Guid,
		VersionGuid:   version.Guid,
		ProjectName:   projectName,
		Content:       content,
		Action:        action,
		Pull:          req.Pull,
		RemoveOrphans: req.RemoveOrphans,
		Services:      cleanStringSlice(req.Services),
		Profiles:      cleanStringSlice(req.Profiles),
	}
	task, err := s.createTask(req.HostGuid, req.ApprovalGuid, domains.TaskTypeDockerConfigDeploy, payload, req.TimeoutSeconds, 900, strings.TrimSpace(req.Operator))
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(releaseAction) == "" {
		releaseAction = action
	}
	release := domains.DeployRelease{
		ConfigGuid:  config.Guid,
		VersionGuid: version.Guid,
		HostGuid:    req.HostGuid,
		AgentGuid:   task.AgentGuid,
		TaskGuid:    task.Guid,
		Action:      releaseAction,
		ProjectName: projectName,
		Status:      task.Status,
		StartedAt:   task.StartedAt,
		FinishedAt:  task.FinishedAt,
	}
	db := s.ReleaseService.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	if err = db.Create(&release).Error; err != nil {
		return nil, err
	}
	return &DeployConfigActionResponse{Task: task, Release: &release, Version: version}, nil
}

func (s DeployConfigService) createTask(hostGuid string, approvalGuid string, taskType string, payload domains.ComposeTaskPayload, timeoutSeconds int, defaultTimeout int, operator string) (*domains.Task, error) {
	if timeoutSeconds <= 0 {
		timeoutSeconds = defaultTimeout
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return TaskServiceApp.Create(TaskCreateRequest{
		HostGuid:       hostGuid,
		ApprovalGuid:   approvalGuid,
		Type:           taskType,
		Payload:        raw,
		TimeoutSeconds: timeoutSeconds,
		ResourceType:   "deploy_config",
		ResourceGuid:   payload.ConfigGuid,
		Operator:       operator,
	})
}

func (s DeployConfigService) getConfig(guid string) (*domains.DeployConfig, error) {
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return nil, errors.New("missing deploy config guid")
	}
	config, err := s.ConfigService.CrudService.GetByGuid(guid)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, errors.New("deploy config not found")
	}
	return config, nil
}

func (s DeployConfigService) getVersion(configGuid string, versionGuid string) (*domains.DeployConfigVersion, error) {
	versionGuid = strings.TrimSpace(versionGuid)
	if versionGuid == "" {
		return nil, errors.New("missing version guid")
	}
	db := s.VersionService.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	var version domains.DeployConfigVersion
	err := db.Where("guid = ? AND config_guid = ?", versionGuid, configGuid).First(&version).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("deploy config version not found")
	}
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (s DeployConfigService) currentVersion(config domains.DeployConfig) (*domains.DeployConfigVersion, error) {
	db := s.VersionService.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	return s.currentVersionWithDB(db, config)
}

func (s DeployConfigService) currentVersionWithDB(db *gorm.DB, config domains.DeployConfig) (*domains.DeployConfigVersion, error) {
	var version domains.DeployConfigVersion
	var err error
	if strings.TrimSpace(config.CurrentVersionGuid) != "" {
		err = db.Where("guid = ? AND config_guid = ?", config.CurrentVersionGuid, config.Guid).First(&version).Error
	} else {
		err = db.Where("config_guid = ?", config.Guid).Order("version_no desc, id desc").First(&version).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (s DeployConfigService) resolveTaskContent(config domains.DeployConfig, versionGuid string, content *string) (*domains.DeployConfigVersion, string, error) {
	if content != nil {
		if err := validateDeployConfigContent(config.Type, *content); err != nil {
			return nil, "", err
		}
		return nil, *content, nil
	}
	version, err := s.GetVersion(config.Guid, versionGuid)
	if err != nil {
		return nil, "", err
	}
	if err = validateDeployConfigContent(config.Type, version.Content); err != nil {
		return nil, "", err
	}
	return version, version.Content, nil
}

func (s DeployConfigService) createVersionInTx(tx *gorm.DB, configGuid string, content string, message string, createdBy string) (domains.DeployConfigVersion, error) {
	var maxVersionNo int
	if err := tx.Model(&domains.DeployConfigVersion{}).Where("config_guid = ?", configGuid).Select("COALESCE(MAX(version_no), 0)").Scan(&maxVersionNo).Error; err != nil {
		return domains.DeployConfigVersion{}, err
	}
	version := domains.DeployConfigVersion{
		ConfigGuid: configGuid,
		VersionNo:  maxVersionNo + 1,
		Content:    content,
		Checksum:   contentChecksum(content),
		Message:    strings.TrimSpace(message),
		CreatedBy:  strings.TrimSpace(createdBy),
	}
	if err := tx.Create(&version).Error; err != nil {
		return domains.DeployConfigVersion{}, err
	}
	return version, nil
}

func validateDeployConfigContent(configType string, content string) error {
	if len([]byte(content)) > maxDeployConfigContentBytes {
		return errors.New("yaml content is too large")
	}
	if strings.TrimSpace(content) == "" {
		return errors.New("missing yaml content")
	}
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(content), &root); err != nil {
		return errors.New("invalid yaml: " + err.Error())
	}
	if normalizeDeployConfigType(configType) != domains.DeployConfigTypeCompose {
		return nil
	}
	node := rootNode(root)
	if node == nil || node.Kind != yaml.MappingNode {
		return errors.New("compose yaml must be a mapping")
	}
	services := mappingValue(node, "services")
	if services == nil {
		return errors.New("compose yaml missing services")
	}
	if services.Kind != yaml.MappingNode || len(services.Content) == 0 {
		return errors.New("compose services must be a non-empty mapping")
	}
	return nil
}

func rootNode(root yaml.Node) *yaml.Node {
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		return root.Content[0]
	}
	return &root
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func contentChecksum(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func normalizeDeployConfigType(configType string) string {
	switch strings.TrimSpace(configType) {
	case domains.DeployConfigTypeCompose, "":
		return domains.DeployConfigTypeCompose
	default:
		return domains.DeployConfigTypeCompose
	}
}

func normalizeComposeAction(action string) string {
	switch strings.TrimSpace(action) {
	case domains.DeployReleaseActionDown:
		return domains.DeployReleaseActionDown
	case domains.DeployReleaseActionRestart:
		return domains.DeployReleaseActionRestart
	case domains.DeployReleaseActionPull:
		return domains.DeployReleaseActionPull
	default:
		return domains.DeployReleaseActionUp
	}
}

func normalizeComposeProjectName(projectName string, fallback string, guid string) string {
	source := strings.TrimSpace(projectName)
	if source == "" {
		source = strings.TrimSpace(fallback)
	}
	if source == "" {
		source = strings.TrimSpace(guid)
	}
	source = strings.ToLower(source)
	var builder strings.Builder
	lastSep := false
	for _, r := range source {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastSep = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastSep = false
		case r == '-' || r == '_':
			if !lastSep {
				builder.WriteRune(r)
				lastSep = true
			}
		default:
			if !lastSep {
				builder.WriteByte('-')
				lastSep = true
			}
		}
	}
	result := strings.Trim(builder.String(), "-_")
	if result == "" {
		result = "nav-docker"
	}
	first := result[0]
	if !((first >= 'a' && first <= 'z') || (first >= '0' && first <= '9')) {
		result = "nav-" + result
	}
	if len(result) > 80 {
		result = strings.Trim(result[:80], "-_")
	}
	return result
}

func cleanStringSlice(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func isDeployTaskType(taskType string) bool {
	switch taskType {
	case domains.TaskTypeDockerConfigValidate,
		domains.TaskTypeDockerConfigDeploy,
		domains.TaskTypeDockerComposeUp,
		domains.TaskTypeDockerComposeDown,
		domains.TaskTypeDockerComposeRestart,
		domains.TaskTypeDockerComposePull:
		return true
	default:
		return false
	}
}

func shouldPromoteReleaseVersion(action string) bool {
	switch action {
	case domains.DeployReleaseActionUp, domains.DeployReleaseActionRollback:
		return true
	default:
		return false
	}
}

func snakeToLowerCamel(value string) string {
	parts := strings.Split(value, "_")
	for index := 1; index < len(parts); index++ {
		if parts[index] == "" {
			continue
		}
		parts[index] = strings.ToUpper(parts[index][:1]) + parts[index][1:]
	}
	return strings.Join(parts, "")
}
