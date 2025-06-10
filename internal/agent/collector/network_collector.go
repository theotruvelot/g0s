package collector

import (
	"github.com/shirou/gopsutil/v4/net"
	"github.com/theotruvelot/g0s/internal/agent/model"
	"go.uber.org/zap"
)

type NetworkCollector struct {
	log *zap.Logger
}

func NewNetworkCollector(log *zap.Logger) *NetworkCollector {
	return &NetworkCollector{
		log: log,
	}
}

func (c *NetworkCollector) Collect() ([]model.NetworkMetrics, error) {
	netStats, err := net.IOCounters(true)
	if err != nil {
		c.log.Error("Failed to collect network interface statistics", zap.Error(err))
		return nil, err
	}

	return c.buildNetworkMetrics(netStats), nil
}

func (c *NetworkCollector) buildNetworkMetrics(netStats []net.IOCountersStat) []model.NetworkMetrics {
	metrics := make([]model.NetworkMetrics, 0, len(netStats))

	for _, iface := range netStats {
		metrics = append(metrics, model.NetworkMetrics{
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
