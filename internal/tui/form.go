package tui

import (
	"fmt"
	"sshbuddy/pkg/models"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type FormModel struct {
	inputs         []textinput.Model
	focused        int
	err            error
	host           *models.Host              // If editing, this is the host being edited
	isEditing      bool                     // True if editing existing host
	validationErrs []models.ValidationError  // Validation errors for current input
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
	var b strings.Builder
	
	// Clean title - show different text for edit vs add
	title := "Add New Host"
	if m.isEditing {
		title = "Edit Host"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")
	
	// Minimal input fields
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
	
	for i, field := range fields {
		// Label
		var labelText string
		if i == m.focused {
			labelText = labelFocusedStyle.Render(field.label)
		} else {
			labelText = labelStyle.Render(field.label)
		}
		
		b.WriteString(labelText)
		b.WriteString("  ")
		b.WriteString(field.input.View())
		b.WriteString("\n")
	}
	
	// Show validation errors if any
	if len(m.validationErrs) > 0 {
		b.WriteString("\n")
		errorStyle := labelStyle.Foreground(errorColor)
		for _, err := range m.validationErrs {
			b.WriteString(errorStyle.Render(fmt.Sprintf("⚠ %s", err.Message)))
			b.WriteString("\n")
		}
	}
	
	// Clean help text
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(
		keyStyle.Render("↑/↓") + descStyle.Render(" navigate  ") +
		keyStyle.Render("enter") + descStyle.Render(" submit  ") +
		keyStyle.Render("esc") + descStyle.Render(" cancel"),
	))
	
	return b.String()
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
