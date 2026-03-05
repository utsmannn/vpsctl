package resource

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kiatkoding/vpsctl/internal/lxd"
)

// Summary represents resource summary
type Summary struct {
	Host      HostResources      `json:"host"`
	Allocated AllocatedResources `json:"allocated"`
	Available AvailableResources `json:"available"`
}

// HostResources represents physical host resources
type HostResources struct {
	CPU       int    `json:"cpu"`
	MemoryMB  int    `json:"memory_mb"`
	DiskGB    int    `json:"disk_gb"`
	MemoryStr string `json:"memory"`
	DiskStr   string `json:"disk"`
}

// AllocatedResources represents total allocated resources across all instances
type AllocatedResources struct {
	CPU      int `json:"cpu"`
	MemoryMB int `json:"memory_mb"`
	DiskGB   int `json:"disk_gb"`
}

// AvailableResources represents available resources (host - allocated)
type AvailableResources struct {
	CPU      int `json:"cpu"`
	MemoryMB int `json:"memory_mb"`
	DiskGB   int `json:"disk_gb"`
}

// Tracker tracks resource allocation
type Tracker struct {
	lxdClient *lxd.Client
}

// NewTracker creates a new resource tracker
func NewTracker(client *lxd.Client) *Tracker {
	return &Tracker{
		lxdClient: client,
	}
}

// GetSummary returns full resource summary
func (t *Tracker) GetSummary() (*Summary, error) {
	host, err := t.GetHostResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get host resources: %w", err)
	}

	allocated, err := t.GetAllocated()
	if err != nil {
		return nil, fmt.Errorf("failed to get allocated resources: %w", err)
	}

	available := t.calculateAvailable(host, allocated)

	return &Summary{
		Host:      *host,
		Allocated: *allocated,
		Available: *available,
	}, nil
}

// GetHostResources returns host physical resources
func (t *Tracker) GetHostResources() (*HostResources, error) {
	// Get host resources from LXD server info
	serverInfo, err := t.lxdClient.InstanceServer().GetServer()
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	// Extract CPU count
	cpu := 0
	if cpuStr, ok := serverInfo.Environment["kernel_cpu_cores"]; ok {
		cpu, _ = strconv.Atoi(cpuStr)
	}

	// Extract memory (comes in bytes)
	memoryMB := 0
	if memStr, ok := serverInfo.Environment["kernel_memory_total"]; ok {
		memBytes, _ := strconv.ParseInt(memStr, 10, 64)
		memoryMB = int(memBytes / (1024 * 1024)) // Convert to MB
	}

	// Get storage pool info for disk
	diskGB := 0
	pools, err := t.lxdClient.InstanceServer().GetStoragePools()
	if err == nil && len(pools) > 0 {
		// Use the default pool or first available
		for _, pool := range pools {
			if pool.Config != nil {
				if sizeStr, ok := pool.Config["size"]; ok {
					diskGB = parseSizeToGB(sizeStr)
					break
				}
			}
		}
	}

	return &HostResources{
		CPU:       cpu,
		MemoryMB:  memoryMB,
		DiskGB:    diskGB,
		MemoryStr: formatMemory(memoryMB),
		DiskStr:   formatDisk(diskGB),
	}, nil
}

// GetAllocated returns total allocated resources across all instances
func (t *Tracker) GetAllocated() (*AllocatedResources, error) {
	instances, err := t.lxdClient.ListInstances()
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	allocated := &AllocatedResources{}

	for _, instance := range instances {
		// CPU - instance.CPU is already parsed
		allocated.CPU += instance.CPU

		// Memory - parse from instance.Memory string
		memMB := parseMemoryStringToMB(instance.Memory)
		allocated.MemoryMB += memMB

		// Disk - parse from instance.Disk string
		diskGB := parseDiskStringToGB(instance.Disk)
		allocated.DiskGB += diskGB
	}

	return allocated, nil
}

// GetAvailable returns available resources (host - allocated)
func (t *Tracker) GetAvailable() (*AvailableResources, error) {
	host, err := t.GetHostResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get host resources: %w", err)
	}

	allocated, err := t.GetAllocated()
	if err != nil {
		return nil, fmt.Errorf("failed to get allocated resources: %w", err)
	}

	return t.calculateAvailable(host, allocated), nil
}

