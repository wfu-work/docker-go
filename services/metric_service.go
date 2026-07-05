package services

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"docker-go/domains"

	commonServices "github.com/wfu-work/nav-common-go-lib/services"
	"gorm.io/gorm"
)

type HostMetricService struct {
	commonServices.CrudService[domains.HostMetric]
}

type ContainerMetricService struct {
	commonServices.CrudService[domains.ContainerMetric]
}

type MetricService struct {
	HostMetricService      HostMetricService
	ContainerMetricService ContainerMetricService
}

type HostOverview struct {
	Host             *domains.Host             `json:"host"`
	LatestMetric     *domains.HostMetric       `json:"latestMetric"`
	ContainerMetrics []domains.ContainerMetric `json:"containerMetrics"`
	ContainerTotal   int64                     `json:"containerTotal"`
	ContainerRunning int64                     `json:"containerRunning"`
	ImageTotal       int64                     `json:"imageTotal"`
}

var MetricServiceApp = new(MetricService)

func (s MetricService) SyncSnapshot(hostGuid string, snapshot domains.MetricsSnapshot) error {
	hostGuid = strings.TrimSpace(hostGuid)
	if hostGuid == "" {
		return errors.New("missing host guid")
	}
	db := s.HostMetricService.DB()
	if db == nil {
		return errors.New("database is not initialized")
	}
	now := time.Now().UnixMilli()
	if snapshot.Host.CollectedAt == 0 {
		snapshot.Host.CollectedAt = now
	}
	return db.Transaction(func(tx *gorm.DB) error {
		hostMetric := domains.HostMetric{
			HostGuid:       hostGuid,
			CPUPercent:     snapshot.Host.CPUPercent,
			CPUCores:       snapshot.Host.CPUCores,
			MemoryTotal:    snapshot.Host.MemoryTotal,
			MemoryUsed:     snapshot.Host.MemoryUsed,
			MemoryPercent:  snapshot.Host.MemoryPercent,
			DiskTotal:      snapshot.Host.DiskTotal,
			DiskUsed:       snapshot.Host.DiskUsed,
			DiskPercent:    snapshot.Host.DiskPercent,
			NetworkRxBytes: snapshot.Host.NetworkRxBytes,
			NetworkTxBytes: snapshot.Host.NetworkTxBytes,
			CollectedAt:    snapshot.Host.CollectedAt,
		}
		if err := tx.Create(&hostMetric).Error; err != nil {
			return err
		}
		for _, item := range snapshot.Containers {
			item.ContainerID = strings.TrimSpace(item.ContainerID)
			if item.ContainerID == "" {
				continue
			}
			if item.CollectedAt == 0 {
				item.CollectedAt = snapshot.Host.CollectedAt
			}
			containerMetric := domains.ContainerMetric{
				HostGuid:        hostGuid,
				ContainerID:     item.ContainerID,
				Name:            item.Name,
				CPUPercent:      item.CPUPercent,
				MemoryUsage:     item.MemoryUsage,
				MemoryLimit:     item.MemoryLimit,
				MemoryPercent:   item.MemoryPercent,
				NetworkRxBytes:  item.NetworkRxBytes,
				NetworkTxBytes:  item.NetworkTxBytes,
				BlockReadBytes:  item.BlockReadBytes,
				BlockWriteBytes: item.BlockWriteBytes,
				PidsCurrent:     item.PidsCurrent,
				CollectedAt:     item.CollectedAt,
			}
			if err := tx.Create(&containerMetric).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s MetricService) SyncFromTaskResult(task domains.Task, result string) error {
	if task.Status != domains.TaskStatusSuccess || task.Type != domains.TaskTypeDockerMetricsSnapshot {
		return nil
	}
	var snapshot domains.MetricsSnapshot
	if err := json.Unmarshal([]byte(result), &snapshot); err != nil {
		return err
	}
	return s.SyncSnapshot(task.HostGuid, snapshot)
}

func (s MetricService) LatestHostMetric(hostGuid string) (*domains.HostMetric, error) {
	hostGuid = strings.TrimSpace(hostGuid)
	if hostGuid == "" {
		return nil, errors.New("missing host guid")
	}
	db := s.HostMetricService.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	var metric domains.HostMetric
	err := db.Where("host_guid = ?", hostGuid).Order("collected_at desc, id desc").First(&metric).Error
	if err != nil {
		return nil, err
	}
	return &metric, nil
}

func (s MetricService) LatestContainerMetrics(hostGuid string) ([]domains.ContainerMetric, error) {
	hostGuid = strings.TrimSpace(hostGuid)
	if hostGuid == "" {
		return nil, errors.New("missing host guid")
	}
	db := s.ContainerMetricService.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	var metrics []domains.ContainerMetric
	err := db.Raw(`
		SELECT m.*
		FROM nav_docker_container_metrics m
		INNER JOIN (
			SELECT container_id, MAX(collected_at) AS collected_at
			FROM nav_docker_container_metrics
			WHERE host_guid = ?
			GROUP BY container_id
		) latest ON latest.container_id = m.container_id AND latest.collected_at = m.collected_at
		WHERE m.host_guid = ?
		ORDER BY m.name ASC, m.container_id ASC
	`, hostGuid, hostGuid).Scan(&metrics).Error
	return metrics, err
}

func (s MetricService) Overview(hostGuid string) (*HostOverview, error) {
	host, err := HostServiceApp.Get(hostGuid)
	if err != nil {
		return nil, err
	}
	db := s.HostMetricService.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	overview := &HostOverview{Host: host}
	overview.LatestMetric, _ = s.LatestHostMetric(hostGuid)
	overview.ContainerMetrics, _ = s.LatestContainerMetrics(hostGuid)
	_ = db.Model(&domains.DockerContainer{}).Where("host_guid = ?", hostGuid).Count(&overview.ContainerTotal).Error
	_ = db.Model(&domains.DockerContainer{}).Where("host_guid = ? AND state = ?", hostGuid, "running").Count(&overview.ContainerRunning).Error
	_ = db.Model(&domains.DockerImage{}).Where("host_guid = ?", hostGuid).Count(&overview.ImageTotal).Error
	return overview, nil
}
