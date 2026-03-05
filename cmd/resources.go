package cmd

import (
	"fmt"
	"os"

	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/kiatkoding/vpsctl/pkg/output"
	"github.com/spf13/cobra"
)

var resourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "Show host and allocated resources",
	Long: `Show host resource information and allocated resources across all instances.

Examples:
  vpsctl resources
  vpsctl resources --format json`,
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")

		client, err := lxd.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to LXD: %v\n", err)
			os.Exit(1)
		}

		// Get resource summary
		summary, err := client.GetResourceSummary()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get resources: %v\n", err)
			os.Exit(1)
		}

		// Calculate usage percentages
		cpuUsagePercent := float64(0)
		if summary.CPUTotal > 0 {
			cpuUsagePercent = float64(summary.CPUUsed) / float64(summary.CPUTotal) * 100
		}

		memoryUsagePercent := float64(0)
		if summary.MemoryTotal > 0 {
			memoryUsagePercent = float64(summary.MemoryUsed) / float64(summary.MemoryTotal) * 100
		}

		diskUsagePercent := float64(0)
		if summary.DiskTotal > 0 {
			diskUsagePercent = float64(summary.DiskUsed) / float64(summary.DiskTotal) * 100
		}

		data := map[string]interface{}{
			"host": map[string]interface{}{
				"cpu_total":    summary.CPUTotal,
				"memory_total": fmt.Sprintf("%d MB", summary.MemoryTotal),
				"disk_total":   fmt.Sprintf("%d GB", summary.DiskTotal),
			},
			"allocated": map[string]interface{}{
				"cpu":    summary.CPUUsed,
				"memory": fmt.Sprintf("%d MB", summary.MemoryUsed),
				"disk":   fmt.Sprintf("%d GB", summary.DiskUsed),
			},
			"usage_percent": map[string]interface{}{
				"cpu":    fmt.Sprintf("%.1f%%", cpuUsagePercent),
				"memory": fmt.Sprintf("%.1f%%", memoryUsagePercent),
				"disk":   fmt.Sprintf("%.1f%%", diskUsagePercent),
			},
		}

		if format == "json" {
			if err := output.PrintJSON(data); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to output JSON: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// Print formatted output
		fmt.Println("=== Host Resources ===")
		fmt.Printf("  CPU:    %d cores\n", summary.CPUTotal)
		fmt.Printf("  Memory: %d MB\n", summary.MemoryTotal)
		fmt.Printf("  Disk:   %d GB\n", summary.DiskTotal)
		fmt.Println()

		fmt.Println("=== Allocated Resources ===")
		fmt.Printf("  CPU:    %d cores (%.1f%%)\n", summary.CPUUsed, cpuUsagePercent)
		fmt.Printf("  Memory: %d MB (%.1f%%)\n", summary.MemoryUsed, memoryUsagePercent)
		fmt.Printf("  Disk:   %d GB (%.1f%%)\n", summary.DiskUsed, diskUsagePercent)
	},
}

func init() {
	rootCmd.AddCommand(resourcesCmd)

	resourcesCmd.Flags().StringP("format", "f", "table", "Output format: table, json")
}
