package collector

import (
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/theotruvelot/g0s/internal/agent/model/metric"
	"go.uber.org/zap"
)

// RAMCollector collects memory usage statistics including virtual and swap memory.
type RAMCollector struct {
	log *zap.Logger
}

// NewRAMCollector creates a new RAMCollector instance.
func NewRAMCollector(log *zap.Logger) *RAMCollector {
	return &RAMCollector{
		log: log,
	}
}

// Collect gathers memory statistics including virtual memory and swap usage.
func (c *RAMCollector) Collect() (metric.RamMetrics, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		c.log.Error("Failed to collect virtual memory statistics", zap.Error(err))
		return metric.RamMetrics{}, err
	}

	sm, err := mem.SwapMemory()
	if err != nil {
		c.log.Error("Failed to collect swap memory statistics", zap.Error(err))
		return metric.RamMetrics{}, err
	}

	return c.buildRAMMetrics(vm, sm), nil
}

// buildRAMMetrics constructs RAM metrics from collected data.
func (c *RAMCollector) buildRAMMetrics(vm *mem.VirtualMemoryStat, sm *mem.SwapMemoryStat) metric.RamMetrics {
	return metric.RamMetrics{
		TotalOctets:     vm.Total,
		UsedOctets:      vm.Used,
		FreeOctets:      vm.Free,
		AvailableOctets: vm.Available,
		UsedPercent:     vm.UsedPercent,
		SwapTotalOctets: sm.Total,
		SwapUsedOctets:  sm.Used,
		SwapUsedPerc:    sm.UsedPercent,
	}
}
