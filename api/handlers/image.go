package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kiatkoding/vpsctl/internal/lxd"
)

// ImageInfo represents an image in API responses
type ImageInfo struct {
	Fingerprint  string `json:"fingerprint"`
	Alias        string `json:"alias"`
	Description  string `json:"description"`
	Architecture string `json:"architecture"`
	Type         string `json:"type"`
	Size         int64  `json:"size"`
	SizeHuman    string `json:"size_human"`
}

// ListImages returns all available images
func ListImages(client *lxd.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		images, err := client.InstanceServer().GetImages()
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "list_failed",
				Message: err.Error(),
			})
			return
		}

		result := make([]ImageInfo, 0, len(images))
		for _, img := range images {
			info := ImageInfo{
				Fingerprint:  img.Fingerprint,
				Architecture: img.Architecture,
				Type:         img.Type,
				Size:         img.Size,
				SizeHuman:     formatBytes(img.Size),
			}

			// Get first alias if available
			if len(img.Aliases) > 0 {
				info.Alias = img.Aliases[0].Name
			}

			// Get description from properties
			if img.Properties != nil {
				var parts []string
				if os, ok := img.Properties["os"]; ok {
					parts = append(parts, os)
				}
				if release, ok := img.Properties["release"]; ok {
					parts = append(parts, release)
				}
				if variant, ok := img.Properties["variant"]; ok {
					parts = append(parts, variant)
				}
				if len(parts) > 0 {
					info.Description = strings.Join(parts, " ")
				}
				if desc, ok := img.Properties["description"]; ok {
					info.Description = desc
				}
			}

			result = append(result, info)
		}

		c.JSON(http.StatusOK, gin.H{
			"images": result,
			"count":  len(result),
		})
	}
}

// GetImage returns a specific image
func GetImage(client *lxd.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		fingerprint := c.Param("fingerprint")
		if fingerprint == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_fingerprint",
				Message: "Image fingerprint is required",
			})
			return
		}

		img, _, err := client.InstanceServer().GetImage(fingerprint)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: err.Error(),
			})
			return
		}

		info := ImageInfo{
			Fingerprint:  img.Fingerprint,
			Architecture: img.Architecture,
			Type:         img.Type,
			Size:         img.Size,
			SizeHuman:     formatBytes(img.Size),
		}

		if len(img.Aliases) > 0 {
			info.Alias = img.Aliases[0].Name
		}

		if img.Properties != nil {
			var parts []string
			if os, ok := img.Properties["os"]; ok {
				parts = append(parts, os)
			}
			if release, ok := img.Properties["release"]; ok {
				parts = append(parts, release)
			}
			if variant, ok := img.Properties["variant"]; ok {
				parts = append(parts, variant)
			}
			if len(parts) > 0 {
				info.Description = strings.Join(parts, " ")
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"image": info,
		})
	}
}

// ListStoragePools returns all storage pools
func ListStoragePools(client *lxd.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		pools, err := client.InstanceServer().GetStoragePools()
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "list_failed",
				Message: err.Error(),
			})
			return
		}

		type StoragePoolResponse struct {
			Name        string `json:"name"`
			Driver      string `json:"driver"`
			TotalBytes  int64  `json:"total_bytes"`
			UsedBytes   int64  `json:"used_bytes"`
			TotalHuman  string `json:"total_human"`
			UsedHuman   string `json:"used_human"`
		}

		result := make([]StoragePoolResponse, 0, len(pools))
		for _, pool := range pools {
			// Note: pool.Resources requires separate API call to GetStoragePoolResources
			// For now, just show basic info
			result = append(result, StoragePoolResponse{
				Name:       pool.Name,
				Driver:     pool.Driver,
				TotalBytes: 0,
				UsedBytes:  0,
				TotalHuman: "-",
				UsedHuman:  "-",
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"pools": result,
			"count": len(result),
		})
	}
}

// ListNetworks returns all networks
func ListNetworks(client *lxd.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		networks, err := client.InstanceServer().GetNetworks()
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "list_failed",
				Message: err.Error(),
			})
			return
		}

		type NetworkResponse struct {
			Name   string   `json:"name"`
			Type   string   `json:"type"`
			UsedBy []string `json:"used_by"`
		}

		result := make([]NetworkResponse, 0, len(networks))
		for _, net := range networks {
			result = append(result, NetworkResponse{
				Name:   net.Name,
				Type:   net.Type,
				UsedBy: net.UsedBy,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"networks": result,
			"count":    len(result),
		})
	}
}
