package collector

import (
	"github.com/shirou/gopsutil/v4/net"
	"github.com/theotruvelot/g0s/internal/agent/model/metric"
	"go.uber.org/zap"
)

// NetworkCollector collects network interface statistics and metrics.
type NetworkCollector struct {
	log *zap.Logger
}

// NewNetworkCollector creates a new NetworkCollector instance.
func NewNetworkCollector(log *zap.Logger) *NetworkCollector {
	return &NetworkCollector{
		log: log,
	}
}

// Collect gathers network interface statistics including bytes and packets transferred.
func (c *NetworkCollector) Collect() ([]metric.NetworkMetrics, error) {
	netStats, err := net.IOCounters(true)
	if err != nil {
		c.log.Error("Failed to collect network interface statistics", zap.Error(err))
		return nil, err
	}

	return c.buildNetworkMetrics(netStats), nil
}

// buildNetworkMetrics constructs network metrics from collected data.
func (c *NetworkCollector) buildNetworkMetrics(netStats []net.IOCountersStat) []metric.NetworkMetrics {
	var metrics []metric.NetworkMetrics

	for _, iface := range netStats {
		metrics = append(metrics, metric.NetworkMetrics{
			InterfaceName: iface.Name,
			BytesSent:     iface.BytesSent,
			BytesRecv:     iface.BytesRecv,
			PacketsSent:   iface.PacketsSent,
			PacketsRecv:   iface.PacketsRecv,
			ErrIn:         iface.Errin,
			ErrOut:        iface.Errout,
		})
	}

	return metrics
}
