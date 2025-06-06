package collector

import (
	"testing"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewNetworkCollector(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewNetworkCollector(logger)

	assert.NotNil(t, collector)
	assert.Equal(t, logger, collector.log)
}

func TestNetworkCollector_Collect(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewNetworkCollector(logger)

	metrics, err := collector.Collect()
	require.NoError(t, err)
	require.NotNil(t, metrics)
	require.NotEmpty(t, metrics)

	// Verify that each metric has the expected fields populated
	for _, m := range metrics {
		assert.NotEmpty(t, m.InterfaceName)
		assert.GreaterOrEqual(t, m.BytesSent, uint64(0))
		assert.GreaterOrEqual(t, m.BytesRecv, uint64(0))
		assert.GreaterOrEqual(t, m.PacketsSent, uint64(0))
		assert.GreaterOrEqual(t, m.PacketsRecv, uint64(0))
		assert.GreaterOrEqual(t, m.ErrIn, uint64(0))
		assert.GreaterOrEqual(t, m.ErrOut, uint64(0))
	}
}

func TestNetworkCollector_buildNetworkMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewNetworkCollector(logger)

	testCases := []struct {
		name     string
		netStats []net.IOCountersStat
		expected int
	}{
		{
			name: "Multiple Interfaces",
			netStats: []net.IOCountersStat{
				{
					Name:        "eth0",
					BytesSent:   1000,
					BytesRecv:   2000,
					PacketsSent: 100,
					PacketsRecv: 200,
					Errin:       5,
					Errout:      3,
				},
				{
					Name:        "wlan0",
					BytesSent:   3000,
					BytesRecv:   4000,
					PacketsSent: 300,
					PacketsRecv: 400,
					Errin:       2,
					Errout:      1,
				},
			},
			expected: 2,
		},
		{
			name: "Single Interface",
			netStats: []net.IOCountersStat{
				{
					Name:        "lo",
					BytesSent:   500,
					BytesRecv:   500,
					PacketsSent: 50,
					PacketsRecv: 50,
					Errin:       0,
					Errout:      0,
				},
			},
			expected: 1,
		},
		{
			name:     "No Interfaces",
			netStats: []net.IOCountersStat{},
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := collector.buildNetworkMetrics(tc.netStats)
			assert.Equal(t, tc.expected, len(metrics))

			for i, metric := range metrics {
				assert.Equal(t, tc.netStats[i].Name, metric.InterfaceName)
				assert.Equal(t, tc.netStats[i].BytesSent, metric.BytesSent)
				assert.Equal(t, tc.netStats[i].BytesRecv, metric.BytesRecv)
				assert.Equal(t, tc.netStats[i].PacketsSent, metric.PacketsSent)
				assert.Equal(t, tc.netStats[i].PacketsRecv, metric.PacketsRecv)
				assert.Equal(t, tc.netStats[i].Errin, metric.ErrIn)
				assert.Equal(t, tc.netStats[i].Errout, metric.ErrOut)
			}
		})
	}
}

func TestNetworkCollector_Collect_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t)
	collector := NewNetworkCollector(logger)

	metrics, err := collector.Collect()

	if err != nil {
		t.Skipf("Network collection not available on this system: %v", err)
	}

	assert.NoError(t, err)

	for _, m := range metrics {
		assert.NotEmpty(t, m.InterfaceName)
		assert.GreaterOrEqual(t, m.BytesSent, uint64(0))
		assert.GreaterOrEqual(t, m.BytesRecv, uint64(0))
		assert.GreaterOrEqual(t, m.PacketsSent, uint64(0))
		assert.GreaterOrEqual(t, m.PacketsRecv, uint64(0))
	}
}
