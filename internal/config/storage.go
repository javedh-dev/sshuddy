package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"sshbuddy/internal/ssh"
	"sshbuddy/internal/termix"
	"sshbuddy/pkg/models"
)

// SourcesConfig holds the enabled/disabled state for each source
type SourcesConfig struct {
	SSHBuddyEnabled  bool `json:"sshbuddyEnabled"`
	SSHConfigEnabled bool `json:"sshConfigEnabled"`
	TermixEnabled    bool `json:"termixEnabled"`
}

// SSHConfig holds SSH config source configuration
type SSHConfig struct {
	Enabled    bool   `json:"enabled"`
	ConfigPath string `json:"configPath,omitempty"` // Custom path, empty means default ~/.ssh/config
}

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
		logError("GetDataPath failed", err)
		return nil, err
	}

	var config models.Config
	
	// Load manual hosts from config file
	if _, err := os.Stat(path); os.IsNotExist(err) {
		config = models.Config{Hosts: []models.Host{}}
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			logError("ReadFile failed", err)
			return nil, err
		}

		if err := json.Unmarshal(data, &config); err != nil {
			logError("Unmarshal config failed", err)
			return nil, err
		}
	}
	
	// Load sources configuration
	sourcesConfig, sourcesErr := LoadSourcesConfig()
	if sourcesErr != nil {
		// Use defaults if can't load
		sourcesConfig = &SourcesConfig{
			SSHBuddyEnabled:  true,
			SSHConfigEnabled: true,
			TermixEnabled:    false,
		}
	}
	
	// Mark manual hosts
	for i := range config.Hosts {
		if config.Hosts[i].Source == "" {
			config.Hosts[i].Source = "manual"
		}
	}
	
	// Track all aliases to avoid duplicates
	existingAliases := make(map[string]bool)
	
	// Only add manual hosts if SSHBuddy source is enabled
	if sourcesConfig.SSHBuddyEnabled {
		for _, host := range config.Hosts {
			existingAliases[host.Alias] = true
		}
	} else {
		// Clear manual hosts if disabled
		config.Hosts = []models.Host{}
	}
	
	// Load hosts from SSH config if enabled
	if sourcesConfig.SSHConfigEnabled {
		sshHosts, err := ssh.LoadHostsFromSSHConfig()
		if err == nil {
			// Mark SSH config hosts
			for i := range sshHosts {
				sshHosts[i].Source = "ssh-config"
			}
			
			// Add SSH config hosts that don't conflict
			for _, sshHost := range sshHosts {
				if !existingAliases[sshHost.Alias] {
					config.Hosts = append(config.Hosts, sshHost)
					existingAliases[sshHost.Alias] = true
				}
			}
		}
	}
	
	// Load hosts from Termix API if enabled
	termixConfig, termixErr := LoadTermixConfig()
	if sourcesConfig.TermixEnabled && termixErr == nil && termixConfig.Enabled && termixConfig.BaseURL != "" {
		logError("Termix config loaded", fmt.Errorf("baseUrl=%s, username=%s", termixConfig.BaseURL, termixConfig.Username))
		
		client := termix.NewClient(termixConfig.BaseURL, termixConfig.Username, termixConfig.Password, termixConfig.JWT)
		termixHosts, termixFetchErr := client.FetchHosts()
		if termixFetchErr != nil {
			// Log the full error
			logError("Termix FetchHosts failed", termixFetchErr)
			
			// Return error to show in UI with config file hint
			configPath, _ := GetTermixConfigPath()
			fullError := fmt.Errorf("%w\n\nCheck your Termix configuration at: %s", termixFetchErr, configPath)
			logError("Returning error to UI", fullError)
			return nil, fullError
		}
		
		logError("Termix hosts fetched successfully", fmt.Errorf("count=%d", len(termixHosts)))
		
		// Add Termix hosts that don't conflict
		for _, termixHost := range termixHosts {
			if !existingAliases[termixHost.Alias] {
				config.Hosts = append(config.Hosts, termixHost)
				existingAliases[termixHost.Alias] = true
			}
		}
		
		// Save the JWT token for future use if it was updated
		if client.GetJWT() != termixConfig.JWT {
			termixConfig.JWT = client.GetJWT()
			SaveTermixConfig(termixConfig)
		}
	}

	return &config, nil
}

