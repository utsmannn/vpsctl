package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpModal is a modal component that displays keyboard shortcuts
type HelpModal struct {
	visible bool
	width   int
	height  int
}

// NewHelpModal creates a new help modal component
func NewHelpModal() HelpModal {
	return HelpModal{
		visible: false,
		width:   50,
		height:  20,
	}
}

// Toggle toggles the visibility of the help modal
func (h *HelpModal) Toggle() {
	h.visible = !h.visible
}

// Show shows the help modal
func (h *HelpModal) Show() {
	h.visible = true
}

// Hide hides the help modal
func (h *HelpModal) Hide() {
	h.visible = false
}

// IsVisible returns whether the modal is visible
func (h HelpModal) IsVisible() bool {
	return h.visible
}

// SetSize sets the dimensions of the modal
func (h *HelpModal) SetSize(width, height int) {
	h.width = width
	h.height = height
}

// View renders the help modal
func (h HelpModal) View() string {
	if !h.visible {
		return ""
	}

	// Styles
	boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(1, 2).
			Width(h.width)

	titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			Align(lipgloss.Center).
			MarginBottom(1)

	sectionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Bold(true).
			MarginTop(1)

	keyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			Width(12)

	descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	footerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Align(lipgloss.Center).
			MarginTop(1)

	// Build content
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("Help - Keyboard Shortcuts"))
	content.WriteString("\n")

	// Navigation section
	content.WriteString(sectionStyle.Render("Navigation:"))
	content.WriteString("\n")
	navKeys := []struct {
		key  string
		desc string
	}{
		{"↑/↓ or k/j", "Move selection"},
		{"Enter", "Select/Open"},
		{"Esc", "Back/Cancel"},
		{"Tab", "Next field"},
		{"Shift+Tab", "Previous field"},
	}

	for _, item := range navKeys {
		line := fmt.Sprintf("  %s %s",
			keyStyle.Render(item.key),
			descStyle.Render(item.desc))
		content.WriteString(line)
		content.WriteString("\n")
	}

	// Actions section
	content.WriteString(sectionStyle.Render("Actions:"))
	content.WriteString("\n")
	actionKeys := []struct {
		key  string
		desc string
	}{
		{"N", "New VPS"},
		{"S", "Start selected"},
		{"X", "Stop selected"},
		{"R", "Restart selected"},
		{"D", "Delete selected"},
		{"E", "Shell/Exec into instance"},
		{"F", "Refresh list"},
	}

	for _, item := range actionKeys {
		line := fmt.Sprintf("  %s %s",
			keyStyle.Render(item.key),
			descStyle.Render(item.desc))
		content.WriteString(line)
		content.WriteString("\n")
	}

	// General section
	content.WriteString(sectionStyle.Render("General:"))
	content.WriteString("\n")
	generalKeys := []struct {
		key  string
		desc string
	}{
		{"H or ?", "Toggle help"},
		{"Q or Ctrl+C", "Quit"},
	}

	for _, item := range generalKeys {
		line := fmt.Sprintf("  %s %s",
			keyStyle.Render(item.key),
			descStyle.Render(item.desc))
		content.WriteString(line)
		content.WriteString("\n")
	}

	// Footer
	content.WriteString("\n")
	content.WriteString(footerStyle.Render("Press Esc or H to close"))

	return boxStyle.Render(content.String())
}

// ViewCompact renders a compact help footer
func ViewCompact() string {
	// Styles
	keyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true)

	separatorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#374151"))

	// Build compact help
	keys := []struct {
		key  string
		desc string
	}{
		{"N", "New"},
		{"S", "Start"},
		{"X", "Stop"},
		{"D", "Delete"},
		{"R", "Refresh"},
		{"H", "Help"},
		{"Q", "Quit"},
	}

	var parts []string
	for _, item := range keys {
		part := fmt.Sprintf("[%s]%s", keyStyle.Render(item.key), item.desc)
		parts = append(parts, part)
	}

	separator := separatorStyle.Render(" | ")
	return separatorStyle.Render(" ") + strings.Join(parts, separator)
}

// ViewStatusBar renders a status bar with help hints
func ViewStatusBar(message string) string {
	// Styles
	barStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("#1F2937")).
			Foreground(lipgloss.Color("#E5E7EB")).
			Padding(0, 1).
			Width(80)

	keyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true)

	// Build status bar
	hints := fmt.Sprintf("[%s]New [%s]Start [%s]Stop [%s]Delete [%s]Refresh [%s]Help [%s]Quit",
		keyStyle.Render("N"),
		keyStyle.Render("S"),
		keyStyle.Render("X"),
		keyStyle.Render("D"),
		keyStyle.Render("F"),
		keyStyle.Render("H"),
		keyStyle.Render("Q"))

	if message != "" {
		return barStyle.Render(message + " | " + hints)
	}

	return barStyle.Render(hints)
}
