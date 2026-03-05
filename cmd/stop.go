package cmd

import (
	"fmt"
	"os"

	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a VPS instance",
	Long: `Stop a running VPS instance.

Examples:
  vpsctl stop myserver
  vpsctl stop webserver --force`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		force, _ := cmd.Flags().GetBool("force")

		client, err := lxd.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to LXD: %v\n", err)
			os.Exit(1)
		}

		if force {
			fmt.Printf("Force stopping instance '%s'...\n", name)
		} else {
			fmt.Printf("Stopping instance '%s'...\n", name)
		}

		if err := client.StopInstance(name, force); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to stop instance: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Instance '%s' stopped successfully.\n", name)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	stopCmd.Flags().BoolP("force", "f", false, "Force stop (SIGKILL)")
}
