package cmd

import (
	"fmt"
	"os"

	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart <name>",
	Short: "Restart a VPS instance",
	Long: `Restart a VPS instance.

Examples:
  vpsctl restart myserver`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		client, err := lxd.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to LXD: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Restarting instance '%s'...\n", name)

		if err := client.RestartInstance(name); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to restart instance: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Instance '%s' restarted successfully.\n", name)

		// Show instance info
		info, err := client.GetInstance(name)
		if err == nil {
			fmt.Printf("Status: %s\n", info.Status)
			if info.IP != "" {
				fmt.Printf("IP Address: %s\n", info.IP)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
