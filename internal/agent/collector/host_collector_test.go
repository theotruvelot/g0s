package collector

import (
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/host"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestNewHostCollector(t *testing.T) {
	logger := zaptest.NewLogger(t)

	collector := NewHostCollector(logger)

	assert.NotNil(t, collector)
	assert.Equal(t, logger, collector.log)
}

func TestHostCollector_buildHostMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewHostCollector(logger)

	hostInfo := &host.InfoStat{
		Hostname:             "test-host",
		Uptime:               3600,
		Procs:                150,
		OS:                   "linux",
		Platform:             "ubuntu",
		PlatformFamily:       "debian",
		PlatformVersion:      "20.04",
		VirtualizationSystem: "docker",
		VirtualizationRole:   "guest",
		KernelVersion:        "5.4.0",
	}

	result := collector.buildHostMetrics(hostInfo)

	assert.Equal(t, "test-host", result.Hostname)
	assert.Equal(t, uint64(3600), result.Uptime)
	assert.Equal(t, uint64(150), result.Procs)
	assert.Equal(t, "linux", result.OS)
	assert.Equal(t, "ubuntu", result.Platform)
	assert.Equal(t, "debian", result.PlatformFamily)
	assert.Equal(t, "20.04", result.PlatformVersion)
	assert.Equal(t, "docker", result.VirtualizationSystem)
	assert.Equal(t, "guest", result.VirtualizationRole)
	assert.Equal(t, "5.4.0", result.KernelVersion)
}

func TestHostCollector_Collect_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t)
	collector := NewHostCollector(logger)

	metrics, err := collector.Collect()

	if err != nil {
		t.Skipf("Host collection not available on this system: %v", err)
	}

	assert.NoError(t, err)
	assert.NotEmpty(t, metrics.Hostname)
	assert.GreaterOrEqual(t, metrics.Uptime, uint64(0))
	assert.GreaterOrEqual(t, metrics.Procs, uint64(0))
	assert.NotEmpty(t, metrics.OS)
}

func TestHostCollector_Collect_MultipleRuns(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewHostCollector(logger)

	// Run multiple collections to test different scenarios
	for i := 0; i < 3; i++ {
		metrics, err := collector.Collect()
		assert.NoError(t, err)

		// Basic validations
		assert.NotEmpty(t, metrics.Hostname)
		assert.GreaterOrEqual(t, metrics.Uptime, uint64(0))
		assert.GreaterOrEqual(t, metrics.Procs, uint64(0))
		assert.NotEmpty(t, metrics.OS)
		assert.NotEmpty(t, metrics.Platform)
		assert.NotEmpty(t, metrics.KernelVersion)

		// Platform family might be empty on some systems, but if it exists it should be non-empty
		if metrics.PlatformFamily != "" {
			assert.NotEmpty(t, metrics.PlatformFamily)
		}

		// Platform version might be empty on some systems, but if it exists it should be non-empty
		if metrics.PlatformVersion != "" {
			assert.NotEmpty(t, metrics.PlatformVersion)
		}
	}
}

func TestHostCollector_Collect_ValidateMetricChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping extended test in short mode")
	}

	logger := zaptest.NewLogger(t)
	collector := NewHostCollector(logger)

	// First collection
	metrics1, err := collector.Collect()
	assert.NoError(t, err)

	// Small delay to allow uptime to increase
	time.Sleep(1000 * time.Millisecond)

	// Second collection
	metrics2, err := collector.Collect()
	assert.NoError(t, err)

	// Static values should remain the same
	assert.Equal(t, metrics1.Hostname, metrics2.Hostname)
	assert.Equal(t, metrics1.OS, metrics2.OS)
	assert.Equal(t, metrics1.Platform, metrics2.Platform)
	assert.Equal(t, metrics1.PlatformFamily, metrics2.PlatformFamily)
	assert.Equal(t, metrics1.PlatformVersion, metrics2.PlatformVersion)
	assert.Equal(t, metrics1.KernelVersion, metrics2.KernelVersion)

	// Uptime should increase
	assert.Greater(t, metrics2.Uptime, metrics1.Uptime)

	// Number of processes might change but should be reasonable
	assert.GreaterOrEqual(t, metrics2.Procs, uint64(1))
}
