package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kiatkoding/vpsctl/internal/lxd"
)

// ResourceSummary represents host resource summary
type ResourceSummary struct {
	Host      HostResources      `json:"host"`
	Allocated AllocatedResources `json:"allocated"`
	Available AvailableResources `json:"available"`
}

// HostResources represents host system resources
type HostResources struct {
	CPU         int    `json:"cpu"`
	CPUUsed     int    `json:"cpu_used"`
	Memory      string `json:"memory"`
	MemoryBytes int64  `json:"memory_bytes"`
	MemoryUsed  string `json:"memory_used"`
	MemoryUsedBytes int64 `json:"memory_used_bytes"`
	Disk        string `json:"disk"`
	DiskBytes   int64  `json:"disk_bytes"`
	DiskUsed    string `json:"disk_used"`
	DiskUsedBytes int64 `json:"disk_used_bytes"`
}

// AllocatedResources represents allocated resources to instances
type AllocatedResources struct {
	CPU         int    `json:"cpu"`
	Memory      string `json:"memory"`
	MemoryBytes int64  `json:"memory_bytes"`
	Disk        string `json:"disk"`
	DiskBytes   int64  `json:"disk_bytes"`
	Instances   int    `json:"instances"`
}

// AvailableResources represents available resources
type AvailableResources struct {
	CPU         int    `json:"cpu"`
	Memory      string `json:"memory"`
	MemoryBytes int64  `json:"memory_bytes"`
	Disk        string `json:"disk"`
	DiskBytes   int64  `json:"disk_bytes"`
}

// GetResources returns host resource summary
func GetResources(client *lxd.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		// Get server resources
		_, _, err := client.InstanceServer().GetServer()
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "server_info_failed",
				Message: err.Error(),
			})
			return
		}

		// Build host resources
		host := HostResources{
			CPU:         0,
			CPUUsed:     0,
			Memory:      "0B",
			MemoryBytes: 0,
			Disk:        "0B",
			DiskBytes:   0,
		}

		// Get server state for detailed resources
		// Note: GetServerState is not available in this LXD client version
		// Use GetResourceSummary from our lxd client instead
		lxdSummary, err := client.GetResourceSummary()
		if err == nil {
			host.CPU = int(lxdSummary.CPUTotal)
			host.CPUUsed = int(lxdSummary.CPUUsed)
			host.MemoryBytes = lxdSummary.MemoryTotal * 1024 * 1024 // Convert MB to bytes
			host.Memory = formatBytes(host.MemoryBytes)
			host.MemoryUsedBytes = lxdSummary.MemoryUsed * 1024 * 1024 // Convert MB to bytes
			host.MemoryUsed = formatBytes(host.MemoryUsedBytes)
			host.DiskBytes = lxdSummary.DiskTotal * 1024 * 1024 * 1024 // Convert GB to bytes
			host.Disk = formatBytes(host.DiskBytes)
			host.DiskUsedBytes = lxdSummary.DiskUsed * 1024 * 1024 * 1024 // Convert GB to bytes
			host.DiskUsed = formatBytes(host.DiskUsedBytes)
		}

		// Get allocated resources
		allocated := calculateAllocatedResources(client)

		// Calculate available resources
		available := AvailableResources{
			CPU:         host.CPU - allocated.CPU,
			MemoryBytes: host.MemoryBytes - allocated.MemoryBytes,
			DiskBytes:   host.DiskBytes - allocated.DiskBytes,
		}
		if available.CPU < 0 {
			available.CPU = 0
		}
		if available.MemoryBytes < 0 {
			available.MemoryBytes = 0
		}
		if available.DiskBytes < 0 {
			available.DiskBytes = 0
		}
		available.Memory = formatBytes(available.MemoryBytes)
		available.Disk = formatBytes(available.DiskBytes)

		summary := ResourceSummary{
			Host:      host,
			Allocated: allocated,
			Available: available,
		}

		c.JSON(http.StatusOK, summary)
	}
}

// GetAllocatedResources returns allocated resources to instances
func GetAllocatedResources(client *lxd.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		allocated := calculateAllocatedResources(client)

		c.JSON(http.StatusOK, allocated)
	}
}

// calculateAllocatedResources calculates resources allocated to all instances
func calculateAllocatedResources(client *lxd.Client) AllocatedResources {
	allocated := AllocatedResources{}

	instances, err := client.ListInstances()
	if err != nil {
		return allocated
	}

	allocated.Instances = len(instances)

	for _, inst := range instances {
		allocated.CPU += inst.CPU
		allocated.MemoryBytes += parseMemoryString(inst.Memory)
		allocated.DiskBytes += parseDiskString(inst.Disk)
	}

	allocated.Memory = formatBytes(allocated.MemoryBytes)
	allocated.Disk = formatBytes(allocated.DiskBytes)

	return allocated
}

// parseMemoryString parses memory string to bytes
func parseMemoryString(mem string) int64 {
	if mem == "" || mem == "unlimited" {
		return 0
	}

	multipliers := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}

	for suffix, mult := range multipliers {
		if len(mem) > len(suffix) {
			if mem[len(mem)-len(suffix):] == suffix {
				numStr := mem[:len(mem)-len(suffix)]
				if num, err := strconv.ParseInt(numStr, 10, 64); err == nil {
					return num * mult
				}
			}
		}
	}

	// Try parsing as plain number
	if num, err := strconv.ParseInt(mem, 10, 64); err == nil {
		return num
	}

	return 0
}

// parseDiskString parses disk string to bytes
func parseDiskString(disk string) int64 {
	if disk == "" {
		return 0
	}

	multipliers := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}

	for suffix, mult := range multipliers {
		if len(disk) > len(suffix) {
			if disk[len(disk)-len(suffix):] == suffix {
				numStr := disk[:len(disk)-len(suffix)]
				if num, err := strconv.ParseInt(numStr, 10, 64); err == nil {
					return num * mult
				}
			}
		}
	}

	// Try parsing as plain number
	if num, err := strconv.ParseInt(disk, 10, 64); err == nil {
		return num
	}

	return 0
}

// formatBytes formats bytes to human readable string
func formatBytes(bytes int64) string {
	if bytes < 0 {
		return "0B"
	}

	const unit = 1024
	if bytes < unit {
		return strconv.FormatInt(bytes, 10) + "B"
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return strconv.FormatFloat(float64(bytes)/float64(div), 'f', 1, 64) + string("KMGTPE"[exp]) + "B"
}
