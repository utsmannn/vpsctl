package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/kiatkoding/vpsctl/tui/components"
)

// App represents the TUI application
type App struct {
	lxdClient *lxd.Client
}

// NewApp creates a new TUI application
func NewApp(client *lxd.Client) *App {
	return &App{
		lxdClient: client,
	}
}

// Run starts the TUI application
func (a *App) Run() error {
	model := initialModel(a.lxdClient)
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}

// Model is the main model for the TUI
type Model struct {
	// LXD client
	lxdClient *lxd.Client

	// State
	currentView string // "dashboard", "create", "help"
	instances   []components.InstanceInfo
	resources   components.ResourceSummary
	selected    int
	width       int
	height      int

	// Components
	dashboard  components.Dashboard
	createForm components.CreateForm
	helpModal  components.HelpModal

	// Messages
	err     error
	loading bool
	status  string
}

// Messages for async operations
type tickMsg time.Time

type instancesLoadedMsg struct {
	instances []lxd.InstanceInfo
	err       error
}

type resourcesLoadedMsg struct {
	resources lxd.ResourceSummary
	err       error
}

type instanceCreatedMsg struct {
	name string
	err  error
}

type instanceActionMsg struct {
	action string
	name   string
	err    error
}

type errorMsg error

// initialModel creates the initial model
func initialModel(client *lxd.Client) Model {
	m := Model{
		lxdClient:   client,
		currentView: "dashboard",
		dashboard:   components.NewDashboard(),
		createForm:  components.NewCreateForm(),
		helpModal:   components.NewHelpModal(),
		loading:     true,
		status:      "Loading...",
	}
	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadInstances(m.lxdClient),
		loadResources(m.lxdClient),
		tickCmd(),
	)
}

// tickCmd sends periodic tick messages for refresh
func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// loadInstances loads instances from LXD
func loadInstances(client *lxd.Client) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return instancesLoadedMsg{err: fmt.Errorf("LXD client not initialized")}
		}

		instances, err := client.ListInstances()
		return instancesLoadedMsg{
			instances: instances,
			err:       err,
		}
	}
}

// loadResources loads resource summary from LXD
func loadResources(client *lxd.Client) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return resourcesLoadedMsg{err: fmt.Errorf("LXD client not initialized")}
		}

		resources, err := client.GetResourceSummary()
		return resourcesLoadedMsg{
			resources: resources,
			err:       err,
		}
	}
}

// createInstanceCmd creates a new instance
func createInstanceCmd(client *lxd.Client, opts components.CreateInstanceOptions) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return instanceCreatedMsg{err: fmt.Errorf("LXD client not initialized")}
		}

		// Parse CPU from string to int
		cpu := 1
		if opts.CPU != "" {
			if parsedCPU, err := strconv.Atoi(opts.CPU); err == nil && parsedCPU > 0 {
				cpu = parsedCPU
			}
		}

		_, err := client.CreateInstance(lxd.CreateInstanceOptions{
			Name:     opts.Name,
			Image:    opts.Image,
			CPU:      cpu,
			Memory:   opts.Memory,
			Disk:     opts.Disk,
			Type:     opts.InstanceType,
			SSHKey:   opts.SSHKey,
			Password: opts.Password,
		})

		return instanceCreatedMsg{
			name: opts.Name,
			err:  err,
		}
	}
}

