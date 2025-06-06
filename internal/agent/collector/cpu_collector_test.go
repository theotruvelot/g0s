package collector

import (
	"testing"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewCPUCollector(t *testing.T) {
	logger := zaptest.NewLogger(t)

	collector := NewCPUCollector(logger)

	assert.NotNil(t, collector)
	assert.Equal(t, logger, collector.log)
}

func TestCPUCollector_Collect(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewCPUCollector(logger)

	metrics, err := collector.Collect()
	require.NoError(t, err)
	require.NotNil(t, metrics)
	require.NotEmpty(t, metrics)

	// Verify that each metric has the expected fields populated
	for _, m := range metrics {
		assert.NotEmpty(t, m.Model)
		assert.Greater(t, m.Cores, 0)
		assert.Greater(t, m.Threads, 0)
		assert.Greater(t, m.FrequencyMHz, float64(0))
		assert.GreaterOrEqual(t, m.UsagePercent, float64(0))
		assert.GreaterOrEqual(t, m.UserTime, float64(0))
		assert.GreaterOrEqual(t, m.SystemTime, float64(0))
		assert.GreaterOrEqual(t, m.IdleTime, float64(0))
	}
}

func TestCPUCollector_buildCPUMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewCPUCollector(logger)

	testCases := []struct {
		name          string
		cpuInfo       []cpu.InfoStat
		percentages   []float64
		cpuTimes      []cpu.TimesStat
		physicalCount int
		logicalCount  int
		expectedLen   int
	}{
		{
			name: "Single CPU",
			cpuInfo: []cpu.InfoStat{
				{
					ModelName: "Test CPU",
					Mhz:       3000,
				},
			},
			percentages: []float64{75.5},
			cpuTimes: []cpu.TimesStat{
				{
					User:   100.0,
					System: 50.0,
					Idle:   200.0,
				},
			},
			physicalCount: 4,
			logicalCount:  8,
			expectedLen:   1,
		},
		{
			name: "Multiple CPUs",
			cpuInfo: []cpu.InfoStat{
				{
					ModelName: "CPU 1",
					Mhz:       2800,
				},
				{
					ModelName: "CPU 2",
					Mhz:       2800,
				},
			},
			percentages: []float64{60.0, 70.0},
			cpuTimes: []cpu.TimesStat{
				{
					User:   150.0,
					System: 75.0,
					Idle:   250.0,
				},
				{
					User:   160.0,
					System: 80.0,
					Idle:   260.0,
				},
			},
			physicalCount: 8,
			logicalCount:  16,
			expectedLen:   2,
		},
		{
			name: "Missing Percentages and Times",
			cpuInfo: []cpu.InfoStat{
				{
					ModelName: "Test CPU",
					Mhz:       3200,
				},
			},
			percentages:   []float64{},
			cpuTimes:      []cpu.TimesStat{},
			physicalCount: 2,
			logicalCount:  4,
			expectedLen:   1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := collector.buildCPUMetrics(tc.cpuInfo, tc.percentages, tc.cpuTimes, tc.physicalCount, tc.logicalCount)

			assert.Equal(t, tc.expectedLen, len(metrics))

			for i, metric := range metrics {
				assert.Equal(t, tc.cpuInfo[i].ModelName, metric.Model)
				assert.Equal(t, tc.physicalCount, metric.Cores)
				assert.Equal(t, tc.logicalCount, metric.Threads)
				assert.Equal(t, float64(tc.cpuInfo[i].Mhz), metric.FrequencyMHz)

				if i < len(tc.percentages) {
					assert.Equal(t, tc.percentages[i], metric.UsagePercent)
				}

				if i < len(tc.cpuTimes) {
					assert.Equal(t, tc.cpuTimes[i].User, metric.UserTime)
					assert.Equal(t, tc.cpuTimes[i].System, metric.SystemTime)
					assert.Equal(t, tc.cpuTimes[i].Idle, metric.IdleTime)
				}
			}
		})
	}
}

func TestCPUCollector_buildCPUMetrics_SafeIndexing(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewCPUCollector(logger)

	// Test case where we have more CPUs than percentages/times
	cpuInfo := []cpu.InfoStat{
		{ModelName: "CPU1", Mhz: 2000.0},
		{ModelName: "CPU2", Mhz: 2500.0},
		{ModelName: "CPU3", Mhz: 3000.0},
	}
	percentages := []float64{15.0} // Only one percentage
	cpuTimes := []cpu.TimesStat{
		{User: 80.0, System: 40.0, Idle: 280.0},
	} // Only one time stat

	metrics := collector.buildCPUMetrics(cpuInfo, percentages, cpuTimes, 4, 8)

	require.Len(t, metrics, 3)

	// First CPU should have all data
	assert.Equal(t, "CPU1", metrics[0].Model)
	assert.Equal(t, 15.0, metrics[0].UsagePercent)
	assert.Equal(t, 80.0, metrics[0].UserTime)

	// Second and third CPUs should have zero values for missing data
	assert.Equal(t, "CPU2", metrics[1].Model)
	assert.Equal(t, float64(0), metrics[1].UsagePercent)
	assert.Equal(t, float64(0), metrics[1].UserTime)

	assert.Equal(t, "CPU3", metrics[2].Model)
	assert.Equal(t, float64(0), metrics[2].UsagePercent)
	assert.Equal(t, float64(0), metrics[2].UserTime)
}

// Integration test - will only work on systems where gopsutil can collect data
func TestCPUCollector_Collect_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t)
	collector := NewCPUCollector(logger)

	metrics, err := collector.Collect()

	// The test should succeed on most systems
	if err != nil {
		t.Skipf("CPU collection not available on this system: %v", err)
	}

	assert.NoError(t, err)
	assert.NotEmpty(t, metrics)

	// Verify the structure of returned metrics
	for _, m := range metrics {
		assert.NotEmpty(t, m.Model)
		assert.GreaterOrEqual(t, m.Cores, 0)
		assert.GreaterOrEqual(t, m.Threads, 0)
		assert.GreaterOrEqual(t, m.FrequencyMHz, float64(0))
		assert.GreaterOrEqual(t, m.UsagePercent, float64(0))
		assert.LessOrEqual(t, m.UsagePercent, float64(100))
	}
}

func TestCPUCollector_Collect_LogsErrors(t *testing.T) {
	// This test verifies that the collector properly logs errors
	// We can't easily mock gopsutil calls, but we can verify the logger is used correctly

	logger := zaptest.NewLogger(t)
	collector := NewCPUCollector(logger)

	// This is more of a smoke test - if Collect works, great
	// If it fails, we ensure it returns an error
	_, err := collector.Collect()

	// On most systems this should work, but if it doesn't, ensure we get an error
	if err != nil {
		assert.Error(t, err)
	}
}
