package converter

import (
	"github.com/theotruvelot/g0s/internal/agent/model"
	pb "github.com/theotruvelot/g0s/pkg/proto/metric"
)

func ConvertHostMetrics(m model.HostMetrics) *pb.HostMetrics {
	return &pb.HostMetrics{
		Hostname:             m.Hostname,
		Uptime:               m.Uptime,
		Procs:                m.Procs,
		Os:                   m.OS,
		Platform:             m.Platform,
		PlatformFamily:       m.PlatformFamily,
		PlatformVersion:      m.PlatformVersion,
		VirtualizationSystem: m.VirtualizationSystem,
		VirtualizationRole:   m.VirtualizationRole,
		KernelVersion:        m.KernelVersion,
	}
}

func ConvertCPUMetrics(metrics []model.CPUMetrics) []*pb.CPUMetrics {
	result := make([]*pb.CPUMetrics, len(metrics))
	for i, m := range metrics {
		result[i] = &pb.CPUMetrics{
			Model:        m.Model,
			Cores:        int32(m.Cores),
			Threads:      int32(m.Threads),
			FrequencyMhz: m.FrequencyMHz,
			UsagePercent: m.UsagePercent,
			UserTime:     m.UserTime,
			SystemTime:   m.SystemTime,
			IdleTime:     m.IdleTime,
			CoreId:       int32(m.CoreID),
			IsTotal:      m.IsTotal,
		}
	}
	return result
}

func ConvertRAMMetrics(m model.RamMetrics) *pb.RAMMetrics {
	return &pb.RAMMetrics{
		TotalOctets:     m.TotalOctets,
		UsedOctets:      m.UsedOctets,
		FreeOctets:      m.FreeOctets,
		UsedPercent:     m.UsedPercent,
		AvailableOctets: m.AvailableOctets,
		SwapTotalOctets: m.SwapTotalOctets,
		SwapUsedOctets:  m.SwapUsedOctets,
		SwapUsedPercent: m.SwapUsedPerc,
	}
}

func ConvertDiskMetrics(metrics []model.DiskMetrics) []*pb.DiskMetrics {
	result := make([]*pb.DiskMetrics, len(metrics))
	for i, m := range metrics {
		result[i] = &pb.DiskMetrics{
			Path:        m.Path,
			Device:      m.Device,
			Fstype:      m.Fstype,
			Total:       m.TotalOctets,
			Used:        m.UsedOctets,
			Free:        m.FreeOctets,
			UsedPercent: m.UsedPercent,
			ReadCount:   m.ReadCount,
			WriteCount:  m.WriteCount,
			ReadOctets:  m.ReadOctets,
			WriteOctets: m.WriteOctets,
		}
	}
	return result
}

func ConvertNetworkMetrics(metrics []model.NetworkMetrics) []*pb.NetworkMetrics {
	result := make([]*pb.NetworkMetrics, len(metrics))
	for i, m := range metrics {
		result[i] = &pb.NetworkMetrics{
			InterfaceName: m.InterfaceName,
			BytesSent:     m.BytesSent,
			BytesRecv:     m.BytesRecv,
			PacketsSent:   m.PacketsSent,
			PacketsRecv:   m.PacketsRecv,
			ErrIn:         m.ErrIn,
			ErrOut:        m.ErrOut,
		}
	}
	return result
}

func ConvertDockerMetrics(metrics []model.DockerMetrics) []*pb.DockerMetrics {
	result := make([]*pb.DockerMetrics, len(metrics))
	for i, m := range metrics {
		result[i] = &pb.DockerMetrics{
			ContainerId:    m.ContainerID,
			ContainerName:  m.ContainerName,
			Image:          m.Image,
			ImageId:        m.ImageID,
			ImageName:      m.ImageName,
			ImageTag:       m.ImageTag,
			ImageDigest:    m.ImageDigest,
			ImageSize:      m.ImageSize,
			CpuMetrics:     ConvertCPUMetrics([]model.CPUMetrics{m.CPUMetrics})[0],
			RamMetrics:     ConvertRAMMetrics(m.RAMMetrics),
			DiskMetrics:    ConvertDiskMetrics([]model.DiskMetrics{m.DiskMetrics})[0],
			NetworkMetrics: ConvertNetworkMetrics([]model.NetworkMetrics{m.NetworkMetrics})[0],
		}
	}
	return result
}