// instanceActionCmd performs an action on an instance
func instanceActionCmd(client *lxd.Client, action, name string) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return instanceActionMsg{
				action: action,
				name:   name,
				err:    fmt.Errorf("LXD client not initialized"),
			}
		}

		var err error
		switch action {
		case "start":
			err = client.StartInstance(name)
		case "stop":
			err = client.StopInstance(name, false)
		case "restart":
			err = client.RestartInstance(name)
		case "delete":
			err = client.DeleteInstance(name, false)
		}

		return instanceActionMsg{
			action: action,
			name:   name,
			err:    err,
		}
	}
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.dashboard.SetSize(msg.Width, msg.Height)
		m.createForm.SetWidth(msg.Width - 10)
		m.helpModal.SetSize(msg.Width-20, 20)

	case tickMsg:
		// Refresh data periodically
		if m.currentView == "dashboard" && !m.loading {
			cmds = append(cmds,
				loadInstances(m.lxdClient),
				loadResources(m.lxdClient),
				tickCmd(),
			)
		} else {
			cmds = append(cmds, tickCmd())
		}

	case instancesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.status = fmt.Sprintf("Error loading instances: %v", msg.err)
		} else {
			m.err = nil
			m.instances = convertInstances(msg.instances)
			m.dashboard.SetInstances(m.instances)
			m.status = fmt.Sprintf("Loaded %d instances", len(m.instances))
		}

	case resourcesLoadedMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Error loading resources: %v", msg.err)
		} else {
			m.resources = convertResources(msg.resources)
			m.dashboard.SetResources(m.resources)
		}

	case instanceCreatedMsg:
		m.loading = false
		if msg.err != nil {
			m.createForm.SetError(msg.err)
			m.status = fmt.Sprintf("Failed to create instance: %v", msg.err)
		} else {
			m.currentView = "dashboard"
			m.status = fmt.Sprintf("Instance '%s' created successfully", msg.name)
			cmds = append(cmds, loadInstances(m.lxdClient))
		}

	case instanceActionMsg:
		m.loading = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Failed to %s instance: %v", msg.action, msg.err)
		} else {
			m.status = fmt.Sprintf("Instance '%s' %sed successfully", msg.name, msg.action)
			cmds = append(cmds, loadInstances(m.lxdClient))
		}

	case errorMsg:
		m.err = msg
		m.status = fmt.Sprintf("Error: %v", msg)

	case tea.KeyMsg:
		// Handle global keys
		switch msg.String() {
		case "ctrl+c", "q":
			if m.currentView != "dashboard" {
				m.currentView = "dashboard"
				return m, nil
			}
			return m, tea.Quit

		case "h", "?":
			m.helpModal.Toggle()
			return m, nil
		}

		// Handle view-specific keys
		switch m.currentView {
		case "dashboard":
			return m.updateDashboard(msg)
		case "create":
			return m.updateCreateForm(msg)
		}

		return m, nil
	}

	// Update components
	var cmd tea.Cmd
	m.dashboard, cmd = m.dashboard.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// updateDashboard handles dashboard-specific key events
func (m Model) updateDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Close help if visible
	if m.helpModal.IsVisible() {
		switch msg.String() {
		case "esc", "h", "?":
			m.helpModal.Hide()
		}
		return m, nil
	}

	switch msg.String() {
	case "n":
		m.currentView = "create"
		m.createForm = components.NewCreateForm()
		m.createForm.SetWidth(m.width - 10)
		return m, nil

	case "s":
		if inst := m.dashboard.GetSelectedInstance(); inst != nil {
			m.loading = true
			m.status = fmt.Sprintf("Starting instance '%s'...", inst.Name)
			return m, instanceActionCmd(m.lxdClient, "start", inst.Name)
		}

	case "x":
		if inst := m.dashboard.GetSelectedInstance(); inst != nil {
			m.loading = true
			m.status = fmt.Sprintf("Stopping instance '%s'...", inst.Name)
			return m, instanceActionCmd(m.lxdClient, "stop", inst.Name)
		}

	case "r":
		if inst := m.dashboard.GetSelectedInstance(); inst != nil {
			m.loading = true
			m.status = fmt.Sprintf("Restarting instance '%s'...", inst.Name)
			return m, instanceActionCmd(m.lxdClient, "restart", inst.Name)
		}

	case "d":
		if inst := m.dashboard.GetSelectedInstance(); inst != nil {
			m.loading = true
			m.status = fmt.Sprintf("Deleting instance '%s'...", inst.Name)
			return m, instanceActionCmd(m.lxdClient, "delete", inst.Name)
		}

	case "f":
		m.loading = true
		m.status = "Refreshing..."
		return m, tea.Batch(
			loadInstances(m.lxdClient),
			loadResources(m.lxdClient),
		)

	case "up", "k":
		m.dashboard.NavigateUp()

	case "down", "j":
		m.dashboard.NavigateDown()

	case "enter":
		// Could open instance details
		if inst := m.dashboard.GetSelectedInstance(); inst != nil {
			m.status = fmt.Sprintf("Selected: %s", inst.Name)
		}
	}

	return m, tea.Batch(cmds...)
}

