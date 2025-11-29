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

// Available themes - optimized for both light and dark backgrounds
var themes = map[string]Theme{
	"purple": {
		Name:        "Purple Dream",
		Primary:     lipgloss.Color("#7C3AED"), // Darker, more saturated purple
		Accent:      lipgloss.Color("#059669"), // Darker green
		Error:       lipgloss.Color("#DC2626"), // Darker red
		Text:        lipgloss.Color("#1F2937"), // Dark gray for text
		Muted:       lipgloss.Color("#6B7280"), // Medium gray
		Dim:         lipgloss.Color("#9CA3AF"), // Light gray
		Border:      lipgloss.Color("#7C3AED"), // Match primary
		PingingWarn: lipgloss.Color("#D97706"), // Darker amber
	},
	"blue": {
		Name:        "Ocean Blue",
		Primary:     lipgloss.Color("#2563EB"), // Darker, more saturated blue
		Accent:      lipgloss.Color("#059669"), // Darker green
		Error:       lipgloss.Color("#DC2626"), // Darker red
		Text:        lipgloss.Color("#1F2937"), // Dark gray for text
		Muted:       lipgloss.Color("#6B7280"), // Medium gray
		Dim:         lipgloss.Color("#9CA3AF"), // Light gray
		Border:      lipgloss.Color("#2563EB"), // Match primary
		PingingWarn: lipgloss.Color("#D97706"), // Darker amber
	},
	"green": {
		Name:        "Matrix Green",
		Primary:     lipgloss.Color("#059669"), // Darker, more saturated green
		Accent:      lipgloss.Color("#0891B2"), // Darker cyan
		Error:       lipgloss.Color("#DC2626"), // Darker red
		Text:        lipgloss.Color("#1F2937"), // Dark gray for text
		Muted:       lipgloss.Color("#6B7280"), // Medium gray
		Dim:         lipgloss.Color("#9CA3AF"), // Light gray
		Border:      lipgloss.Color("#059669"), // Match primary
		PingingWarn: lipgloss.Color("#D97706"), // Darker amber
	},
	"pink": {
		Name:        "Bubblegum Pink",
		Primary:     lipgloss.Color("#DB2777"), // Darker, more saturated pink
		Accent:      lipgloss.Color("#7C3AED"), // Darker purple
		Error:       lipgloss.Color("#DC2626"), // Darker red
		Text:        lipgloss.Color("#1F2937"), // Dark gray for text
		Muted:       lipgloss.Color("#6B7280"), // Medium gray
		Dim:         lipgloss.Color("#9CA3AF"), // Light gray
		Border:      lipgloss.Color("#DB2777"), // Match primary
		PingingWarn: lipgloss.Color("#D97706"), // Darker amber
	},
	"amber": {
		Name:        "Sunset Amber",
		Primary:     lipgloss.Color("#D97706"), // Darker, more saturated amber
		Accent:      lipgloss.Color("#059669"), // Darker green
		Error:       lipgloss.Color("#DC2626"), // Darker red
		Text:        lipgloss.Color("#1F2937"), // Dark gray for text
		Muted:       lipgloss.Color("#6B7280"), // Medium gray
		Dim:         lipgloss.Color("#9CA3AF"), // Light gray
		Border:      lipgloss.Color("#D97706"), // Match primary
		PingingWarn: lipgloss.Color("#D97706"), // Match primary
	},
	"cyan": {
		Name:        "Cyber Cyan",
		Primary:     lipgloss.Color("#0891B2"), // Darker, more saturated cyan
		Accent:      lipgloss.Color("#7C3AED"), // Darker purple
		Error:       lipgloss.Color("#DC2626"), // Darker red
		Text:        lipgloss.Color("#1F2937"), // Dark gray for text
		Muted:       lipgloss.Color("#6B7280"), // Medium gray
		Dim:         lipgloss.Color("#9CA3AF"), // Light gray
		Border:      lipgloss.Color("#0891B2"), // Match primary
		PingingWarn: lipgloss.Color("#D97706"), // Darker amber
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