// calculateAvailable calculates available resources
func (t *Tracker) calculateAvailable(host *HostResources, allocated *AllocatedResources) *AvailableResources {
	availableCPU := host.CPU - allocated.CPU
	if availableCPU < 0 {
		availableCPU = 0
	}

	availableMemory := host.MemoryMB - allocated.MemoryMB
	if availableMemory < 0 {
		availableMemory = 0
	}

	availableDisk := host.DiskGB - allocated.DiskGB
	if availableDisk < 0 {
		availableDisk = 0
	}

	return &AvailableResources{
		CPU:      availableCPU,
		MemoryMB: availableMemory,
		DiskGB:   availableDisk,
	}
}

// Helper functions

// parseMemoryStringToMB parses memory string (e.g., "1.0GB", "512MB", "unlimited") to MB
func parseMemoryStringToMB(s string) int {
	s = strings.ToLower(strings.TrimSpace(s))

	if s == "" || s == "unlimited" {
		return 0
	}

	// Remove any whitespace
	s = strings.ReplaceAll(s, " ", "")

	// Extract numeric part and unit
	var value float64
	var unit string

	i := 0
	for i < len(s) && (s[i] >= '0' && s[i] <= '9' || s[i] == '.') {
		i++
	}

	if i > 0 {
		val, err := strconv.ParseFloat(s[:i], 64)
		if err != nil {
			return 0
		}
		value = val
	}

	if i < len(s) {
		unit = s[i:]
	}

	// Convert based on unit
	switch unit {
	case "b", "":
		return int(value / (1024 * 1024))
	case "kb", "k":
		return int(value / 1024)
	case "mb", "m":
		return int(value)
	case "gb", "g":
		return int(value * 1024)
	case "tb", "t":
		return int(value * 1024 * 1024)
	default:
		return int(value)
	}
}

// parseDiskStringToGB parses disk string (e.g., "10GB") to GB
func parseDiskStringToGB(s string) int {
	s = strings.ToLower(strings.TrimSpace(s))

	if s == "" {
		return 0
	}

	s = strings.ReplaceAll(s, " ", "")

	var value float64
	var unit string

	i := 0
	for i < len(s) && (s[i] >= '0' && s[i] <= '9' || s[i] == '.') {
		i++
	}

	if i > 0 {
		val, err := strconv.ParseFloat(s[:i], 64)
		if err != nil {
			return 0
		}
		value = val
	}

	if i < len(s) {
		unit = s[i:]
	}

	switch unit {
	case "b", "":
		return int(value / (1024 * 1024 * 1024))
	case "kb", "k":
		return int(value / (1024 * 1024))
	case "mb", "m":
		return int(value / 1024)
	case "gb", "g":
		return int(value)
	case "tb", "t":
		return int(value * 1024)
	default:
		return int(value)
	}
}

// parseSizeToGB parses size string to GB
func parseSizeToGB(s string) int {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")

	var value int
	var unit string

	i := 0
	for i < len(s) && (s[i] >= '0' && s[i] <= '9' || s[i] == '.') {
		i++
	}

	valueStr := s[:i]
	if i < len(s) {
		unit = s[i:]
	}

	val, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0
	}
	value = int(val)

	switch unit {
	case "b", "":
		return value / (1024 * 1024 * 1024)
	case "kb", "k":
		return value / (1024 * 1024)
	case "mb", "m":
		return value / 1024
	case "gb", "g":
		return value
	case "tb", "t":
		return value * 1024
	default:
		return value
	}
}

// formatMemory formats memory in MB to human readable string
func formatMemory(mb int) string {
	if mb >= 1024 {
		gb := mb / 1024
		return fmt.Sprintf("%dGB", gb)
	}
	return fmt.Sprintf("%dMB", mb)
}

// formatDisk formats disk in GB to human readable string
func formatDisk(gb int) string {
	if gb >= 1024 {
		tb := gb / 1024
		return fmt.Sprintf("%dTB", tb)
	}
	return fmt.Sprintf("%dGB", gb)
}
