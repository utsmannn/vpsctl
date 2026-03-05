package lxd

import (
	"context"
	"fmt"

	lxdapi "github.com/canonical/lxd/shared/api"
)

// ImageInfo represents an LXD image
type ImageInfo struct {
	Fingerprint string `json:"fingerprint"`
	Alias       string `json:"alias"`
	Description string `json:"description"`
	Architecture string `json:"architecture"`
	Type        string `json:"type"`
	Size        int64  `json:"size"`
}

// ListImages returns all available images
func (c *Client) ListImages(ctx context.Context) ([]ImageInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to LXD")
	}

	images, err := c.instanceServer.GetImages()
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	result := make([]ImageInfo, 0, len(images))
	for _, img := range images {
		info := ImageInfo{
			Fingerprint:  img.Fingerprint,
			Architecture: img.Architecture,
			Type:         img.Type,
			Size:         img.Size,
		}

		// Get first alias if available
		if len(img.Aliases) > 0 {
			info.Alias = img.Aliases[0].Name
		}

		// Get description from properties
		if img.Properties != nil {
			if desc, ok := img.Properties["description"]; ok {
				info.Description = desc
			}
		}

		result = append(result, info)
	}

	return result, nil
}

// GetImage returns a specific image
func (c *Client) GetImage(ctx context.Context, fingerprint string) (*ImageInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to LXD")
	}

	img, _, err := c.instanceServer.GetImage(fingerprint)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	info := &ImageInfo{
		Fingerprint:  img.Fingerprint,
		Architecture: img.Architecture,
		Type:         img.Type,
		Size:         img.Size,
	}

	if len(img.Aliases) > 0 {
		info.Alias = img.Aliases[0].Name
	}

	if img.Properties != nil {
		if desc, ok := img.Properties["description"]; ok {
			info.Description = desc
		}
	}

	return info, nil
}

// GetImageAliases returns all image aliases
func (c *Client) GetImageAliases(ctx context.Context) ([]lxdapi.ImageAliasesEntry, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to LXD")
	}

	aliases, err := c.instanceServer.GetImageAliases()
	if err != nil {
		return nil, fmt.Errorf("failed to get image aliases: %w", err)
	}

	return aliases, nil
}

// CopyImage copies an image from a remote server
func (c *Client) CopyImage(ctx context.Context, source string, aliases []string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return fmt.Errorf("not connected to LXD")
	}

	// Parse source (e.g., "ubuntu:24.04")
	// This would need to connect to a remote server and copy the image
	// For now, we'll return an error as this requires more complex implementation
	return fmt.Errorf("image copying not yet implemented")
}

// DeleteImage deletes an image
func (c *Client) DeleteImage(ctx context.Context, fingerprint string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return fmt.Errorf("not connected to LXD")
	}

	op, err := c.instanceServer.DeleteImage(fingerprint)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return op.WaitContext(ctx)
}
