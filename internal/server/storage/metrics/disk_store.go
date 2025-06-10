package metrics

import (
	"fmt"
	"strings"

	pb "github.com/theotruvelot/g0s/pkg/proto/metric"
	"go.uber.org/zap"
)

type DiskStore struct {
	logger     *zap.Logger
	vmEndpoint string
}

func NewDiskStore(vmEndpoint string, logger *zap.Logger) *DiskStore {
	return &DiskStore{
		logger:     logger,
		vmEndpoint: vmEndpoint,
	}
}

func (s *DiskStore) Format(metrics *pb.MetricsPayload, timestamp int64) []string {
	var lines []string

	for _, disk := range metrics.Disk {
		lines = append(lines, fmt.Sprintf(
			"disk_total{host=\"%s\",device=\"%s\",path=\"%s\",fstype=\"%s\"} %d %d\n",
			metrics.Host.Hostname,
			disk.Device,
			disk.Path,
			disk.Fstype,
			disk.Total,
			timestamp,
		))
		lines = append(lines, fmt.Sprintf(
			"disk_used{host=\"%s\",device=\"%s\",path=\"%s\",fstype=\"%s\"} %d %d\n",
			metrics.Host.Hostname,
			disk.Device,
			disk.Path,
			disk.Fstype,
			disk.Used,
			timestamp,
		))
		lines = append(lines, fmt.Sprintf(
			"disk_used_percent{host=\"%s\",device=\"%s\",path=\"%s\",fstype=\"%s\"} %f %d\n",
			metrics.Host.Hostname,
			disk.Device,
			disk.Path,
			disk.Fstype,
			disk.UsedPercent,
			timestamp,
		))
	}

	return lines
}

func (s *DiskStore) Store(data []string) error {
	if len(data) == 0 {
		return nil
	}

	payload := strings.Join(data, "")
	endpoint := fmt.Sprintf("%s/api/v1/import/prometheus", s.vmEndpoint)

	if err := sendWithRetry(endpoint, payload, s.logger, "Disk"); err != nil {
		return err
	}

	s.logger.Debug("Disk metrics stored successfully", zap.Int("metrics_count", len(data)))
	return nil
}
