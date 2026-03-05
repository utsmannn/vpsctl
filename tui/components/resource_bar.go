package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ResourceBar displays a progress bar for resource usage
type ResourceBar struct {
	label       string
	used        int64
	total       int64
	width       int
	showPercent bool
	unit        string
}

// NewResourceBar creates a new resource bar with the given label
func NewResourceBar(label string) ResourceBar {
	return ResourceBar{
		label:       label,
		used:        0,
		total:       100,
		width:       20,
		showPercent: true,
		unit:        "",
	}
}

// SetValues sets the used and total values for the resource bar
func (rb *ResourceBar) SetValues(used, total int64) {
	rb.used = used
	rb.total = total
}

// SetWidth sets the width of the progress bar
func (rb *ResourceBar) SetWidth(width int) {
	rb.width = width
}

// SetUnit sets the unit suffix (e.g., "cores", "GB", "MB")
func (rb *ResourceBar) SetUnit(unit string) {
	rb.unit = unit
}

// SetShowPercent sets whether to show percentage
func (rb *ResourceBar) SetShowPercent(show bool) {
	rb.showPercent = show
}

// Percentage calculates the current percentage
func (rb ResourceBar) Percentage() float64 {
	if rb.total == 0 {
		return 0
	}
	return float64(rb.used) / float64(rb.total) * 100
}

// View renders the resource bar
func (rb ResourceBar) View() string {
	// Styles
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E5E7EB")).
		Width(6).
		Align(lipgloss.Left)

	barBgStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#374151"))

	pctStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB")).
			Width(6).
			Align(lipgloss.Right)

	detailStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	// Calculate percentage
	pct := rb.Percentage()

	// Determine color based on percentage
	var barFilledStyle lipgloss.Style
	switch {
	case pct >= 80:
		barFilledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
	case pct >= 50:
		barFilledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
	default:
		barFilledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	}

	// Calculate filled width
	filledWidth := int(float64(rb.width) * pct / 100)
	if filledWidth > rb.width {
		filledWidth = rb.width
	}
	if filledWidth < 0 {
		filledWidth = 0
	}

	// Build the bar
	filled := barFilledStyle.Render(string(repeatChar('█', filledWidth)))
	empty := barBgStyle.Render(string(repeatChar('░', rb.width-filledWidth)))
	bar := fmt.Sprintf("[%s%s]", filled, empty)

	// Build percentage display
	pctDisplay := ""
	if rb.showPercent {
		pctDisplay = pctStyle.Render(fmt.Sprintf("%3.0f%%", pct))
	}

	// Build detail display
	detailDisplay := ""
	if rb.unit != "" {
		detailDisplay = detailStyle.Render(fmt.Sprintf("(%d/%d %s)", rb.used, rb.total, rb.unit))
	}

	// Combine all parts
	return fmt.Sprintf("%s %s %s %s",
		labelStyle.Render(rb.label),
		bar,
		pctDisplay,
		detailDisplay,
	)
}

// repeatChar creates a slice of repeated characters
func repeatChar(char rune, count int) []rune {
	result := make([]rune, count)
	for i := range result {
		result[i] = char
	}
	return result
}

// ResourceBarGroup manages multiple resource bars
type ResourceBarGroup struct {
	cpuBar    ResourceBar
	memoryBar ResourceBar
	diskBar   ResourceBar
	width     int
}

// NewResourceBarGroup creates a new group of resource bars
func NewResourceBarGroup() ResourceBarGroup {
	cpuBar := NewResourceBar("CPU")
	cpuBar.SetUnit("cores")

	memoryBar := NewResourceBar("RAM")
	memoryBar.SetUnit("GB")

	diskBar := NewResourceBar("Disk")
	diskBar.SetUnit("GB")

	return ResourceBarGroup{
		cpuBar:    cpuBar,
		memoryBar: memoryBar,
		diskBar:   diskBar,
		width:     20,
	}
}

// SetWidth sets the width for all bars
func (rbg *ResourceBarGroup) SetWidth(width int) {
	rbg.width = width
	rbg.cpuBar.SetWidth(width)
	rbg.memoryBar.SetWidth(width)
	rbg.diskBar.SetWidth(width)
}

// SetCPUValues sets the CPU values
func (rbg *ResourceBarGroup) SetCPUValues(used, total int64) {
	rbg.cpuBar.SetValues(used, total)
}

// SetMemoryValues sets the memory values (in GB)
func (rbg *ResourceBarGroup) SetMemoryValues(used, total int64) {
	rbg.memoryBar.SetValues(used, total)
}

// SetDiskValues sets the disk values (in GB)
func (rbg *ResourceBarGroup) SetDiskValues(used, total int64) {
	rbg.diskBar.SetValues(used, total)
}

// View renders all resource bars
func (rbg ResourceBarGroup) View() string {
	return fmt.Sprintf("%s\n%s\n%s",
		rbg.cpuBar.View(),
		rbg.memoryBar.View(),
		rbg.diskBar.View(),
	)
}