// updateCreateForm handles create form-specific key events
func (m Model) updateCreateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.currentView = "dashboard"
		m.createForm.ClearError()
		return m, nil

	case "enter":
		if err := m.createForm.Validate(); err != nil {
			m.createForm.SetError(err)
			return m, nil
		}

		opts := m.createForm.GetValues()
		if strings.TrimSpace(opts.Name) == "" {
			m.createForm.SetError(fmt.Errorf("name is required"))
			return m, nil
		}

		m.loading = true
		m.status = fmt.Sprintf("Creating instance '%s'...", opts.Name)
		return m, createInstanceCmd(m.lxdClient, opts)
	}

	m.createForm, cmd = m.createForm.Update(msg)
	return m, cmd
}

// View renders the current view
func (m Model) View() string {
	// Styles
	loadingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

	errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	// Show loading overlay
	if m.loading && m.currentView == "dashboard" {
		return m.renderLoading()
	}

	var content string
	switch m.currentView {
	case "dashboard":
		content = m.renderDashboard()
	case "create":
		content = m.renderCreateForm()
	default:
		content = m.renderDashboard()
	}

	// Show help modal overlay
	if m.helpModal.IsVisible() {
		// Center the help modal
		helpView := m.helpModal.View()
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().
				Background(lipgloss.Color("#111827")).
				Render(helpView),
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceBackground(lipgloss.Color("#111827")),
		)
	}

	// Show status bar at bottom
	if m.err != nil {
		statusBar := errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
		content = content + "\n" + statusBar
	} else if m.status != "" {
		statusStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280")).
				Italic(true)
		statusBar := statusStyle.Render(m.status)
		content = content + "\n" + statusBar
	}

	if m.loading {
		loadingBar := loadingStyle.Render("Loading...")
		content = content + "\n" + loadingBar
	}

	return content
}

// renderDashboard renders the dashboard view
func (m Model) renderDashboard() string {
	return m.dashboard.View()
}

// renderCreateForm renders the create form view
func (m Model) renderCreateForm() string {
	// Center the form
	formView := m.createForm.View()

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		formView,
	)
}

// renderLoading renders a loading overlay
func (m Model) renderLoading() string {
	loadingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			Align(lipgloss.Center)

	spinner := "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
	frame := int(time.Now().UnixNano()/100000000) % len(spinner)

	loadingText := fmt.Sprintf("%c Loading...", spinner[frame])

	boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(2, 4)

	content := loadingStyle.Render(loadingText)
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		boxStyle.Render(content),
	)
}

// convertInstances converts LXD instances to component instances
func convertInstances(instances []lxd.InstanceInfo) []components.InstanceInfo {
	result := make([]components.InstanceInfo, len(instances))
	for i, inst := range instances {
		result[i] = components.InstanceInfo{
			Name:   inst.Name,
			Status: inst.Status,
			CPU:    inst.CPU,
			Memory: inst.Memory,
			Disk:   inst.Disk,
			Type:   inst.Type,
			IP:     inst.IP,
		}
	}
	return result
}

// convertResources converts LXD resources to component resources
func convertResources(resources lxd.ResourceSummary) components.ResourceSummary {
	return components.ResourceSummary{
		CPUTotal:    resources.CPUTotal,
		CPUUsed:     resources.CPUUsed,
		MemoryTotal: resources.MemoryTotal,
		MemoryUsed:  resources.MemoryUsed,
		DiskTotal:   resources.DiskTotal,
		DiskUsed:    resources.DiskUsed,
	}
}
