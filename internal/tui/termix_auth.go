package tui

import (
	"fmt"
	"strings"
	"sshbuddy/internal/config"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TermixAuthModel struct {
	inputs    []textinput.Model
	focused   int
	err       error
	width     int
	height    int
	authError string
}

func NewTermixAuthModel() TermixAuthModel {
	var inputs []textinput.Model = make([]textinput.Model, 2)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Username"
	inputs[0].Focus()
	inputs[0].CharLimit = 50
	inputs[0].Width = 40

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Password"
	inputs[1].EchoMode = textinput.EchoPassword
	inputs[1].EchoCharacter = '•'
	inputs[1].CharLimit = 100
	inputs[1].Width = 40

	return TermixAuthModel{
		inputs:  inputs,
		focused: 0,
	}
}

func (m TermixAuthModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TermixAuthModel) Update(msg tea.Msg) (TermixAuthModel, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab, tea.KeyDown:
			m.focused++
			if m.focused >= len(m.inputs) {
				m.focused = 0
			}
		case tea.KeyShiftTab, tea.KeyUp:
			m.focused--
			if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}
		case tea.KeyEnter:
			// Submit credentials
			username := strings.TrimSpace(m.inputs[0].Value())
			password := m.inputs[1].Value()
			
			if username == "" || password == "" {
				m.authError = "Username and password are required"
				return m, nil
			}
			
			// Attempt authentication
			err := config.AuthenticateTermix(username, password)
			if err != nil {
				m.authError = fmt.Sprintf("Authentication failed: %v", err)
				// Clear password field on error
				m.inputs[1].SetValue("")
				return m, nil
			}
			
			// Success - return message to reload config
			return m, func() tea.Msg { return TermixAuthSuccessMsg{} }
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

func (m TermixAuthModel) View() string {
	const boxWidth = 60
	
	// Header
	title := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render("Termix Authentication Required")
	
	subtitle := lipgloss.NewStyle().
		Foreground(mutedColor).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render("Your session has expired. Please log in again.")
	
	separator := lipgloss.NewStyle().
		Foreground(dimColor).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render(strings.Repeat("─", boxWidth-4))
	
	header := lipgloss.JoinVertical(lipgloss.Left, title, subtitle, separator)
	
	// Form fields
	fields := []struct {
		label string
		input textinput.Model
	}{
		{"Username", m.inputs[0]},
		{"Password", m.inputs[1]},
	}
	
	var formFields []string
	for i, field := range fields {
		isFocused := i == m.focused
		
		// Label
		labelStyle := lipgloss.NewStyle().Foreground(textColor).Bold(true)
		if isFocused {
			labelStyle = labelStyle.Foreground(primaryColor)
		}
		labelText := labelStyle.Render(field.label + ":")
		
		// Input
		inputView := field.input.View()
		
		fieldView := lipgloss.JoinVertical(lipgloss.Left, labelText, inputView)
		formFields = append(formFields, fieldView)
	}
	
	formContent := lipgloss.JoinVertical(lipgloss.Left, formFields...)
	
	// Show error if any
	var errorMsg string
	if m.authError != "" {
		errorMsg = lipgloss.NewStyle().
			Foreground(errorColor).
			Render("✗ " + m.authError)
	}
	
	// Footer
	keyBindings := []string{
		keyStyle.Render("↑↓/tab") + descStyle.Render(":navigate "),
		keyStyle.Render("enter") + descStyle.Render(":login "),
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
	
	// Wrap in a box
	mainBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Width(boxWidth).
		Padding(1, 2).
		Render(content)
	
	// Center the box
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, mainBox)
}

type TermixAuthSuccessMsg struct{}
