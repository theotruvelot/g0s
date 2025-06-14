package metrics

import (
	"fmt"
	"github.com/theotruvelot/g0s/pkg/logger"
	"strings"

	pb "github.com/theotruvelot/g0s/pkg/proto/metric"
	"go.uber.org/zap"
)

type NetworkStore struct {
	vmEndpoint string
}

func NewNetworkStore(vmEndpoint string) *NetworkStore {
	return &NetworkStore{
		vmEndpoint: vmEndpoint,
	}
}

func (s *NetworkStore) Format(metrics *pb.MetricsPayload, timestamp int64) []string {
	var lines []string

	for _, net := range metrics.Network {
		lines = append(lines, fmt.Sprintf(
			"network_bytes_sent{host=\"%s\",interface=\"%s\"} %d %d\n",
			metrics.Host.Hostname,
			net.InterfaceName,
			net.BytesSent,
			timestamp,
		))
		lines = append(lines, fmt.Sprintf(
			"network_bytes_recv{host=\"%s\",interface=\"%s\"} %d %d\n",
			metrics.Host.Hostname,
			net.InterfaceName,
			net.BytesRecv,
			timestamp,
		))
		lines = append(lines, fmt.Sprintf(
			"network_packets_sent{host=\"%s\",interface=\"%s\"} %d %d\n",
			metrics.Host.Hostname,
			net.InterfaceName,
			net.PacketsSent,
			timestamp,
		))
		lines = append(lines, fmt.Sprintf(
			"network_packets_recv{host=\"%s\",interface=\"%s\"} %d %d\n",
			metrics.Host.Hostname,
			net.InterfaceName,
			net.PacketsRecv,
			timestamp,
		))
	}

	return lines
}

func (s *NetworkStore) Store(data []string) error {
	if len(data) == 0 {
		return nil
	}

	payload := strings.Join(data, "")
	endpoint := fmt.Sprintf("%s/api/v1/import/prometheus", s.vmEndpoint)

	if err := sendWithRetry(endpoint, payload, "Network"); err != nil {
		return err
	}

	logger.Debug("Network metrics stored successfully", zap.Int("metrics_count", len(data)))
	return nil
}
