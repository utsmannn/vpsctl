package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ResourceSummary represents host resource summary
type ResourceSummary struct {
	CPUTotal    int64
	CPUUsed     int64
	MemoryTotal int64 // in MB
	MemoryUsed  int64 // in MB
	DiskTotal   int64 // in GB
	DiskUsed    int64 // in GB
}

// Dashboard is the main dashboard view component
type Dashboard struct {
	width       int
	height      int
	resourceBar ResourceBarGroup
	instanceList InstanceList
	resources   ResourceSummary
	focusedArea string // "resources" or "instances"
}

// NewDashboard creates a new dashboard component
func NewDashboard() Dashboard {
	return Dashboard{
		width:       80,
		height:      24,
		resourceBar: NewResourceBarGroup(),
		instanceList: NewInstanceList(),
		focusedArea: "instances",
	}
}

// SetSize sets the dimensions of the dashboard
func (d *Dashboard) SetSize(width, height int) {
	d.width = width
	d.height = height

	// Calculate component sizes
	listWidth := width - 6
	listHeight := height - 18

	d.resourceBar.SetWidth(listWidth - 20)
	d.instanceList.SetSize(listWidth, listHeight)
}

// SetResources sets the resource summary data
func (d *Dashboard) SetResources(resources ResourceSummary) {
	d.resources = resources
	d.resourceBar.SetCPUValues(resources.CPUUsed, resources.CPUTotal)
	d.resourceBar.SetMemoryValues(resources.MemoryUsed/1024, resources.MemoryTotal/1024)
	d.resourceBar.SetDiskValues(resources.DiskUsed, resources.DiskTotal)
}

// SetInstances sets the instance list data
func (d *Dashboard) SetInstances(instances []InstanceInfo) {
	d.instanceList.SetInstances(instances)
}

// SetFocusedArea sets which area is focused
func (d *Dashboard) SetFocusedArea(area string) {
	d.focusedArea = area
	d.instanceList.SetFocused(area == "instances")
}

// InstanceList returns a pointer to the instance list for direct access
func (d *Dashboard) InstanceList() *InstanceList {
	return &d.instanceList
}

// Update handles messages for the dashboard
func (d Dashboard) Update(msg tea.Msg) (Dashboard, tea.Cmd) {
	var cmd tea.Cmd

	// Update instance list if it's focused
	if d.focusedArea == "instances" {
		d.instanceList, cmd = d.instanceList.Update(msg)
	}

	return d, cmd
}

// View renders the dashboard
func (d Dashboard) View() string {
	// Styles
	mainBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(1, 1)

	titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			Align(lipgloss.Center).
			Width(d.width - 4)

	sectionTitleStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#3B82F6")).
				Bold(true)

	footerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Background(lipgloss.Color("#1F2937")).
			Padding(0, 1).
			Width(d.width - 4)

	keyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true)

	// Build title
	title := titleStyle.Render("VPSCTL Dashboard")
	separator := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#374151")).
			Render(strings.Repeat("─", d.width-6))

	// Build resources section
	resourcesTitle := sectionTitleStyle.Render("Resources:")
	resourcesContent := d.resourceBar.View()

	// Build instances section
	runningCount := d.instanceList.RunningCount()
	totalCount := d.instanceList.Count()
	instancesTitle := sectionTitleStyle.Render(fmt.Sprintf("Instances: (%d running / %d total)", runningCount, totalCount))
	instancesContent := d.instanceList.View()

	// Build footer
	footerKeys := fmt.Sprintf("[%s]New [%s]Start [%s]Stop [%s]Delete [%s]Refresh [%s]Help [%s]Quit",
		keyStyle.Render("N"),
		keyStyle.Render("S"),
		keyStyle.Render("X"),
		keyStyle.Render("D"),
		keyStyle.Render("F"),
		keyStyle.Render("H"),
		keyStyle.Render("Q"))
	footer := footerStyle.Render(footerKeys)

	// Combine all parts
	var parts []string
	parts = append(parts, title)
	parts = append(parts, separator)
	parts = append(parts, resourcesTitle)
	parts = append(parts, resourcesContent)
	parts = append(parts, separator)
	parts = append(parts, instancesTitle)
	parts = append(parts, instancesContent)
	parts = append(parts, separator)
	parts = append(parts, footer)

	content := strings.Join(parts, "\n")

	return mainBoxStyle.Render(content)
}

// ViewSplit renders the dashboard in a split layout
func (d Dashboard) ViewSplit() string {
	// Styles
	mainBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(1, 1).
			Width(d.width)

	titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			Align(lipgloss.Center).
			Width(d.width - 4)

	sectionStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")).
			Padding(1, 1)

	resourceBoxStyle := sectionStyle.Width(d.width/2 - 4)
	instanceBoxStyle := sectionStyle.Width(d.width/2 - 4)

	// Title
	title := titleStyle.Render("VPSCTL Dashboard")

	// Resources box
	resourcesTitle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Bold(true).
			Render("Resources")
	resourcesContent := d.resourceBar.View()
	resourcesBox := resourceBoxStyle.Render(resourcesTitle + "\n\n" + resourcesContent)

	// Instances box
	runningCount := d.instanceList.RunningCount()
	totalCount := d.instanceList.Count()
	instancesTitle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Bold(true).
			Render(fmt.Sprintf("Instances (%d/%d)", runningCount, totalCount))
	instancesBox := instanceBoxStyle.Render(instancesTitle + "\n\n" + d.instanceList.ViewCompact())

	// Combine side by side
	row := lipgloss.JoinHorizontal(lipgloss.Top, resourcesBox, "  ", instancesBox)

	// Footer
	footer := ViewCompact()

	// Final layout
	return mainBoxStyle.Render(title + "\n\n" + row + "\n\n" + footer)
}

// GetSelectedInstance returns the currently selected instance
func (d Dashboard) GetSelectedInstance() *InstanceInfo {
	return d.instanceList.Selected()
}

// NavigateUp moves selection up
func (d *Dashboard) NavigateUp() {
	if d.focusedArea == "instances" {
		d.instanceList, _ = d.instanceList.Update(tea.KeyMsg{Type: tea.KeyUp})
	}
}

// NavigateDown moves selection down
func (d *Dashboard) NavigateDown() {
	if d.focusedArea == "instances" {
		d.instanceList, _ = d.instanceList.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
}
