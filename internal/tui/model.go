package tui

import (
	"fmt"
	"strings"
	"sshbuddy/internal/config"
	"sshbuddy/pkg/models"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	stateList sessionState = iota
	stateForm
	stateConfirmDelete
	stateConfigError
)

type item struct {
	host     models.Host
	status   string // Ping status indicator
	pinging  bool   // Is currently being pinged
	pingTime string // Ping time in ms
}

func (i item) Title() string { 
	// Colored dot based on ping status
	var statusText string
	if i.pinging {
		// Yellow dot for pinging in progress
		statusText = statusPingingStyle.Render("●")
	} else {
		switch i.status {
		case "🟢":
			statusText = statusOnlineStyle.Render("●")
		case "🔴":
			statusText = statusOfflineStyle.Render("●")
		default:
			statusText = statusUnknownStyle.Render("○")
		}
	}
	
	// Add ping time if available
	if i.pingTime != "" {
		return fmt.Sprintf("%s %s %s", statusText, i.host.Alias, 
			lipgloss.NewStyle().Foreground(dimColor).Render(fmt.Sprintf("(%s)", i.pingTime)))
	}
	return fmt.Sprintf("%s %s", statusText, i.host.Alias)
}

func (i item) Description() string {
	port := i.host.Port
	if port == "" {
		port = "22"
	}
	return fmt.Sprintf("%s@%s:%s", i.host.User, i.host.Hostname, port)
}

func (i item) FilterValue() string { return i.host.Alias + i.host.Hostname }

type Model struct {
	list              list.Model
	form              FormModel
	state             sessionState
	config            *models.Config
	pingStatus        map[string]bool          // track ping status for each host
	pinging           map[string]bool          // track which hosts are currently being pinged
	pingTimes         map[string]string        // track ping times for each host
	width             int
	height            int
	selectedHost      *models.Host              // Host to connect to after quitting
	editingIndex      int                      // Index of host being edited (-1 if adding new)
	deleteConfirmHost *models.Host              // Host pending deletion confirmation
	deleteConfirmIdx  int                      // Index of host pending deletion
	configErrors      []models.ValidationError  // Config validation errors
}

func NewModel() Model {
	cfg, err := config.LoadConfig()
	var validationErrors []models.ValidationError
	
	if err != nil {
		// Convert error to validation error for display
		validationErrors = []models.ValidationError{
			{
				Field:   "Config",
				Message: err.Error(),
				Index:   -1,
			},
		}
		cfg = &models.Config{Hosts: []models.Host{}}
	} else {
		// Validate config
		validationErrors = cfg.Validate()
	}
	
	// Apply saved theme or default to purple
	themeName := cfg.Theme
	if themeName == "" {
		themeName = "purple"
	}
	ApplyTheme(themeName)
	
	items := []list.Item{}
	for _, h := range cfg.Hosts {
		items = append(items, item{host: h, status: "⚪"})
	}

	// Custom delegate with original styling
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(3) // Three lines per item (title + description + tags)
	delegate.SetSpacing(0)
	
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(primaryColor).
		Bold(true).
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(primaryColor).
		Padding(0, 0, 0, 1)
	
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(mutedColor).
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(primaryColor).
		Padding(0, 0, 0, 1)
	
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(textColor).
		Padding(0, 0, 0, 2)
	
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.
		Foreground(dimColor).
		Padding(0, 0, 0, 2)

	l := list.New(items, delegate, 0, 0)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle()
	l.Styles.StatusBar = lipgloss.NewStyle()

	m := Model{
		list:         l,
		form:         NewFormModel(),
		state:        stateList,
		config:       cfg,
		pingStatus:   make(map[string]bool),
		pinging:      make(map[string]bool),
		pingTimes:    make(map[string]string),
		editingIndex: -1,
		configErrors: validationErrors,
	}
	
	// If there are validation errors, show error state
	if len(validationErrors) > 0 {
		m.state = stateConfigError
	}
	
	return m
}

