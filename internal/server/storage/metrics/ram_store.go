package metrics

import (
	"fmt"
	"github.com/theotruvelot/g0s/pkg/logger"
	"strings"

	pb "github.com/theotruvelot/g0s/pkg/proto/metric"
	"go.uber.org/zap"
)

type RAMStore struct {
	vmEndpoint string
}

func NewRAMStore(vmEndpoint string) *RAMStore {
	return &RAMStore{
		vmEndpoint: vmEndpoint,
	}
}

func (s *RAMStore) Format(metrics *pb.MetricsPayload, timestamp int64) []string {
	var lines []string

	lines = append(lines, fmt.Sprintf(
		"ram_total_octets{host=\"%s\"} %d %d\n",
		metrics.Host.Hostname,
		metrics.Ram.TotalOctets,
		timestamp,
	))
	lines = append(lines, fmt.Sprintf(
		"ram_used_octets{host=\"%s\"} %d %d\n",
		metrics.Host.Hostname,
		metrics.Ram.UsedOctets,
		timestamp,
	))
	lines = append(lines, fmt.Sprintf(
		"ram_used_percent{host=\"%s\"} %f %d\n",
		metrics.Host.Hostname,
		metrics.Ram.UsedPercent,
		timestamp,
	))

	return lines
}

func (s *RAMStore) Store(data []string) error {
	if len(data) == 0 {
		return nil
	}

	payload := strings.Join(data, "")
	endpoint := fmt.Sprintf("%s/api/v1/import/prometheus", s.vmEndpoint)

	if err := sendWithRetry(endpoint, payload, "RAM"); err != nil {
		return err
	}

	logger.Debug("RAM metrics stored successfully", zap.Int("metrics_count", len(data)))
	return nil
}
