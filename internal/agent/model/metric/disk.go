package metric

type DiskMetrics struct {
	Path        string  `json:"path"`
	Device      string  `json:"device"`
	Fstype      string  `json:"fstype"`
	TotalOctets uint64  `json:"total"`
	UsedOctets  uint64  `json:"used"`
	FreeOctets  uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
	ReadCount   uint64  `json:"read_count"`
	WriteCount  uint64  `json:"write_count"`
	ReadOctets  uint64  `json:"read_octets"`
	WriteOctets uint64  `json:"write_octets"`
}
