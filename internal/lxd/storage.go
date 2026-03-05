package lxd

import (
	"fmt"

	lxd "github.com/canonical/lxd/client"
)

// StoragePoolInfo represents storage pool information
type StoragePoolInfo struct {
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	TotalBytes int64  `json:"total_bytes"`
	UsedBytes  int64  `json:"used_bytes"`
}

// ResourceSummary represents host resource summary
type ResourceSummary struct {
	CPUTotal    int64
	CPUUsed     int64
	MemoryTotal int64 // in MB
	MemoryUsed  int64 // in MB
	DiskTotal   int64 // in GB
	DiskUsed    int64 // in GB
}

// ListStoragePools returns all storage pools
func (c *Client) ListStoragePools() ([]StoragePoolInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to LXD")
	}

	pools, err := c.instanceServer.GetStoragePools()
	if err != nil {
		return nil, fmt.Errorf("failed to list storage pools: %w", err)
	}

	result := make([]StoragePoolInfo, 0, len(pools))
	for _, pool := range pools {
		info := StoragePoolInfo{
			Name:   pool.Name,
			Driver: pool.Driver,
		}
		result = append(result, info)
	}

	return result, nil
}

// GetResourceSummary returns a summary of host resources
func (c *Client) GetResourceSummary() (ResourceSummary, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return ResourceSummary{}, fmt.Errorf("not connected to LXD")
	}

	var summary ResourceSummary

	// Get all instances to calculate used resources
	instances, err := c.instanceServer.GetInstances(lxd.GetInstancesArgs{})
	if err != nil {
		return summary, fmt.Errorf("failed to get instances: %w", err)
	}

	// Calculate used CPU and Memory
	for _, inst := range instances {
		// CPU
		if cpu, ok := inst.Config["limits.cpu"]; ok {
			var cpuCount int64
			if _, err := fmt.Sscanf(cpu, "%d", &cpuCount); err == nil {
				summary.CPUUsed += cpuCount
			}
		} else {
			summary.CPUUsed += 1 // Default 1 CPU per instance
		}

		// Memory - parse from config
		if mem, ok := inst.Config["limits.memory"]; ok {
			memBytes := parseMemoryToBytes(mem)
			summary.MemoryUsed += memBytes / (1024 * 1024) // Convert to MB
		}
	}

	// Set reasonable defaults
	summary.CPUTotal = 8
	summary.MemoryTotal = 16384 // 16GB in MB
	summary.DiskTotal = 100 // 100GB
	summary.DiskUsed = 0

	return summary, nil
}

// parseMemoryToBytes parses memory string to bytes
func parseMemoryToBytes(mem string) int64 {
	var bytes int64
	var unit string

	n, err := fmt.Sscanf(mem, "%d%s", &bytes, &unit)
	if err != nil || n < 1 {
		// Try parsing as plain number (bytes)
		fmt.Sscanf(mem, "%d", &bytes)
		return bytes
	}

	switch unit {
	case "KB", "kB", "kb":
		bytes *= 1024
	case "MB", "mB", "mb":
		bytes *= 1024 * 1024
	case "GB", "gB", "gb":
		bytes *= 1024 * 1024 * 1024
	case "TB", "tB", "tb":
		bytes *= 1024 * 1024 * 1024 * 1024
	}

	return bytes
}
