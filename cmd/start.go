package cmd

import (
	"fmt"
	"os"

	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start a VPS instance",
	Long: `Start a stopped VPS instance.

Examples:
  vpsctl start myserver
  vpsctl start webserver`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		client, err := lxd.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to LXD: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Starting instance '%s'...\n", name)

		if err := client.StartInstance(name); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start instance: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Instance '%s' started successfully.\n", name)

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
	rootCmd.AddCommand(startCmd)
}
