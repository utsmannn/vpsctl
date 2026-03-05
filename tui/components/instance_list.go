package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InstanceInfo represents instance information for the list
type InstanceInfo struct {
	Name   string
	Status string
	CPU    int
	Memory string
	Disk   string
	Type   string // "container" or "vm"
	IP     string
}

// InstanceList is a component that displays a list of instances
type InstanceList struct {
	instances []InstanceInfo
	selected  int
	width     int
	height    int
	focused   bool
}

// NewInstanceList creates a new instance list component
func NewInstanceList() InstanceList {
	return InstanceList{
		instances: []InstanceInfo{},
		selected:  0,
		width:     60,
		height:    10,
		focused:   false,
	}
}

// SetInstances sets the list of instances to display
func (il *InstanceList) SetInstances(instances []InstanceInfo) {
	il.instances = instances
	if il.selected >= len(instances) && len(instances) > 0 {
		il.selected = len(instances) - 1
	}
}

// SetSize sets the dimensions of the component
func (il *InstanceList) SetSize(width, height int) {
	il.width = width
	il.height = height
}

// SetFocused sets whether the list is focused
func (il *InstanceList) SetFocused(focused bool) {
	il.focused = focused
}

// Selected returns the currently selected instance, or nil if none
func (il InstanceList) Selected() *InstanceInfo {
	if len(il.instances) == 0 || il.selected < 0 || il.selected >= len(il.instances) {
		return nil
	}
	return &il.instances[il.selected]
}

// SelectedIndex returns the index of the selected item
func (il InstanceList) SelectedIndex() int {
	return il.selected
}

// Count returns the number of instances
func (il InstanceList) Count() int {
	return len(il.instances)
}

// RunningCount returns the number of running instances
func (il InstanceList) RunningCount() int {
	count := 0
	for _, inst := range il.instances {
		if strings.ToLower(inst.Status) == "running" {
			count++
		}
	}
	return count
}

// Update handles messages for the instance list
func (il InstanceList) Update(msg tea.Msg) (InstanceList, tea.Cmd) {
	if !il.focused {
		return il, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if il.selected > 0 {
				il.selected--
			}
		case "down", "j":
			if il.selected < len(il.instances)-1 {
				il.selected++
			}
		case "home", "g":
			il.selected = 0
		case "end", "G":
			if len(il.instances) > 0 {
				il.selected = len(il.instances) - 1
			}
		}
	}

	return il, nil
}

// View renders the instance list
func (il InstanceList) View() string {
	// Styles
	boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6B7280")).
			Padding(0, 1).
			Width(il.width - 2)

	if il.focused {
		boxStyle = boxStyle.BorderForeground(lipgloss.Color("#7C3AED"))
	}

	headerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true)

	statusRunningStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981")).
				Width(10)

	statusStoppedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				Width(10)

	statusPendingStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F59E0B")).
				Width(10)

	selectedStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("#374151")).
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true)

	emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true)

	// Build header
	header := fmt.Sprintf("%-15s %-10s %-12s %-8s",
		"NAME", "STATUS", "RESOURCES", "TYPE")
	header = headerStyle.Render(header)

	// Build instance rows
	var rows []string
	rows = append(rows, header)
	rows = append(rows, strings.Repeat("─", il.width-4))

	if len(il.instances) == 0 {
		emptyMsg := emptyStyle.Render("No instances found. Press [N] to create one.")
		rows = append(rows, emptyMsg)
	} else {
		for i, inst := range il.instances {
			// Format resources
			resources := fmt.Sprintf("%dc/%s", inst.CPU, inst.Memory)
			if inst.Disk != "" {
				resources = fmt.Sprintf("%dc/%s/%s", inst.CPU, inst.Memory, inst.Disk)
			}

			// Format status
			var statusDisplay string
			statusLower := strings.ToLower(inst.Status)
			switch statusLower {
			case "running":
				statusDisplay = statusRunningStyle.Render("RUNNING")
			case "stopped":
				statusDisplay = statusStoppedStyle.Render("STOPPED")
			default:
				statusDisplay = statusPendingStyle.Render(strings.ToUpper(inst.Status))
			}

			// Format type
			instType := "CT"
			if strings.ToLower(inst.Type) == "vm" {
				instType = "VM"
			}

			// Build row
			row := fmt.Sprintf("%-15s %s %-12s %-8s",
				truncate(inst.Name, 15),
				statusDisplay,
				resources,
				instType)

			// Apply selection style
			if i == il.selected && il.focused {
				row = selectedStyle.Render("→ " + row)
			} else if i == il.selected {
				row = "  " + row
			} else {
				row = "  " + row
			}

			rows = append(rows, row)
		}
	}

	// Build content
	content := strings.Join(rows, "\n")

	return boxStyle.Render(content)
}

// truncate truncates a string to the given length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// ViewCompact renders a compact version of the instance list
func (il InstanceList) ViewCompact() string {
	// Styles
	nameStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	statusRunningStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981"))

	statusStoppedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444"))

	detailStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	if len(il.instances) == 0 {
		return detailStyle.Render("No instances")
	}

	var lines []string
	for i, inst := range il.instances {
		prefix := "  "
		if i == il.selected {
			prefix = "→ "
		}

		var statusDot string
		if strings.ToLower(inst.Status) == "running" {
			statusDot = statusRunningStyle.Render("●")
		} else {
			statusDot = statusStoppedStyle.Render("○")
		}

		line := fmt.Sprintf("%s%s %s %s",
			prefix,
			statusDot,
			nameStyle.Render(inst.Name),
			detailStyle.Render(fmt.Sprintf("(%dc/%s)", inst.CPU, inst.Memory)))

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
