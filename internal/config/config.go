package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Load reads and parses a configuration file
func Load(path string) (*Config, error) {
	cleanPath := filepath.Clean(path)
	data, err := os.ReadFile(cleanPath) // #nosec G304 -- path is user-provided config file
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	expanded := expandEnvVars(string(data))

	// Start with defaults
	cfg := DefaultConfig()

	// Parse YAML
	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// expandEnvVars replaces ${VAR} patterns with environment variable values
func expandEnvVars(input string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name
		varName := strings.TrimPrefix(strings.TrimSuffix(match, "}"), "${")
		return os.Getenv(varName)
	})
}

// parseRelativeDate parses relative date strings like "-90d", "-2w", "-3m"
// Returns the parsed time or nil if not a relative format
func parseRelativeDate(s string) *time.Time {
	if !strings.HasPrefix(s, "-") && !strings.HasPrefix(s, "+") {
		return nil
	}

	// Parse the number and unit
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return nil
	}

	unit := s[len(s)-1]
	numStr := s[1 : len(s)-1] // Skip the +/- prefix and unit suffix

	num := 0
	for _, c := range numStr {
		if c < '0' || c > '9' {
			return nil
		}
		num = num*10 + int(c-'0')
	}

	if s[0] == '-' {
		num = -num
	}

	now := time.Now()
	var result time.Time

	switch unit {
	case 'd': // days
		result = now.AddDate(0, 0, num)
	case 'w': // weeks
		result = now.AddDate(0, 0, num*7)
	case 'm': // months
		result = now.AddDate(0, num, 0)
	case 'y': // years
		result = now.AddDate(num, 0, 0)
	default:
		return nil
	}

	// Normalize to start of day
	result = time.Date(result.Year(), result.Month(), result.Day(), 0, 0, 0, 0, result.Location())
	return &result
}

// GetParsedDateRange parses and returns the date range with defaults
// Supports both absolute dates (2024-01-01) and relative dates (-90d, -2w, -3m, -1y)
func (c *Config) GetParsedDateRange() (*ParsedDateRange, error) {
	result := &ParsedDateRange{}

	if c.DateRange.Start != "" {
		// Try relative date first
		if t := parseRelativeDate(c.DateRange.Start); t != nil {
			result.Start = t
		} else {
			// Try absolute date
			t, err := time.Parse("2006-01-02", c.DateRange.Start)
			if err != nil {
				return nil, fmt.Errorf("invalid start date format (use YYYY-MM-DD or -Nd/-Nw/-Nm/-Ny): %w", err)
			}
			result.Start = &t
		}
	}

	if c.DateRange.End != "" {
		// Try relative date first
		if t := parseRelativeDate(c.DateRange.End); t != nil {
			// Set end to end of day
			endOfDay := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			result.End = &endOfDay
		} else {
			// Try absolute date
			t, err := time.Parse("2006-01-02", c.DateRange.End)
			if err != nil {
				return nil, fmt.Errorf("invalid end date format (use YYYY-MM-DD or -Nd/-Nw/-Nm/-Ny): %w", err)
			}
			// Set end to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			result.End = &t
		}
	} else {
		// Default to now
		now := time.Now()
		result.End = &now
	}

	return result, nil
}

// GetCacheTTL returns the cache TTL as a time.Duration
func (c *Config) GetCacheTTL() (time.Duration, error) {
	if c.Cache.TTL == "" {
		return 24 * time.Hour, nil
	}
	return time.ParseDuration(c.Cache.TTL)
}

// HasGithubToken returns true if token authentication is configured
func (c *Config) HasGithubToken() bool {
	return c.Auth.GithubToken != ""
}

// HasGithubApp returns true if GitHub App authentication is configured
func (c *Config) HasGithubApp() bool {
	return c.Auth.GithubApp != nil &&
		c.Auth.GithubApp.AppID > 0 &&
		c.Auth.GithubApp.InstallationID > 0 &&
		(c.Auth.GithubApp.PrivateKey != "" || c.Auth.GithubApp.PrivateKeyPath != "")
}

// GetGithubAppPrivateKey returns the GitHub App private key content
func (c *Config) GetGithubAppPrivateKey() ([]byte, error) {
	if c.Auth.GithubApp == nil {
		return nil, fmt.Errorf("GitHub App not configured")
	}

	if c.Auth.GithubApp.PrivateKey != "" {
		return []byte(c.Auth.GithubApp.PrivateKey), nil
	}

	if c.Auth.GithubApp.PrivateKeyPath != "" {
		cleanPath := filepath.Clean(c.Auth.GithubApp.PrivateKeyPath)
		return os.ReadFile(cleanPath) // #nosec G304 -- path is user-provided config value
	}

	return nil, fmt.Errorf("no private key configured")
}

// GetTeamForUser returns the team configuration for a given username
func (c *Config) GetTeamForUser(username string) *TeamConfig {
	for i := range c.Teams {
		for _, member := range c.Teams[i].Members {
			if strings.EqualFold(member, username) {
				return &c.Teams[i]
			}
		}
	}
	return nil
}

// IsBot checks if a username matches bot patterns
func (c *Config) IsBot(username string) bool {
	if c.Options.IncludeBots {
		return false
	}

	lower := strings.ToLower(username)
	for _, pattern := range c.Options.BotPatterns {
		pattern = strings.ToLower(pattern)
		if matchPattern(lower, pattern) {
			return true
		}
	}
	return false
}

// matchPattern performs simple glob-style pattern matching
func matchPattern(s, pattern string) bool {
	// Handle exact match
	if !strings.Contains(pattern, "*") {
		return s == pattern
	}

	// Handle prefix match (pattern*)
	if strings.HasSuffix(pattern, "*") && !strings.HasPrefix(pattern, "*") {
		return strings.HasPrefix(s, strings.TrimSuffix(pattern, "*"))
	}

	// Handle suffix match (*pattern)
	if strings.HasPrefix(pattern, "*") && !strings.HasSuffix(pattern, "*") {
		return strings.HasSuffix(s, strings.TrimPrefix(pattern, "*"))
	}

	// Handle contains match (*pattern*)
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		inner := strings.TrimPrefix(strings.TrimSuffix(pattern, "*"), "*")
		return strings.Contains(s, inner)
	}

	return false
}

// GetCustomPeriods returns parsed custom periods
func (c *Config) GetCustomPeriods() ([]ParsedCustomPeriod, error) {
	var periods []ParsedCustomPeriod

	for _, cp := range c.CustomPeriods {
		start, err := time.Parse("2006-01-02", cp.Start)
		if err != nil {
			return nil, fmt.Errorf("invalid start date for period %s: %w", cp.Name, err)
		}

		end, err := time.Parse("2006-01-02", cp.End)
		if err != nil {
			return nil, fmt.Errorf("invalid end date for period %s: %w", cp.Name, err)
		}

		// Set end to end of day
		end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

		periods = append(periods, ParsedCustomPeriod{
			Name:  cp.Name,
			Start: start,
			End:   end,
		})
	}

	return periods, nil
}

// ParsedCustomPeriod represents a parsed custom time period
type ParsedCustomPeriod struct {
	Name  string
	Start time.Time
	End   time.Time
}
