package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// sizeUnits maps size suffixes to their byte multipliers
var sizeUnits = map[string]int64{
	"B":   1,
	"KB":  1024,
	"K":   1024,
	"MB":  1024 * 1024,
	"M":   1024 * 1024,
	"GB":  1024 * 1024 * 1024,
	"G":   1024 * 1024 * 1024,
	"TB":  1024 * 1024 * 1024 * 1024,
	"T":   1024 * 1024 * 1024 * 1024,
	"PB":  1024 * 1024 * 1024 * 1024 * 1024,
	"P":   1024 * 1024 * 1024 * 1024 * 1024,
	"EB":  1024 * 1024 * 1024 * 1024 * 1024 * 1024,
	"E":   1024 * 1024 * 1024 * 1024 * 1024 * 1024,
}

// ParseSize parses size string like "1GB", "512MB", "1024KB" to bytes
// Supports case-insensitive suffixes: B, K/KB, M/MB, G/GB, T/TB, P/PB, E/EB
// Also supports plain numbers (assumed to be bytes)
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Convert to uppercase for case-insensitive matching
	sUpper := strings.ToUpper(s)

	// Try to parse as plain number first (bytes)
	if val, err := strconv.ParseInt(s, 10, 64); err == nil {
		return val, nil
	}

	// Try each suffix from longest to shortest to avoid partial matches
	suffixes := []string{"KB", "MB", "GB", "TB", "PB", "EB", "K", "M", "G", "T", "P", "E", "B"}

	for _, suffix := range suffixes {
		if strings.HasSuffix(sUpper, suffix) {
			numStr := sUpper[:len(sUpper)-len(suffix)]
			num, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number in size string '%s': %w", s, err)
			}

			multiplier := sizeUnits[suffix]
			result := int64(num * float64(multiplier))

			if result < 0 {
				return 0, fmt.Errorf("size overflow: %s", s)
			}

			return result, nil
		}
	}

	return 0, fmt.Errorf("invalid size format: %s (expected format like 1GB, 512MB, etc)", s)
}

// FormatSize formats bytes to human readable string
// Uses the largest appropriate unit (B, KB, MB, GB, TB, PB, EB)
func FormatSize(bytes int64) string {
	if bytes < 0 {
		return "Invalid"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	// Format with 1 decimal place if not whole number
	value := float64(bytes) / float64(div)
	if value == float64(int64(value)) {
		return fmt.Sprintf("%d%cB", int64(value), "KMGTPE"[exp])
	}
	return fmt.Sprintf("%.1f%cB", value, "KMGTPE"[exp])
}

// ParseMemoryToMB parses memory string to megabytes
// Returns the value in MB (rounded to integer)
func ParseMemoryToMB(s string) (int, error) {
	bytes, err := ParseSize(s)
	if err != nil {
		return 0, err
	}

	// Convert to MB
	mb := bytes / (1024 * 1024)

	// Round up if there's a remainder
	if bytes%(1024*1024) > 0 {
		mb++
	}

	return int(mb), nil
}

// ParseDiskToGB parses disk string to gigabytes
// Returns the value in GB (rounded to integer)
func ParseDiskToGB(s string) (int, error) {
	bytes, err := ParseSize(s)
	if err != nil {
		return 0, err
	}

	// Convert to GB
	gb := bytes / (1024 * 1024 * 1024)

	// Round up if there's a remainder
	if bytes%(1024*1024*1024) > 0 {
		gb++
	}

	return int(gb), nil
}

// MustParseSize is like ParseSize but panics on error
// Use only for hardcoded values that are known to be valid
func MustParseSize(s string) int64 {
	val, err := ParseSize(s)
	if err != nil {
		panic(fmt.Sprintf("MustParseSize(%q): %v", s, err))
	}
	return val
}

// ParseSizeOrDefault parses size string, returns default value on error
func ParseSizeOrDefault(s string, defaultValue int64) int64 {
	val, err := ParseSize(s)
	if err != nil {
		return defaultValue
	}
	return val
}

// ValidateSize validates that a size string is within min and max bounds (in bytes)
func ValidateSize(s string, minBytes, maxBytes int64) error {
	bytes, err := ParseSize(s)
	if err != nil {
		return err
	}

	if bytes < minBytes {
		return fmt.Errorf("size %s is below minimum %s", s, FormatSize(minBytes))
	}

	if bytes > maxBytes {
		return fmt.Errorf("size %s exceeds maximum %s", s, FormatSize(maxBytes))
	}

	return nil
}

// BytesToMB converts bytes to megabytes
func BytesToMB(bytes int64) int {
	return int(bytes / (1024 * 1024))
}

// BytesToGB converts bytes to gigabytes
func BytesToGB(bytes int64) int {
	return int(bytes / (1024 * 1024 * 1024))
}

// MBToBytes converts megabytes to bytes
func MBToBytes(mb int) int64 {
	return int64(mb) * 1024 * 1024
}

// GBToBytes converts gigabytes to bytes
func GBToBytes(gb int) int64 {
	return int64(gb) * 1024 * 1024 * 1024
}