func (m Model) Init() tea.Cmd {
	// Mark all hosts as pinging on startup
	for _, h := range m.config.Hosts {
		key := GetHostKey(h)
		m.pinging[key] = true
	}
	return StartPingAll(m.config.Hosts)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		
		if m.state == stateList {
			// Check if we're in search/filter mode - if so, only allow escape and let list handle other keys
			filterState := m.list.FilterState()
			isSearching := filterState == list.Filtering
			
			// Only process shortcuts when NOT in search mode
			if !isSearching {
				switch msg.String() {
				case "t":
					// Cycle through themes
					themeNames := GetThemeNames()
					currentThemeName := m.config.Theme
					if currentThemeName == "" {
						currentThemeName = "purple"
					}
					
					// Find current theme index and move to next
					currentIdx := 0
					for i, name := range themeNames {
						if name == currentThemeName {
							currentIdx = i
							break
						}
					}
					
					nextIdx := (currentIdx + 1) % len(themeNames)
					newTheme := themeNames[nextIdx]
					
					// Apply and save theme
					ApplyTheme(newTheme)
					m.config.Theme = newTheme
					config.SaveConfig(m.config)
					
					// Force refresh of list to apply new colors
					m.refreshList()
					return m, nil
				case "n":
					m.state = stateForm
					m.form = NewFormModel() // Reset form
					m.editingIndex = -1     // -1 means adding new
					return m, m.form.Init()
				case "p":
					// Ping all servers - mark all as pinging
					for _, h := range m.config.Hosts {
						key := GetHostKey(h)
						m.pinging[key] = true
					}
					m.refreshList()
					return m, StartPingAll(m.config.Hosts)
				case "enter":
					// Connect to selected host
					if selectedItem, ok := m.list.SelectedItem().(item); ok {
						// Return a command that will execute SSH after quitting
						return m, func() tea.Msg {
							return ConnectMsg{Host: selectedItem.host}
						}
					}
				case "left":
					// Move left in row-wise layout (decrement by 1 if on odd index)
					currentIdx := m.list.Index()
					if currentIdx%2 == 1 { // If on right column
						m.list.Select(currentIdx - 1)
					}
					return m, nil
				case "right":
					// Move right in row-wise layout (increment by 1 if on even index)
					currentIdx := m.list.Index()
					totalItems := len(m.list.Items())
					if currentIdx%2 == 0 && currentIdx+1 < totalItems { // If on left column
						m.list.Select(currentIdx + 1)
					}
					return m, nil
				case "e":
					// Edit selected host (only if not from SSH config)
					if selectedItem, ok := m.list.SelectedItem().(item); ok {
						if selectedItem.host.Source == "ssh-config" {
							// Cannot edit SSH config hosts
							return m, nil
						}
						m.state = stateForm
						m.form = NewFormModelWithHost(selectedItem.host)
						m.editingIndex = m.list.Index()
						return m, m.form.Init()
					}
				case "c":
					// Duplicate selected host
					if selectedItem, ok := m.list.SelectedItem().(item); ok {
						m.state = stateForm
						duplicatedHost := selectedItem.host
						// Append " (copy)" to the alias to avoid duplicates
						duplicatedHost.Alias = duplicatedHost.Alias + " (copy)"
						m.form = NewFormModelWithHost(duplicatedHost)
						m.editingIndex = -1 // -1 means adding new (not editing)
						return m, m.form.Init()
					}
				case "d", "delete":
					// Show delete confirmation (only if not from SSH config)
					if selectedItem, ok := m.list.SelectedItem().(item); ok {
						if selectedItem.host.Source == "ssh-config" {
							// Cannot delete SSH config hosts
							return m, nil
						}
						currentIdx := m.list.Index()
						if currentIdx >= 0 && currentIdx < len(m.config.Hosts) {
							m.deleteConfirmHost = &selectedItem.host
							m.deleteConfirmIdx = currentIdx
							m.state = stateConfirmDelete
						}
					}
					return m, nil
				}
			}
		} else if m.state == stateForm {
			if msg.String() == "esc" {
				m.state = stateList
				return m, nil
			}
		} else if m.state == stateConfirmDelete {
			switch msg.String() {
			case "y", "Y":
				// Confirm deletion
				if m.deleteConfirmIdx >= 0 && m.deleteConfirmIdx < len(m.config.Hosts) {
					m.config.Hosts = append(m.config.Hosts[:m.deleteConfirmIdx], m.config.Hosts[m.deleteConfirmIdx+1:]...)
					config.SaveConfig(m.config)
					m.refreshList()
					// Adjust selection if needed
					if m.deleteConfirmIdx >= len(m.config.Hosts) && len(m.config.Hosts) > 0 {
						m.list.Select(len(m.config.Hosts) - 1)
					}
				}
				m.deleteConfirmHost = nil
				m.state = stateList
				return m, nil
			case "n", "N", "esc":
				// Cancel deletion
				m.deleteConfirmHost = nil
				m.state = stateList
				return m, nil
			}
		} else if m.state == stateConfigError {
			switch msg.String() {
			case "e", "E":
				// Open config file for editing
				m.state = stateList
				return m, nil
			case "i", "I":
				// Ignore errors and continue
				m.configErrors = nil
				m.state = stateList
				return m, nil
			case "q", "Q":
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Fixed width box for 2-column layout
		const boxWidth = 80
		listWidth := boxWidth - 8 // Account for padding and borders
		listHeight := 20 // Height for scrollable list
		m.list.SetSize(listWidth, listHeight)

	case PingResultMsg:
		// Update ping status, time, and clear pinging state
		key := GetHostKey(msg.Host)
		m.pingStatus[key] = msg.Status
		m.pingTimes[key] = msg.PingTime
		m.pinging[key] = false
		m.refreshList()
		return m, nil

	case FormSubmittedMsg:
		if m.editingIndex >= 0 && m.editingIndex < len(m.config.Hosts) {
			// Editing existing host
			m.config.Hosts[m.editingIndex] = msg.Host
		} else {
			// Adding new host
			m.config.Hosts = append(m.config.Hosts, msg.Host)
		}
		config.SaveConfig(m.config)
		m.state = stateList
		m.editingIndex = -1
		m.refreshList()
		// Ping the host
		return m, PingHost(msg.Host)

	case ConnectMsg:
		// Store the host and quit the TUI
		m.selectedHost = &msg.Host
		return m, tea.Quit
	}

	if m.state == stateList {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.state == stateForm {
		m.form, cmd = m.form.Update(msg)
		cmds = append(cmds, cmd)
	}
	// No update needed for stateConfirmDelete

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	// Fixed width box for 2-column layout
	const boxWidth = 80
	const minHeight = 24
	
	// Check if terminal is too small
	if m.width < boxWidth+4 || m.height < minHeight {
		errorMsg := lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			Align(lipgloss.Center).
			Render("⚠ Terminal Too Small ⚠")
		
		instruction := lipgloss.NewStyle().
			Foreground(mutedColor).
			Align(lipgloss.Center).
			Render(fmt.Sprintf("Please resize your terminal to at least %dx%d", boxWidth+4, minHeight))
		
		currentSize := lipgloss.NewStyle().
			Foreground(dimColor).
			Align(lipgloss.Center).
			Render(fmt.Sprintf("Current: %dx%d", m.width, m.height))
		
		errorBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(errorColor).
			Padding(2, 4).
			Render(lipgloss.JoinVertical(lipgloss.Center, errorMsg, "", instruction, "", currentSize))
		
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, errorBox)
	}
	
	if m.state == stateForm {
		// Form view with centered box
		formView := m.form.View()
		boxed := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(2, 4).
			Align(lipgloss.Center).
			Render(formView)
		
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, boxed)
	}
	
	if m.state == stateConfirmDelete {
		// Confirmation dialog
		return m.renderDeleteConfirmation()
	}
	
	if m.state == stateConfigError {
		// Config error view
		return m.renderConfigError()
	}
	
	// ASCII art header
	asciiArt := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render(`╔═╗┌─┐┬ ┬  ╔╗ ┬ ┬┌┬┐┌┬┐┬ ┬
╚═╗└─┐├─┤  ╠╩╗│ │ ││ ││└┬┘
╚═╝└─┘┴ ┴  ╚═╝└─┘─┴┘─┴┘ ┴`)
	
	// Theme indicator
	theme := GetCurrentTheme()
	themeIndicator := lipgloss.NewStyle().
		Foreground(dimColor).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render(fmt.Sprintf("Theme: %s", theme.Name))
	
	separator := lipgloss.NewStyle().
		Foreground(dimColor).
		Width(boxWidth - 4).
		Align(lipgloss.Center).
		Render(strings.Repeat("─", boxWidth-4))
	
	header := lipgloss.JoinVertical(lipgloss.Left, asciiArt, themeIndicator, separator)
	
	// Footer with key bindings including ping command and theme switcher
	keyBindings := []string{
		keyStyle.Render("↵") + descStyle.Render(":connect "),
		keyStyle.Render("n") + descStyle.Render(":new "),
		keyStyle.Render("e") + descStyle.Render(":edit "),
		keyStyle.Render("c") + descStyle.Render(":copy "),
		keyStyle.Render("d") + descStyle.Render(":del "),
		keyStyle.Render("p") + descStyle.Render(":ping "),
		keyStyle.Render("t") + descStyle.Render(":theme "),
		keyStyle.Render("/") + descStyle.Render(":search "),
		keyStyle.Render("q") + descStyle.Render(":quit"),
	}
	footer := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(borderColor).
		Width(boxWidth - 4).
		Padding(0, 0).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, keyBindings...))
	
	// Render list in 2 columns
	listView := m.renderTwoColumnList()
	
	// Add search bar if filtering is active or has filter value
	var searchBar string
	filterState := m.list.FilterState()
	searchQuery := m.list.FilterValue()
	
	// Show search bar when filtering or when there's a filter value
	if filterState == list.Filtering || filterState == list.FilterApplied {
		if searchQuery == "" {
			searchQuery = "_" // Show cursor when empty
		}
		searchBar = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 2).
			Render(fmt.Sprintf("Search: %s", searchQuery))
		
		searchBar = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(primaryColor).
			Width(boxWidth - 4).
			Render(searchBar)
	}
	
	// Combine all elements
	var content string
	if searchBar != "" {
		content = lipgloss.JoinVertical(lipgloss.Left,
			header,
			searchBar,
			listView,
			footer,
		)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			listView,
			footer,
		)
	}
	
	// Wrap in a fixed-width box
	mainBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Width(boxWidth).
		Padding(0, 2).
		Render(content)
	
	// Center the fixed box on screen
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, mainBox)
}

