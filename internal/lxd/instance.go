package lxd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	lxd "github.com/canonical/lxd/client"
	lxdapi "github.com/canonical/lxd/shared/api"
)

// InstanceInfo represents VPS instance information
type InstanceInfo struct {
	Name       string
	Status     string // RUNNING, STOPPED, etc
	Type       string // container or virtual-machine
	CPU        int
	Memory     string
	Disk       string
	IP         string
	CreatedAt  time.Time
}

// CreateInstanceOptions for creating new instance
type CreateInstanceOptions struct {
	Name       string
	Image      string // e.g. "ubuntu:24.04"
	CPU        int
	Memory     string // e.g. "1GB"
	Disk       string // e.g. "10GB"
	Type       string // "container" or "virtual-machine"
	SSHKey     string // optional SSH public key
	Password   string // optional root password
}

// CreateInstance creates a new VPS instance
func (c *Client) CreateInstance(opts CreateInstanceOptions) (*InstanceInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to LXD")
	}

	if opts.Name == "" {
		return nil, fmt.Errorf("instance name is required")
	}
	if opts.Image == "" {
		return nil, fmt.Errorf("image is required")
	}

	// Set default values
	if opts.CPU <= 0 {
		opts.CPU = 1
	}
	if opts.Memory == "" {
		opts.Memory = "512MB"
	}
	if opts.Disk == "" {
		opts.Disk = "10GB"
	}
	if opts.Type == "" {
		opts.Type = "container"
	}

	// Parse memory and disk sizes
	memoryBytes, err := parseSizeToBytes(opts.Memory)
	if err != nil {
		return nil, fmt.Errorf("invalid memory size: %w", err)
	}

	diskBytes, err := parseSizeToBytes(opts.Disk)
	if err != nil {
		return nil, fmt.Errorf("invalid disk size: %w", err)
	}

	// Create instance request
	req := lxdapi.InstancesPost{
		Name: opts.Name,
		Source: lxdapi.InstanceSource{
			Type:        "image",
			Alias:       opts.Image,
			Server:      "",
			Protocol:    "simplestreams",
		},
		Type: lxdapi.InstanceType(opts.Type),
	}

	// Set instance config
	req.Config = map[string]string{
		"limits.cpu":           fmt.Sprintf("%d", opts.CPU),
		"limits.memory":        fmt.Sprintf("%d", memoryBytes),
		"security.privileged":  "false",
	}

	// Set root disk size
	req.Devices = map[string]map[string]string{
		"root": {
			"type": "disk",
			"path": "/",
			"size": fmt.Sprintf("%dB", diskBytes),
		},
	}

	// Add SSH key if provided
	if opts.SSHKey != "" {
		req.Config["cloud-init.user-data"] = fmt.Sprintf(`#cloud-config
ssh_authorized_keys:
  - %s
`, opts.SSHKey)
	}

	// Add password if provided
	if opts.Password != "" {
		if req.Config["cloud-init.user-data"] == "" {
			req.Config["cloud-init.user-data"] = "#cloud-config\n"
		}
		req.Config["cloud-init.user-data"] += fmt.Sprintf(`
chpasswd:
  list: |
    root:%s
  expire: false
`, opts.Password)
	}

	// Create the instance
	op, err := c.instanceServer.CreateInstance(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	// Wait for creation to complete with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err = op.WaitContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed waiting for instance creation: %w", err)
	}

	// Start the instance
	c.mu.Unlock()
	err = c.startInstanceInternal(opts.Name)
	c.mu.Lock()
	if err != nil {
		return nil, fmt.Errorf("instance created but failed to start: %w", err)
	}

	// Get instance info
	return c.getInstanceInternal(opts.Name)
}

// ListInstances lists all instances
func (c *Client) ListInstances() ([]InstanceInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to LXD")
	}

	instances, err := c.instanceServer.GetInstances(lxd.GetInstancesArgs{})
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	var result []InstanceInfo
	for _, inst := range instances {
		info := convertToInstanceInfo(inst)
		result = append(result, info)
	}

	return result, nil
}

