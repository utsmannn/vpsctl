package lxd

import (
	"context"
	"fmt"
)

// NetworkInfo represents network information
type NetworkInfo struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	UsedBy []string `json:"used_by"`
}

// ListNetworks returns all networks
func (c *Client) ListNetworks(ctx context.Context) ([]NetworkInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to LXD")
	}

	networks, err := c.client.GetNetworks()
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	result := make([]NetworkInfo, 0, len(networks))
	for _, net := range networks {
		result = append(result, NetworkInfo{
			Name:   net.Name,
			Type:   net.Type,
			UsedBy: net.UsedBy,
		})
	}

	return result, nil
}