func (m *Model) refreshList() {
	items := []list.Item{}
	for _, h := range m.config.Hosts {
		key := GetHostKey(h)
		status := "⚪" // Default - unknown
		if pingStatus, exists := m.pingStatus[key]; exists {
			status = GetHostStatus(pingStatus)
		}
		isPinging := m.pinging[key]
		pingTime := m.pingTimes[key]
		items = append(items, item{host: h, status: status, pinging: isPinging, pingTime: pingTime})
	}
	m.list.SetItems(items)
}

func (m *Model) renderTwoColumnList() string {
	items := m.list.VisibleItems()
	if len(items) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true).
			Padding(2, 0).
			Render("No hosts configured. Press 'n' to add a new host.")
		return emptyMsg
	}

	const columnWidth = 34 // Each column width
	const columnGap = 2    // Gap between columns
	const itemHeight = 3   // Title + Description + Tags
	const listHeight = 3  // Number of items visible per column
	
	var leftColumn, rightColumn []string
	
	// Get the current cursor position
	cursor := m.list.Index()
	startIdx := 0
	
	// Calculate scroll offset to keep cursor visible
	itemsPerScreen := listHeight * 2 // Two columns
	if cursor >= itemsPerScreen {
		startIdx = ((cursor / itemsPerScreen) * itemsPerScreen)
	}
	
	// Split items into two columns with scrolling
	endIdx := min(startIdx+itemsPerScreen, len(items))
	
	// Helper function to render an item or empty placeholder
	renderItemAtIndex := func(i int) string {
		// Check if we have an actual item at this position
		if i >= len(items) {
			// Return empty placeholder
			return lipgloss.NewStyle().
				Width(columnWidth).
				Height(itemHeight).
				Render("")
		}
		
		if itm, ok := items[i].(item); ok {
			isSelected := i == cursor
			
			// Format the item with status
			var statusText string
			if itm.pinging {
				statusText = statusPingingStyle.Render("●")
			} else {
				switch itm.status {
				case "🟢":
					statusText = statusOnlineStyle.Render("●")
				case "🔴":
					statusText = statusOfflineStyle.Render("●")
				default:
					statusText = statusUnknownStyle.Render("○")
				}
			}
			
			// Title line - build with alias and ping time
			alias := itm.host.Alias
			pingTimeStr := ""
			if itm.pingTime != "" {
				pingTimeStr = lipgloss.NewStyle().Foreground(dimColor).Render(fmt.Sprintf(" (%s)", itm.pingTime))
			}
			
			// Truncate alias to fit with ping time
			maxAliasLen := 15
			if len(alias) > maxAliasLen {
				alias = alias[:maxAliasLen-3] + "..."
			}
			
			port := itm.host.Port
			if port == "" {
				port = "22"
			}
			
			// Description line - truncate to fit
			hostInfo := fmt.Sprintf("%s@%s:%s", itm.host.User, itm.host.Hostname, port)
			if len(hostInfo) > 28 {
				hostInfo = hostInfo[:25] + "..."
			}
			
			// Source line - render with colors
			sourceLine := renderSource(itm.host.Source, columnWidth-2, isSelected)
			
			var titleLine, descLine string
			if isSelected {
				// Selected item with border - need to account for border width
				titleLine = lipgloss.NewStyle().
					Foreground(primaryColor).
					Bold(true).
					BorderLeft(true).
					BorderStyle(lipgloss.NormalBorder()).
					BorderForeground(primaryColor).
					Padding(0, 0, 0, 1).
					Width(columnWidth - 2). // Subtract border + padding
					Render(fmt.Sprintf("%s %s%s", statusText, alias, pingTimeStr))
				
				descLine = lipgloss.NewStyle().
					Foreground(mutedColor).
					BorderLeft(true).
					BorderStyle(lipgloss.NormalBorder()).
					BorderForeground(primaryColor).
					Padding(0, 0, 0, 1).
					Width(columnWidth - 2). // Subtract border + padding
					Render(hostInfo)
				
				sourceLine = lipgloss.NewStyle().
					BorderLeft(true).
					BorderStyle(lipgloss.NormalBorder()).
					BorderForeground(primaryColor).
					Padding(0, 0, 0, 1).
					Width(columnWidth - 2).
					Render(sourceLine)
			} else {
				// Normal item without border - use full width with padding
				titleLine = lipgloss.NewStyle().
					Foreground(textColor).
					Padding(0, 0, 0, 2).
					Width(columnWidth - 2). // Subtract padding
					Render(fmt.Sprintf("%s %s%s", statusText, alias, pingTimeStr))
				
				descLine = lipgloss.NewStyle().
					Foreground(dimColor).
					Padding(0, 0, 0, 2).
					Width(columnWidth - 2). // Subtract padding
					Render(hostInfo)
				
				sourceLine = lipgloss.NewStyle().
					Padding(0, 0, 0, 2).
					Width(columnWidth - 2).
					Render(sourceLine)
			}
			
			// Wrap in a fixed-width container to prevent shifting
			titleLine = lipgloss.NewStyle().Width(columnWidth).Render(titleLine)
			descLine = lipgloss.NewStyle().Width(columnWidth).Render(descLine)
			sourceLine = lipgloss.NewStyle().Width(columnWidth).Render(sourceLine)
			
			return lipgloss.JoinVertical(lipgloss.Left, titleLine, descLine, sourceLine)
		}
		
		return lipgloss.NewStyle().Width(columnWidth).Height(itemHeight).Render("")
	}
	
	// Render items row-wise: fill left column first, then right column for each row
	for row := 0; row < listHeight; row++ {
		leftIdx := startIdx + (row * 2)     // 0, 2, 4, 6...
		rightIdx := startIdx + (row * 2) + 1 // 1, 3, 5, 7...
		
		leftColumn = append(leftColumn, renderItemAtIndex(leftIdx))
		rightColumn = append(rightColumn, renderItemAtIndex(rightIdx))
	}
	
	// Create gap between columns
	// gap := lipgloss.NewStyle().Width(columnGap).Render("")
	
	// Join columns side by side with gap
	var rows []string
	for i := 0; i < len(leftColumn); i++ {
		row := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn[i], rightColumn[i])
		row_space := lipgloss.JoinHorizontal(lipgloss.Top, "")
		rows = append(rows, row)
		rows = append(rows, row_space)
	}
	
	listContent := lipgloss.JoinVertical(lipgloss.Left, rows...)
	
	// Add scroll indicator if needed
	if len(items) > itemsPerScreen {
		scrollInfo := lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true).
			Render(fmt.Sprintf("  %d-%d of %d (↑↓ scroll)", startIdx+1, min(endIdx, len(items)), len(items)))
		listContent = lipgloss.JoinVertical(lipgloss.Left, listContent, scrollInfo)
	}
	
	return listContent
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetSelectedHost returns the host selected for SSH connection
func (m Model) GetSelectedHost() *models.Host {
	return m.selectedHost
}

