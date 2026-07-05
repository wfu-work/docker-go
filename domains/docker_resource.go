package domains

import commonDomains "github.com/wfu-work/nav-common-go-lib/domains"

type DockerContainer struct {
	commonDomains.BaseDataEntity
	HostGuid    string `gorm:"column:host_guid;size:50;uniqueIndex:idx_host_container;index" json:"hostGuid"`
	ContainerID string `gorm:"column:container_id;size:100;uniqueIndex:idx_host_container" json:"containerId"`
	Names       string `gorm:"column:names;type:text" json:"names"`
	Image       string `gorm:"column:image;size:255;index" json:"image"`
	ImageID     string `gorm:"column:image_id;size:150" json:"imageId"`
	Command     string `gorm:"column:command;size:1000" json:"command"`
	State       string `gorm:"column:state;size:50;index" json:"state"`
	Status      string `gorm:"column:status;size:255" json:"status"`
	Ports       string `gorm:"column:ports;type:text" json:"ports"`
	Labels      string `gorm:"column:labels;type:text" json:"labels"`
	CreatedAt   int64  `gorm:"column:created_at" json:"createdAt"`
	SyncedAt    int64  `gorm:"column:synced_at;index" json:"syncedAt"`
}

func (DockerContainer) TableName() string {
	return "nav_docker_containers"
}

type DockerImage struct {
	commonDomains.BaseDataEntity
	HostGuid    string `gorm:"column:host_guid;size:50;uniqueIndex:idx_host_image;index" json:"hostGuid"`
	ImageID     string `gorm:"column:image_id;size:150;uniqueIndex:idx_host_image" json:"imageId"`
	RepoTags    string `gorm:"column:repo_tags;type:text" json:"repoTags"`
	RepoDigests string `gorm:"column:repo_digests;type:text" json:"repoDigests"`
	Labels      string `gorm:"column:labels;type:text" json:"labels"`
	Size        int64  `gorm:"column:size" json:"size"`
	Containers  int64  `gorm:"column:containers" json:"containers"`
	CreatedAt   int64  `gorm:"column:created_at" json:"createdAt"`
	SyncedAt    int64  `gorm:"column:synced_at;index" json:"syncedAt"`
}

func (DockerImage) TableName() string {
	return "nav_docker_images"
}

type DockerContainerSnapshot struct {
	ContainerID string            `json:"containerId"`
	Names       []string          `json:"names"`
	Image       string            `json:"image"`
	ImageID     string            `json:"imageId"`
	Command     string            `json:"command"`
	State       string            `json:"state"`
	Status      string            `json:"status"`
	Ports       any               `json:"ports"`
	Labels      map[string]string `json:"labels"`
	CreatedAt   int64             `json:"createdAt"`
}

type DockerImageSnapshot struct {
	ImageID     string            `json:"imageId"`
	RepoTags    []string          `json:"repoTags"`
	RepoDigests []string          `json:"repoDigests"`
	Labels      map[string]string `json:"labels"`
	Size        int64             `json:"size"`
	Containers  int64             `json:"containers"`
	CreatedAt   int64             `json:"createdAt"`
}

type DockerResourceSnapshot struct {
	Containers []DockerContainerSnapshot `json:"containers,omitempty"`
	Images     []DockerImageSnapshot     `json:"images,omitempty"`
}
