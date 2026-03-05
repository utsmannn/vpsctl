package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kiatkoding/vpsctl/internal/lxd"
)

// CreateInstanceRequest represents the request body for creating an instance
type CreateInstanceRequest struct {
	Name     string `json:"name" binding:"required"`
	Image    string `json:"image" binding:"required"`
	CPU      int    `json:"cpu"`
	Memory   string `json:"memory"`
	Disk     string `json:"disk"`
	Type     string `json:"type"`
	SSHKey   string `json:"ssh_key"`
	Password string `json:"password"`
}

// UpdateInstanceRequest represents the request body for updating an instance
type UpdateInstanceRequest struct {
	CPU    *int    `json:"cpu"`
	Memory *string `json:"memory"`
	Disk   *string `json:"disk"`
}

// InstanceResponse represents an instance in API responses
type InstanceResponse struct {
	Name       string            `json:"name"`
	Status     string            `json:"status"`
	Type       string            `json:"type"`
	CPU        int               `json:"cpu"`
	Memory     string            `json:"memory"`
	MemoryBytes int64            `json:"memory_bytes"`
	Disk       string            `json:"disk"`
	DiskBytes  int64             `json:"disk_bytes"`
	IP         string            `json:"ip"`
	CreatedAt  string            `json:"created_at"`
	Config     map[string]string `json:"config,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ListInstances returns all instances
func ListInstances(client *lxd.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		instances, err := client.ListInstances()
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "list_failed",
				Message: err.Error(),
			})
			return
		}

		// Convert to response format
		response := make([]InstanceResponse, 0, len(instances))
		for _, inst := range instances {
			// Get IP address for each instance
			ip, _ := client.GetInstanceIP(inst.Name)

			response = append(response, InstanceResponse{
				Name:        inst.Name,
				Status:      inst.Status,
				Type:        inst.Type,
				CPU:         inst.CPU,
				Memory:      inst.Memory,
				MemoryBytes: parseMemoryToBytes(inst.Memory),
				Disk:        inst.Disk,
				DiskBytes:   parseDiskToBytes(inst.Disk),
				IP:          ip,
				CreatedAt:   inst.CreatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"instances": response,
			"count":     len(response),
		})
	}
}

// CreateInstance creates a new instance
func CreateInstance(client *lxd.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		var req CreateInstanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_request",
				Message: err.Error(),
			})
			return
		}

		// Validate name
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_name",
				Message: "Instance name is required",
			})
			return
		}

		// Set defaults
		if req.CPU == 0 {
			req.CPU = 1
		}
		if req.Memory == "" {
			req.Memory = "512MB"
		}
		if req.Disk == "" {
			req.Disk = "10GB"
		}
		if req.Type == "" {
			req.Type = "container"
		}

		// Create instance
		opts := lxd.CreateInstanceOptions{
			Name:     req.Name,
			Image:    req.Image,
			CPU:      req.CPU,
			Memory:   req.Memory,
			Disk:     req.Disk,
			Type:     req.Type,
			SSHKey:   req.SSHKey,
			Password: req.Password,
		}

		inst, err := client.CreateInstance(opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "create_failed",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"instance": InstanceResponse{
				Name:      inst.Name,
				Status:    inst.Status,
				Type:      inst.Type,
				CPU:       inst.CPU,
				Memory:    inst.Memory,
				Disk:      inst.Disk,
				CreatedAt: inst.CreatedAt.Format("2006-01-02T15:04:05Z"),
			},
		})
	}
}

// GetInstance returns a specific instance
func GetInstance(client *lxd.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_name",
				Message: "Instance name is required",
			})
			return
		}

		inst, err := client.GetInstance(name)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: err.Error(),
			})
			return
		}

		// Get IP address
		ip, _ := client.GetInstanceIP(name)

		c.JSON(http.StatusOK, gin.H{
			"instance": InstanceResponse{
				Name:        inst.Name,
				Status:      inst.Status,
				Type:        inst.Type,
				CPU:         inst.CPU,
				Memory:      inst.Memory,
				MemoryBytes: parseMemoryToBytes(inst.Memory),
				Disk:        inst.Disk,
				DiskBytes:   parseDiskToBytes(inst.Disk),
				IP:          ip,
				CreatedAt:   inst.CreatedAt.Format("2006-01-02T15:04:05Z"),
			},
		})
	}
}

// UpdateInstance updates an instance's resources
func UpdateInstance(client *lxd.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_name",
				Message: "Instance name is required",
			})
			return
		}

		var req UpdateInstanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_request",
				Message: err.Error(),
			})
			return
		}

		// Update instance
		err := client.ResizeInstance(name, derefInt(req.CPU), derefString(req.Memory), derefString(req.Disk))
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "update_failed",
				Message: err.Error(),
			})
			return
		}

		// Get updated instance
		inst, err := client.GetInstance(name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "get_failed",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"instance": InstanceResponse{
				Name:      inst.Name,
				Status:    inst.Status,
				Type:      inst.Type,
				CPU:       inst.CPU,
				Memory:    inst.Memory,
				Disk:      inst.Disk,
				CreatedAt: inst.CreatedAt.Format("2006-01-02T15:04:05Z"),
			},
		})
	}
}

// DeleteInstance deletes an instance
func DeleteInstance(client *lxd.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_name",
				Message: "Instance name is required",
			})
			return
		}

		force := c.Query("force") == "true"

		err := client.DeleteInstance(name, force)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "delete_failed",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Instance deleted successfully",
			"name":    name,
		})
	}
}

// StartInstance starts an instance
func StartInstance(client *lxd.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_name",
				Message: "Instance name is required",
			})
			return
		}

		err := client.StartInstance(name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "start_failed",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Instance started successfully",
			"name":    name,
		})
	}
}

// StopInstance stops an instance
func StopInstance(client *lxd.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_name",
				Message: "Instance name is required",
			})
			return
		}

		force := c.Query("force") == "true"

		err := client.StopInstance(name, force)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "stop_failed",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Instance stopped successfully",
			"name":    name,
		})
	}
}

// RestartInstance restarts an instance
func RestartInstance(client *lxd.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_name",
				Message: "Instance name is required",
			})
			return
		}

		err := client.RestartInstance(name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "restart_failed",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Instance restarted successfully",
			"name":    name,
		})
	}
}

// GetInstanceConsole returns console connection info
func GetInstanceConsole(client *lxd.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_name",
				Message: "Instance name is required",
			})
			return
		}

		// Check if instance exists
		_, err := client.GetInstance(name)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Console endpoint ready",
			"name":     name,
			"protocol": "websocket",
		})
	}
}

// Helper functions

func derefInt(val *int) int {
	if val == nil {
		return 0
	}
	return *val
}

func derefString(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}

func parseMemoryToBytes(mem string) int64 {
	if mem == "" || mem == "unlimited" {
		return 0
	}

	mem = strings.ToUpper(strings.TrimSpace(mem))

	multipliers := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}

	for suffix, mult := range multipliers {
		if strings.HasSuffix(mem, suffix) {
			numStr := strings.TrimSuffix(mem, suffix)
			if num, err := strconv.ParseInt(numStr, 10, 64); err == nil {
				return num * mult
			}
		}
	}

	return 0
}

func parseDiskToBytes(disk string) int64 {
	if disk == "" {
		return 0
	}

	disk = strings.ToUpper(strings.TrimSpace(disk))

	multipliers := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}

	for suffix, mult := range multipliers {
		if strings.HasSuffix(disk, suffix) {
			numStr := strings.TrimSuffix(disk, suffix)
			if num, err := strconv.ParseInt(numStr, 10, 64); err == nil {
				return num * mult
			}
		}
	}

	return 0
}
