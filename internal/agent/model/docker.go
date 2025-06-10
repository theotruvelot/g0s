package model

type DockerMetrics struct {
	ContainerID    string         `json:"container_id"`
	ContainerName  string         `json:"container_name"`
	Image          string         `json:"image"`
	ImageID        string         `json:"image_id"`
	ImageName      string         `json:"image_name"`
	ImageTag       string         `json:"image_tag"`
	ImageDigest    string         `json:"image_digest"`
	ImageSize      string         `json:"image_size"`
	CPUMetrics     CPUMetrics     `json:"cpu_metrics"`
	RAMMetrics     RamMetrics     `json:"ram_metrics"`
	DiskMetrics    DiskMetrics    `json:"disk_metrics"`
	NetworkMetrics NetworkMetrics `json:"network_metrics"`
}
