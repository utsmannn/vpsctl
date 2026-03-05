package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kiatkoding/vpsctl/internal/lxd"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Padding(0, 1)

	runningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	stoppedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	frozenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	rowStyle = lipgloss.NewStyle().
			Padding(0, 1)
)

// PrintTable prints data as formatted table
func PrintTable(headers []string, rows [][]string) {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	headerParts := make([]string, len(headers))
	for i, h := range headers {
		headerParts[i] = headerStyle.Render(padRight(h, widths[i]))
	}
	fmt.Println(strings.Join(headerParts, ""))

	// Print separator
	separatorParts := make([]string, len(headers))
	for i, w := range widths {
		separatorParts[i] = strings.Repeat("-", w+2)
	}
	fmt.Println(strings.Join(separatorParts, " "))

	// Print rows
	for _, row := range rows {
		rowParts := make([]string, len(row))
		for i, cell := range row {
			rowParts[i] = rowStyle.Render(padRight(cell, widths[i]))
		}
		fmt.Println(strings.Join(rowParts, ""))
	}
}

// PrintInstanceTable prints instances as table
func PrintInstanceTable(instances []lxd.InstanceInfo, detailed bool) {
	if len(instances) == 0 {
		fmt.Println("No instances found.")
		return
	}

	if detailed {
		printDetailedInstanceTable(instances)
	} else {
		printSimpleInstanceTable(instances)
	}
}

func printSimpleInstanceTable(instances []lxd.InstanceInfo) {
	headers := []string{"NAME", "STATUS", "TYPE", "IP ADDRESS", "CPU", "MEMORY"}
	rows := make([][]string, len(instances))

	for i, inst := range instances {
		status := formatStatus(inst.Status)
		ip := inst.IP
		if ip == "" {
			ip = "-"
		}

		rows[i] = []string{
			inst.Name,
			status,
			inst.Type,
			ip,
			fmt.Sprintf("%d", inst.CPU),
			inst.Memory,
		}
	}

	PrintTable(headers, rows)
}

func printDetailedInstanceTable(instances []lxd.InstanceInfo) {
	headers := []string{"NAME", "STATUS", "TYPE", "IP ADDRESS", "CPU", "MEMORY", "DISK", "CREATED"}
	rows := make([][]string, len(instances))

	for i, inst := range instances {
		status := formatStatus(inst.Status)
		ip := inst.IP
		if ip == "" {
			ip = "-"
		}

		rows[i] = []string{
			inst.Name,
			status,
			inst.Type,
			ip,
			fmt.Sprintf("%d", inst.CPU),
			inst.Memory,
			inst.Disk,
			inst.CreatedAt.Format("2006-01-02 15:04"),
		}
	}

	PrintTable(headers, rows)
}

// PrintInstanceCSV prints instances as CSV
func PrintInstanceCSV(instances []lxd.InstanceInfo, detailed bool) {
	if detailed {
		fmt.Println("name,status,type,ip_address,cpu,memory,disk,created")
		for _, inst := range instances {
			ip := inst.IP
			if ip == "" {
				ip = "-"
			}
			fmt.Printf("%s,%s,%s,%s,%d,%s,%s,%s\n",
				inst.Name,
				inst.Status,
				inst.Type,
				ip,
				inst.CPU,
				inst.Memory,
				inst.Disk,
				inst.CreatedAt.Format("2006-01-02 15:04"),
			)
		}
	} else {
		fmt.Println("name,status,type,ip_address,cpu,memory")
		for _, inst := range instances {
			ip := inst.IP
			if ip == "" {
				ip = "-"
			}
			fmt.Printf("%s,%s,%s,%s,%d,%s\n",
				inst.Name,
				inst.Status,
				inst.Type,
				ip,
				inst.CPU,
				inst.Memory,
			)
		}
	}
}

// PrintSnapshotTable prints snapshots as table
func PrintSnapshotTable(snapshots []lxd.SnapshotInfo) {
	if len(snapshots) == 0 {
		fmt.Println("No snapshots found.")
		return
	}

	headers := []string{"NAME", "CREATED", "SIZE"}
	rows := make([][]string, len(snapshots))

	for i, snap := range snapshots {
		rows[i] = []string{
			snap.Name,
			snap.CreatedAt.Format("2006-01-02 15:04"),
			formatBytes(snap.Size),
		}
	}

	PrintTable(headers, rows)
}

func formatStatus(status string) string {
	switch strings.ToLower(status) {
	case "running":
		return runningStyle.Render("Running")
	case "stopped":
		return stoppedStyle.Render("Stopped")
	case "frozen":
		return frozenStyle.Render("Frozen")
	default:
		return status
	}
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.0fMB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.0fKB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// PrintError prints an error message with styling
func PrintError(message string) {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	fmt.Fprintln(os.Stderr, errorStyle.Render("Error: ")+message)
}

// PrintSuccess prints a success message with styling
func PrintSuccess(message string) {
	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("42")).
		Bold(true)

	fmt.Println(successStyle.Render("✓ " + message))
}

// PrintWarning prints a warning message with styling
func PrintWarning(message string) {
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	fmt.Println(warningStyle.Render("⚠ " + message))
}
