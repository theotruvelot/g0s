package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewDockerCollector(t *testing.T) {
	logger := zaptest.NewLogger(t)

	collector, err := NewDockerCollector(logger)
	require.NoError(t, err)
	assert.NotNil(t, collector)
	assert.Equal(t, logger, collector.log)
	assert.NotNil(t, collector.client)
}

func TestDockerCollector_Collect_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t)
	collector, err := NewDockerCollector(logger)
	require.NoError(t, err)

	metrics, err := collector.Collect()
	if err != nil {
		// Skip if Docker daemon is not available
		t.Skipf("Docker collection not available on this system: %v", err)
	}

	// Even if no containers are running, we should get an empty slice, not nil
	assert.NotNil(t, metrics)

	// If there are running containers, verify their metrics
	for _, m := range metrics {
		assert.NotEmpty(t, m.ContainerID)
		assert.NotEmpty(t, m.ContainerName)
		assert.NotEmpty(t, m.Image)
		assert.NotEmpty(t, m.ImageID)

		// CPU metrics
		assert.GreaterOrEqual(t, m.CPUMetrics.UsagePercent, float64(0))
		assert.LessOrEqual(t, m.CPUMetrics.UsagePercent, float64(100))
		assert.GreaterOrEqual(t, m.CPUMetrics.UserTime, float64(0))
		assert.GreaterOrEqual(t, m.CPUMetrics.SystemTime, float64(0))

		// RAM metrics
		assert.Greater(t, m.RAMMetrics.TotalOctets, uint64(0))
		assert.GreaterOrEqual(t, m.RAMMetrics.UsedOctets, uint64(0))
		assert.GreaterOrEqual(t, m.RAMMetrics.AvailableOctets, uint64(0))
		assert.GreaterOrEqual(t, m.RAMMetrics.UsedPercent, float64(0))
		assert.LessOrEqual(t, m.RAMMetrics.UsedPercent, float64(100))

		// Network metrics
		assert.GreaterOrEqual(t, m.NetworkMetrics.BytesRecv, uint64(0))
		assert.GreaterOrEqual(t, m.NetworkMetrics.BytesSent, uint64(0))
	}
}
