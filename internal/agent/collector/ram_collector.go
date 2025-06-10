package collector

import (
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/theotruvelot/g0s/internal/agent/model"
	"go.uber.org/zap"
)

type RAMCollector struct {
	log *zap.Logger
}

func NewRAMCollector(log *zap.Logger) *RAMCollector {
	return &RAMCollector{
		log: log,
	}
}

func (c *RAMCollector) Collect() (model.RamMetrics, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		c.log.Error("Failed to collect virtual memory statistics", zap.Error(err))
		return model.RamMetrics{}, err
	}

	sm, err := mem.SwapMemory()
	if err != nil {
		c.log.Error("Failed to collect swap memory statistics", zap.Error(err))
		return model.RamMetrics{}, err
	}

	return c.buildRAMMetrics(vm, sm), nil
}

func (c *RAMCollector) buildRAMMetrics(vm *mem.VirtualMemoryStat, sm *mem.SwapMemoryStat) model.RamMetrics {
	return model.RamMetrics{
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
