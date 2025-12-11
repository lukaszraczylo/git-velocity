package config

import (
	"fmt"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// Validate checks the configuration for errors
func Validate(cfg *Config) error {
	var errs ValidationErrors

	// Validate authentication
	if !cfg.HasGithubToken() && !cfg.HasGithubApp() {
		errs = append(errs, ValidationError{
			Field:   "auth",
			Message: "either github_token or github_app must be configured",
		})
	}

	// Validate repositories
	if len(cfg.Repositories) == 0 {
		errs = append(errs, ValidationError{
			Field:   "repositories",
			Message: "at least one repository must be specified",
		})
	}

	for i, repo := range cfg.Repositories {
		if repo.Owner == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("repositories[%d].owner", i),
				Message: "owner is required",
			})
		}
		if repo.Name == "" && repo.Pattern == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("repositories[%d]", i),
				Message: "either name or pattern must be specified",
			})
		}
	}

	// Validate date range
	if cfg.DateRange.Start != "" {
		if _, err := cfg.GetParsedDateRange(); err != nil {
			errs = append(errs, ValidationError{
				Field:   "date_range",
				Message: err.Error(),
			})
		}
	}

	// Validate granularity
	validGranularities := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}
	for _, g := range cfg.Granularity {
		if !validGranularities[g] {
			errs = append(errs, ValidationError{
				Field:   "granularity",
				Message: fmt.Sprintf("invalid granularity: %s (must be daily, weekly, or monthly)", g),
			})
		}
	}

	// Validate teams
	for i, team := range cfg.Teams {
		if team.Name == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("teams[%d].name", i),
				Message: "team name is required",
			})
		}
		if len(team.Members) == 0 {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("teams[%d].members", i),
				Message: "team must have at least one member",
			})
		}
	}

	// Validate scoring
	if cfg.Scoring.Enabled {
		if cfg.Scoring.Points.Commit < 0 {
			errs = append(errs, ValidationError{
				Field:   "scoring.points.commit",
				Message: "point values cannot be negative",
			})
		}
		// Additional point validations can be added here
	}

	// Note: Achievements are hardcoded and not user-configurable to prevent manipulation

	// Validate output
	if cfg.Output.Directory == "" {
		errs = append(errs, ValidationError{
			Field:   "output.directory",
			Message: "output directory is required",
		})
	}

	validFormats := map[string]bool{"html": true, "json": true}
	for _, format := range cfg.Output.Format {
		if !validFormats[format] {
			errs = append(errs, ValidationError{
				Field:   "output.format",
				Message: fmt.Sprintf("invalid format: %s (must be html or json)", format),
			})
		}
	}

	// Validate cache
	if cfg.Cache.Enabled {
		if cfg.Cache.Directory == "" {
			errs = append(errs, ValidationError{
				Field:   "cache.directory",
				Message: "cache directory is required when caching is enabled",
			})
		}
		if _, err := cfg.GetCacheTTL(); err != nil {
			errs = append(errs, ValidationError{
				Field:   "cache.ttl",
				Message: fmt.Sprintf("invalid TTL duration: %v", err),
			})
		}
	}

	// Validate options
	if cfg.Options.ConcurrentRequests < 1 {
		errs = append(errs, ValidationError{
			Field:   "options.concurrent_requests",
			Message: "must be at least 1",
		})
	}
	if cfg.Options.ConcurrentRequests > 20 {
		errs = append(errs, ValidationError{
			Field:   "options.concurrent_requests",
			Message: "should not exceed 20 to avoid rate limiting",
		})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