// renderSource renders the source label with muted color
func renderSource(source string, maxWidth int, isSelected bool) string {
	if source == "" {
		source = "sshbuddy"
	}
	
	// Map source names to display names
	displayName := source
	switch source {
	case "manual":
		displayName = "sshbuddy"
	case "ssh-config":
		displayName = "config"
	case "termix":
		displayName = "termix"
	}
	
	// Use dimColor for source (same as ping time)
	sourceStyle := lipgloss.NewStyle().
		Foreground(dimColor).
		Bold(false)
	
	return sourceStyle.Render("source: " + displayName)
}

// renderDeleteConfirmation renders the delete confirmation dialog
func (m Model) renderDeleteConfirmation() string {
	if m.deleteConfirmHost == nil {
		return ""
	}
	
	host := m.deleteConfirmHost
	
	// Warning icon and title
	warningIcon := lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true).
		Render("⚠ Delete Host?")
	
	// Host details
	hostDetails := lipgloss.NewStyle().
		Foreground(textColor).
		MarginTop(1).
		MarginBottom(1).
		Render(fmt.Sprintf("Alias: %s\nHost: %s@%s", host.Alias, host.User, host.Hostname))
	
	// Confirmation message
	confirmMsg := lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true).
		Render("This action cannot be undone.")
	
	// Action buttons
	yesButton := lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true).
		Render("Y")
	
	noButton := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		Render("N")
	
	actions := lipgloss.NewStyle().
		MarginTop(1).
		Render(yesButton + descStyle.Render(" Yes  ") + noButton + descStyle.Render(" No  ") + 
			keyStyle.Render("esc") + descStyle.Render(" Cancel"))
	
	// Combine all elements
	content := lipgloss.JoinVertical(lipgloss.Left,
		warningIcon,
		hostDetails,
		confirmMsg,
		actions,
	)
	
	// Wrap in a dialog box
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(errorColor).
		Padding(2, 4).
		Width(50).
		Render(content)
	
	// Center on screen
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

