package lxd

import (
	"fmt"
	"os"
	"sync"

	lxd "github.com/canonical/lxd/client"
)

// Client wraps LXD connection
type Client struct {
	mu             sync.RWMutex
	instanceServer lxd.InstanceServer
	client         lxd.InstanceServer // alias for backward compatibility
	connected      bool
}

// socketPaths contains common LXD socket paths to try
var socketPaths = []string{
	"/var/snap/lxd/common/lxd/unix.socket", // Snap installation
	"/var/lib/lxd/unix.socket",              // Native installation
	"/run/lxd/unix.socket",                  // Alternative path
	"/home/.config/lxc/unix.socket",         // User socket
}

// NewClient creates new LXD client connection via unix socket
// It tries multiple common socket paths until one succeeds
func NewClient() (*Client, error) {
	var lastErr error

	for _, socketPath := range socketPaths {
		// Expand environment variables in path
		expandedPath := os.ExpandEnv(socketPath)

		// Check if socket exists
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			continue
		}

		// Try to connect using ConnectLXD with remote "unix:/path"
		args := &lxd.ConnectionArgs{}
		instanceServer, err := lxd.ConnectLXD(fmt.Sprintf("unix:%s", expandedPath), args)
		if err != nil {
			lastErr = fmt.Errorf("failed to connect to LXD socket at %s: %w", expandedPath, err)
			continue
		}

		// Verify connection by getting server info
		_, _, err = instanceServer.GetServer()
		if err != nil {
			lastErr = fmt.Errorf("failed to verify LXD connection at %s: %w", expandedPath, err)
			continue
		}

		return &Client{
			instanceServer: instanceServer,
			client:         instanceServer,
			connected:      true,
		}, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("could not connect to LXD daemon: %w", lastErr)
	}

	return nil, fmt.Errorf("could not find LXD socket, tried: %v", socketPaths)
}

// NewClientWithPath creates new LXD client with specific socket path
func NewClientWithPath(socketPath string) (*Client, error) {
	expandedPath := os.ExpandEnv(socketPath)

	// Check if socket exists
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("LXD socket not found at %s", expandedPath)
	}

	// Connect to LXD
	args := &lxd.ConnectionArgs{}
	instanceServer, err := lxd.ConnectLXD(fmt.Sprintf("unix:%s", expandedPath), args)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LXD socket at %s: %w", expandedPath, err)
	}

	// Verify connection
	_, _, err = instanceServer.GetServer()
	if err != nil {
		return nil, fmt.Errorf("failed to verify LXD connection: %w", err)
	}

	return &Client{
		instanceServer: instanceServer,
		client:         instanceServer,
		connected:      true,
	}, nil
}

// InstanceServer returns the underlying instance server
func (c *Client) InstanceServer() lxd.InstanceServer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.instanceServer
}

// Close closes the LXD client connection
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.instanceServer != nil {
		c.instanceServer.Disconnect()
		c.connected = false
	}
}

// GetServerVersion returns the LXD server version
func (c *Client) GetServerVersion() (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return "", fmt.Errorf("not connected to LXD")
	}

	server, _, err := c.instanceServer.GetServer()
	if err != nil {
		return "", fmt.Errorf("failed to get server info: %w", err)
	}
	return server.Environment.ServerVersion, nil
}

// IsConnected checks if the client is still connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}
