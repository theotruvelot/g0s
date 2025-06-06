package collector

import (
	"testing"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewDiskCollector(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewDiskCollector(logger)

	assert.NotNil(t, collector)
	assert.Equal(t, logger, collector.log)
}

func TestDiskCollector_Collect(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewDiskCollector(logger)

	metrics, err := collector.Collect()
	require.NoError(t, err)
	require.NotNil(t, metrics)
	require.NotEmpty(t, metrics)

	// Verify that each metric has the expected fields populated
	for _, m := range metrics {
		assert.NotEmpty(t, m.Path)
		assert.NotEmpty(t, m.Device)
		assert.NotEmpty(t, m.Fstype)
		assert.GreaterOrEqual(t, m.TotalOctets, uint64(0))
		assert.GreaterOrEqual(t, m.UsedOctets, uint64(0))
		assert.GreaterOrEqual(t, m.FreeOctets, uint64(0))
		assert.GreaterOrEqual(t, m.UsedPercent, float64(0))
	}
}

func TestDiskCollector_buildDiskMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewDiskCollector(logger)

	testCases := []struct {
		name      string
		usage     *disk.UsageStat
		ioStats   map[string]disk.IOCountersStat
		partition disk.PartitionStat
	}{
		{
			name: "With IO Stats",
			usage: &disk.UsageStat{
				Path:        "/test",
				Total:       1000000,
				Used:        500000,
				Free:        500000,
				UsedPercent: 50.0,
			},
			ioStats: map[string]disk.IOCountersStat{
				"test": {
					ReadCount:  100,
					WriteCount: 200,
					ReadBytes:  1000,
					WriteBytes: 2000,
				},
			},
			partition: disk.PartitionStat{
				Device:     "/dev/test",
				Mountpoint: "/test",
				Fstype:     "ext4",
			},
		},
		{
			name: "Without IO Stats",
			usage: &disk.UsageStat{
				Path:        "/test2",
				Total:       2000000,
				Used:        1000000,
				Free:        1000000,
				UsedPercent: 50.0,
			},
			ioStats: map[string]disk.IOCountersStat{},
			partition: disk.PartitionStat{
				Device:     "/dev/test2",
				Mountpoint: "/test2",
				Fstype:     "apfs",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metric := collector.buildDiskMetrics(tc.usage, tc.ioStats, tc.partition)

			assert.Equal(t, tc.usage.Path, metric.Path)
			assert.Equal(t, tc.partition.Device, metric.Device)
			assert.Equal(t, tc.partition.Fstype, metric.Fstype)
			assert.Equal(t, tc.usage.Total, metric.TotalOctets)
			assert.Equal(t, tc.usage.Used, metric.UsedOctets)
			assert.Equal(t, tc.usage.Free, metric.FreeOctets)
			assert.Equal(t, tc.usage.UsedPercent, metric.UsedPercent)

			deviceName := tc.partition.Device[5:] // Remove "/dev/" prefix
			if ioStat, exists := tc.ioStats[deviceName]; exists {
				assert.Equal(t, ioStat.ReadCount, metric.ReadCount)
				assert.Equal(t, ioStat.WriteCount, metric.WriteCount)
				assert.Equal(t, ioStat.ReadBytes, metric.ReadOctets)
				assert.Equal(t, ioStat.WriteBytes, metric.WriteOctets)
			} else {
				assert.Zero(t, metric.ReadCount)
				assert.Zero(t, metric.WriteCount)
				assert.Zero(t, metric.ReadOctets)
				assert.Zero(t, metric.WriteOctets)
			}
		})
	}
}

func TestDiskCollector_collectPartitionMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewDiskCollector(logger)

	// Test with a real partition that exists on the system
	partitions, err := disk.Partitions(false)
	require.NoError(t, err)
	require.NotEmpty(t, partitions)

	// Test with a valid partition
	t.Run("Valid Partition", func(t *testing.T) {
		metrics, err := collector.collectPartitionMetrics(partitions[0])
		require.NoError(t, err)
		assert.NotEmpty(t, metrics.Path)
		assert.NotEmpty(t, metrics.Device)
		assert.NotEmpty(t, metrics.Fstype)
		assert.GreaterOrEqual(t, metrics.TotalOctets, uint64(0))
	})

	// Test with an invalid partition
	t.Run("Invalid Partition", func(t *testing.T) {
		invalidPartition := disk.PartitionStat{
			Device:     "/dev/invalid",
			Mountpoint: "/invalid",
			Fstype:     "invalid",
			Opts:       []string{},
		}

		metrics, err := collector.collectPartitionMetrics(invalidPartition)
		assert.Error(t, err)
		assert.Empty(t, metrics.Path)
	})
}
