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
	
	// Load config from file
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Initialize with defaults
		config = models.Config{
			Hosts: []models.Host{},
			Sources: models.SourcesConfig{
				SSHBuddyEnabled:  true,
				SSHConfigEnabled: true,
				TermixEnabled:    false,
			},
			Termix: models.TermixConfig{
				Enabled: false,
			},
			SSH: models.SSHConfig{
				Enabled: true,
			},
		}
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
	
	// Mark manual hosts
	for i := range config.Hosts {
		if config.Hosts[i].Source == "" {
			config.Hosts[i].Source = "manual"
		}
	}
	
	// Track all aliases to avoid duplicates
	existingAliases := make(map[string]bool)
	
	// Only add manual hosts if SSHBuddy source is enabled
	if config.Sources.SSHBuddyEnabled {
		for _, host := range config.Hosts {
			existingAliases[host.Alias] = true
		}
	} else {
		// Clear manual hosts if disabled
		config.Hosts = []models.Host{}
	}
	
	// Load hosts from SSH config if enabled
	if config.Sources.SSHConfigEnabled && config.SSH.Enabled {
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
	if config.Sources.TermixEnabled && config.Termix.Enabled && config.Termix.BaseURL != "" {
		logError("Termix config loaded", fmt.Errorf("baseUrl=%s", config.Termix.BaseURL))
		
		client := termix.NewClient(config.Termix.BaseURL, config.Termix.JWT, config.Termix.JWTExpiry)
		
		// Try to fetch hosts without credentials first (using cached token)
		termixHosts, termixFetchErr := client.FetchHosts("", "")
		
		// If auth is required, return a special error that the TUI can handle
		if termixFetchErr != nil {
			if _, isAuthError := termixFetchErr.(*termix.AuthError); isAuthError {
				// Return auth error to trigger credential prompt in TUI
				return nil, termixFetchErr
			}
			
			// Log other errors
			logError("Termix FetchHosts failed", termixFetchErr)
			
			// Return error to show in UI with config file hint
			configPath, _ := GetDataPath()
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
		
		// Save the JWT token and expiry if they were updated
		if client.GetJWT() != config.Termix.JWT || client.GetJWTExpiry() != config.Termix.JWTExpiry {
			config.Termix.JWT = client.GetJWT()
			config.Termix.JWTExpiry = client.GetJWTExpiry()
			SaveConfig(&config)
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
	saveConfig := &models.Config{
		Theme:   config.Theme,
		Sources: config.Sources,
		Termix:  config.Termix,
		SSH:     config.SSH,
		Hosts:   []models.Host{},
	}
	
	for _, host := range config.Hosts {
		if host.Source != "ssh-config" && host.Source != "termix" {
			saveConfig.Hosts = append(saveConfig.Hosts, host)
		}
	}

	data, err := json.MarshalIndent(saveConfig, "", "  ")
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



// LoadConfigRaw loads the config file without fetching external sources (SSH config, Termix)
func LoadConfigRaw() (*models.Config, error) {
	path, err := GetDataPath()
	if err != nil {
		return nil, err
	}

	var config models.Config
	
	// Load config from file
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Initialize with defaults
		config = models.Config{
			Hosts: []models.Host{},
			Sources: models.SourcesConfig{
				SSHBuddyEnabled:  true,
				SSHConfigEnabled: true,
				TermixEnabled:    false,
			},
			Termix: models.TermixConfig{
				Enabled: false,
			},
			SSH: models.SSHConfig{
				Enabled: true,
			},
		}
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, &config); err != nil {
			return nil, err
		}
	}
	
	return &config, nil
}

// AuthenticateTermix authenticates with Termix using provided credentials and updates the config
func AuthenticateTermix(username, password string) error {
	// Load config without fetching Termix hosts to avoid circular dependency
	config, err := LoadConfigRaw()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	if !config.Termix.Enabled || config.Termix.BaseURL == "" {
		return fmt.Errorf("termix is not enabled or baseUrl is not configured")
	}
	
	client := termix.NewClient(config.Termix.BaseURL, "", 0)
	jwt, expiry, err := client.Authenticate(username, password)
	if err != nil {
		return err
	}
	
	// Update config with new token and expiry
	config.Termix.JWT = jwt
	config.Termix.JWTExpiry = expiry
	
	return SaveConfig(config)
}