func SaveConfig(config *models.Config) error {
	path, err := GetDataPath()
	if err != nil {
		return err
	}

	// Only save manual hosts (not SSH config or termix hosts)
	manualConfig := &models.Config{
		Theme: config.Theme,
		Hosts: []models.Host{},
	}
	
	for _, host := range config.Hosts {
		if host.Source != "ssh-config" && host.Source != "termix" {
			manualConfig.Hosts = append(manualConfig.Hosts, host)
		}
	}

	data, err := json.MarshalIndent(manualConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetTermixConfigPath returns the path to the Termix config file
func GetTermixConfigPath() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	
	sshbuddyDir := filepath.Join(configDir, "sshbuddy")
	if err := os.MkdirAll(sshbuddyDir, 0755); err != nil {
		return "", err
	}
	
	return filepath.Join(sshbuddyDir, "termix.json"), nil
}

// LoadTermixConfig loads the Termix API configuration
func LoadTermixConfig() (*termix.Config, error) {
	path, err := GetTermixConfigPath()
	if err != nil {
		return nil, err
	}

	// Return default disabled config if file doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &termix.Config{Enabled: false}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config termix.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveTermixConfig saves the Termix API configuration
func SaveTermixConfig(config *termix.Config) error {
	path, err := GetTermixConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// logError logs errors to a debug file for troubleshooting
func logError(context string, err error) {
	logPath := "/tmp/sshbuddy-debug.log"
	
	logFile, fileErr := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		return // Silently fail if we can't log
	}
	defer logFile.Close()
	
	timestamp := fmt.Sprintf("[%s]", os.Getenv("USER"))
	logLine := fmt.Sprintf("%s %s: %v\n", timestamp, context, err)
	logFile.WriteString(logLine)
}

// GetSourcesConfigPath returns the path to the sources config file
func GetSourcesConfigPath() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	
	sshbuddyDir := filepath.Join(configDir, "sshbuddy")
	if err := os.MkdirAll(sshbuddyDir, 0755); err != nil {
		return "", err
	}
	
	return filepath.Join(sshbuddyDir, "sources.json"), nil
}

// LoadSourcesConfig loads the sources configuration
func LoadSourcesConfig() (*SourcesConfig, error) {
	path, err := GetSourcesConfigPath()
	if err != nil {
		return nil, err
	}

	// Return default config if file doesn't exist (all enabled by default)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &SourcesConfig{
			SSHBuddyEnabled:  true,
			SSHConfigEnabled: true,
			TermixEnabled:    false,
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config SourcesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveSourcesConfig saves the sources configuration
func SaveSourcesConfig(config *SourcesConfig) error {
	path, err := GetSourcesConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetSSHConfigPath returns the path to the SSH config file
func GetSSHConfigPath() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	
	sshbuddyDir := filepath.Join(configDir, "sshbuddy")
	if err := os.MkdirAll(sshbuddyDir, 0755); err != nil {
		return "", err
	}
	
	return filepath.Join(sshbuddyDir, "sshconfig.json"), nil
}

// LoadSSHConfig loads the SSH config source configuration
func LoadSSHConfig() (*SSHConfig, error) {
	path, err := GetSSHConfigPath()
	if err != nil {
		return nil, err
	}

	// Return default config if file doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &SSHConfig{Enabled: true, ConfigPath: ""}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config SSHConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveSSHConfig saves the SSH config source configuration
func SaveSSHConfig(config *SSHConfig) error {
	path, err := GetSSHConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
