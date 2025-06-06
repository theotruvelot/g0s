package collector

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/theotruvelot/g0s/internal/agent/model/metric"
	"go.uber.org/zap"
)

// CPUCollector collects CPU usage and information metrics.
type CPUCollector struct {
	log *zap.Logger
}

// NewCPUCollector creates a new CPUCollector instance.
func NewCPUCollector(log *zap.Logger) *CPUCollector {
	return &CPUCollector{
		log: log,
	}
}

// Collect gathers CPU metrics including usage percentages, timing information, and hardware details.
func (c *CPUCollector) Collect() ([]metric.CPUMetrics, error) {
	percentages, err := cpu.Percent(0, true)
	if err != nil {
		c.log.Error("Failed to collect CPU usage percentages", zap.Error(err))
		return nil, err
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		c.log.Error("Failed to collect CPU information", zap.Error(err))
		return nil, err
	}

	logicalCount, err := cpu.Counts(true)
	if err != nil {
		c.log.Error("Failed to get logical CPU count", zap.Error(err))
		return nil, err
	}

	physicalCount, err := cpu.Counts(false)
	if err != nil {
		c.log.Error("Failed to get physical CPU count", zap.Error(err))
		return nil, err
	}

	cpuTimes, err := cpu.Times(true)
	if err != nil {
		c.log.Error("Failed to get CPU times", zap.Error(err))
		return nil, err
	}

	return c.buildCPUMetrics(cpuInfo, percentages, cpuTimes, physicalCount, logicalCount), nil
}

// buildCPUMetrics constructs CPU metrics from collected data.
func (c *CPUCollector) buildCPUMetrics(
	cpuInfo []cpu.InfoStat,
	percentages []float64,
	cpuTimes []cpu.TimesStat,
	physicalCount, logicalCount int,
) []metric.CPUMetrics {
	var metrics []metric.CPUMetrics

	for i := 0; i < len(cpuInfo); i++ {
		cpuMetric := metric.CPUMetrics{
			Model:        cpuInfo[i].ModelName,
			Cores:        physicalCount,
			Threads:      logicalCount,
			FrequencyMHz: float64(cpuInfo[i].Mhz),
		}

		// Safely assign usage percentage if available
		if i < len(percentages) {
			cpuMetric.UsagePercent = percentages[i]
		}

		// Safely assign CPU times if available
		if i < len(cpuTimes) {
			cpuMetric.UserTime = cpuTimes[i].User
			cpuMetric.SystemTime = cpuTimes[i].System
			cpuMetric.IdleTime = cpuTimes[i].Idle
		}

		metrics = append(metrics, cpuMetric)
	}

	return metrics
}
