package tui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	primaryColor   = lipgloss.Color("#7C3AED")
	secondaryColor = lipgloss.Color("#3B82F6")
	successColor   = lipgloss.Color("#10B981")
	warningColor   = lipgloss.Color("#F59E0B")
	errorColor     = lipgloss.Color("#EF4444")
	textColor      = lipgloss.Color("#E5E7EB")
	mutedColor     = lipgloss.Color("#6B7280")
	bgColor        = lipgloss.Color("#1F2937")
	darkBgColor    = lipgloss.Color("#111827")
	cardBgColor    = lipgloss.Color("#374151")
)

// Base styles
var (
	// TitleStyle for main titles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	// SubtitleStyle for subtitles
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Padding(0, 1)

	// BoxStyle for bordered boxes
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// CardStyle for card-like containers
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2).
			Background(cardBgColor)

	// ErrorBoxStyle for error messages
	ErrorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(errorColor).
			Padding(1, 2)

	// SuccessBoxStyle for success messages
	SuccessBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(successColor).
			Padding(1, 2)
)

// Text styles
var (
	// TextStyle for normal text
	TextStyle = lipgloss.NewStyle().
			Foreground(textColor)

	// BoldTextStyle for emphasized text
	BoldTextStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor)

	// MutedTextStyle for less important text
	MutedTextStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// ErrorTextStyle for error text
	ErrorTextStyle = lipgloss.NewStyle().
			Foreground(errorColor)

	// SuccessTextStyle for success text
	SuccessTextStyle = lipgloss.NewStyle().
			Foreground(successColor)

	// WarningTextStyle for warning text
	WarningTextStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	// HighlightTextStyle for highlighted text
	HighlightTextStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)
)

// Status styles
var (
	// StatusRunningStyle for running status
	StatusRunningStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	// StatusStoppedStyle for stopped status
	StatusStoppedStyle = lipgloss.NewStyle().
				Foreground(errorColor)

	// StatusPendingStyle for pending status
	StatusPendingStyle = lipgloss.NewStyle().
				Foreground(warningColor)
)

// Instance list styles
var (
	// ListItemStyle for list items
	ListItemStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 1)

	// SelectedListItemStyle for selected list items
	SelectedListItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				Background(darkBgColor).
				Padding(0, 1)

	// InstanceNameStyle for instance names
	InstanceNameStyle = lipgloss.NewStyle().
				Foreground(textColor).
				Bold(true)

	// InstanceDetailStyle for instance details
	InstanceDetailStyle = lipgloss.NewStyle().
				Foreground(mutedColor)
)

// Form styles
var (
	// FormLabelStyle for form labels
	FormLabelStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Width(12).
			Align(lipgloss.Right)

	// FormInputStyle for form input fields
	FormInputStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(darkBgColor).
			Padding(0, 1)

	// FocusedFormInputStyle for focused input fields
	FocusedFormInputStyle = lipgloss.NewStyle().
				Foreground(textColor).
				Background(cardBgColor).
				Border(lipgloss.NormalBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	// RadioButtonStyle for radio buttons
	RadioButtonStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	// SelectedRadioButtonStyle for selected radio buttons
	SelectedRadioButtonStyle = lipgloss.NewStyle().
					Foreground(primaryColor).
					Bold(true)
)

// Help styles
var (
	// HelpKeyStyle for keyboard shortcuts
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// HelpDescStyle for help descriptions
	HelpDescStyle = lipgloss.NewStyle().
			Foreground(mutedColor)
)

// Resource bar styles
var (
	// ResourceBarLabelStyle for resource labels
	ResourceBarLabelStyle = lipgloss.NewStyle().
				Foreground(textColor).
				Width(6).
				Align(lipgloss.Left)

	// ResourceBarStyle for the bar background
	ResourceBarStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	// ResourceBarFilledStyle for filled portion
	ResourceBarFilledStyle = lipgloss.NewStyle().
				Foreground(successColor)

	// ResourceBarWarningStyle for warning level (50-80%)
	ResourceBarWarningStyle = lipgloss.NewStyle().
				Foreground(warningColor)

	// ResourceBarCriticalStyle for critical level (>80%)
	ResourceBarCriticalStyle = lipgloss.NewStyle().
				Foreground(errorColor)
)

// Footer styles
var (
	// FooterStyle for the footer bar
	FooterStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Background(darkBgColor).
			Padding(0, 1)

	// KeyHelpStyle for key hints in footer
	KeyHelpStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// SeparatorStyle for separators
	SeparatorStyle = lipgloss.NewStyle().
			Foreground(mutedColor)
)

// GetBoxStyle returns a box style with custom width
func GetBoxStyle(width int) lipgloss.Style {
	return BoxStyle.Width(width)
}

// GetCardStyle returns a card style with custom dimensions
func GetCardStyle(width, height int) lipgloss.Style {
	return CardStyle.Width(width).Height(height)
}

// ColorForPercentage returns the appropriate color based on percentage
func ColorForPercentage(pct float64) lipgloss.Color {
	switch {
	case pct >= 80:
		return errorColor
	case pct >= 50:
		return warningColor
	default:
		return successColor
	}
}

// StyleForPercentage returns the appropriate style based on percentage
func StyleForPercentage(pct float64) lipgloss.Style {
	switch {
	case pct >= 80:
		return ResourceBarCriticalStyle
	case pct >= 50:
		return ResourceBarWarningStyle
	default:
		return ResourceBarFilledStyle
	}
}
