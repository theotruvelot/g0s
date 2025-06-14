package collector

import (
	"fmt"
	"github.com/theotruvelot/g0s/internal/agent/model"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"go.uber.org/zap"
)

type CPUCollector struct {
	log             *zap.Logger
	lastCPUTimes    []cpu.TimesStat
	lastTotalTimes  cpu.TimesStat
	lastCollectTime time.Time

	mu             sync.RWMutex
	cachedCPUInfo  []cpu.InfoStat
	cachedLogical  int
	cachedPhysical int
	cacheExpiry    time.Time
	cacheDuration  time.Duration
}

func NewCPUCollector(log *zap.Logger) *CPUCollector {
	return &CPUCollector{
		log:           log,
		cacheDuration: 5 * time.Minute,
	}
}

func (c *CPUCollector) getCachedStaticData() ([]cpu.InfoStat, int, int, error) {
	c.mu.RLock()
	if time.Now().Before(c.cacheExpiry) && c.cachedCPUInfo != nil {
		info := c.cachedCPUInfo
		logical := c.cachedLogical
		physical := c.cachedPhysical
		c.mu.RUnlock()
		return info, logical, physical, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().Before(c.cacheExpiry) && c.cachedCPUInfo != nil {
		return c.cachedCPUInfo, c.cachedLogical, c.cachedPhysical, nil
	}

	info, err := cpu.Info()
	if err != nil {
		c.log.Error("Failed to collect CPU information", zap.Error(err))
		return nil, 0, 0, err
	}

	logicalCount, err := cpu.Counts(true)
	if err != nil {
		c.log.Error("Failed to get logical CPU count", zap.Error(err))
		return nil, 0, 0, err
	}

	physicalCount, err := cpu.Counts(false)
	if err != nil {
		c.log.Error("Failed to get physical CPU count", zap.Error(err))
		return nil, 0, 0, err
	}

	c.cachedCPUInfo = info
	c.cachedLogical = logicalCount
	c.cachedPhysical = physicalCount
	c.cacheExpiry = time.Now().Add(c.cacheDuration)

	return info, logicalCount, physicalCount, nil
}

func (c *CPUCollector) Collect() ([]model.CPUMetrics, error) {
	cpuInfo, logicalCount, physicalCount, err := c.getCachedStaticData()
	if err != nil {
		return nil, err
	}

	totalTimes, err := cpu.Times(false)
	if err != nil || len(totalTimes) == 0 {
		c.log.Error("Failed to get total CPU times", zap.Error(err))
		return nil, err
	}
	now := time.Now()

	if c.lastCollectTime.IsZero() {
		c.lastTotalTimes = totalTimes[0]
		c.lastCollectTime = now

		currentCPUTimes, _ := cpu.Times(true)
		c.lastCPUTimes = currentCPUTimes

		c.log.Debug("Initial CPU sample collected, retrying after warm-up")
		time.Sleep(200 * time.Millisecond)
		return c.Collect()
	}

	timeDelta := now.Sub(c.lastCollectTime).Seconds()
	totalUsagePercent := 0.0
	if timeDelta > 0 && len(c.lastTotalTimes.CPU) > 0 {
		totalUsagePercent = calculateCPUUsagePercentage(totalTimes[0], c.lastTotalTimes, timeDelta)
	}

	currentCPUTimes, err := cpu.Times(true)
	var perCorePercentages []float64

	if err == nil && len(c.lastCPUTimes) > 0 {
		perCorePercentages = make([]float64, len(currentCPUTimes))
		for i := range currentCPUTimes {
			if i < len(c.lastCPUTimes) {
				perCorePercentages[i] = calculateCPUUsagePercentage(currentCPUTimes[i], c.lastCPUTimes[i], timeDelta)
			}
		}
	} else if err != nil {
		c.log.Warn("Failed to get per-core CPU times", zap.Error(err))
	}

	c.lastCPUTimes = currentCPUTimes
	c.lastTotalTimes = totalTimes[0]
	c.lastCollectTime = now

	return c.buildCPUMetrics(
		cpuInfo,
		perCorePercentages,
		totalUsagePercent,
		currentCPUTimes,
		physicalCount,
		logicalCount,
	), nil
}

func calculateCPUUsagePercentage(current, last cpu.TimesStat, timeDelta float64) float64 {
	totalDiff := (current.User + current.System + current.Nice + current.Iowait +
		current.Irq + current.Softirq + current.Steal) -
		(last.User + last.System + last.Nice + last.Iowait +
			last.Irq + last.Softirq + last.Steal)

	idleDiff := (current.Idle + current.Iowait) - (last.Idle + last.Iowait)
	totalTime := totalDiff + idleDiff

	if totalTime == 0 {
		return 0.0
	}
	return (totalTime - idleDiff) / totalTime * 100.0
}

func (c *CPUCollector) buildCPUMetrics(
	cpuInfo []cpu.InfoStat,
	perCorePercentages []float64,
	totalUsagePercent float64,
	cpuTimes []cpu.TimesStat,
	physicalCount, logicalCount int,
) []model.CPUMetrics {
	metrics := make([]model.CPUMetrics, 0, len(cpuInfo)+1)

	var defaultFrequencyMHz float64
	if len(cpuInfo) > 0 {
		defaultFrequencyMHz = float64(cpuInfo[0].Mhz)
	}

	totalMetric := model.CPUMetrics{
		Model:        cpuInfo[0].ModelName,
		Cores:        physicalCount,
		Threads:      logicalCount,
		FrequencyMHz: defaultFrequencyMHz,
		UsagePercent: totalUsagePercent,
		UserTime:     cpuTimes[0].User,
		SystemTime:   cpuTimes[0].System,
		IdleTime:     cpuTimes[0].Idle,
		IsTotal:      true,
	}
	metrics = append(metrics, totalMetric)

	for i := 0; i < physicalCount && i < len(cpuTimes); i++ {
		m := model.CPUMetrics{
			Model:        fmt.Sprintf("CPU %d", i+1),
			Cores:        1,
			Threads:      1,
			FrequencyMHz: defaultFrequencyMHz,
			UserTime:     cpuTimes[i].User,
			SystemTime:   cpuTimes[i].System,
			IdleTime:     cpuTimes[i].Idle,
			UsagePercent: 0.0,
			CoreID:       i + 1,
			IsTotal:      false,
		}
		if i < len(perCorePercentages) {
			m.UsagePercent = perCorePercentages[i]
		}
		metrics = append(metrics, m)
	}

	return metrics
}
