package services

import (
	"errors"
	"strings"

	"docker-go/domains"

	commonServices "github.com/wfu-work/nav-common-go-lib/services"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
	"gorm.io/gorm/clause"
)

type SettingService struct {
	commonServices.CrudService[domains.Setting]
}

var SettingServiceApp = new(SettingService)

// List returns paginated runtime settings.
func (s SettingService) List(params map[string]string) (interface{}, int64, error) {
	return s.CrudService.List(commonUtils.ToPageInfo(params), "key,description")
}

// Save creates or updates a runtime setting by key.
func (s SettingService) Save(setting domains.Setting) error {
	setting.Key = strings.TrimSpace(setting.Key)
	if setting.Key == "" {
		return errors.New("missing setting key")
	}
	return s.DB().Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "key"}},
		DoUpdates: clause.Assignments(map[string]any{
			"value":       setting.Value,
			"description": setting.Description,
			"update_time": setting.UpdateTime,
		}),
	}).Create(&setting).Error
}

// Delete soft-deletes one setting by guid.
func (s SettingService) Delete(guid string) error {
	if guid == "" {
		return errors.New("missing setting guid")
	}
	return s.CrudService.DeleteByGuid(guid)
}
