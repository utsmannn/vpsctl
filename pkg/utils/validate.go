package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateName validates instance name
// Rules:
// - Must be between 1 and 63 characters
// - Must start with a letter
// - Can only contain lowercase letters, numbers, and hyphens
// - Cannot end with a hyphen
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("instance name cannot be empty")
	}

	if len(name) > 63 {
		return fmt.Errorf("instance name cannot be longer than 63 characters")
	}

	// Must start with a letter
	if !regexp.MustCompile(`^[a-z]`).MatchString(name) {
		return fmt.Errorf("instance name must start with a lowercase letter")
	}

	// Can only contain lowercase letters, numbers, and hyphens
	validName := regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("instance name can only contain lowercase letters, numbers, and hyphens")
	}

	// Cannot end with a hyphen
	if strings.HasSuffix(name, "-") {
		return fmt.Errorf("instance name cannot end with a hyphen")
	}

	// Reserved names
	reservedNames := []string{"root", "admin", "test", "localhost", "all"}
	for _, reserved := range reservedNames {
		if strings.EqualFold(name, reserved) {
			return fmt.Errorf("'%s' is a reserved name", name)
		}
	}

	return nil
}

// ValidateImage validates image string format
// Format: <distribution>:<version> or <alias>
// Examples: ubuntu:24.04, alpine:3.19, images:ubuntu/22.04
func ValidateImage(image string) error {
	if image == "" {
		return fmt.Errorf("image cannot be empty")
	}

	// Check for valid image format
	// Format 1: distribution:version (e.g., ubuntu:24.04)
	// Format 2: remote:image (e.g., images:ubuntu/22.04)
	// Format 3: simple alias (e.g., ubuntu)

	validImage := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]*(:[a-zA-Z0-9][a-zA-Z0-9._/-]*)?$`)
	if !validImage.MatchString(image) {
		return fmt.Errorf("invalid image format: %s (expected format: distribution:version or alias)", image)
	}

	return nil
}

// ValidateCPU validates CPU count
func ValidateCPU(cpu int) error {
	if cpu < 1 {
		return fmt.Errorf("CPU count must be at least 1")
	}
	if cpu > 128 {
		return fmt.Errorf("CPU count cannot exceed 128")
	}
	return nil
}

// ValidateMemory validates memory size string
func ValidateMemory(memory string) error {
	if memory == "" {
		return fmt.Errorf("memory cannot be empty")
	}

	_, err := ParseSize(memory)
	if err != nil {
		return fmt.Errorf("invalid memory size: %w", err)
	}

	return nil
}

// ValidateDisk validates disk size string
func ValidateDisk(disk string) error {
	if disk == "" {
		return fmt.Errorf("disk size cannot be empty")
	}

	size, err := ParseSize(disk)
	if err != nil {
		return fmt.Errorf("invalid disk size: %w", err)
	}

	// Minimum disk size: 1GB
	minDisk := int64(1024 * 1024 * 1024)
	if size < minDisk {
		return fmt.Errorf("disk size must be at least 1GB")
	}

	return nil
}

// ValidateInstanceType validates instance type
func ValidateInstanceType(instanceType string) error {
	validTypes := []string{"container", "vm"}

	for _, t := range validTypes {
		if strings.EqualFold(instanceType, t) {
			return nil
		}
	}

	return fmt.Errorf("invalid instance type: %s (must be 'container' or 'vm')", instanceType)
}

// ValidatePort validates port number
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}

// ValidateSSHKey validates SSH public key
func ValidateSSHKey(key string) error {
	if key == "" {
		return nil // Optional field
	}

	key = strings.TrimSpace(key)

	// Check for common SSH key prefixes
	validPrefixes := []string{
		"ssh-rsa",
		"ssh-ed25519",
		"ssh-dss",
		"ecdsa-sha2-nistp256",
		"ecdsa-sha2-nistp384",
		"ecdsa-sha2-nistp521",
	}

	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(key, prefix) {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		return fmt.Errorf("invalid SSH key format (expected ssh-rsa, ssh-ed25519, or ecdsa)")
	}

	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if password == "" {
		return nil // Optional field
	}

	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}

	return nil
}
