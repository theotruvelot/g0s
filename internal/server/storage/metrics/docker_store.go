package metrics

import (
	"fmt"
	"strings"

	pb "github.com/theotruvelot/g0s/pkg/proto/metric"
	"go.uber.org/zap"
)

type DockerStore struct {
	logger     *zap.Logger
	vmEndpoint string
}

func NewDockerStore(vmEndpoint string, logger *zap.Logger) *DockerStore {
	return &DockerStore{
		logger:     logger,
		vmEndpoint: vmEndpoint,
	}
}

func (s *DockerStore) Format(metrics *pb.MetricsPayload, timestamp int64) []string {
	var lines []string

	for _, docker := range metrics.Docker {
		lines = append(lines, fmt.Sprintf(
			"docker_cpu_usage_percent{host=\"%s\",container_id=\"%s\",container_name=\"%s\",image=\"%s\"} %f %d\n",
			metrics.Host.Hostname,
			docker.ContainerId,
			docker.ContainerName,
			docker.Image,
			docker.CpuMetrics.UsagePercent,
			timestamp,
		))
		lines = append(lines, fmt.Sprintf(
			"docker_memory_used_percent{host=\"%s\",container_id=\"%s\",container_name=\"%s\",image=\"%s\"} %f %d\n",
			metrics.Host.Hostname,
			docker.ContainerId,
			docker.ContainerName,
			docker.Image,
			docker.RamMetrics.UsedPercent,
			timestamp,
		))
		lines = append(lines, fmt.Sprintf(
			"docker_network_bytes_sent{host=\"%s\",container_id=\"%s\",container_name=\"%s\",image=\"%s\"} %d %d\n",
			metrics.Host.Hostname,
			docker.ContainerId,
			docker.ContainerName,
			docker.Image,
			docker.NetworkMetrics.BytesSent,
			timestamp,
		))
	}

	return lines
}

func (s *DockerStore) Store(data []string) error {
	if len(data) == 0 {
		return nil
	}

	payload := strings.Join(data, "")
	endpoint := fmt.Sprintf("%s/api/v1/import/prometheus", s.vmEndpoint)

	if err := sendWithRetry(endpoint, payload, s.logger, "Docker"); err != nil {
		return err
	}

	s.logger.Debug("Docker metrics stored successfully", zap.Int("metrics_count", len(data)))
	return nil
}
