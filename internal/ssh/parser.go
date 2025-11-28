package ssh

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sshbuddy/pkg/models"
)

// SSHConfigHost represents a host entry from SSH config
type SSHConfigHost struct {
	Host             string
	HostName         string
	User             string
	Port             string
	IdentityFile     string
	ProxyJump        string
	ForwardAgent     string
	LocalForward     string
	RemoteForward    string
	DynamicForward   string
	ServerAliveInterval string
}

// ParseSSHConfig reads and parses the SSH config file
func ParseSSHConfig() ([]SSHConfigHost, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".ssh", "config")
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return []SSHConfigHost{}, nil
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var hosts []SSHConfigHost
	var currentHost *SSHConfigHost

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split into key and value
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.ToLower(parts[0])
		value := strings.Join(parts[1:], " ")

		switch key {
		case "host":
			// Save previous host if exists
			if currentHost != nil && currentHost.Host != "*" {
				hosts = append(hosts, *currentHost)
			}
			// Start new host
			currentHost = &SSHConfigHost{
				Host: value,
			}
		case "hostname":
			if currentHost != nil {
				currentHost.HostName = value
			}
		case "user":
			if currentHost != nil {
				currentHost.User = value
			}
		case "port":
			if currentHost != nil {
				currentHost.Port = value
			}
		case "identityfile":
			if currentHost != nil {
				// Expand ~ to home directory
				if strings.HasPrefix(value, "~/") {
					value = filepath.Join(homeDir, value[2:])
				}
				currentHost.IdentityFile = value
			}
		case "proxyjump":
			if currentHost != nil {
				currentHost.ProxyJump = value
			}
		case "forwardagent":
			if currentHost != nil {
				currentHost.ForwardAgent = value
			}
		case "localforward":
			if currentHost != nil {
				currentHost.LocalForward = value
			}
		case "remoteforward":
			if currentHost != nil {
				currentHost.RemoteForward = value
			}
		case "dynamicforward":
			if currentHost != nil {
				currentHost.DynamicForward = value
			}
		case "serveraliveinterval":
			if currentHost != nil {
				currentHost.ServerAliveInterval = value
			}
		}
	}

	// Add the last host
	if currentHost != nil && currentHost.Host != "*" {
		hosts = append(hosts, *currentHost)
	}

	return hosts, scanner.Err()
}

// ConvertToHost converts an SSHConfigHost to a models.Host
func ConvertToHost(sshHost SSHConfigHost) models.Host {
	hostname := sshHost.HostName
	if hostname == "" {
		hostname = sshHost.Host
	}

	port := sshHost.Port
	if port == "" {
		port = "22"
	}

	user := sshHost.User
	if user == "" {
		// Try to get current user
		if currentUser := os.Getenv("USER"); currentUser != "" {
			user = currentUser
		} else {
			user = "root"
		}
	}

	// Build tags based on SSH config properties
	var tags []string
	tags = append(tags, "ssh-config")
	
	if sshHost.IdentityFile != "" {
		tags = append(tags, "key-auth")
	}
	if sshHost.ProxyJump != "" {
		tags = append(tags, "proxy")
	}
	if sshHost.LocalForward != "" || sshHost.RemoteForward != "" || sshHost.DynamicForward != "" {
		tags = append(tags, "forwarding")
	}

	return models.Host{
		Alias:        sshHost.Host,
		Hostname:     hostname,
		User:         user,
		Port:         port,
		Tags:         tags,
		IdentityFile: sshHost.IdentityFile,
		ProxyJump:    sshHost.ProxyJump,
	}
}

// LoadHostsFromSSHConfig loads all hosts from SSH config
func LoadHostsFromSSHConfig() ([]models.Host, error) {
	sshHosts, err := ParseSSHConfig()
	if err != nil {
		return nil, err
	}

	var hosts []models.Host
	for _, sshHost := range sshHosts {
		hosts = append(hosts, ConvertToHost(sshHost))
	}

	return hosts, nil
}
