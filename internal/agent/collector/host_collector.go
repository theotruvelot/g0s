package collector

import (
	"github.com/shirou/gopsutil/v4/host"
	"github.com/theotruvelot/g0s/internal/agent/model/metric"
	"go.uber.org/zap"
)

// HostCollector collects host system information and statistics.
type HostCollector struct {
	log *zap.Logger
}

// NewHostCollector creates a new HostCollector instance.
func NewHostCollector(log *zap.Logger) *HostCollector {
	return &HostCollector{
		log: log,
	}
}

// Collect gathers host system information including OS details, uptime, and virtualization info.
func (c *HostCollector) Collect() (metric.HostMetrics, error) {
	hostInfo, err := host.Info()
	if err != nil {
		c.log.Error("Failed to collect host information", zap.Error(err))
		return metric.HostMetrics{}, err
	}

	return c.buildHostMetrics(hostInfo), nil
}

// buildHostMetrics constructs host metrics from collected data.
func (c *HostCollector) buildHostMetrics(hostInfo *host.InfoStat) metric.HostMetrics {
	return metric.HostMetrics{
		Hostname:             hostInfo.Hostname,
		Uptime:               hostInfo.Uptime,
		Procs:                hostInfo.Procs,
		OS:                   hostInfo.OS,
		Platform:             hostInfo.Platform,
		PlatformFamily:       hostInfo.PlatformFamily,
		PlatformVersion:      hostInfo.PlatformVersion,
		VirtualizationSystem: hostInfo.VirtualizationSystem,
		VirtualizationRole:   hostInfo.VirtualizationRole,
		KernelVersion:        hostInfo.KernelVersion,
	}
}
