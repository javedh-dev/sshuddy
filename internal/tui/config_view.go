package tui

import (
	"fmt"
	"strings"
	"sshbuddy/internal/config"
	"sshbuddy/internal/termix"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SourceConfig represents configuration for a data source
type SourceConfig struct {
	Name        string
	Enabled     bool
	Description string
	Configurable bool // Whether this source has additional config
}

// ConfigViewModel handles the configuration UI
type ConfigViewModel struct {
	sources         []SourceConfig
	termixConfig    *termix.Config
	sshConfig       *config.SSHConfig
	focusIndex      int // Which source is focused
	editingTermix   bool
	editingSSHConfig bool
	termixInputs    []textinput.Model
	sshConfigInputs []textinput.Model
	termixFocus     int
	sshConfigFocus  int
	width           int
	height          int
	saved           bool
	errorMsg        string
}

// NewConfigViewModel creates a new configuration view model
func NewConfigViewModel() ConfigViewModel {
	// Load current Termix config
	termixCfg, err := config.LoadTermixConfig()
	if err != nil {
		termixCfg = &termix.Config{Enabled: false}
	}

	// Load current SSH config
	sshCfg, err := config.LoadSSHConfig()
	if err != nil {
		sshCfg = &config.SSHConfig{Enabled: true, ConfigPath: ""}
	}

	// Load sources config
	sourcesCfg, err := config.LoadSourcesConfig()
	if err != nil {
		sourcesCfg = &config.SourcesConfig{
			SSHBuddyEnabled:  true,
			SSHConfigEnabled: true,
			TermixEnabled:    false,
		}
	}

	// Define sources
	sources := []SourceConfig{
		{
			Name:         "SSHBuddy",
			Enabled:      sourcesCfg.SSHBuddyEnabled,
			Description:  "Hosts added manually through SSHBuddy",
			Configurable: true,
		},
		{
			Name:         "SSH Config",
			Enabled:      sourcesCfg.SSHConfigEnabled,
			Description:  "Hosts from ~/.ssh/config",
			Configurable: true,
		},
		{
			Name:         "Termix",
			Enabled:      sourcesCfg.TermixEnabled,
			Description:  "Hosts from Termix API server",
			Configurable: true,
		},
	}

	// Create Termix input fields
	termixInputs := make([]textinput.Model, 3)
	
	// Base URL input
	termixInputs[0] = textinput.New()
	termixInputs[0].Placeholder = "https://termix.example.com/api"
	termixInputs[0].SetValue(termixCfg.BaseURL)
	termixInputs[0].CharLimit = 200
	termixInputs[0].Width = 50
	
	// Username input
	termixInputs[1] = textinput.New()
	termixInputs[1].Placeholder = "username"
	termixInputs[1].SetValue(termixCfg.Username)
	termixInputs[1].CharLimit = 100
	termixInputs[1].Width = 50
	
	// Password input
	termixInputs[2] = textinput.New()
	termixInputs[2].Placeholder = "password"
	termixInputs[2].SetValue(termixCfg.Password)
	termixInputs[2].EchoMode = textinput.EchoPassword
	termixInputs[2].EchoCharacter = '•'
	termixInputs[2].CharLimit = 100
	termixInputs[2].Width = 50

	// Create SSH Config input fields
	sshConfigInputs := make([]textinput.Model, 1)
	
	// Config Path input
	sshConfigInputs[0] = textinput.New()
	sshConfigInputs[0].Placeholder = "~/.ssh/config (leave empty for default)"
	sshConfigInputs[0].SetValue(sshCfg.ConfigPath)
	sshConfigInputs[0].CharLimit = 300
	sshConfigInputs[0].Width = 50

	return ConfigViewModel{
		sources:          sources,
		termixConfig:     termixCfg,
		sshConfig:        sshCfg,
		focusIndex:       0,
		editingTermix:    false,
		editingSSHConfig: false,
		termixInputs:     termixInputs,
		sshConfigInputs:  sshConfigInputs,
		termixFocus:      0,
		sshConfigFocus:   0,
	}
}

func (m ConfigViewModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ConfigViewModel) Update(msg tea.Msg) (ConfigViewModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If editing SSH Config
		if m.editingSSHConfig {
			switch msg.String() {
			case "esc":
				m.editingSSHConfig = false
				m.errorMsg = ""
				return m, nil
			case "tab", "shift+tab", "up", "down":
				// Only one input for SSH Config, so no navigation needed
				return m, nil
			case "enter":
				// Save SSH Config
				m.sshConfig.ConfigPath = strings.TrimSpace(m.sshConfigInputs[0].Value())
				
				// Save to file
				if err := config.SaveSSHConfig(m.sshConfig); err != nil {
					m.errorMsg = fmt.Sprintf("Failed to save: %v", err)
					return m, nil
				}
				
				m.editingSSHConfig = false
				m.saved = true
				m.errorMsg = ""
				return m, nil
			}
			
			// Update the input
			m.sshConfigInputs[0], cmd = m.sshConfigInputs[0].Update(msg)
			return m, cmd
		}
		
		// If editing Termix config
		if m.editingTermix {
			switch msg.String() {
			case "esc":
				m.editingTermix = false
				m.errorMsg = ""
				return m, nil
			case "tab", "shift+tab", "up", "down":
				// Navigate between inputs
				if msg.String() == "up" || msg.String() == "shift+tab" {
					m.termixFocus--
				} else {
					m.termixFocus++
				}
				
				if m.termixFocus < 0 {
					m.termixFocus = len(m.termixInputs) - 1
				} else if m.termixFocus >= len(m.termixInputs) {
					m.termixFocus = 0
				}
				
				// Update focus
				for i := range m.termixInputs {
					if i == m.termixFocus {
						m.termixInputs[i].Focus()
					} else {
						m.termixInputs[i].Blur()
					}
				}
				return m, nil
			case "enter":
				// Save Termix config
				m.termixConfig.BaseURL = strings.TrimSpace(m.termixInputs[0].Value())
				m.termixConfig.Username = strings.TrimSpace(m.termixInputs[1].Value())
				m.termixConfig.Password = m.termixInputs[2].Value()
				
				// Validate
				if m.termixConfig.Enabled && m.termixConfig.BaseURL == "" {
					m.errorMsg = "Base URL is required when Termix is enabled"
					return m, nil
				}
				
				// Save to file
				if err := config.SaveTermixConfig(m.termixConfig); err != nil {
					m.errorMsg = fmt.Sprintf("Failed to save: %v", err)
					return m, nil
				}
				
				m.editingTermix = false
				m.saved = true
				m.errorMsg = ""
				return m, nil
			}
			
			// Update the focused input
			m.termixInputs[m.termixFocus], cmd = m.termixInputs[m.termixFocus].Update(msg)
			return m, cmd
		}
		
		// Normal navigation
		switch msg.String() {
		case "up", "k":
			if m.focusIndex > 0 {
				m.focusIndex--
			}
			m.saved = false
			m.errorMsg = ""
		case "down", "j":
			if m.focusIndex < len(m.sources)-1 {
				m.focusIndex++
			}
			m.saved = false
			m.errorMsg = ""
		case " ":
			// Toggle enabled state (space key only)
			if m.sources[m.focusIndex].Configurable {
				m.sources[m.focusIndex].Enabled = !m.sources[m.focusIndex].Enabled
				
				// Save sources configuration
				sourcesConfig := &config.SourcesConfig{
					SSHBuddyEnabled:  m.sources[0].Enabled,
					SSHConfigEnabled: m.sources[1].Enabled,
					TermixEnabled:    m.sources[2].Enabled,
				}
				
				if err := config.SaveSourcesConfig(sourcesConfig); err != nil {
					m.errorMsg = fmt.Sprintf("Failed to save: %v", err)
					m.saved = false
				} else {
					// Also update Termix config if it's the Termix source
					if m.sources[m.focusIndex].Name == "Termix" {
						m.termixConfig.Enabled = m.sources[m.focusIndex].Enabled
						config.SaveTermixConfig(m.termixConfig)
					}
					m.saved = true
					m.errorMsg = ""
				}
			}
		case "enter":
			// Toggle enabled state (enter key only)
			if m.sources[m.focusIndex].Configurable {
				m.sources[m.focusIndex].Enabled = !m.sources[m.focusIndex].Enabled
				
				// Save sources configuration
				sourcesConfig := &config.SourcesConfig{
					SSHBuddyEnabled:  m.sources[0].Enabled,
					SSHConfigEnabled: m.sources[1].Enabled,
					TermixEnabled:    m.sources[2].Enabled,
				}
				
				if err := config.SaveSourcesConfig(sourcesConfig); err != nil {
					m.errorMsg = fmt.Sprintf("Failed to save: %v", err)
					m.saved = false
				} else {
					// Also update Termix config if it's the Termix source
					if m.sources[m.focusIndex].Name == "Termix" {
						m.termixConfig.Enabled = m.sources[m.focusIndex].Enabled
						config.SaveTermixConfig(m.termixConfig)
					}
					m.saved = true
					m.errorMsg = ""
				}
			}
		case "e":
			// Edit configuration for the selected source
			if m.sources[m.focusIndex].Configurable {
				if m.sources[m.focusIndex].Name == "Termix" {
					m.editingTermix = true
					m.termixFocus = 0
					m.termixInputs[0].Focus()
					m.saved = false
				} else if m.sources[m.focusIndex].Name == "SSH Config" {
					m.editingSSHConfig = true
					m.sshConfigFocus = 0
					m.sshConfigInputs[0].Focus()
					m.saved = false
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Update all inputs for blinking cursor
	for i := range m.termixInputs {
		m.termixInputs[i], cmd = m.termixInputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	
	for i := range m.sshConfigInputs {
		m.sshConfigInputs[i], cmd = m.sshConfigInputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m ConfigViewModel) View() string {
	if m.editingTermix {
		return m.renderTermixEdit()
	}
	
	if m.editingSSHConfig {
		return m.renderSSHConfigEdit()
	}
	
	const boxWidth = 80
	
	// ASCII art header (same as main screen)
	asciiArt := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render(`╔═╗┌─┐┬ ┬  ╔╗ ┬ ┬┌┬┐┌┬┐┬ ┬
╚═╗└─┐├─┤  ╠╩╗│ │ ││ ││└┬┘
╚═╝└─┘┴ ┴  ╚═╝└─┘─┴┘─┴┘ ┴`)
	
	// Configuration subheading
	subheading := lipgloss.NewStyle().
		Foreground(dimColor).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render("Configuration")
	
	separator := lipgloss.NewStyle().
		Foreground(dimColor).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render(strings.Repeat("─", boxWidth-4))
	
	header := lipgloss.JoinVertical(lipgloss.Left, asciiArt, subheading, separator)
	
	// Sources list
	var sourceItems []string
	for i, source := range m.sources {
		isSelected := i == m.focusIndex
		sourceItems = append(sourceItems, m.renderSource(source, isSelected))
	}
	
	sourcesList := lipgloss.JoinVertical(lipgloss.Left, sourceItems...)
	
	// Status message
	var statusMsg string
	if m.saved {
		statusMsg = lipgloss.NewStyle().
			Foreground(accentColor).
			Render("✓ Configuration saved")
	} else if m.errorMsg != "" {
		statusMsg = lipgloss.NewStyle().
			Foreground(errorColor).
			Render("✗ " + m.errorMsg)
	}
	
	// Footer
	keyBindings := []string{
		keyStyle.Render("↑↓") + descStyle.Render(":navigate "),
		keyStyle.Render("space") + descStyle.Render(":toggle "),
		keyStyle.Render("e") + descStyle.Render(":edit "),
		keyStyle.Render("esc/q") + descStyle.Render(":back"),
	}
	footer := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(borderColor).
		Width(boxWidth - 4).
		Padding(0, 0).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, keyBindings...))
	
	// Combine all elements
	var content string
	if statusMsg != "" {
		content = lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			sourcesList,
			"",
			statusMsg,
			"",
			footer,
		)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			sourcesList,
			"",
			footer,
		)
	}
	
	// Wrap in a fixed-width box - match main app styling
	mainBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Width(boxWidth).
		Padding(0, 2).
		Render(content)
	
	// Center the box
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, mainBox)
}

func (m ConfigViewModel) renderSource(source SourceConfig, isSelected bool) string {
	// Status indicator
	var statusIcon string
	if source.Enabled {
		statusIcon = lipgloss.NewStyle().Foreground(accentColor).Render("✓")
	} else {
		statusIcon = lipgloss.NewStyle().Foreground(dimColor).Render("○")
	}
	
	// Source name
	nameStyle := lipgloss.NewStyle().Foreground(textColor).Bold(true)
	if isSelected {
		nameStyle = nameStyle.Foreground(primaryColor)
	}
	name := nameStyle.Render(source.Name)
	
	// Description
	desc := lipgloss.NewStyle().
		Foreground(dimColor).
		Render(source.Description)
	
	// Configurable indicator
	var configIndicator string
	if source.Configurable && isSelected {
		if source.Name == "Termix" || source.Name == "SSH Config" {
			configIndicator = lipgloss.NewStyle().
				Foreground(mutedColor).
				Render(" (press 'e' to edit)")
		}
	}
	
	// Title line
	titleLine := fmt.Sprintf("%s %s%s", statusIcon, name, configIndicator)
	
	// Add selection indicator
	if isSelected {
		titleLine = lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(primaryColor).
			Padding(0, 0, 0, 1).
			Render(titleLine)
		
		desc = lipgloss.NewStyle().
			Foreground(dimColor).
			BorderLeft(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(primaryColor).
			Padding(0, 0, 0, 1).
			Render(desc)
	} else {
		titleLine = lipgloss.NewStyle().Padding(0, 0, 0, 2).Render(titleLine)
		desc = lipgloss.NewStyle().Padding(0, 0, 0, 2).Render(desc)
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, titleLine, desc, "")
}

func (m ConfigViewModel) renderTermixEdit() string {
	const boxWidth = 80
	
	// ASCII art header (same as main screen)
	asciiArt := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render(`╔═╗┌─┐┬ ┬  ╔╗ ┬ ┬┌┬┐┌┬┐┬ ┬
╚═╗└─┐├─┤  ╠╩╗│ │ ││ ││└┬┘
╚═╝└─┘┴ ┴  ╚═╝└─┘─┴┘─┴┘ ┴`)
	
	// Termix Configuration subheading
	subheading := lipgloss.NewStyle().
		Foreground(dimColor).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render("Termix Configuration")
	
	separator := lipgloss.NewStyle().
		Foreground(dimColor).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render(strings.Repeat("─", boxWidth-4))
	
	header := lipgloss.JoinVertical(lipgloss.Left, asciiArt, subheading, separator)
	
	// Form fields
	fields := []string{
		m.renderField("Base URL", m.termixInputs[0], 0, "API endpoint (e.g., https://termix.example.com/api)"),
		m.renderField("Username", m.termixInputs[1], 1, "Your Termix username"),
		m.renderField("Password", m.termixInputs[2], 2, "Your Termix password"),
	}
	
	formContent := lipgloss.JoinVertical(lipgloss.Left, fields...)
	
	// Error message
	var errorMsg string
	if m.errorMsg != "" {
		errorMsg = lipgloss.NewStyle().
			Foreground(errorColor).
			Render("✗ " + m.errorMsg)
	}
	
	// Footer
	keyBindings := []string{
		keyStyle.Render("↑↓/tab") + descStyle.Render(":navigate "),
		keyStyle.Render("enter") + descStyle.Render(":save "),
		keyStyle.Render("esc") + descStyle.Render(":cancel"),
	}
	footer := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(borderColor).
		Width(boxWidth - 4).
		Padding(0, 0).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, keyBindings...))
	
	// Combine all elements
	var content string
	if errorMsg != "" {
		content = lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			formContent,
			"",
			errorMsg,
			"",
			footer,
		)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			formContent,
			"",
			footer,
		)
	}
	
	// Wrap in a fixed-width box - match main app styling
	mainBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Width(boxWidth).
		Padding(0, 2).
		Render(content)
	
	// Center the box
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, mainBox)
}

