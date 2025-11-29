package termix

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sshbuddy/pkg/models"
	"time"
)

// TermixHost represents the API response structure from Termix
type TermixHost struct {
	ID                         int      `json:"id"`
	UserID                     string   `json:"userId"`
	Name                       string   `json:"name"`
	IP                         string   `json:"ip"`
	Port                       int      `json:"port"`
	Username                   string   `json:"username"`
	Folder                     string   `json:"folder"`
	Tags                       []string `json:"tags"`
	Pin                        bool     `json:"pin"`
	AuthType                   string   `json:"authType"`
	ForceKeyboardInteractive   bool     `json:"forceKeyboardInteractive"`
	Password                   *string  `json:"password"`
	Key                        *string  `json:"key"`
	KeyPassword                *string  `json:"key_password"`
	KeyType                    string   `json:"keyType"`
	AutostartPassword          *string  `json:"autostartPassword"`
	AutostartKey               *string  `json:"autostartKey"`
	AutostartKeyPassword       *string  `json:"autostartKeyPassword"`
	CredentialID               *int     `json:"credentialId"`
	OverrideCredentialUsername *string  `json:"overrideCredentialUsername"`
	EnableTerminal             bool     `json:"enableTerminal"`
	EnableTunnel               bool     `json:"enableTunnel"`
	TunnelConnections          []any    `json:"tunnelConnections"`
	JumpHosts                  []any    `json:"jumpHosts"`
	EnableFileManager          bool     `json:"enableFileManager"`
	DefaultPath                string   `json:"defaultPath"`
	QuickActions               []any    `json:"quickActions"`
	CreatedAt                  string   `json:"createdAt"`
	UpdatedAt                  string   `json:"updatedAt"`
}

