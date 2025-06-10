package collector

import (
	"github.com/theotruvelot/g0s/internal/agent/model"
	"strings"

	"github.com/shirou/gopsutil/v4/disk"
	"go.uber.org/zap"
)

type DiskCollector struct {
	log *zap.Logger
}

func NewDiskCollector(log *zap.Logger) *DiskCollector {
	return &DiskCollector{
		log: log,
	}
}

func (c *DiskCollector) isRelevantPartition(partition disk.PartitionStat) bool {
	// Skip virtual or system partitions
	if strings.HasPrefix(partition.Mountpoint, "/System/Volumes") ||
		strings.HasPrefix(partition.Mountpoint, "/dev") {
		return false
	}

	// Only include physical disks and user data partitions
	return partition.Fstype != "devfs" &&
		partition.Fstype != "autofs" &&
		partition.Fstype != "none"
}

// Collect gathers disk metrics including usage and I/O statistics for relevant mounted partitions.
func (c *DiskCollector) Collect() ([]model.DiskMetrics, error) {
	// Get all physical partitions (false means don't include virtual partitions)
	partitions, err := disk.Partitions(true)
	if err != nil {
		c.log.Error("Failed to get disk partitions", zap.Error(err))
		return nil, err
	}

	var metrics []model.DiskMetrics
	for _, partition := range partitions {
		// Skip irrelevant partitions
		if !c.isRelevantPartition(partition) {
			continue
		}

		diskMetric, err := c.collectPartitionMetrics(partition)
		if err != nil {
			c.log.Warn("Failed to collect metrics for partition",
				zap.String("mountpoint", partition.Mountpoint),
				zap.Error(err))
			continue
		}

		if diskMetric.TotalOctets > 0 {
			metrics = append(metrics, diskMetric)
		}
	}

	return metrics, nil
}

func (c *DiskCollector) collectPartitionMetrics(partition disk.PartitionStat) (model.DiskMetrics, error) {
	usage, err := disk.Usage(partition.Mountpoint)
	if err != nil {
		return model.DiskMetrics{}, err
	}

	ioStats, err := disk.IOCounters()
	if err != nil {
		c.log.Debug("Failed to collect disk IO stats",
			zap.String("device", partition.Device),
			zap.Error(err))
	}

	return c.buildDiskMetrics(usage, ioStats, partition), nil
}

func (c *DiskCollector) buildDiskMetrics(usage *disk.UsageStat, ioStats map[string]disk.IOCountersStat, partition disk.PartitionStat) model.DiskMetrics {
	diskMetrics := model.DiskMetrics{
		Path:        usage.Path,
		Device:      partition.Device,
		Fstype:      partition.Fstype,
		TotalOctets: usage.Total,
		UsedOctets:  usage.Used,
		FreeOctets:  usage.Free,
		UsedPercent: usage.UsedPercent,
	}

	deviceName := strings.TrimPrefix(partition.Device, "/dev/")
	if ioStat, exists := ioStats[deviceName]; exists {
		diskMetrics.ReadCount = ioStat.ReadCount
		diskMetrics.WriteCount = ioStat.WriteCount
		diskMetrics.ReadOctets = ioStat.ReadBytes
		diskMetrics.WriteOctets = ioStat.WriteBytes
	}

	return diskMetrics
}
