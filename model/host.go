package model

import (
	"fmt"
	"strconv"
	"strings"
)

type Host struct {
	Alias    string   `json:"alias"`
	Hostname string   `json:"hostname"`
	User     string   `json:"user"`
	Port     string   `json:"port"`
	Tags     []string `json:"tags"`
}

type Config struct {
	Hosts []Host `json:"hosts"`
	Theme string `json:"theme,omitempty"`
}

// ValidationError represents a config validation error
type ValidationError struct {
	Field   string
	Message string
	Index   int // -1 for config-level errors, >= 0 for host-specific errors
}

func (e ValidationError) Error() string {
	if e.Index >= 0 {
		return fmt.Sprintf("Host #%d (%s): %s", e.Index+1, e.Field, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate checks if a host configuration is valid
func (h *Host) Validate() []ValidationError {
	var errors []ValidationError

	// Alias is required
	if strings.TrimSpace(h.Alias) == "" {
		errors = append(errors, ValidationError{
			Field:   "Alias",
			Message: "alias is required",
			Index:   -1,
		})
	}

	// Hostname is required
	if strings.TrimSpace(h.Hostname) == "" {
		errors = append(errors, ValidationError{
			Field:   "Hostname",
			Message: "hostname is required",
			Index:   -1,
		})
	}

	// User is required
	if strings.TrimSpace(h.User) == "" {
		errors = append(errors, ValidationError{
			Field:   "User",
			Message: "user is required",
			Index:   -1,
		})
	}

	// Port validation (if provided)
	if h.Port != "" {
		port, err := strconv.Atoi(h.Port)
		if err != nil {
			errors = append(errors, ValidationError{
				Field:   "Port",
				Message: "port must be a number",
				Index:   -1,
			})
		} else if port < 1 || port > 65535 {
			errors = append(errors, ValidationError{
				Field:   "Port",
				Message: "port must be between 1 and 65535",
				Index:   -1,
			})
		}
	}

	return errors
}

// Validate checks if the entire config is valid
func (c *Config) Validate() []ValidationError {
	var errors []ValidationError

	// Check for duplicate aliases
	aliasMap := make(map[string]int)
	for i, host := range c.Hosts {
		alias := strings.TrimSpace(host.Alias)
		if alias != "" {
			if firstIdx, exists := aliasMap[alias]; exists {
				errors = append(errors, ValidationError{
					Field:   "Alias",
					Message: fmt.Sprintf("duplicate alias '%s' (also used in host #%d)", alias, firstIdx+1),
					Index:   i,
				})
			} else {
				aliasMap[alias] = i
			}
		}

		// Validate each host
		hostErrors := host.Validate()
		for _, err := range hostErrors {
			err.Index = i
			errors = append(errors, err)
		}
	}

	// Validate theme if provided
	if c.Theme != "" {
		validThemes := []string{"purple", "blue", "green", "pink", "amber", "cyan"}
		isValid := false
		for _, valid := range validThemes {
			if c.Theme == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			errors = append(errors, ValidationError{
				Field:   "Theme",
				Message: fmt.Sprintf("invalid theme '%s' (valid: %s)", c.Theme, strings.Join(validThemes, ", ")),
				Index:   -1,
			})
		}
	}

	return errors
}
