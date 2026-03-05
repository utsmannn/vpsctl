package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a VPS instance",
	Long: `Delete a VPS instance permanently.
WARNING: This action is irreversible and will delete all data in the instance.

Examples:
  vpsctl delete myserver
  vpsctl delete webserver --force`,
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"rm"},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		force, _ := cmd.Flags().GetBool("force")

		client, err := lxd.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to LXD: %v\n", err)
			os.Exit(1)
		}

		// Check if instance exists
		info, err := client.GetInstance(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Instance '%s' not found: %v\n", name, err)
			os.Exit(1)
		}

		// Confirm deletion if not forced
		if !force {
			fmt.Printf("WARNING: You are about to delete instance '%s'\n", name)
			fmt.Printf("  Status: %s\n", info.Status)
			fmt.Printf("  Type: %s\n", info.Type)
			fmt.Println("\nThis action is IRREVERSIBLE and will delete all data!")

			reader := bufio.NewReader(os.Stdin)
			fmt.Printf("\nType '%s' to confirm deletion: ", name)
			confirmation, _ := reader.ReadString('\n')
			confirmation = strings.TrimSpace(confirmation)

			if confirmation != name {
				fmt.Println("Deletion cancelled.")
				return
			}
		}

		fmt.Printf("Deleting instance '%s'...\n", name)

		if err := client.DeleteInstance(name, force); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete instance: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Instance '%s' deleted successfully.\n", name)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")
}
