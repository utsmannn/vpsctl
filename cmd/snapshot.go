package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/kiatkoding/vpsctl/pkg/output"
	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot <name>",
	Short: "Manage snapshots for a VPS instance",
	Long: `Create, list, or restore snapshots for a VPS instance.

Examples:
  # Create a snapshot
  vpsctl snapshot myserver --name backup-2024

  # List snapshots
  vpsctl snapshot myserver --list

  # Restore from snapshot
  vpsctl snapshot myserver --restore backup-2024`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		snapshotName, _ := cmd.Flags().GetString("name")
		listSnapshots, _ := cmd.Flags().GetBool("list")
		restoreSnapshot, _ := cmd.Flags().GetString("restore")

		client, err := lxd.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to LXD: %v\n", err)
			os.Exit(1)
		}

		// Check if instance exists
		_, err = client.GetInstance(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Instance '%s' not found: %v\n", name, err)
			os.Exit(1)
		}

		ctx := context.Background()

		// List snapshots
		if listSnapshots {
			snapshots, err := client.ListSnapshots(ctx, name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to list snapshots: %v\n", err)
				os.Exit(1)
			}

			if len(snapshots) == 0 {
				fmt.Printf("No snapshots found for instance '%s'.\n", name)
				return
			}

			fmt.Printf("Snapshots for instance '%s':\n\n", name)
			output.PrintSnapshotTable(snapshots)
			return
		}

		// Restore snapshot
		if restoreSnapshot != "" {
			fmt.Printf("Restoring snapshot '%s' for instance '%s'...\n", restoreSnapshot, name)

			if err := client.RestoreSnapshot(ctx, name, restoreSnapshot); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to restore snapshot: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Snapshot '%s' restored successfully.\n", restoreSnapshot)
			return
		}

		// Create snapshot
		if snapshotName == "" {
			fmt.Fprintln(os.Stderr, "Error: Snapshot name is required. Use --name flag.")
			os.Exit(1)
		}

		fmt.Printf("Creating snapshot '%s' for instance '%s'...\n", snapshotName, name)

		if err := client.CreateSnapshot(ctx, name, snapshotName); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create snapshot: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Snapshot '%s' created successfully.\n", snapshotName)
	},
}

func init() {
	rootCmd.AddCommand(snapshotCmd)

	snapshotCmd.Flags().StringP("name", "n", "", "Snapshot name")
	snapshotCmd.Flags().Bool("list", false, "List all snapshots")
	snapshotCmd.Flags().String("restore", "", "Restore from snapshot")
}