// renderConfigError renders the config validation error screen
func (m Model) renderConfigError() string {
	// Error icon and title
	errorIcon := lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true).
		Render("⚠ Configuration Errors")
	
	// Error count - determine source of error
	errorSource := "configuration"
	if len(m.configErrors) > 0 {
		// Check if error is from termix by looking at the error message
		firstError := m.configErrors[0].Error()
		if strings.Contains(firstError, "termix") {
			errorSource = "Termix"
		} else if strings.Contains(firstError, "Config:") {
			errorSource = "configuration"
		} else {
			errorSource = "sshbuddy.json"
		}
	}
	
	errorCount := lipgloss.NewStyle().
		Foreground(mutedColor).
		MarginTop(1).
		Render(fmt.Sprintf("Found %d error(s) in %s:", len(m.configErrors), errorSource))
	
	// List errors (limit to first 10)
	var errorLines []string
	maxErrors := 10
	for i, err := range m.configErrors {
		if i >= maxErrors {
			remaining := len(m.configErrors) - maxErrors
			errorLines = append(errorLines, lipgloss.NewStyle().
				Foreground(dimColor).
				Italic(true).
				Render(fmt.Sprintf("... and %d more error(s)", remaining)))
			break
		}
		
		errorLine := lipgloss.NewStyle().
			Foreground(errorColor).
			Render(fmt.Sprintf("• %s", err.Error()))
		errorLines = append(errorLines, errorLine)
	}
	
	errorList := lipgloss.NewStyle().
		MarginTop(1).
		MarginBottom(1).
		Render(strings.Join(errorLines, "\n"))
	
	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render("Please fix the errors in your config file.")
	
	// Action buttons
	ignoreButton := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		Render("I")
	
	quitButton := lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true).
		Render("Q")
	
	actions := lipgloss.NewStyle().
		MarginTop(1).
		Render(ignoreButton + descStyle.Render(" Ignore & Continue  ") + 
			quitButton + descStyle.Render(" Quit"))
	
	// Combine all elements
	content := lipgloss.JoinVertical(lipgloss.Left,
		errorIcon,
		errorCount,
		errorList,
		instructions,
		actions,
	)
	
	// Wrap in a dialog box
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(errorColor).
		Padding(2, 4).
		Width(70).
		Render(content)
	
	// Center on screen
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

