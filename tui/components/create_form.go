package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CreateInstanceOptions holds the form values for creating a new instance
type CreateInstanceOptions struct {
	Name         string
	Image        string
	CPU          string
	Memory       string
	Disk         string
	InstanceType string // "container" or "vm"
	SSHKey       string
	Password     string
}

// CreateForm is a form component for creating new VPS instances
type CreateForm struct {
	focusedIndex  int
	inputs        []textinput.Model
	labels        []string
	instanceType  int // 0 = container, 1 = vm
	width         int
	focused       bool
	submitLabel   string
	cancelLabel   string
	err           error
}

// NewCreateForm creates a new create form component
func NewCreateForm() CreateForm {
	// Create input fields
	inputs := make([]textinput.Model, 7)
	labels := []string{
		"Name",
		"Image",
		"CPU",
		"Memory",
		"Disk",
		"SSH Key",
		"Password",
	}

	// Name input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "my-vps"
	inputs[0].Focus()
	inputs[0].CharLimit = 32
	inputs[0].Width = 30

	// Image input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "ubuntu:24.04"
	inputs[1].SetValue("ubuntu:24.04")
	inputs[1].CharLimit = 64
	inputs[1].Width = 30

	// CPU input
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "1"
	inputs[2].SetValue("1")
	inputs[2].CharLimit = 4
	inputs[2].Width = 30

	// Memory input
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "512MB"
	inputs[3].SetValue("512MB")
	inputs[3].CharLimit = 16
	inputs[3].Width = 30

	// Disk input
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "10GB"
	inputs[4].SetValue("10GB")
	inputs[4].CharLimit = 16
	inputs[4].Width = 30

	// SSH Key input (optional)
	inputs[5] = textinput.New()
	inputs[5].Placeholder = "ssh-rsa AAAA... (optional)"
	inputs[5].CharLimit = 512
	inputs[5].Width = 30
	inputs[5].EchoMode = textinput.EchoNormal

	// Password input
	inputs[6] = textinput.New()
	inputs[6].Placeholder = "root password (optional)"
	inputs[6].CharLimit = 64
	inputs[6].Width = 30
	inputs[6].EchoMode = textinput.EchoPassword
	inputs[6].EchoCharacter = '*'

	return CreateForm{
		focusedIndex: 0,
		inputs:       inputs,
		labels:       labels,
		instanceType: 0,
		width:        50,
		focused:      true,
		submitLabel:  "Create",
		cancelLabel:  "Cancel",
	}
}

// SetWidth sets the form width
func (cf *CreateForm) SetWidth(width int) {
	cf.width = width
	for i := range cf.inputs {
		cf.inputs[i].Width = width - 20
	}
}

// SetFocused sets whether the form is focused
func (cf *CreateForm) SetFocused(focused bool) {
	cf.focused = focused
}

// Update handles messages for the form
func (cf CreateForm) Update(msg tea.Msg) (CreateForm, tea.Cmd) {
	var cmds []tea.Cmd

	if !cf.focused {
		return cf, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			cf.focusedIndex++
			if cf.focusedIndex > len(cf.inputs) {
				cf.focusedIndex = 0
			}
			cf.updateFocus()
		case "shift+tab", "up":
			cf.focusedIndex--
			if cf.focusedIndex < 0 {
				cf.focusedIndex = len(cf.inputs)
			}
			cf.updateFocus()
		case "left":
			if cf.focusedIndex == len(cf.inputs) {
				cf.instanceType = 0
			}
		case "right":
			if cf.focusedIndex == len(cf.inputs) {
				cf.instanceType = 1
			}
		case "enter":
			// Handle form submission in parent component
		case "esc":
			// Handle cancel in parent component
		}
	}

	// Update the focused input
	if cf.focusedIndex >= 0 && cf.focusedIndex < len(cf.inputs) {
		var cmd tea.Cmd
		cf.inputs[cf.focusedIndex], cmd = cf.inputs[cf.focusedIndex].Update(msg)
		cmds = append(cmds, cmd)
	}

	return cf, tea.Batch(cmds...)
}

// updateFocus updates which input is focused
func (cf *CreateForm) updateFocus() {
	for i := range cf.inputs {
		if i == cf.focusedIndex {
			cf.inputs[i].Focus()
		} else {
			cf.inputs[i].Blur()
		}
	}
}

// GetValues returns the current form values
func (cf CreateForm) GetValues() CreateInstanceOptions {
	instanceType := "container"
	if cf.instanceType == 1 {
		instanceType = "vm"
	}

	return CreateInstanceOptions{
		Name:         cf.inputs[0].Value(),
		Image:        cf.inputs[1].Value(),
		CPU:          cf.inputs[2].Value(),
		Memory:       cf.inputs[3].Value(),
		Disk:         cf.inputs[4].Value(),
		InstanceType: instanceType,
		SSHKey:       cf.inputs[5].Value(),
		Password:     cf.inputs[6].Value(),
	}
}

