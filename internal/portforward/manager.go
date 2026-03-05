package portforward

import (
	"fmt"
	"sync"

	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/sirupsen/logrus"
)

// PortForward represents a port forwarding configuration
type PortForward struct {
	HostPort      int    `json:"host_port"`
	ContainerPort int    `json:"container_port"`
	Protocol      string `json:"protocol"`
	Label         string `json:"label"`
	Status        string `json:"status"`
}

// Manager manages port forwarding operations
type Manager struct {
	lxdClient *lxd.Client
	forwards  map[string][]PortForward // instance name -> forwards
	mu        sync.RWMutex
}

// NewManager creates a new port forward manager
func NewManager(client *lxd.Client) *Manager {
	return &Manager{
		lxdClient: client,
		forwards:  make(map[string][]PortForward),
	}
}

// List returns all port forwards for an instance
func (m *Manager) List(instanceName string) ([]PortForward, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if instanceName == "" {
		return nil, fmt.Errorf("instance name is required")
	}

	// Get instance to verify it exists
	_, err := m.lxdClient.GetInstance(instanceName)
	if err != nil {
		return nil, fmt.Errorf("instance not found: %w", err)
	}

	forwards, ok := m.forwards[instanceName]
	if ok {
		return forwards, nil
	}

	return []PortForward{}, nil
}

// Add creates a new port forward
func (m *Manager) Add(instanceName string, hostPort, containerPort int, protocol string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if instanceName == "" {
		return fmt.Errorf("instance name is required")
	}

	// Validate inputs
	if hostPort <= 0 || hostPort > 65535 {
		return fmt.Errorf("invalid host port: must be between 1 and 65535")
	}
	if containerPort <= 0 || containerPort > 65535 {
		return fmt.Errorf("invalid container port: must be between 1 and 65535")
	}
	if protocol == "" {
		protocol = "tcp"
	}
	if protocol != "tcp" && protocol != "udp" {
		return fmt.Errorf("protocol must be tcp or udp")
	}

	// Get instance IP
	_, err := m.lxdClient.GetInstanceIP(instanceName)
	if err != nil {
		return fmt.Errorf("could not get instance IP: %w", err)
	}

	// Check if host port is already in use
	for _, forwards := range m.forwards {
		for _, f := range forwards {
			if f.HostPort == hostPort {
				return fmt.Errorf("host port %d is already in use", hostPort)
			}
		}
	}

	// Create port forward
	pf := PortForward{
		HostPort:      hostPort,
		ContainerPort: containerPort,
		Protocol:      protocol,
		Label:         fmt.Sprintf("%s:%d->%d", instanceName, hostPort, containerPort),
		Status:        "active",
	}

	// Initialize list if needed
	if m.forwards[instanceName] == nil {
		m.forwards[instanceName] = make([]PortForward, 0)
	}
	m.forwards[instanceName] = append(m.forwards[instanceName], pf)

	// Configure iptables rule (this would be executed externally)
	// For now, just store the configuration
	logrus.Infof("Added port forward: host=%d -> container %s:%d (%s)", hostPort, instanceName, containerPort, protocol)

	return nil
}

// Remove removes a port forward
func (m *Manager) Remove(instanceName string, hostPort int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if instanceName == "" {
		return fmt.Errorf("instance name is required")
	}

	forwards, ok := m.forwards[instanceName]
	if !ok {
		return fmt.Errorf("no port forwards found for instance %s", instanceName)
	}

	// Find and remove the port forward
	var found bool
	for i, pf := range forwards {
		if pf.HostPort == hostPort {
			m.forwards[instanceName] = append(forwards[:i], forwards[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("port forward %d not found for instance %s", hostPort, instanceName)
	}

	// Remove iptables rule (this would be executed externally)
	logrus.Infof("Removed port forward: host=%d for instance %s", hostPort, instanceName)

	return nil
}

// ScanListeningPorts scans the instance for listening ports
func (m *Manager) ScanListeningPorts(instanceName string) ([]int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if instanceName == "" {
		return nil, fmt.Errorf("instance name is required")
	}

	// Get instance state
	state, err := m.lxdClient.GetInstanceState(instanceName)
	if err != nil {
		return nil, fmt.Errorf("could not get instance state: %w", err)
	}

	// Extract listening ports from network state
	// This is a simplified version - in production you would actually
	// execute 'netstat' or 'ss' inside the container
	listeningPorts := make([]int, 0)

	if state.Network != nil {
		// For now, return common ports
		// In production, you would exec into container
		commonPorts := []int{
			22,     // SSH
			80,     // HTTP
			443,    // HTTPS
			3306,   // MySQL
			5432,   // PostgreSQL
			6379,   // Redis
			27017,  // MongoDB
			8080,   // HTTP Alternate
		}

		// Check if instance is running
		if state.Status == "Running" {
			listeningPorts = commonPorts
		}
	}

	return listeningPorts, nil
}
