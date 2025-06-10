package model

type RamMetrics struct {
	TotalOctets     uint64  `json:"total_octets"`
	UsedOctets      uint64  `json:"used_octets"`
	FreeOctets      uint64  `json:"free_octets"`
	UsedPercent     float64 `json:"used_percent"`
	AvailableOctets uint64  `json:"available_octets"`
	SwapTotalOctets uint64  `json:"swap_total_octets"`
	SwapUsedOctets  uint64  `json:"swap_used_octets"`
	SwapUsedPerc    float64 `json:"swap_used_percent"`
}
