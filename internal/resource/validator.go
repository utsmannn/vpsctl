package resource

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// validNameRegex validates instance names (alphanumeric, hyphens, underscores)
var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidationResult contains validation result
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`
}

// ValidateRequest validates resource request against available resources
// Returns a list of validation messages (empty if all valid)
func ValidateRequest(reqCPU int, reqMemMB int, reqDiskGB int, avail *AvailableResources) []string {
	var messages []string

	if reqCPU > avail.CPU {
		messages = append(messages, fmt.Sprintf("Insufficient CPU: requested %d, available %d", reqCPU, avail.CPU))
	}

	if reqMemMB > avail.MemoryMB {
		messages = append(messages, fmt.Sprintf("Insufficient memory: requested %dMB, available %dMB", reqMemMB, avail.MemoryMB))
	}

	if reqDiskGB > avail.DiskGB {
		messages = append(messages, fmt.Sprintf("Insufficient disk: requested %dGB, available %dGB", reqDiskGB, avail.DiskGB))
	}

	// Warnings for high utilization
	if avail.CPU > 0 && reqCPU > 0 {
		utilization := float64(reqCPU) / float64(avail.CPU) * 100
		if utilization > 80 {
			messages = append(messages, fmt.Sprintf("Warning: CPU allocation will be at %.0f%% of available", utilization))
		}
	}

	if avail.MemoryMB > 0 && reqMemMB > 0 {
		utilization := float64(reqMemMB) / float64(avail.MemoryMB) * 100
		if utilization > 80 {
			messages = append(messages, fmt.Sprintf("Warning: Memory allocation will be at %.0f%% of available", utilization))
		}
	}

	if avail.DiskGB > 0 && reqDiskGB > 0 {
		utilization := float64(reqDiskGB) / float64(avail.DiskGB) * 100
		if utilization > 80 {
			messages = append(messages, fmt.Sprintf("Warning: Disk allocation will be at %.0f%% of available", utilization))
		}
	}

	return messages
}

// CanAllocate checks if resources can be allocated
func CanAllocate(reqCPU int, reqMemMB int, reqDiskGB int, avail *AvailableResources) bool {
	return reqCPU <= avail.CPU && reqMemMB <= avail.MemoryMB && reqDiskGB <= avail.DiskGB
}

// Validate validates resource request with string inputs
// memory and disk can be strings like "1GB", "512MB", etc.
func Validate(cpu int, memory, disk string, avail *AvailableResources) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Warnings: []string{},
		Errors:   []string{},
	}

	// Validate CPU
	if cpu <= 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "CPU must be greater than 0")
	}

	if cpu > avail.CPU {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Insufficient CPU: requested %d, available %d", cpu, avail.CPU))
	} else if avail.CPU > 0 {
		utilization := float64(cpu) / float64(avail.CPU) * 100
		if utilization > 80 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("CPU allocation will be at %.0f%% of available", utilization))
		}
	}

	// Parse and validate memory
	memMB, err := parseMemoryString(memory)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid memory format: %s", memory))
	} else if memMB <= 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "Memory must be greater than 0")
	} else if memMB > avail.MemoryMB {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Insufficient memory: requested %s (%dMB), available %dMB", memory, memMB, avail.MemoryMB))
	} else if avail.MemoryMB > 0 {
		utilization := float64(memMB) / float64(avail.MemoryMB) * 100
		if utilization > 80 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Memory allocation will be at %.0f%% of available", utilization))
		}
	}

	// Parse and validate disk
	diskGB, err := parseDiskString(disk)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid disk format: %s", disk))
	} else if diskGB <= 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "Disk must be greater than 0")
	} else if diskGB > avail.DiskGB {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Insufficient disk: requested %s (%dGB), available %dGB", disk, diskGB, avail.DiskGB))
	} else if avail.DiskGB > 0 {
		utilization := float64(diskGB) / float64(avail.DiskGB) * 100
		if utilization > 80 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Disk allocation will be at %.0f%% of available", utilization))
		}
	}

	// Minimum resource recommendations
	if memMB > 0 && memMB < 256 {
		result.Warnings = append(result.Warnings, "Memory less than 256MB may cause instability")
	}

	if diskGB > 0 && diskGB < 5 {
		result.Warnings = append(result.Warnings, "Disk less than 5GB may be insufficient for most workloads")
	}

	return result, nil
}

// parseMemoryString parses memory string (e.g., "1GB", "512MB") to MB
func parseMemoryString(s string) (int, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")

	if s == "" {
		return 0, fmt.Errorf("empty memory string")
	}

	// Extract numeric part
	i := 0
	for i < len(s) && (s[i] >= '0' && s[i] <= '9' || s[i] == '.') {
		i++
	}

	valueStr := s[:i]
	unit := ""
	if i < len(s) {
		unit = s[i:]
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %s", valueStr)
	}

	// Convert based on unit
	switch unit {
	case "", "b":
		return int(value / (1024 * 1024)), nil
	case "kb", "k":
		return int(value / 1024), nil
	case "mb", "m":
		return int(value), nil
	case "gb", "g":
		return int(value * 1024), nil
	case "tb", "t":
		return int(value * 1024 * 1024), nil
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
}

// parseDiskString parses disk string (e.g., "10GB", "1TB") to GB
func parseDiskString(s string) (int, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")

	if s == "" {
		return 0, fmt.Errorf("empty disk string")
	}

	// Extract numeric part
	i := 0
	for i < len(s) && (s[i] >= '0' && s[i] <= '9' || s[i] == '.') {
		i++
	}

	valueStr := s[:i]
	unit := ""
	if i < len(s) {
		unit = s[i:]
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %s", valueStr)
	}

	// Convert based on unit
	switch unit {
	case "", "b":
		return int(value / (1024 * 1024 * 1024)), nil
	case "kb", "k":
		return int(value / (1024 * 1024)), nil
	case "mb", "m":
		return int(value / 1024), nil
	case "gb", "g":
		return int(value), nil
	case "tb", "t":
		return int(value * 1024), nil
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
}

// ValidateInstanceName validates instance name
func ValidateInstanceName(name string) error {
	if name == "" {
		return fmt.Errorf("instance name cannot be empty")
	}

	if len(name) > 63 {
		return fmt.Errorf("instance name cannot exceed 63 characters")
	}

	if !validNameRegex.MatchString(name) {
		return fmt.Errorf("instance name can only contain alphanumeric characters, hyphens, and underscores")
	}

	// Check for reserved names
	reserved := []string{"root", "admin", "system", "default", "lxd"}
	nameLower := strings.ToLower(name)
	for _, r := range reserved {
		if nameLower == r {
			return fmt.Errorf("'%s' is a reserved name", name)
		}
	}

	return nil
}

// QuickValidate performs quick validation without available resources
func QuickValidate(cpu int, memory, disk string) error {
	if cpu <= 0 {
		return fmt.Errorf("CPU must be greater than 0")
	}

	memMB, err := parseMemoryString(memory)
	if err != nil {
		return fmt.Errorf("invalid memory: %w", err)
	}
	if memMB <= 0 {
		return fmt.Errorf("memory must be greater than 0")
	}

	diskGB, err := parseDiskString(disk)
	if err != nil {
		return fmt.Errorf("invalid disk: %w", err)
	}
	if diskGB <= 0 {
		return fmt.Errorf("disk must be greater than 0")
	}

	return nil
}
