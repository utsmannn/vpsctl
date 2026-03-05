package cmd

import (
	"fmt"
	"os"

	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/kiatkoding/vpsctl/pkg/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all VPS instances",
	Long: `List all VPS instances managed by LXD.

Examples:
  vpsctl list
  vpsctl list --format json
  vpsctl list --all
  vpsctl list -f csv`,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		detailed, _ := cmd.Flags().GetBool("all")

		client, err := lxd.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to LXD: %v\n", err)
			os.Exit(1)
		}

		instances, err := client.ListInstances()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list instances: %v\n", err)
			os.Exit(1)
		}

		if len(instances) == 0 {
			fmt.Println("No instances found.")
			return
		}

		switch format {
		case "json":
			if err := output.PrintJSON(instances); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to output JSON: %v\n", err)
				os.Exit(1)
			}
		case "csv":
			output.PrintInstanceCSV(instances, detailed)
		default:
			output.PrintInstanceTable(instances, detailed)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringP("format", "f", "table", "Output format: table, json, csv")
	listCmd.Flags().BoolP("all", "a", false, "Show all details")
}
