package api

import (
	"github.com/gin-gonic/gin"
	"github.com/kiatkoding/vpsctl/api/handlers"
)

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Instances endpoints
		instances := v1.Group("/instances")
		{
			instances.GET("", handlers.ListInstances(s.lxdClient))
			instances.POST("", handlers.CreateInstance(s.lxdClient))
			instances.GET("/:name", handlers.GetInstance(s.lxdClient))
			instances.PUT("/:name", handlers.UpdateInstance(s.lxdClient))
			instances.DELETE("/:name", handlers.DeleteInstance(s.lxdClient))
			instances.POST("/:name/start", handlers.StartInstance(s.lxdClient))
			instances.POST("/:name/stop", handlers.StopInstance(s.lxdClient))
			instances.POST("/:name/restart", handlers.RestartInstance(s.lxdClient))
			instances.GET("/:name/metrics", handlers.GetInstanceMetrics(s.lxdClient))
			instances.GET("/:name/console", handlers.GetInstanceConsole(s.lxdClient))

			// Port forwarding endpoints
			instances.GET("/:name/ports", handlers.ListPortForwards(s.lxdClient, s.portManager))
			instances.POST("/:name/ports", handlers.AddPortForward(s.lxdClient, s.portManager))
			instances.DELETE("/:name/ports/:port", handlers.RemovePortForward(s.lxdClient, s.portManager))
		}

		// Resources endpoints
		v1.GET("/resources", handlers.GetResources(s.lxdClient))
		v1.GET("/resources/allocated", handlers.GetAllocatedResources(s.lxdClient))

		// Images endpoints
		v1.GET("/images", handlers.ListImages(s.lxdClient))
		v1.GET("/images/:fingerprint", handlers.GetImage(s.lxdClient))

		// Storage endpoints
		v1.GET("/storage", handlers.ListStoragePools(s.lxdClient))

		// Network endpoints
		v1.GET("/networks", handlers.ListNetworks(s.lxdClient))
	}
}