func (m ConfigViewModel) renderField(label string, input textinput.Model, index int, hint string) string {
	isFocused := m.termixFocus == index
	
	// Label
	labelStyle := lipgloss.NewStyle().Foreground(textColor).Bold(true)
	if isFocused {
		labelStyle = labelStyle.Foreground(primaryColor)
	}
	labelText := labelStyle.Render(label + ":")
	
	// Input
	inputView := input.View()
	
	// Hint
	hintText := lipgloss.NewStyle().
		Foreground(dimColor).
		Italic(true).
		Render(hint)
	
	return lipgloss.JoinVertical(lipgloss.Left,
		labelText,
		inputView,
		hintText,
		"",
	)
}

func (m ConfigViewModel) renderSSHConfigEdit() string {
	const boxWidth = 80
	
	// ASCII art header (same as main screen)
	asciiArt := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render(`╔═╗┌─┐┬ ┬  ╔╗ ┬ ┬┌┬┐┌┬┐┬ ┬
╚═╗└─┐├─┤  ╠╩╗│ │ ││ ││└┬┘
╚═╝└─┘┴ ┴  ╚═╝└─┘─┴┘─┴┘ ┴`)
	
	// SSH Config Configuration subheading
	subheading := lipgloss.NewStyle().
		Foreground(dimColor).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render("SSH Config Configuration")
	
	separator := lipgloss.NewStyle().
		Foreground(dimColor).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render(strings.Repeat("─", boxWidth-4))
	
	header := lipgloss.JoinVertical(lipgloss.Left, asciiArt, subheading, separator)
	
	// Form field
	field := m.renderField("Config Path", m.sshConfigInputs[0], 0, "Path to SSH config file (leave empty for default ~/.ssh/config)")
	
	formContent := lipgloss.JoinVertical(lipgloss.Left, field)
	
	// Error message
	var errorMsg string
	if m.errorMsg != "" {
		errorMsg = lipgloss.NewStyle().
			Foreground(errorColor).
			Render("✗ " + m.errorMsg)
	}
	
	// Footer
	keyBindings := []string{
		keyStyle.Render("enter") + descStyle.Render(":save "),
		keyStyle.Render("esc") + descStyle.Render(":cancel"),
	}
	footer := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(borderColor).
		Width(boxWidth - 4).
		Padding(0, 0).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, keyBindings...))
	
	// Combine all elements
	var content string
	if errorMsg != "" {
		content = lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			formContent,
			"",
			errorMsg,
			"",
			footer,
		)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			formContent,
			"",
			footer,
		)
	}
	
	// Wrap in a fixed-width box - match main app styling
	mainBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Width(boxWidth).
		Padding(0, 2).
		Render(content)
	
	// Center the box
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, mainBox)
}
