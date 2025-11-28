package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"sshbuddy/internal/ssh"
	"sshbuddy/pkg/models"
)

func GetDataPath() (string, error) {
	// Use XDG_CONFIG_HOME if set, otherwise default to ~/.config
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	
	// Create sshbuddy config directory
	sshbuddyDir := filepath.Join(configDir, "sshbuddy")
	if err := os.MkdirAll(sshbuddyDir, 0755); err != nil {
		return "", err
	}
	
	return filepath.Join(sshbuddyDir, "config.json"), nil
}

func LoadConfig() (*models.Config, error) {
	path, err := GetDataPath()
	if err != nil {
		return nil, err
	}

	var config models.Config
	
	// Load manual hosts from config file
	if _, err := os.Stat(path); os.IsNotExist(err) {
		config = models.Config{Hosts: []models.Host{}}
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, &config); err != nil {
			return nil, err
		}
	}
	
	// Mark manual hosts
	for i := range config.Hosts {
		if config.Hosts[i].Source == "" {
			config.Hosts[i].Source = "manual"
		}
	}
	
	// Load hosts from SSH config
	sshHosts, err := ssh.LoadHostsFromSSHConfig()
	if err == nil {
		// Mark SSH config hosts
		for i := range sshHosts {
			sshHosts[i].Source = "ssh-config"
		}
		
		// Merge with manual hosts (manual hosts take precedence)
		manualAliases := make(map[string]bool)
		for _, host := range config.Hosts {
			manualAliases[host.Alias] = true
		}
		
		// Add SSH config hosts that don't conflict with manual hosts
		for _, sshHost := range sshHosts {
			if !manualAliases[sshHost.Alias] {
				config.Hosts = append(config.Hosts, sshHost)
			}
		}
	}

	return &config, nil
}

func SaveConfig(config *models.Config) error {
	path, err := GetDataPath()
	if err != nil {
		return err
	}

	// Only save manual hosts (not SSH config hosts)
	manualConfig := &models.Config{
		Theme: config.Theme,
		Hosts: []models.Host{},
	}
	
	for _, host := range config.Hosts {
		if host.Source != "ssh-config" {
			manualConfig.Hosts = append(manualConfig.Hosts, host)
		}
	}

	data, err := json.MarshalIndent(manualConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
