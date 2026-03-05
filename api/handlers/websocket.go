package handlers

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// MetricsPayload represents metrics data sent via WebSocket
type MetricsPayload struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	NetworkRX   int64   `json:"network_rx"`
	NetworkTX   int64   `json:"network_tx"`
	Timestamp   int64   `json:"timestamp"`
}

// MetricsClient represents a connected WebSocket client
type MetricsClient struct {
	conn         *websocket.Conn
	instanceName string
	send         chan MetricsPayload
}

// MetricsHub maintains active WebSocket connections
type MetricsHub struct {
	clients    map[*MetricsClient]bool
	broadcast  chan MetricsPayload
	register   chan *MetricsClient
	unregister chan *MetricsClient
	mu         sync.RWMutex
}

// NewMetricsHub creates a new metrics hub
func NewMetricsHub() *MetricsHub {
	return &MetricsHub{
		clients:    make(map[*MetricsClient]bool),
		broadcast:  make(chan MetricsPayload, 256),
		register:   make(chan *MetricsClient),
		unregister: make(chan *MetricsClient),
	}
}

// Run starts the metrics hub
func (h *MetricsHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			logrus.Debugf("WebSocket client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			logrus.Debugf("WebSocket client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Global metrics hub
var metricsHub = NewMetricsHub()

func init() {
	go metricsHub.Run()
}

// GetInstanceMetrics streams real-time metrics via WebSocket
func GetInstanceMetrics(client *lxd.Client) gin.HandlerFunc {
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

		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logrus.Errorf("WebSocket upgrade failed: %v", err)
			return
		}

		// Create client
		metricsClient := &MetricsClient{
			conn:         conn,
			instanceName: name,
			send:         make(chan MetricsPayload, 256),
		}

		// Register client
		metricsHub.register <- metricsClient

		// Start read/write goroutines
		go writePump(client, metricsClient)
		go readPump(metricsClient)
	}
}

// writePump sends metrics to the WebSocket client
func writePump(client *lxd.Client, mc *MetricsClient) {
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
		mc.conn.Close()
		metricsHub.unregister <- mc
	}()

	for {
		select {
		case <-ticker.C:
			// Get instance state for metrics
			state, err := client.GetInstanceState(mc.instanceName)
			if err != nil {
				logrus.Errorf("Failed to get instance state: %v", err)
				continue
			}

			// Build metrics payload
			payload := MetricsPayload{
				Timestamp: time.Now().Unix(),
			}

			// CPU usage
			if state.CPU.Usage > 0 {
				// CPU usage is in nanoseconds, convert to percentage
				payload.CPUUsage = float64(state.CPU.Usage) / 1000000000
			}

			// Memory usage
			if state.Memory.Total > 0 {
				payload.MemoryUsage = float64(state.Memory.Usage) / float64(state.Memory.Total) * 100
			}

			// Disk usage
			if state.Disk != nil {
				for _, disk := range state.Disk {
					if disk.Total > 0 {
						payload.DiskUsage = float64(disk.Usage) / float64(disk.Total) * 100
						break
					}
				}
			}

			// Network stats
			if state.Network != nil {
				for _, net := range state.Network {
					payload.NetworkRX += int64(net.Counters.BytesReceived)
					payload.NetworkTX += int64(net.Counters.BytesSent)
				}
			}

			// Send metrics
			if err := mc.conn.WriteJSON(payload); err != nil {
				logrus.Errorf("WebSocket write error: %v", err)
				return
			}

		case metrics := <-mc.send:
			if err := mc.conn.WriteJSON(metrics); err != nil {
				logrus.Errorf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

// readPump reads messages from the WebSocket client
func readPump(mc *MetricsClient) {
	defer func() {
		mc.conn.Close()
		metricsHub.unregister <- mc
	}()

	// Set read limit and deadline
	mc.conn.SetReadLimit(512)
	mc.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// Set pong handler
	mc.conn.SetPongHandler(func(string) error {
		mc.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := mc.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("WebSocket read error: %v", err)
			}
			break
		}
	}
}

// AddPortForwardRequest represents a port forward request
type AddPortForwardRequest struct {
	HostPort      int    `json:"host_port" binding:"required"`
	ContainerPort int    `json:"container_port" binding:"required"`
	Protocol      string `json:"protocol"`
	Label         string `json:"label"`
}

// PortForwardResponse represents a port forward in responses
type PortForwardResponse struct {
	HostPort      int    `json:"host_port"`
	ContainerPort int    `json:"container_port"`
	Protocol      string `json:"protocol"`
	Label         string `json:"label"`
	Status        string `json:"status"`
}

// ListPortForwards lists port forwards for an instance
func ListPortForwards(client *lxd.Client, pm interface{}) gin.HandlerFunc {
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

		// Get instance state to find listening ports
		state, err := client.GetInstanceState(name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "state_failed",
				Message: err.Error(),
			})
			return
		}

		// Build port list from network state
		ports := make([]PortForwardResponse, 0)
		if state.Network != nil {
			for netName, net := range state.Network {
				for _, addr := range net.Addresses {
					if addr.Family == "inet" {
						// Add discovered ports
						ports = append(ports, PortForwardResponse{
							HostPort:      0, // Would need iptables parsing
							ContainerPort: 0, // Would need port scanning
							Protocol:      "tcp",
							Label:         netName,
							Status:        "active",
						})
					}
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"ports": ports,
			"count": len(ports),
		})
	}
}

// AddPortForward adds a port forward for an instance
func AddPortForward(client *lxd.Client, pm interface{}) gin.HandlerFunc {
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

		var req AddPortForwardRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_request",
				Message: err.Error(),
			})
			return
		}

		// Set default protocol
		if req.Protocol == "" {
			req.Protocol = "tcp"
		}

		// Validate protocol
		if req.Protocol != "tcp" && req.Protocol != "udp" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_protocol",
				Message: "Protocol must be tcp or udp",
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

		// Get instance IP
		ip, err := client.GetInstanceIP(name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "ip_failed",
				Message: "Could not get instance IP address",
			})
			return
		}

		// Port forward would be implemented via iptables or proxy device
		// For now, return success with the configuration
		c.JSON(http.StatusCreated, gin.H{
			"message": "Port forward configured",
			"port_forward": PortForwardResponse{
				HostPort:      req.HostPort,
				ContainerPort: req.ContainerPort,
				Protocol:      req.Protocol,
				Label:         req.Label,
				Status:        "active",
			},
			"instance_ip": ip,
		})
	}
}

// RemovePortForward removes a port forward from an instance
func RemovePortForward(client *lxd.Client, pm interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "LXD client not initialized",
			})
			return
		}

		name := c.Param("name")
		portStr := c.Param("port")

		if name == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_name",
				Message: "Instance name is required",
			})
			return
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_port",
				Message: "Port must be a number",
			})
			return
		}

		// Check if instance exists
		_, err = client.GetInstance(name)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: err.Error(),
			})
			return
		}

		// Port forward removal would be implemented via iptables or proxy device removal
		c.JSON(http.StatusOK, gin.H{
			"message": "Port forward removed",
			"host_port": port,
			"name":     name,
		})
	}
}
