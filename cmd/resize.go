package cmd

import (
	"fmt"
	"os"

	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/spf13/cobra"
)

var resizeCmd = &cobra.Command{
	Use:   "resize <name>",
	Short: "Resize a VPS instance",
	Long: `Resize CPU, memory, or disk for a VPS instance.

Examples:
  vpsctl resize myserver --cpu 4
  vpsctl resize myserver --memory 2GB
  vpsctl resize myserver --disk 50GB
  vpsctl resize myserver --cpu 2 --memory 4GB`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		cpu, _ := cmd.Flags().GetInt("cpu")
		memory, _ := cmd.Flags().GetString("memory")
		disk, _ := cmd.Flags().GetString("disk")

		// Check if at least one flag is provided
		if cpu == 0 && memory == "" && disk == "" {
			fmt.Fprintln(os.Stderr, "Error: At least one resize option is required (--cpu, --memory, or --disk)")
			os.Exit(1)
		}

		client, err := lxd.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to LXD: %v\n", err)
			os.Exit(1)
		}

		// Get current instance info
		info, err := client.GetInstance(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Instance '%s' not found: %v\n", name, err)
			os.Exit(1)
		}

		fmt.Printf("Resizing instance '%s'...\n", name)
		fmt.Printf("Current configuration:\n")
		fmt.Printf("  CPU: %d cores\n", info.CPU)
		fmt.Printf("  Memory: %s\n", info.Memory)
		fmt.Printf("  Disk: %s\n", info.Disk)
		fmt.Println()

		// Use defaults from current config if not specified
		if cpu == 0 {
			cpu = info.CPU
		} else {
			fmt.Printf("New CPU: %d cores\n", cpu)
		}

		if memory != "" {
			fmt.Printf("New Memory: %s\n", memory)
		}

		if disk != "" {
			fmt.Printf("New Disk: %s\n", disk)
		}

		if err := client.ResizeInstance(name, cpu, memory, disk); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to resize instance: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nInstance '%s' resized successfully.\n", name)

		// Show updated info
		updatedInfo, err := client.GetInstance(name)
		if err == nil {
			fmt.Println("\nUpdated configuration:")
			fmt.Printf("  CPU: %d cores\n", updatedInfo.CPU)
			fmt.Printf("  Memory: %s\n", updatedInfo.Memory)
			fmt.Printf("  Disk: %s\n", updatedInfo.Disk)
		}
	},
}

func init() {
	rootCmd.AddCommand(resizeCmd)

	resizeCmd.Flags().Int("cpu", 0, "New CPU cores limit")
	resizeCmd.Flags().String("memory", "", "New memory limit (e.g., 2GB)")
	resizeCmd.Flags().String("disk", "", "New disk size (e.g., 50GB)")
}
