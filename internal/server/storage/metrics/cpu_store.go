package metrics

import (
	"fmt"
	"strings"

	pb "github.com/theotruvelot/g0s/pkg/proto/metric"
	"go.uber.org/zap"
)

type CPUStore struct {
	logger     *zap.Logger
	vmEndpoint string
}

func NewCPUStore(vmEndpoint string, logger *zap.Logger) *CPUStore {
	return &CPUStore{
		logger:     logger,
		vmEndpoint: vmEndpoint,
	}
}

func (s *CPUStore) Format(metrics *pb.MetricsPayload, timestamp int64) []string {
	var lines []string

	for _, cpu := range metrics.Cpu {
		if cpu.IsTotal {
			lines = append(lines, fmt.Sprintf(
				"cpu_usage_percent_avg{host=\"%s\"} %f %d\n",
				metrics.Host.Hostname,
				cpu.UsagePercent,
				timestamp,
			))
		} else {
			lines = append(lines, fmt.Sprintf(
				"cpu_usage_percent{host=\"%s\",model=\"%s\",core_id=\"%d\"} %f %d\n",
				metrics.Host.Hostname,
				cpu.Model,
				cpu.CoreId,
				cpu.UsagePercent,
				timestamp,
			))
			lines = append(lines, fmt.Sprintf(
				"cpu_user_time{host=\"%s\",model=\"%s\",core_id=\"%d\"} %f %d\n",
				metrics.Host.Hostname,
				cpu.Model,
				cpu.CoreId,
				cpu.UserTime,
				timestamp,
			))
			lines = append(lines, fmt.Sprintf(
				"cpu_system_time{host=\"%s\",model=\"%s\",core_id=\"%d\"} %f %d\n",
				metrics.Host.Hostname,
				cpu.Model,
				cpu.CoreId,
				cpu.SystemTime,
				timestamp,
			))
			lines = append(lines, fmt.Sprintf(
				"cpu_idle_time{host=\"%s\",model=\"%s\",core_id=\"%d\"} %f %d\n",
				metrics.Host.Hostname,
				cpu.Model,
				cpu.CoreId,
				cpu.IdleTime,
				timestamp,
			))
		}
	}

	return lines
}

func (s *CPUStore) Store(data []string) error {
	if len(data) == 0 {
		return nil
	}

	payload := strings.Join(data, "")
	endpoint := fmt.Sprintf("%s/api/v1/import/prometheus", s.vmEndpoint)

	if err := sendWithRetry(endpoint, payload, s.logger, "CPU"); err != nil {
		return err
	}

	s.logger.Debug("CPU metrics stored successfully", zap.Int("metrics_count", len(data)))
	return nil
}
