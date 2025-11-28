package ui

import (
	"fmt"
	"os"
	"os/exec"
	"sshbuddy/model"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ConnectToHost initiates an SSH connection and exits the TUI
func ConnectToHost(host model.Host) tea.Cmd {
	return func() tea.Msg {
		return ConnectMsg{host}
	}
}

type ConnectMsg struct {
	Host model.Host
}

// ExecuteSSH executes SSH connection in the foreground
func ExecuteSSH(host model.Host) error {
	port := host.Port
	if port == "" {
		port = "22"
	}

	var args []string
	args = append(args, "-p", port)
	args = append(args, fmt.Sprintf("%s@%s", host.User, host.Hostname))

	cmd := exec.Command("ssh", args...)
	
	// Connect to current terminal for interactive SSH session
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run SSH in foreground and wait for it to complete
	return cmd.Run()
}

// PingHost checks if a host is reachable using a simple ping
func PingHost(host model.Host) tea.Cmd {
	return func() tea.Msg {
		// Use ping with 1 count and 1 second timeout
		cmd := exec.Command("ping", "-c", "1", "-W", "1", host.Hostname)
		output, err := cmd.CombinedOutput()
		
		// Parse ping time from output
		pingTime := ""
		if err == nil {
			// Extract time from ping output (e.g., "time=12.3 ms")
			outputStr := string(output)
			
			// Try to find "time=" pattern
			if idx := strings.Index(outputStr, "time="); idx != -1 {
				timeStr := outputStr[idx+5:]
				// Find the end of the time value (space or newline)
				endIdx := strings.IndexAny(timeStr, " \n\r")
				if endIdx != -1 {
					timeValue := strings.TrimSpace(timeStr[:endIdx])
					pingTime = timeValue
					// Add "ms" if not already present
					if !strings.HasSuffix(pingTime, "ms") {
						pingTime = pingTime + "ms"
					}
				}
			}
		}
		
		return PingResultMsg{
			Host:     host,
			Status:   err == nil,
			PingTime: pingTime,
		}
	}
}

type PingResultMsg struct {
	Host     model.Host
	Status   bool   // true if reachable
	PingTime string // ping time in ms
}

// StartPingAll starts background ping for all hosts
func StartPingAll(hosts []model.Host) tea.Cmd {
	var cmds []tea.Cmd
	for _, host := range hosts {
		cmds = append(cmds, PingHost(host))
	}
	return tea.Batch(cmds...)
}

// GetHostStatus returns a visual indicator for host status
func GetHostStatus(status bool) string {
	if status {
		return "🟢" // Green dot - reachable
	}
	return "🔴" // Red dot - unreachable
}

// GetHostKey creates a unique key for a host (for tracking ping status)
func GetHostKey(host model.Host) string {
	return strings.ToLower(host.Hostname + ":" + host.User)
}
