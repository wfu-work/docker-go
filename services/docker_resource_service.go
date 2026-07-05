package services

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"docker-go/domains"

	commonServices "github.com/wfu-work/nav-common-go-lib/services"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DockerContainerService struct {
	commonServices.CrudService[domains.DockerContainer]
}

type DockerImageService struct {
	commonServices.CrudService[domains.DockerImage]
}

type DockerResourceService struct {
	ContainerService DockerContainerService
	ImageService     DockerImageService
}

var DockerResourceServiceApp = new(DockerResourceService)

func (s DockerResourceService) ListContainers(hostGuid string, params map[string]string) (interface{}, int64, error) {
	hostGuid = strings.TrimSpace(hostGuid)
	if hostGuid == "" {
		return nil, 0, errors.New("missing host guid")
	}
	pageInfo := commonUtils.ToPageInfo(params)
	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.Size <= 0 {
		pageInfo.Size = 20
	}
	if pageInfo.Size > 200 {
		pageInfo.Size = 200
	}
	keyword := strings.TrimSpace(params["keyword"])
	if keyword == "" {
		keyword = strings.TrimSpace(params["content"])
	}

	db := s.ContainerService.DB()
	if db == nil {
		return nil, 0, errors.New("database is not initialized")
	}
	query := db.Model(&domains.DockerContainer{}).Where("host_guid = ?", hostGuid)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("container_id LIKE ? OR names LIKE ? OR image LIKE ? OR state LIKE ? OR status LIKE ?", like, like, like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domains.DockerContainer
	err := query.Order("id desc").Limit(pageInfo.Size).Offset(pageInfo.Size * (pageInfo.Page - 1)).Find(&items).Error
	return items, total, err
}

func (s DockerResourceService) ListImages(hostGuid string, params map[string]string) (interface{}, int64, error) {
	hostGuid = strings.TrimSpace(hostGuid)
	if hostGuid == "" {
		return nil, 0, errors.New("missing host guid")
	}
	pageInfo := commonUtils.ToPageInfo(params)
	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.Size <= 0 {
		pageInfo.Size = 20
	}
	if pageInfo.Size > 200 {
		pageInfo.Size = 200
	}
	keyword := strings.TrimSpace(params["keyword"])
	if keyword == "" {
		keyword = strings.TrimSpace(params["content"])
	}

	db := s.ImageService.DB()
	if db == nil {
		return nil, 0, errors.New("database is not initialized")
	}
	query := db.Model(&domains.DockerImage{}).Where("host_guid = ?", hostGuid)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("image_id LIKE ? OR repo_tags LIKE ? OR repo_digests LIKE ?", like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domains.DockerImage
	err := query.Order("id desc").Limit(pageInfo.Size).Offset(pageInfo.Size * (pageInfo.Page - 1)).Find(&items).Error
	return items, total, err
}

func (s DockerResourceService) SyncContainers(hostGuid string, containers []domains.DockerContainerSnapshot) error {
	hostGuid = strings.TrimSpace(hostGuid)
	if hostGuid == "" {
		return errors.New("missing host guid")
	}
	db := s.ContainerService.DB()
	if db == nil {
		return errors.New("database is not initialized")
	}
	now := time.Now().UnixMilli()
	return db.Transaction(func(tx *gorm.DB) error {
		for _, item := range containers {
			item.ContainerID = strings.TrimSpace(item.ContainerID)
			if item.ContainerID == "" {
				continue
			}
			names, err := marshalJSONText(item.Names)
			if err != nil {
				return err
			}
			ports, err := marshalJSONText(item.Ports)
			if err != nil {
				return err
			}
			labels, err := marshalJSONText(item.Labels)
			if err != nil {
				return err
			}
			entity := domains.DockerContainer{
				HostGuid:    hostGuid,
				ContainerID: item.ContainerID,
				Names:       names,
				Image:       strings.TrimSpace(item.Image),
				ImageID:     strings.TrimSpace(item.ImageID),
				Command:     item.Command,
				State:       strings.TrimSpace(item.State),
				Status:      strings.TrimSpace(item.Status),
				Ports:       ports,
				Labels:      labels,
				CreatedAt:   item.CreatedAt,
				SyncedAt:    now,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "host_guid"}, {Name: "container_id"}},
				DoUpdates: clause.Assignments(map[string]any{
					"names":       entity.Names,
					"image":       entity.Image,
					"image_id":    entity.ImageID,
					"command":     entity.Command,
					"state":       entity.State,
					"status":      entity.Status,
					"ports":       entity.Ports,
					"labels":      entity.Labels,
					"created_at":  entity.CreatedAt,
					"synced_at":   entity.SyncedAt,
					"update_time": now,
				}),
			}).Create(&entity).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s DockerResourceService) SyncImages(hostGuid string, images []domains.DockerImageSnapshot) error {
	hostGuid = strings.TrimSpace(hostGuid)
	if hostGuid == "" {
		return errors.New("missing host guid")
	}
	db := s.ImageService.DB()
	if db == nil {
		return errors.New("database is not initialized")
	}
	now := time.Now().UnixMilli()
	return db.Transaction(func(tx *gorm.DB) error {
		for _, item := range images {
			item.ImageID = strings.TrimSpace(item.ImageID)
			if item.ImageID == "" {
				continue
			}
			repoTags, err := marshalJSONText(item.RepoTags)
			if err != nil {
				return err
			}
			repoDigests, err := marshalJSONText(item.RepoDigests)
			if err != nil {
				return err
			}
			labels, err := marshalJSONText(item.Labels)
			if err != nil {
				return err
			}
			entity := domains.DockerImage{
				HostGuid:    hostGuid,
				ImageID:     item.ImageID,
				RepoTags:    repoTags,
				RepoDigests: repoDigests,
				Labels:      labels,
				Size:        item.Size,
				Containers:  item.Containers,
				CreatedAt:   item.CreatedAt,
				SyncedAt:    now,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "host_guid"}, {Name: "image_id"}},
				DoUpdates: clause.Assignments(map[string]any{
					"repo_tags":    entity.RepoTags,
					"repo_digests": entity.RepoDigests,
					"labels":       entity.Labels,
					"size":         entity.Size,
					"containers":   entity.Containers,
					"created_at":   entity.CreatedAt,
					"synced_at":    entity.SyncedAt,
					"update_time":  now,
				}),
			}).Create(&entity).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s DockerResourceService) SyncFromTaskResult(task domains.Task, result string) error {
	if task.Status != domains.TaskStatusSuccess {
		return nil
	}
	switch task.Type {
	case domains.TaskTypeDockerContainerList:
		var snapshot domains.DockerResourceSnapshot
		if err := json.Unmarshal([]byte(result), &snapshot); err != nil {
			return err
		}
		return s.SyncContainers(task.HostGuid, snapshot.Containers)
	case domains.TaskTypeDockerImageList:
		var snapshot domains.DockerResourceSnapshot
		if err := json.Unmarshal([]byte(result), &snapshot); err != nil {
			return err
		}
		return s.SyncImages(task.HostGuid, snapshot.Images)
	default:
		return nil
	}
}

func marshalJSONText(value any) (string, error) {
	if value == nil {
		return "null", nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