// SetError sets an error message to display
func (cf *CreateForm) SetError(err error) {
	cf.err = err
}

// ClearError clears the error message
func (cf *CreateForm) ClearError() {
	cf.err = nil
}

// Validate validates the form input
func (cf CreateForm) Validate() error {
	values := cf.GetValues()

	if strings.TrimSpace(values.Name) == "" {
		return fmt.Errorf("name is required")
	}

	if strings.TrimSpace(values.Image) == "" {
		return fmt.Errorf("image is required")
	}

	// Validate name format
	for _, c := range values.Name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return fmt.Errorf("name can only contain letters, numbers, hyphens, and underscores")
		}
	}

	return nil
}

// View renders the create form
func (cf CreateForm) View() string {
	// Styles
	boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(1, 2).
			Width(cf.width)

	titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB")).
			Width(10).
			Align(lipgloss.Right)

	inputStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	focusedLabelStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7C3AED")).
				Bold(true).
				Width(10).
				Align(lipgloss.Right)

	errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			MarginTop(1)

	footerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1)

	// Build title
	title := titleStyle.Render("Create New VPS")

	// Build form fields
	var fields []string
	for i, input := range cf.inputs {
		var labelText string
		if i == cf.focusedIndex {
			labelText = focusedLabelStyle.Render(cf.labels[i] + ":")
		} else {
			 labelText = labelStyle.Render(cf.labels[i] + ":")
		}

		field := fmt.Sprintf("%s %s",
			labelText,
			inputStyle.Render(input.View()))
		fields = append(fields, field)
	}

	// Build instance type selector
	typeLabel := labelStyle.Render("Type:")
	typeFocused := cf.focusedIndex == len(cf.inputs)

	containerStyle := lipgloss.NewStyle()
	vmStyle := lipgloss.NewStyle()

	if typeFocused {
		if cf.instanceType == 0 {
			containerStyle = containerStyle.
				Foreground(lipgloss.Color("#7C3AED")).
				Bold(true)
			vmStyle = vmStyle.Foreground(lipgloss.Color("#6B7280"))
		} else {
			containerStyle = containerStyle.Foreground(lipgloss.Color("#6B7280"))
			vmStyle = vmStyle.
				Foreground(lipgloss.Color("#7C3AED")).
				Bold(true)
		}
	} else {
		if cf.instanceType == 0 {
			containerStyle = containerStyle.Foreground(lipgloss.Color("#10B981"))
			vmStyle = vmStyle.Foreground(lipgloss.Color("#6B7280"))
		} else {
			containerStyle = containerStyle.Foreground(lipgloss.Color("#6B7280"))
			vmStyle = vmStyle.Foreground(lipgloss.Color("#10B981"))
		}
	}

	typeSelector := fmt.Sprintf("%s (•) Container  ( ) VM", typeLabel)
	if cf.instanceType == 1 {
		typeSelector = fmt.Sprintf("%s ( ) Container  (•) VM", typeLabel)
	}

	if typeFocused {
		typeSelector = fmt.Sprintf("%s %s  %s",
			focusedLabelStyle.Render("Type:"),
			containerStyle.Render("(•) Container"),
			vmStyle.Render("( ) VM"))
		if cf.instanceType == 1 {
			typeSelector = fmt.Sprintf("%s %s  %s",
				focusedLabelStyle.Render("Type:"),
				containerStyle.Render("( ) Container"),
				vmStyle.Render("(•) VM"))
		}
	}

	fields = append(fields, typeSelector)

	// Build error message
	var errorMsg string
	if cf.err != nil {
		errorMsg = errorStyle.Render(fmt.Sprintf("Error: %v", cf.err))
	}

	// Build footer
	footer := footerStyle.Render("[Enter] Create  [Esc] Cancel  [Tab] Next Field")

	// Combine all parts
	content := title + "\n\n"
	content += strings.Join(fields, "\n")
	if errorMsg != "" {
		content += "\n" + errorMsg
	}
	content += "\n\n" + footer

	return boxStyle.Render(content)
}

// IsSubmitFocused returns true if the submit button area is focused
func (cf CreateForm) IsSubmitFocused() bool {
	return cf.focusedIndex == len(cf.inputs)
}

// FocusFirst focuses the first input
func (cf *CreateForm) FocusFirst() {
	cf.focusedIndex = 0
	cf.updateFocus()
}
