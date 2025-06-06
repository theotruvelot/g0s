package metric

type NetworkMetrics struct {
	InterfaceName string `json:"interface_name"`
	BytesSent     uint64 `json:"bytes_sent"`
	BytesRecv     uint64 `json:"bytes_recv"`
	PacketsSent   uint64 `json:"packets_sent"`
	PacketsRecv   uint64 `json:"packets_recv"`
	ErrIn         uint64 `json:"err_in"`
	ErrOut        uint64 `json:"err_out"`
}