// GetInstance gets instance by name
func (c *Client) GetInstance(name string) (*InstanceInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.getInstanceInternal(name)
}

// getInstanceInternal is the internal implementation without locking
func (c *Client) getInstanceInternal(name string) (*InstanceInfo, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to LXD")
	}

	if name == "" {
		return nil, fmt.Errorf("instance name is required")
	}

	inst, _, err := c.instanceServer.GetInstance(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance %s: %w", name, err)
	}

	info := convertToInstanceInfo(*inst)
	return &info, nil
}

// StartInstance starts an instance
func (c *Client) StartInstance(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.startInstanceInternal(name)
}

// startInstanceInternal is the internal implementation without locking
func (c *Client) startInstanceInternal(name string) error {
	if !c.connected {
		return fmt.Errorf("not connected to LXD")
	}

	if name == "" {
		return fmt.Errorf("instance name is required")
	}

	req := lxdapi.InstanceStatePut{
		Action:  "start",
		Timeout: 30,
	}

	op, err := c.instanceServer.UpdateInstanceState(name, req, "")
	if err != nil {
		return fmt.Errorf("failed to start instance %s: %w", name, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = op.WaitContext(ctx)
	if err != nil {
		return fmt.Errorf("failed waiting for instance %s to start: %w", name, err)
	}

	return nil
}

// StopInstance stops an instance
func (c *Client) StopInstance(name string, force bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return fmt.Errorf("not connected to LXD")
	}

	if name == "" {
		return fmt.Errorf("instance name is required")
	}

	req := lxdapi.InstanceStatePut{
		Action:  "stop",
		Timeout: 30,
		Force:   force,
	}

	op, err := c.instanceServer.UpdateInstanceState(name, req, "")
	if err != nil {
		return fmt.Errorf("failed to stop instance %s: %w", name, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = op.WaitContext(ctx)
	if err != nil {
		return fmt.Errorf("failed waiting for instance %s to stop: %w", name, err)
	}

	return nil
}

// RestartInstance restarts an instance
func (c *Client) RestartInstance(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return fmt.Errorf("not connected to LXD")
	}

	if name == "" {
		return fmt.Errorf("instance name is required")
	}

	req := lxdapi.InstanceStatePut{
		Action:  "restart",
		Timeout: 30,
	}

	op, err := c.instanceServer.UpdateInstanceState(name, req, "")
	if err != nil {
		return fmt.Errorf("failed to restart instance %s: %w", name, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = op.WaitContext(ctx)
	if err != nil {
		return fmt.Errorf("failed waiting for instance %s to restart: %w", name, err)
	}

	return nil
}

// DeleteInstance deletes an instance
func (c *Client) DeleteInstance(name string, force bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return fmt.Errorf("not connected to LXD")
	}

	if name == "" {
		return fmt.Errorf("instance name is required")
	}

	// If force is true, stop the instance first
	if force {
		c.mu.Unlock()
		_ = c.StopInstance(name, true)
		c.mu.Lock()
	}

	op, err := c.instanceServer.DeleteInstance(name, false)
	if err != nil {
		return fmt.Errorf("failed to delete instance %s: %w", name, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = op.WaitContext(ctx)
	if err != nil {
		return fmt.Errorf("failed waiting for instance %s to delete: %w", name, err)
	}

	return nil
}

// ResizeInstance resizes instance resources
func (c *Client) ResizeInstance(name string, cpu int, memory, disk string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return fmt.Errorf("not connected to LXD")
	}

	if name == "" {
		return fmt.Errorf("instance name is required")
	}

	// Get current instance
	inst, etag, err := c.instanceServer.GetInstance(name)
	if err != nil {
		return fmt.Errorf("failed to get instance %s: %w", name, err)
	}

	// Update config
	if inst.Config == nil {
		inst.Config = make(map[string]string)
	}

	if cpu > 0 {
		inst.Config["limits.cpu"] = fmt.Sprintf("%d", cpu)
	}

	if memory != "" {
		memoryBytes, err := parseSizeToBytes(memory)
		if err != nil {
			return fmt.Errorf("invalid memory size: %w", err)
		}
		inst.Config["limits.memory"] = fmt.Sprintf("%d", memoryBytes)
	}

	// Update disk if specified
	if disk != "" {
		diskBytes, err := parseSizeToBytes(disk)
		if err != nil {
			return fmt.Errorf("invalid disk size: %w", err)
		}

		if inst.Devices == nil {
			inst.Devices = make(map[string]map[string]string)
		}

		if rootDev, ok := inst.Devices["root"]; ok {
			rootDev["size"] = fmt.Sprintf("%dB", diskBytes)
			inst.Devices["root"] = rootDev
		} else {
			inst.Devices["root"] = map[string]string{
				"type": "disk",
				"path": "/",
				"size": fmt.Sprintf("%dB", diskBytes),
			}
		}
	}

	// Update instance
	op, err := c.instanceServer.UpdateInstance(name, inst.Writable(), etag)
	if err != nil {
		return fmt.Errorf("failed to resize instance %s: %w", name, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = op.WaitContext(ctx)
	if err != nil {
		return fmt.Errorf("failed waiting for instance %s resize: %w", name, err)
	}

	return nil
}

// GetInstanceState gets the current state of an instance
func (c *Client) GetInstanceState(name string) (*lxdapi.InstanceState, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to LXD")
	}

	if name == "" {
		return nil, fmt.Errorf("instance name is required")
	}

	state, _, err := c.instanceServer.GetInstanceState(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance %s state: %w", name, err)
	}

	return state, nil
}

// convertToInstanceInfo converts LXD instance to InstanceInfo
func convertToInstanceInfo(inst lxdapi.Instance) InstanceInfo {
	info := InstanceInfo{
		Name:      inst.Name,
		Status:    strings.ToUpper(inst.Status),
		Type:      string(inst.Type),
		CreatedAt: inst.CreatedAt,
	}

	// Parse CPU
	if cpu, ok := inst.Config["limits.cpu"]; ok {
		info.CPU, _ = strconv.Atoi(cpu)
	} else {
		info.CPU = 1 // Default
	}

	// Parse Memory
	if mem, ok := inst.Config["limits.memory"]; ok {
		info.Memory = formatMemory(mem)
	} else {
		info.Memory = "unlimited"
	}

	// Parse Disk
	if devices, ok := inst.Devices["root"]; ok {
		if size, ok := devices["size"]; ok {
			info.Disk = size
		}
	}

	// Get IP address from state (this is approximate, actual IP requires state call)
	// For now, we'll set it to empty as it requires a separate API call
	info.IP = ""

	return info
}

// GetInstanceIP gets the IP address of an instance
func (c *Client) GetInstanceIP(name string) (string, error) {
	state, err := c.GetInstanceState(name)
	if err != nil {
		return "", err
	}

	for _, network := range state.Network {
		for _, addr := range network.Addresses {
			if addr.Family == "inet" && addr.Scope == "global" {
				return addr.Address, nil
			}
		}
	}

	return "", fmt.Errorf("no IP address found for instance %s", name)
}

// parseSizeToBytes parses size string to bytes (internal helper)
func parseSizeToBytes(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Handle pure number (assume bytes)
	if val, err := strconv.ParseInt(s, 10, 64); err == nil {
		return val, nil
	}

	// Define multipliers
	multipliers := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
		"PB": 1024 * 1024 * 1024 * 1024 * 1024,
	}

	for suffix, mult := range multipliers {
		if strings.HasSuffix(s, suffix) {
			numStr := strings.TrimSuffix(s, suffix)
			num, err := strconv.ParseInt(strings.TrimSpace(numStr), 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number in size string: %s", s)
			}
			return num * mult, nil
		}
	}

	return 0, fmt.Errorf("invalid size format: %s", s)
}

// formatMemory formats memory value to human readable
func formatMemory(mem string) string {
	// If it's already in bytes (just a number), convert to human readable
	if val, err := strconv.ParseInt(mem, 10, 64); err == nil {
		return formatBytes(val)
	}
	// Otherwise return as is (might already be formatted)
	return mem
}

// formatBytes formats bytes to human readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
