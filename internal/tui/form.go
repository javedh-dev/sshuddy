package tui

import (
	"fmt"
	"sshbuddy/pkg/models"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Lipgloss helper functions (no aliases needed, use lipgloss directly)

type FormModel struct {
	inputs         []textinput.Model
	focused        int
	err            error
	host           *models.Host              // If editing, this is the host being edited
	isEditing      bool                     // True if editing existing host
	validationErrs []models.ValidationError  // Validation errors for current input
	width          int
	height         int
}

func NewFormModel() FormModel {
	var inputs []textinput.Model = make([]textinput.Model, 7)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Alias"
	inputs[0].Focus()
	inputs[0].CharLimit = 20
	inputs[0].Width = 30

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Hostname/IP"
	inputs[1].CharLimit = 50
	inputs[1].Width = 30

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "User"
	inputs[2].CharLimit = 20
	inputs[2].Width = 30

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Port (22)"
	inputs[3].CharLimit = 5
	inputs[3].Width = 30

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Identity File (optional)"
	inputs[4].CharLimit = 100
	inputs[4].Width = 30

	inputs[5] = textinput.New()
	inputs[5].Placeholder = "Proxy Jump (optional)"
	inputs[5].CharLimit = 50
	inputs[5].Width = 30

	inputs[6] = textinput.New()
	inputs[6].Placeholder = "Tags (comma separated)"
	inputs[6].CharLimit = 50
	inputs[6].Width = 30

	return FormModel{
		inputs:  inputs,
		focused: 0,
	}
}

func NewFormModelWithHost(host models.Host) FormModel {
	fm := NewFormModel()
	fm.isEditing = true

	// Pre-fill with existing host data
	fm.inputs[0].SetValue(host.Alias)
	fm.inputs[1].SetValue(host.Hostname)
	fm.inputs[2].SetValue(host.User)
	fm.inputs[3].SetValue(host.Port)
	fm.inputs[4].SetValue(host.IdentityFile)
	fm.inputs[5].SetValue(host.ProxyJump)

	// Convert tags array to comma-separated string
	if len(host.Tags) > 0 {
		fm.inputs[6].SetValue(strings.Join(host.Tags, ", "))
	}

	return fm
}

func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab, tea.KeyDown, tea.KeyEnter:
			if msg.Type == tea.KeyEnter && m.focused == len(m.inputs)-1 {
				// Validate before submitting
				host := m.GetHost()
				validationErrs := host.Validate()
				if len(validationErrs) > 0 {
					m.validationErrs = validationErrs
					return m, nil
				}
				// Submit
				m.validationErrs = nil
				return m, func() tea.Msg { return FormSubmittedMsg{host} }
			}
			m.focused++
			if m.focused >= len(m.inputs) {
				m.focused = 0
			}
		case tea.KeyShiftTab, tea.KeyUp:
			m.focused--
			if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}
		case tea.KeyRight:
			// Move to corresponding field in right column (add 4 if in left column)
			if m.focused < 4 {
				// In left column, move to right column
				newFocus := m.focused + 4
				if newFocus < len(m.inputs) {
					m.focused = newFocus
				}
			}
		case tea.KeyLeft:
			// Move to corresponding field in left column (subtract 4 if in right column)
			if m.focused >= 4 {
				// In right column, move to left column
				m.focused = m.focused - 4
			}
		}
	}

	for i := range m.inputs {
		m.inputs[i].Blur()
		if i == m.focused {
			m.inputs[i].Focus()
		}
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m FormModel) View() string {
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
	
	// Subheading - show different text for edit vs add
	subheadingText := "Add New Host"
	if m.isEditing {
		subheadingText = "Edit Host"
	}
	subheading := lipgloss.NewStyle().
		Foreground(dimColor).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render(subheadingText)
	
	separator := lipgloss.NewStyle().
		Foreground(dimColor).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render(strings.Repeat("─", boxWidth-4))
	
	header := lipgloss.JoinVertical(lipgloss.Left, asciiArt, subheading, separator)
	
	// Form fields - 2-column layout
	fields := []struct {
		label string
		input textinput.Model
	}{
		{"Alias", m.inputs[0]},
		{"Hostname", m.inputs[1]},
		{"User", m.inputs[2]},
		{"Port", m.inputs[3]},
		{"Identity File", m.inputs[4]},
		{"Proxy Jump", m.inputs[5]},
		{"Tags", m.inputs[6]},
	}
	
	// Render each field
	renderField := func(i int, field struct {
		label string
		input textinput.Model
	}) string {
		isFocused := i == m.focused
		
		// Label
		labelStyle := lipgloss.NewStyle().Foreground(textColor).Bold(true)
		if isFocused {
			labelStyle = labelStyle.Foreground(primaryColor)
		}
		labelText := labelStyle.Render(field.label + ":")
		
		// Input
		inputView := field.input.View()
		
		return lipgloss.JoinVertical(lipgloss.Left,
			labelText,
			inputView,
		)
	}
	
	// Split into two columns (first 4 fields in left, last 3 in right)
	const columnWidth = 35
	
	var leftColumn []string
	var rightColumn []string
	
	// Left column: Alias, Hostname, User, Port
	for i := 0; i < 4 && i < len(fields); i++ {
		fieldView := renderField(i, fields[i])
		leftColumn = append(leftColumn, lipgloss.NewStyle().Width(columnWidth).Render(fieldView))
		leftColumn = append(leftColumn, "") // spacing
	}
	
	// Right column: Identity File, Proxy Jump, Tags
	for i := 4; i < len(fields); i++ {
		fieldView := renderField(i, fields[i])
		rightColumn = append(rightColumn, lipgloss.NewStyle().Width(columnWidth).Render(fieldView))
		rightColumn = append(rightColumn, "") // spacing
	}
	
	// Pad right column to match left column height
	for len(rightColumn) < len(leftColumn) {
		rightColumn = append(rightColumn, "")
	}
	
	// Join columns side by side
	leftContent := lipgloss.JoinVertical(lipgloss.Left, leftColumn...)
	rightContent := lipgloss.JoinVertical(lipgloss.Left, rightColumn...)
	
	formContent := lipgloss.JoinHorizontal(lipgloss.Top, leftContent, rightContent)
	
	// Show validation errors if any
	var errorMsg string
	if len(m.validationErrs) > 0 {
		var errorLines []string
		for _, err := range m.validationErrs {
			errorLines = append(errorLines, fmt.Sprintf("• %s", err.Message))
		}
		errorMsg = lipgloss.NewStyle().
			Foreground(errorColor).
			Render("✗ " + strings.Join(errorLines, "\n  "))
	}
	
	// Footer
	keyBindings := []string{
		keyStyle.Render("↑↓/tab") + descStyle.Render(":navigate "),
		keyStyle.Render("←→") + descStyle.Render(":columns "),
		keyStyle.Render("enter") + descStyle.Render(":save "),
		keyStyle.Render("esc/q") + descStyle.Render(":cancel"),
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

func (m FormModel) GetHost() models.Host {
	// Parse tags from comma-separated string
	var tags []string
	tagsInput := strings.TrimSpace(m.inputs[6].Value())
	if tagsInput != "" {
		tagsParts := strings.Split(tagsInput, ",")
		for _, tag := range tagsParts {
			trimmed := strings.TrimSpace(tag)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	return models.Host{
		Alias:        m.inputs[0].Value(),
		Hostname:     m.inputs[1].Value(),
		User:         m.inputs[2].Value(),
		Port:         m.inputs[3].Value(),
		IdentityFile: strings.TrimSpace(m.inputs[4].Value()),
		ProxyJump:    strings.TrimSpace(m.inputs[5].Value()),
		Tags:         tags,
		Source:       "manual",
	}
}

type FormSubmittedMsg struct {
	Host models.Host
}
