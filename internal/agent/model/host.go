package model

type HostMetrics struct {
	Hostname             string `json:"hostname"`
	Uptime               uint64 `json:"uptime"`
	Procs                uint64 `json:"procs"`
	OS                   string `json:"os"`
	Platform             string `json:"platform"`
	PlatformFamily       string `json:"platform_family"`
	PlatformVersion      string `json:"platform_version"`
	VirtualizationSystem string `json:"virtualization_system"`
	VirtualizationRole   string `json:"virtualization_role"`
	KernelVersion        string `json:"kernel_version"`
}
