package collector

import (
	"github.com/theotruvelot/g0s/internal/agent/model"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/host"
	"go.uber.org/zap"
)

// HostCollector collects host system information and statistics.
type HostCollector struct {
	log *zap.Logger

	// Cache for static data that never changes
	mu               sync.RWMutex
	cachedStaticInfo *staticHostInfo
	cacheExpiry      time.Time
	cacheDuration    time.Duration
}

type staticHostInfo struct {
	hostname             string
	os                   string
	platform             string
	platformFamily       string
	platformVersion      string
	virtualizationSystem string
	virtualizationRole   string
	kernelVersion        string
}

// NewHostCollector creates a new HostCollector instance.
func NewHostCollector(log *zap.Logger) *HostCollector {
	return &HostCollector{
		log:           log,
		cacheDuration: 10 * time.Minute, // Cache static data for 10 minutes
	}
}

// getCachedStaticData returns cached static host data or fetches it if cache is expired
func (c *HostCollector) getCachedStaticData() (*staticHostInfo, error) {
	c.mu.RLock()
	if time.Now().Before(c.cacheExpiry) && c.cachedStaticInfo != nil {
		info := c.cachedStaticInfo
		c.mu.RUnlock()
		return info, nil
	}
	c.mu.RUnlock()

	// Cache expired, fetch new data
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double check in case another goroutine updated the cache
	if time.Now().Before(c.cacheExpiry) && c.cachedStaticInfo != nil {
		return c.cachedStaticInfo, nil
	}

	hostInfo, err := host.Info()
	if err != nil {
		c.log.Error("Failed to collect host information", zap.Error(err))
		return nil, err
	}

	// Cache static data that never changes
	c.cachedStaticInfo = &staticHostInfo{
		hostname:             hostInfo.Hostname,
		os:                   hostInfo.OS,
		platform:             hostInfo.Platform,
		platformFamily:       hostInfo.PlatformFamily,
		platformVersion:      hostInfo.PlatformVersion,
		virtualizationSystem: hostInfo.VirtualizationSystem,
		virtualizationRole:   hostInfo.VirtualizationRole,
		kernelVersion:        hostInfo.KernelVersion,
	}
	c.cacheExpiry = time.Now().Add(c.cacheDuration)

	return c.cachedStaticInfo, nil
}

// Collect gathers host system information including OS details, uptime, and virtualization info.
func (c *HostCollector) Collect() (model.HostMetrics, error) {
	// Get cached static data
	staticInfo, err := c.getCachedStaticData()
	if err != nil {
		return model.HostMetrics{}, err
	}

	// Get dynamic data that changes frequently
	hostInfo, err := host.Info()
	if err != nil {
		c.log.Error("Failed to collect host information", zap.Error(err))
		return model.HostMetrics{}, err
	}

	return c.buildHostMetrics(staticInfo, hostInfo), nil
}

// buildHostMetrics constructs host metrics from collected data.
func (c *HostCollector) buildHostMetrics(staticInfo *staticHostInfo, hostInfo *host.InfoStat) model.HostMetrics {
	return model.HostMetrics{
		Hostname:             staticInfo.hostname,
		Uptime:               hostInfo.Uptime, // Dynamic - changes
		Procs:                hostInfo.Procs,  // Dynamic - changes
		OS:                   staticInfo.os,
		Platform:             staticInfo.platform,
		PlatformFamily:       staticInfo.platformFamily,
		PlatformVersion:      staticInfo.platformVersion,
		VirtualizationSystem: staticInfo.virtualizationSystem,
		VirtualizationRole:   staticInfo.virtualizationRole,
		KernelVersion:        staticInfo.kernelVersion,
	}
}
