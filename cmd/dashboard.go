package cmd

import (
	"fmt"
	"os"

	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/kiatkoding/vpsctl/tui"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Launch the TUI dashboard",
	Long: `Launch an interactive terminal-based dashboard for VPS management.

The dashboard provides:
- Real-time instance list with status
- Resource usage visualization
- Quick actions (create, start, stop, delete)
- Instance details view

Examples:
  vpsctl dashboard`,
	Aliases: []string{"ui", "tui"},
	Run: func(cmd *cobra.Command, args []string) {
		// Create LXD client
		client, err := lxd.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to LXD: %v\n", err)
			os.Exit(1)
		}

		app := tui.NewApp(client)
		if err := app.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Dashboard error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
