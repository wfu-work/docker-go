package domains

import commonDomains "github.com/wfu-work/nav-common-go-lib/domains"

type HostMetric struct {
	commonDomains.BaseDataEntity
	HostGuid       string  `gorm:"column:host_guid;size:50;index" json:"hostGuid"`
	CPUPercent     float64 `gorm:"column:cpu_percent" json:"cpuPercent"`
	CPUCores       int     `gorm:"column:cpu_cores" json:"cpuCores"`
	MemoryTotal    uint64  `gorm:"column:memory_total" json:"memoryTotal"`
	MemoryUsed     uint64  `gorm:"column:memory_used" json:"memoryUsed"`
	MemoryPercent  float64 `gorm:"column:memory_percent" json:"memoryPercent"`
	DiskTotal      uint64  `gorm:"column:disk_total" json:"diskTotal"`
	DiskUsed       uint64  `gorm:"column:disk_used" json:"diskUsed"`
	DiskPercent    float64 `gorm:"column:disk_percent" json:"diskPercent"`
	NetworkRxBytes uint64  `gorm:"column:network_rx_bytes" json:"networkRxBytes"`
	NetworkTxBytes uint64  `gorm:"column:network_tx_bytes" json:"networkTxBytes"`
	CollectedAt    int64   `gorm:"column:collected_at;index" json:"collectedAt"`
}

func (HostMetric) TableName() string {
	return "nav_docker_host_metrics"
}

type ContainerMetric struct {
	commonDomains.BaseDataEntity
	HostGuid        string  `gorm:"column:host_guid;size:50;index" json:"hostGuid"`
	ContainerID     string  `gorm:"column:container_id;size:100;index" json:"containerId"`
	Name            string  `gorm:"column:name;size:255" json:"name"`
	CPUPercent      float64 `gorm:"column:cpu_percent" json:"cpuPercent"`
	MemoryUsage     uint64  `gorm:"column:memory_usage" json:"memoryUsage"`
	MemoryLimit     uint64  `gorm:"column:memory_limit" json:"memoryLimit"`
	MemoryPercent   float64 `gorm:"column:memory_percent" json:"memoryPercent"`
	NetworkRxBytes  uint64  `gorm:"column:network_rx_bytes" json:"networkRxBytes"`
	NetworkTxBytes  uint64  `gorm:"column:network_tx_bytes" json:"networkTxBytes"`
	BlockReadBytes  uint64  `gorm:"column:block_read_bytes" json:"blockReadBytes"`
	BlockWriteBytes uint64  `gorm:"column:block_write_bytes" json:"blockWriteBytes"`
	PidsCurrent     uint64  `gorm:"column:pids_current" json:"pidsCurrent"`
	CollectedAt     int64   `gorm:"column:collected_at;index" json:"collectedAt"`
}

func (ContainerMetric) TableName() string {
	return "nav_docker_container_metrics"
}

type HostMetricSnapshot struct {
	CPUPercent     float64 `json:"cpuPercent"`
	CPUCores       int     `json:"cpuCores"`
	MemoryTotal    uint64  `json:"memoryTotal"`
	MemoryUsed     uint64  `json:"memoryUsed"`
	MemoryPercent  float64 `json:"memoryPercent"`
	DiskTotal      uint64  `json:"diskTotal"`
	DiskUsed       uint64  `json:"diskUsed"`
	DiskPercent    float64 `json:"diskPercent"`
	NetworkRxBytes uint64  `json:"networkRxBytes"`
	NetworkTxBytes uint64  `json:"networkTxBytes"`
	CollectedAt    int64   `json:"collectedAt"`
}

type ContainerMetricSnapshot struct {
	ContainerID     string  `json:"containerId"`
	Name            string  `json:"name"`
	CPUPercent      float64 `json:"cpuPercent"`
	MemoryUsage     uint64  `json:"memoryUsage"`
	MemoryLimit     uint64  `json:"memoryLimit"`
	MemoryPercent   float64 `json:"memoryPercent"`
	NetworkRxBytes  uint64  `json:"networkRxBytes"`
	NetworkTxBytes  uint64  `json:"networkTxBytes"`
	BlockReadBytes  uint64  `json:"blockReadBytes"`
	BlockWriteBytes uint64  `json:"blockWriteBytes"`
	PidsCurrent     uint64  `json:"pidsCurrent"`
	CollectedAt     int64   `json:"collectedAt"`
}

type MetricsSnapshot struct {
	Host       HostMetricSnapshot        `json:"host"`
	Containers []ContainerMetricSnapshot `json:"containers"`
}
