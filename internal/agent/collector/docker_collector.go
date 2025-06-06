package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/theotruvelot/g0s/internal/agent/model/metric"
	"go.uber.org/zap"
)

// DockerCollector collects Docker container metrics
type DockerCollector struct {
	log    *zap.Logger
	client *client.Client
}

// NewDockerCollector creates a new DockerCollector instance
func NewDockerCollector(log *zap.Logger) (*DockerCollector, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &DockerCollector{
		log:    log,
		client: cli,
	}, nil
}

// Collect gathers metrics from all running Docker containers
func (d *DockerCollector) Collect() ([]metric.DockerMetrics, error) {
	ctx := context.Background()

	containers, err := d.client.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var wg sync.WaitGroup
	metricsChan := make(chan metric.DockerMetrics, len(containers))
	errorsChan := make(chan error, len(containers))

	// Launch a goroutine for each container
	for _, container := range containers {
		wg.Add(1)
		c := container
		go func() {
			defer wg.Done()

			stats, err := d.collectContainerStats(ctx, c.ID)
			if err != nil {
				d.log.Error("Failed to collect stats for container",
					zap.String("containerID", c.ID),
					zap.Error(err))
				errorsChan <- err
				return
			}

			// Extract image details
			imageName, imageTag := parseImageName(c.Image)

			containerMetrics := metric.DockerMetrics{
				ContainerID:   c.ID,
				ContainerName: strings.TrimPrefix(c.Names[0], "/"),
				Image:         c.Image,
				ImageID:       c.ImageID,
				ImageName:     imageName,
				ImageTag:      imageTag,
				CPUMetrics: metric.CPUMetrics{
					UsagePercent: calculateCPUPercentage(stats),
					UserTime:     float64(stats.CPUStats.CPUUsage.UsageInUsermode),
					SystemTime:   float64(stats.CPUStats.CPUUsage.UsageInKernelmode),
					Cores:        int(stats.CPUStats.OnlineCPUs),
					Threads:      int(stats.CPUStats.CPUUsage.TotalUsage),
				},
				RAMMetrics: metric.RamMetrics{
					TotalOctets:     stats.MemoryStats.Limit,
					UsedOctets:      stats.MemoryStats.Usage,
					AvailableOctets: stats.MemoryStats.Limit - stats.MemoryStats.Usage,
					UsedPercent:     float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100,
				},
				DiskMetrics:    collectDiskMetrics(stats),
				NetworkMetrics: collectNetworkMetrics(stats),
			}

			metricsChan <- containerMetrics
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(metricsChan)
	close(errorsChan)

	// Collect results
	var metrics []metric.DockerMetrics
	for metric := range metricsChan {
		metrics = append(metrics, metric)
	}

	// Check for errors
	var errs []error
	for err := range errorsChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return metrics, fmt.Errorf("errors collecting container stats: %v", errs)
	}

	return metrics, nil
}

// collectContainerStats gets stats for a specific container
func (d *DockerCollector) collectContainerStats(ctx context.Context, containerID string) (*container.StatsResponse, error) {
	d.log.Debug("Collecting stats for container", zap.String("containerID", containerID))

	stats, err := d.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer stats.Body.Close()

	var statsJSON container.StatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&statsJSON); err != nil {
		return nil, fmt.Errorf("failed to decode stats JSON: %w", err)
	}

	d.log.Debug("Raw Docker stats",
		zap.String("containerID", containerID),
		zap.Any("cpuStats", statsJSON.CPUStats),
		zap.Any("memoryStats", statsJSON.MemoryStats),
		zap.Any("networks", statsJSON.Networks))

	return &statsJSON, nil
}

func calculateCPUPercentage(stats *container.StatsResponse) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)

	if systemDelta <= 0 || cpuDelta < 0 {
		return 0.0
	}

	numCPUs := float64(stats.CPUStats.OnlineCPUs)
	if numCPUs == 0 {
		numCPUs = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
		if numCPUs == 0 {
			numCPUs = 1.0
		}
	}

	cpuPercent := (cpuDelta / systemDelta) * numCPUs * 100.0
	if cpuPercent > 100.0 {
		cpuPercent = 100.0
	} else if cpuPercent < 0.0 {
		cpuPercent = 0.0
	}

	return cpuPercent
}

func collectDiskMetrics(stats *container.StatsResponse) metric.DiskMetrics {
	var readOctets, writeOctets uint64
	var readCount, writeCount uint64

	for _, blkio := range stats.BlkioStats.IoServiceBytesRecursive {
		switch blkio.Op {
		case "Read":
			readOctets += blkio.Value
		case "Write":
			writeOctets += blkio.Value
		}
	}

	for _, blkio := range stats.BlkioStats.IoServicedRecursive {
		switch blkio.Op {
		case "Read":
			readCount += blkio.Value
		case "Write":
			writeCount += blkio.Value
		}
	}

	return metric.DiskMetrics{
		Path:        "/", // Container root filesystem
		Device:      "",  // Device info not available from container stats
		Fstype:      "",  // Filesystem type not available from container stats
		TotalOctets: 0,   // Total space not available from container stats
		UsedOctets:  0,   // Used space not available from container stats
		FreeOctets:  0,   // Free space not available from container stats
		UsedPercent: 0,   // Usage percent not available from container stats
		ReadCount:   readCount,
		WriteCount:  writeCount,
		ReadOctets:  readOctets,
		WriteOctets: writeOctets,
	}
}

func collectNetworkMetrics(stats *container.StatsResponse) metric.NetworkMetrics {
	var rx, tx uint64
	var rxPackets, txPackets uint64
	var rxErrors, txErrors uint64

	for _, network := range stats.Networks {
		rx += network.RxBytes
		tx += network.TxBytes
		rxPackets += network.RxPackets
		txPackets += network.TxPackets
		rxErrors += network.RxErrors
		txErrors += network.TxErrors
	}

	return metric.NetworkMetrics{
		BytesRecv:   rx,
		BytesSent:   tx,
		PacketsRecv: rxPackets,
		PacketsSent: txPackets,
		ErrIn:       rxErrors,
		ErrOut:      txErrors,
	}
}

// parseImageName extracts image name and tag from a Docker image string
func parseImageName(image string) (string, string) {
	parts := strings.Split(image, ":")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return parts[0], "latest" // tag par défaut si pas spécifié
}