// Config holds Termix API configuration
type Config struct {
	Enabled  bool   `json:"enabled"`
	BaseURL  string `json:"baseUrl"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	JWT      string `json:"jwt,omitempty"` // Cached JWT token
}

// Client handles communication with Termix API
type Client struct {
	baseURL  string
	username string
	password string
	jwt      string
	client   *http.Client
}

// NewClient creates a new Termix API client
func NewClient(baseURL, username, password, jwt string) *Client {
	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		jwt:      jwt,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Authenticate logs in to Termix and returns the JWT token
func (c *Client) Authenticate() (string, error) {
	loginURL := c.baseURL + "/users/login"
	logDebug("Termix Authenticate", fmt.Sprintf("URL: %s, Username: %s", loginURL, c.username))
	
	loginData := map[string]string{
		"username": c.username,
		"password": c.password,
	}
	
	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return "", fmt.Errorf("termix: failed to marshal login data: %w", err)
	}

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("termix: failed to create auth request (check baseUrl): %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		logDebug("Termix Auth Request Failed", err.Error())
		return "", fmt.Errorf("termix: connection failed (check baseUrl and network): %w", err)
	}
	defer resp.Body.Close()

	logDebug("Termix Auth Response", fmt.Sprintf("Status: %d", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyPreview := string(body)
		logDebug("Termix Auth Failed Body", bodyPreview)
		if len(bodyPreview) > 200 {
			bodyPreview = bodyPreview[:200] + "..."
		}
		return "", fmt.Errorf("termix: authentication failed (status %d, check username/password): %s", resp.StatusCode, bodyPreview)
	}

	// Extract JWT from Set-Cookie header
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "jwt" {
			c.jwt = cookie.Value
			return cookie.Value, nil
		}
	}

	return "", fmt.Errorf("termix: JWT cookie not found (server may not be Termix API)")
}

// FetchHosts retrieves hosts from the Termix API
func (c *Client) FetchHosts() ([]models.Host, error) {
	logDebug("Termix FetchHosts", fmt.Sprintf("Starting, JWT present: %v", c.jwt != ""))
	
	// Authenticate if we don't have a JWT
	if c.jwt == "" {
		jwt, err := c.Authenticate()
		if err != nil {
			logDebug("Termix FetchHosts Auth Failed", err.Error())
			return nil, err
		}
		c.jwt = jwt
		logDebug("Termix FetchHosts Auth Success", "JWT obtained")
	}

	hostsURL := c.baseURL + "/ssh/db/host"
	logDebug("Termix FetchHosts URL", hostsURL)
	req, err := http.NewRequest("GET", hostsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("termix: failed to create request: %w", err)
	}

	// Add JWT as cookie along with i18nextLng
	req.AddCookie(&http.Cookie{
		Name:  "jwt",
		Value: c.jwt,
	})
	req.AddCookie(&http.Cookie{
		Name:  "i18nextLng",
		Value: "en",
	})

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("termix: failed to fetch hosts (check baseUrl): %w", err)
	}
	defer resp.Body.Close()

	// If unauthorized, try to re-authenticate
	if resp.StatusCode == http.StatusUnauthorized {
		jwt, err := c.Authenticate()
		if err != nil {
			return nil, err
		}
		c.jwt = jwt
		
		// Retry the request with new JWT
		req.Header.Del("Cookie")
		req.AddCookie(&http.Cookie{
			Name:  "jwt",
			Value: c.jwt,
		})
		req.AddCookie(&http.Cookie{
			Name:  "i18nextLng",
			Value: "en",
		})
		
		resp, err = c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("termix: failed to fetch hosts after re-auth: %w", err)
		}
		defer resp.Body.Close()
	}

	logDebug("Termix FetchHosts Response", fmt.Sprintf("Status: %d", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyPreview := string(body)
		logDebug("Termix FetchHosts Error Body", bodyPreview)
		if len(bodyPreview) > 200 {
			bodyPreview = bodyPreview[:200] + "..."
		}
		return nil, fmt.Errorf("termix: API returned status %d: %s", resp.StatusCode, bodyPreview)
	}

	// Read the body first for logging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logDebug("Termix FetchHosts Read Body Failed", err.Error())
		return nil, fmt.Errorf("termix: failed to read response body: %w", err)
	}
	
	logDebug("Termix FetchHosts Response Body", string(bodyBytes)[:min(len(bodyBytes), 500)])

	var termixHosts []TermixHost
	if err := json.Unmarshal(bodyBytes, &termixHosts); err != nil {
		bodyPreview := string(bodyBytes)
		logDebug("Termix FetchHosts JSON Decode Failed", fmt.Sprintf("Error: %v, Body: %s", err, bodyPreview))
		if len(bodyPreview) > 100 {
			bodyPreview = bodyPreview[:100] + "..."
		}
		return nil, fmt.Errorf("termix API returned invalid JSON (check baseUrl in termix.json): %s", bodyPreview)
	}
	
	logDebug("Termix FetchHosts Success", fmt.Sprintf("Decoded %d hosts", len(termixHosts)))

	// Convert Termix hosts to sshbuddy hosts
	hosts := make([]models.Host, 0, len(termixHosts))
	for _, th := range termixHosts {
		host := convertTermixHost(th)
		hosts = append(hosts, host)
	}

	return hosts, nil
}

// convertTermixHost converts a Termix host to sshbuddy host format
func convertTermixHost(th TermixHost) models.Host {
	host := models.Host{
		Alias:    th.Name,
		Hostname: th.IP,
		User:     th.Username,
		Port:     strconv.Itoa(th.Port),
		Tags:     th.Tags,
		Source:   "termix",
	}

	// Handle SSH key if present
	if th.Key != nil && *th.Key != "" {
		// Note: Termix stores the key content, but sshbuddy expects a file path
		// We'll need to handle this appropriately - for now, we'll skip it
		// In a production scenario, you might want to write the key to a temp file
	}

	return host
}


// GetJWT returns the current JWT token
func (c *Client) GetJWT() string {
	return c.jwt
}

// logDebug logs debug information to a file for troubleshooting
func logDebug(context string, message string) {
	logPath := "/tmp/sshbuddy-debug.log"
	
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return // Silently fail if we can't log
	}
	defer logFile.Close()
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] %s: %s\n", timestamp, context, message)
	logFile.WriteString(logLine)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
