package collector

import (
	"testing"

	"github.com/shirou/gopsutil/v4/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewRAMCollector(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewRAMCollector(logger)

	assert.NotNil(t, collector)
	assert.Equal(t, logger, collector.log)
}

func TestRAMCollector_Collect(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewRAMCollector(logger)

	metrics, err := collector.Collect()
	require.NoError(t, err)

	// Verify that the metrics have valid values
	assert.Greater(t, metrics.TotalOctets, uint64(0))
	assert.GreaterOrEqual(t, metrics.UsedOctets, uint64(0))
	assert.GreaterOrEqual(t, metrics.FreeOctets, uint64(0))
	assert.GreaterOrEqual(t, metrics.AvailableOctets, uint64(0))
	assert.GreaterOrEqual(t, metrics.UsedPercent, float64(0))
	assert.LessOrEqual(t, metrics.UsedPercent, float64(100))

	assert.GreaterOrEqual(t, metrics.SwapTotalOctets, uint64(0))
	assert.GreaterOrEqual(t, metrics.SwapUsedOctets, uint64(0))
	assert.GreaterOrEqual(t, metrics.SwapUsedPerc, float64(0))
	assert.LessOrEqual(t, metrics.SwapUsedPerc, float64(100))

	// Verify that the total memory is greater than or equal to used + free
	assert.GreaterOrEqual(t, metrics.TotalOctets, metrics.UsedOctets+metrics.FreeOctets)

	// Verify that the swap total is greater than or equal to swap used
	assert.GreaterOrEqual(t, metrics.SwapTotalOctets, metrics.SwapUsedOctets)
}

func TestRAMCollector_buildRAMMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewRAMCollector(logger)

	testCases := []struct {
		name string
		vm   *mem.VirtualMemoryStat
		sm   *mem.SwapMemoryStat
	}{
		{
			name: "Normal Memory Usage",
			vm: &mem.VirtualMemoryStat{
				Total:       16 * 1024 * 1024 * 1024, // 16GB
				Used:        8 * 1024 * 1024 * 1024,  // 8GB
				Free:        8 * 1024 * 1024 * 1024,  // 8GB
				Available:   10 * 1024 * 1024 * 1024, // 10GB
				UsedPercent: 50.0,
			},
			sm: &mem.SwapMemoryStat{
				Total:       8 * 1024 * 1024 * 1024, // 8GB
				Used:        1 * 1024 * 1024 * 1024, // 1GB
				UsedPercent: 12.5,
			},
		},
		{
			name: "High Memory Usage",
			vm: &mem.VirtualMemoryStat{
				Total:       8 * 1024 * 1024 * 1024, // 8GB
				Used:        7 * 1024 * 1024 * 1024, // 7GB
				Free:        1 * 1024 * 1024 * 1024, // 1GB
				Available:   1 * 1024 * 1024 * 1024, // 1GB
				UsedPercent: 87.5,
			},
			sm: &mem.SwapMemoryStat{
				Total:       4 * 1024 * 1024 * 1024, // 4GB
				Used:        3 * 1024 * 1024 * 1024, // 3GB
				UsedPercent: 75.0,
			},
		},
		{
			name: "Low Memory Usage",
			vm: &mem.VirtualMemoryStat{
				Total:       32 * 1024 * 1024 * 1024, // 32GB
				Used:        4 * 1024 * 1024 * 1024,  // 4GB
				Free:        28 * 1024 * 1024 * 1024, // 28GB
				Available:   28 * 1024 * 1024 * 1024, // 28GB
				UsedPercent: 12.5,
			},
			sm: &mem.SwapMemoryStat{
				Total:       16 * 1024 * 1024 * 1024, // 16GB
				Used:        0,                       // 0GB
				UsedPercent: 0.0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := collector.buildRAMMetrics(tc.vm, tc.sm)

			// Virtual Memory assertions
			assert.Equal(t, tc.vm.Total, metrics.TotalOctets)
			assert.Equal(t, tc.vm.Used, metrics.UsedOctets)
			assert.Equal(t, tc.vm.Free, metrics.FreeOctets)
			assert.Equal(t, tc.vm.Available, metrics.AvailableOctets)
			assert.Equal(t, tc.vm.UsedPercent, metrics.UsedPercent)

			// Swap Memory assertions
			assert.Equal(t, tc.sm.Total, metrics.SwapTotalOctets)
			assert.Equal(t, tc.sm.Used, metrics.SwapUsedOctets)
			assert.Equal(t, tc.sm.UsedPercent, metrics.SwapUsedPerc)

			// Additional validations
			assert.GreaterOrEqual(t, metrics.TotalOctets, metrics.UsedOctets+metrics.FreeOctets)
			assert.GreaterOrEqual(t, metrics.SwapTotalOctets, metrics.SwapUsedOctets)
			assert.GreaterOrEqual(t, float64(100), metrics.UsedPercent)
			assert.GreaterOrEqual(t, float64(100), metrics.SwapUsedPerc)
		})
	}
}

func TestRAMCollector_Collect_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t)
	collector := NewRAMCollector(logger)

	metrics, err := collector.Collect()

	if err != nil {
		t.Skipf("RAM collection not available on this system: %v", err)
	}

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, metrics.TotalOctets, uint64(0))
	assert.GreaterOrEqual(t, metrics.UsedPercent, float64(0))
	assert.LessOrEqual(t, metrics.UsedPercent, float64(100))
}
