package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/theotruvelot/g0s/internal/agent/model"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

type DockerCollector struct {
	log    *zap.Logger
	client *client.Client
}

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

func (d *DockerCollector) Collect() ([]model.DockerMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	containers, err := d.client.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		return []model.DockerMetrics{}, nil
	}

	numWorkers := d.calculateOptimalWorkers(len(containers))

	d.log.Debug("Starting container metrics collection",
		zap.Int("containers", len(containers)),
		zap.Int("workers", numWorkers))

	jobs := make(chan types.Container, len(containers))
	results := make(chan containerResult, len(containers))

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go d.worker(ctx, jobs, results, &wg)
	}

	for _, container := range containers {
		jobs <- container
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	metrics := make([]model.DockerMetrics, 0, len(containers))
	errorCount := 0

	for result := range results {
		if result.err != nil {
			errorCount++
			d.log.Debug("Failed to collect container metrics",
				zap.String("containerID", result.containerID),
				zap.Error(result.err))
		} else {
			metrics = append(metrics, result.metrics)
		}
	}

	if errorCount > 0 {
		d.log.Warn("Some container metrics collection failed",
			zap.Int("errors", errorCount),
			zap.Int("successful", len(metrics)))
	}

	return metrics, nil
}

func (d *DockerCollector) calculateOptimalWorkers(containerCount int) int {
	const (
		minWorkers          = 1
		maxWorkers          = 8
		containersPerWorker = 3
	)

	if containerCount <= 0 {
		return minWorkers
	}

	optimalWorkers := (containerCount + containersPerWorker - 1) / containersPerWorker

	if optimalWorkers > maxWorkers {
		return maxWorkers
	}

	if optimalWorkers < minWorkers {
		return minWorkers
	}

	return optimalWorkers
}

type containerResult struct {
	metrics     model.DockerMetrics
	containerID string
	err         error
}

func (d *DockerCollector) worker(ctx context.Context, jobs <-chan types.Container, results chan<- containerResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for container := range jobs {
		if ctx.Err() != nil {
			results <- containerResult{
				containerID: container.ID[:12],
				err:         ctx.Err(),
			}
			continue
		}

		metrics, err := d.processContainer(ctx, container)
		results <- containerResult{
			metrics:     metrics,
			containerID: container.ID[:12],
			err:         err,
		}
	}
}

func (d *DockerCollector) processContainer(ctx context.Context, c types.Container) (model.DockerMetrics, error) {
	stats, err := d.collectContainerStats(ctx, c.ID)
	if err != nil {
		return model.DockerMetrics{}, err
	}

	imageName, imageTag := parseImageName(c.Image)

	return model.DockerMetrics{
		ContainerID:    c.ID,
		ContainerName:  strings.TrimPrefix(c.Names[0], "/"),
		Image:          c.Image,
		ImageID:        c.ImageID,
		ImageName:      imageName,
		ImageTag:       imageTag,
		CPUMetrics:     d.buildCPUMetrics(stats),
		RAMMetrics:     d.buildRAMMetrics(stats),
		DiskMetrics:    d.buildDiskMetrics(stats),
		NetworkMetrics: d.buildNetworkMetrics(stats),
	}, nil
}

func (d *DockerCollector) buildCPUMetrics(stats *container.StatsResponse) model.CPUMetrics {
	return model.CPUMetrics{
		UsagePercent: calculateCPUPercentage(stats),
		UserTime:     float64(stats.CPUStats.CPUUsage.UsageInUsermode),
		SystemTime:   float64(stats.CPUStats.CPUUsage.UsageInKernelmode),
		Cores:        int(stats.CPUStats.OnlineCPUs),
		Threads:      int(stats.CPUStats.CPUUsage.TotalUsage),
	}
}

func (d *DockerCollector) buildRAMMetrics(stats *container.StatsResponse) model.RamMetrics {
	usage := stats.MemoryStats.Usage
	limit := stats.MemoryStats.Limit

	return model.RamMetrics{
		TotalOctets:     limit,
		UsedOctets:      usage,
		AvailableOctets: limit - usage,
		UsedPercent:     calculateMemoryPercentage(usage, limit),
	}
}

func (d *DockerCollector) buildDiskMetrics(stats *container.StatsResponse) model.DiskMetrics {
	var readBytes, writeBytes, readOps, writeOps uint64

	for _, blkio := range stats.BlkioStats.IoServiceBytesRecursive {
		if blkio.Op == "Read" {
			readBytes += blkio.Value
		} else if blkio.Op == "Write" {
			writeBytes += blkio.Value
		}
	}

	for _, blkio := range stats.BlkioStats.IoServicedRecursive {
		if blkio.Op == "Read" {
			readOps += blkio.Value
		} else if blkio.Op == "Write" {
			writeOps += blkio.Value
		}
	}

	return model.DiskMetrics{
		Path:        "/",
		ReadCount:   readOps,
		WriteCount:  writeOps,
		ReadOctets:  readBytes,
		WriteOctets: writeBytes,
	}
}

func (d *DockerCollector) buildNetworkMetrics(stats *container.StatsResponse) model.NetworkMetrics {
	var metrics model.NetworkMetrics

	for _, network := range stats.Networks {
		metrics.BytesRecv += network.RxBytes
		metrics.BytesSent += network.TxBytes
		metrics.PacketsRecv += network.RxPackets
		metrics.PacketsSent += network.TxPackets
		metrics.ErrIn += network.RxErrors
		metrics.ErrOut += network.TxErrors
	}

	return metrics
}

func (d *DockerCollector) Close() {
	if d.client != nil {
		d.client.Close()
	}
}

func (d *DockerCollector) collectContainerStats(ctx context.Context, containerID string) (*container.StatsResponse, error) {
	stats, err := d.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer stats.Body.Close()

	var statsJSON container.StatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&statsJSON); err != nil {
		return nil, fmt.Errorf("failed to decode stats JSON: %w", err)
	}

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
		if len(stats.CPUStats.CPUUsage.PercpuUsage) > 0 {
			numCPUs = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
		} else {
			numCPUs = 1.0
		}
	}

	cpuPercent := (cpuDelta / systemDelta) * numCPUs * 100.0

	if cpuPercent > 100.0 {
		return 100.0
	} else if cpuPercent < 0.0 {
		return 0.0
	}

	return cpuPercent
}

func calculateMemoryPercentage(used, limit uint64) float64 {
	if limit == 0 {
		return 0.0
	}

	percent := float64(used) / float64(limit) * 100.0
	if percent > 100.0 {
		return 100.0
	}

	return percent
}

func parseImageName(image string) (string, string) {
	if idx := strings.LastIndex(image, ":"); idx != -1 {
		return image[:idx], image[idx+1:]
	}
	return image, "latest"
}
