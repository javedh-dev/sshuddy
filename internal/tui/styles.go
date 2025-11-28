package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a color scheme
type Theme struct {
	Name        string
	Primary     lipgloss.Color
	Accent      lipgloss.Color
	Error       lipgloss.Color
	Text        lipgloss.Color
	Muted       lipgloss.Color
	Dim         lipgloss.Color
	Border      lipgloss.Color
	PingingWarn lipgloss.Color
}

// Available themes
var themes = map[string]Theme{
	"purple": {
		Name:        "Purple Dream",
		Primary:     lipgloss.Color("#A78BFA"),
		Accent:      lipgloss.Color("#34D399"),
		Error:       lipgloss.Color("#F87171"),
		Text:        lipgloss.Color("#E5E7EB"),
		Muted:       lipgloss.Color("#9CA3AF"),
		Dim:         lipgloss.Color("#6B7280"),
		Border:      lipgloss.Color("#374151"),
		PingingWarn: lipgloss.Color("#FBBF24"),
	},
	"blue": {
		Name:        "Ocean Blue",
		Primary:     lipgloss.Color("#60A5FA"),
		Accent:      lipgloss.Color("#34D399"),
		Error:       lipgloss.Color("#F87171"),
		Text:        lipgloss.Color("#E5E7EB"),
		Muted:       lipgloss.Color("#9CA3AF"),
		Dim:         lipgloss.Color("#6B7280"),
		Border:      lipgloss.Color("#1E3A8A"),
		PingingWarn: lipgloss.Color("#FBBF24"),
	},
	"green": {
		Name:        "Matrix Green",
		Primary:     lipgloss.Color("#10B981"),
		Accent:      lipgloss.Color("#34D399"),
		Error:       lipgloss.Color("#F87171"),
		Text:        lipgloss.Color("#D1FAE5"),
		Muted:       lipgloss.Color("#6EE7B7"),
		Dim:         lipgloss.Color("#059669"),
		Border:      lipgloss.Color("#064E3B"),
		PingingWarn: lipgloss.Color("#FBBF24"),
	},
	"pink": {
		Name:        "Bubblegum Pink",
		Primary:     lipgloss.Color("#F472B6"),
		Accent:      lipgloss.Color("#A78BFA"),
		Error:       lipgloss.Color("#F87171"),
		Text:        lipgloss.Color("#FCE7F3"),
		Muted:       lipgloss.Color("#F9A8D4"),
		Dim:         lipgloss.Color("#DB2777"),
		Border:      lipgloss.Color("#831843"),
		PingingWarn: lipgloss.Color("#FBBF24"),
	},
	"amber": {
		Name:        "Sunset Amber",
		Primary:     lipgloss.Color("#F59E0B"),
		Accent:      lipgloss.Color("#10B981"),
		Error:       lipgloss.Color("#EF4444"),
		Text:        lipgloss.Color("#FEF3C7"),
		Muted:       lipgloss.Color("#FCD34D"),
		Dim:         lipgloss.Color("#D97706"),
		Border:      lipgloss.Color("#78350F"),
		PingingWarn: lipgloss.Color("#FBBF24"),
	},
	"cyan": {
		Name:        "Cyber Cyan",
		Primary:     lipgloss.Color("#06B6D4"),
		Accent:      lipgloss.Color("#A78BFA"),
		Error:       lipgloss.Color("#F87171"),
		Text:        lipgloss.Color("#CFFAFE"),
		Muted:       lipgloss.Color("#67E8F9"),
		Dim:         lipgloss.Color("#0891B2"),
		Border:      lipgloss.Color("#164E63"),
		PingingWarn: lipgloss.Color("#FBBF24"),
	},
}

var currentTheme = themes["purple"]

var (
	// Minimal color palette
	primaryColor   = currentTheme.Primary
	accentColor    = currentTheme.Accent
	errorColor     = currentTheme.Error
	textColor      = currentTheme.Text
	mutedColor     = currentTheme.Muted
	dimColor       = currentTheme.Dim
	borderColor    = currentTheme.Border

	// Clean title style
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginBottom(1)

	// Form label style
	labelStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Width(10)

	labelFocusedStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				Width(10)

	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			MarginTop(1)

	// Minimal box style for forms
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2)

	// Focused item style
	focusedStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// Instructions style
	instructionsStyle = lipgloss.NewStyle().
				Foreground(dimColor).
				Padding(1, 0)

	// Key binding styles
	keyStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	descStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	// Status indicator styles (text-based)
	statusOnlineStyle = lipgloss.NewStyle().
				Foreground(accentColor)

	statusOfflineStyle = lipgloss.NewStyle().
				Foreground(errorColor)

	statusUnknownStyle = lipgloss.NewStyle().
				Foreground(dimColor)

	statusPingingStyle = lipgloss.NewStyle().
				Foreground(currentTheme.PingingWarn)
)

// ApplyTheme updates all styles with the selected theme
func ApplyTheme(themeName string) {
	theme, exists := themes[themeName]
	if !exists {
		theme = themes["purple"] // Default fallback
	}
	
	currentTheme = theme
	
	// Update color variables
	primaryColor = theme.Primary
	accentColor = theme.Accent
	errorColor = theme.Error
	textColor = theme.Text
	mutedColor = theme.Muted
	dimColor = theme.Dim
	borderColor = theme.Border
	
	// Update all styles
	titleStyle = titleStyle.Foreground(primaryColor)
	subtitleStyle = subtitleStyle.Foreground(mutedColor)
	labelStyle = labelStyle.Foreground(textColor)
	labelFocusedStyle = labelFocusedStyle.Foreground(primaryColor)
	helpStyle = helpStyle.Foreground(dimColor)
	boxStyle = boxStyle.BorderForeground(borderColor)
	focusedStyle = focusedStyle.Foreground(primaryColor)
	instructionsStyle = instructionsStyle.Foreground(dimColor)
	keyStyle = keyStyle.Foreground(primaryColor)
	descStyle = descStyle.Foreground(dimColor)
	statusOnlineStyle = statusOnlineStyle.Foreground(accentColor)
	statusOfflineStyle = statusOfflineStyle.Foreground(errorColor)
	statusUnknownStyle = statusUnknownStyle.Foreground(dimColor)
	statusPingingStyle = statusPingingStyle.Foreground(theme.PingingWarn)
}

// GetThemeNames returns a list of available theme names
func GetThemeNames() []string {
	return []string{"purple", "blue", "green", "pink", "amber", "cyan"}
}

// GetCurrentTheme returns the current theme
func GetCurrentTheme() Theme {
	return currentTheme
}
