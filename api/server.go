package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/kiatkoding/vpsctl/internal/portforward"
	"github.com/sirupsen/logrus"
)

// Server represents the API server
type Server struct {
	router      *gin.Engine
	lxdClient   *lxd.Client
	portManager *portforward.Manager
	port        int
	socket      string
	authToken   string
	httpServer  *http.Server
}

// Config holds server configuration
type Config struct {
	Port      int
	Socket    string
	AuthToken string
	LXDClient *lxd.Client
}

// NewServer creates a new API server
func NewServer(cfg Config) *Server {
	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Create port forward manager
	var portManager *portforward.Manager
	if cfg.LXDClient != nil {
		portManager = portforward.NewManager(cfg.LXDClient)
	}

	server := &Server{
		router:      router,
		lxdClient:   cfg.LXDClient,
		portManager: portManager,
		port:        cfg.Port,
		socket:      cfg.Socket,
		authToken:   cfg.AuthToken,
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

// setupMiddleware configures global middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// CORS middleware
	s.router.Use(s.corsMiddleware())

	// Logging middleware
	s.router.Use(loggingMiddleware())

	// Auth middleware (if token is set)
	if s.authToken != "" {
		s.router.Use(authMiddleware(s.authToken))
	}
}

// corsMiddleware returns CORS middleware
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// loggingMiddleware returns logging middleware
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Log request
		latency := time.Since(start)
		status := c.Writer.Status()

		logrus.WithFields(logrus.Fields{
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
			"status":  status,
			"latency": latency,
			"ip":      c.ClientIP(),
		}).Info("HTTP Request")
	}
}

// authMiddleware returns authentication middleware
func authMiddleware(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header required",
			})
			return
		}

		// Check Bearer token
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid authorization header format",
			})
			return
		}

		providedToken := authHeader[7:]
		if providedToken != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid token",
			})
			return
		}

		c.Next()
	}
}

// Run starts the API server
func (s *Server) Run() error {
	var listener net.Listener
	var err error

	// Create HTTP server
	s.httpServer = &http.Server{
		Handler: s.router,
	}

	// Listen on socket or port
	if s.socket != "" {
		// Remove existing socket file
		os.Remove(s.socket)

		// Create Unix socket listener
		listener, err = net.Listen("unix", s.socket)
		if err != nil {
			return fmt.Errorf("failed to listen on socket %s: %w", s.socket, err)
		}

		// Set socket permissions
		os.Chmod(s.socket, 0660)

		logrus.Infof("API server listening on socket: %s", s.socket)
	} else {
		// Create TCP listener
		addr := fmt.Sprintf(":%d", s.port)
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
		}

		logrus.Infof("API server listening on port: %d", s.port)
	}

	// Handle graceful shutdown
	go s.handleShutdown()

	// Start server
	if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// handleShutdown handles graceful shutdown signals
func (s *Server) handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logrus.Info("Shutting down API server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		logrus.Errorf("Error shutting down server: %v", err)
	}

	// Clean up socket file
	if s.socket != "" {
		os.Remove(s.socket)
	}

	logrus.Info("API server stopped")
}

// Stop stops the API server
func (s *Server) Stop() error {
	if s.httpServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down server: %w", err)
	}

	// Clean up socket file
	if s.socket != "" {
		os.Remove(s.socket)
	}

	return nil
}

// GetRouter returns the Gin router (useful for testing)
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
