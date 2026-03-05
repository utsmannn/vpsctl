package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kiatkoding/vpsctl/api"
	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	Long: `Start the REST API server for remote VPS management.

The API provides endpoints for:
- Instance management (create, list, start, stop, delete)
- Resource monitoring
- WebSocket for real-time metrics

Examples:
  vpsctl serve
  vpsctl serve --port 3000
  vpsctl serve --socket /var/run/vpsctl.sock
  vpsctl serve --token my-secret-token`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		socket, _ := cmd.Flags().GetString("socket")
		token, _ := cmd.Flags().GetString("token")

		// Create LXD client
		lxdClient, err := lxd.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to LXD: %v\n", err)
			os.Exit(1)
		}

		serverConfig := api.Config{
			Port:      port,
			Socket:    socket,
			AuthToken: token,
			LXDClient: lxdClient,
		}

		server := api.NewServer(serverConfig)

		// Handle graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			fmt.Println("\nShutting down server...")
			server.Stop()
			os.Exit(0)
		}()

		fmt.Printf("Starting vpsctl API server...\n")

		if socket != "" {
			fmt.Printf("Listening on unix socket: %s\n", socket)
		} else {
			fmt.Printf("Listening on port: %d\n", port)
		}

		if token != "" {
			fmt.Println("Authentication: enabled")
		}

		fmt.Println("\nAPI Endpoints:")
		fmt.Println("  GET    /api/v1/instances          - List instances")
		fmt.Println("  POST   /api/v1/instances          - Create instance")
		fmt.Println("  GET    /api/v1/instances/:name    - Get instance")
		fmt.Println("  DELETE /api/v1/instances/:name    - Delete instance")
		fmt.Println("  POST   /api/v1/instances/:name/start  - Start instance")
		fmt.Println("  POST   /api/v1/instances/:name/stop   - Stop instance")
		fmt.Println("  GET    /api/v1/resources          - Get resources")
		fmt.Println("  GET    /api/v1/images             - List images")
		fmt.Println("\nPress Ctrl+C to stop")

		if err := server.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntP("port", "p", 8080, "API server port")
	serveCmd.Flags().String("socket", "", "Unix socket path (overrides port)")
	serveCmd.Flags().String("token", "", "API token for authentication")
}
