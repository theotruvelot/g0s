package metric

import "time"

type MetricsPayload struct {
	Host      HostMetrics      `json:"host"`
	CPU       []CPUMetrics     `json:"cpu"`
	RAM       RamMetrics       `json:"ram"`
	Disk      []DiskMetrics    `json:"disk"`
	Network   []NetworkMetrics `json:"network"`
	Docker    []DockerMetrics  `json:"docker"`
	Timestamp time.Time        `json:"timestamp"`
}
