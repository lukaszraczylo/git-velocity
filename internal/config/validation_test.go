package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorField  string
	}{
		{
			name: "valid config with token",
			config: &Config{
				Auth: AuthConfig{
					GithubToken: "ghp_test123",
				},
				Repositories: []RepositoryConfig{
					{Owner: "testorg", Name: "testrepo"},
				},
				Granularity: []string{"daily", "weekly"},
				Output: OutputConfig{
					Directory: "./dist",
					Format:    []string{"html", "json"},
				},
				Cache: CacheConfig{
					Enabled:   true,
					Directory: "./.cache",
					TTL:       "24h",
				},
				Options: OptionsConfig{
					ConcurrentRequests: 5,
				},
			},
			expectError: false,
		},
		{
			name: "valid config with github app",
			config: &Config{
				Auth: AuthConfig{
					GithubApp: &GithubAppConfig{
						AppID:          12345,
						InstallationID: 67890,
						PrivateKey:     "key-content",
					},
				},
				Repositories: []RepositoryConfig{
					{Owner: "testorg", Name: "testrepo"},
				},
				Granularity: []string{"daily"},
				Output: OutputConfig{
					Directory: "./dist",
					Format:    []string{"html"},
				},
				Options: OptionsConfig{
					ConcurrentRequests: 5,
				},
			},
			expectError: false,
		},
		{
			name: "missing authentication",
			config: &Config{
				Repositories: []RepositoryConfig{
					{Owner: "testorg", Name: "testrepo"},
				},
				Granularity: []string{"daily"},
				Output: OutputConfig{
					Directory: "./dist",
					Format:    []string{"html"},
				},
				Options: OptionsConfig{
					ConcurrentRequests: 5,
				},
			},
			expectError: true,
			errorField:  "auth",
		},
		{
			name: "no repositories",
			config: &Config{
				Auth: AuthConfig{
					GithubToken: "ghp_test123",
				},
				Repositories: []RepositoryConfig{},
				Granularity:  []string{"daily"},
				Output: OutputConfig{
					Directory: "./dist",
					Format:    []string{"html"},
				},
				Options: OptionsConfig{
					ConcurrentRequests: 5,
				},
			},
			expectError: true,
			errorField:  "repositories",
		},
		{
			name: "repository missing owner",
			config: &Config{
				Auth: AuthConfig{
					GithubToken: "ghp_test123",
				},
				Repositories: []RepositoryConfig{
					{Name: "testrepo"},
				},
				Granularity: []string{"daily"},
				Output: OutputConfig{
					Directory: "./dist",
					Format:    []string{"html"},
				},
				Options: OptionsConfig{
					ConcurrentRequests: 5,
				},
			},
			expectError: true,
			errorField:  "repositories[0].owner",
		},
		{
			name: "repository missing name and pattern",
			config: &Config{
				Auth: AuthConfig{
					GithubToken: "ghp_test123",
				},
				Repositories: []RepositoryConfig{
					{Owner: "testorg"},
				},
				Granularity: []string{"daily"},
				Output: OutputConfig{
					Directory: "./dist",
					Format:    []string{"html"},
				},
				Options: OptionsConfig{
					ConcurrentRequests: 5,
				},
			},
			expectError: true,
			errorField:  "repositories[0]",
		},
		{
			name: "repository with pattern instead of name is valid",
			config: &Config{
				Auth: AuthConfig{
					GithubToken: "ghp_test123",
				},
				Repositories: []RepositoryConfig{
					{Owner: "testorg", Pattern: "*"},
				},
				Granularity: []string{"daily"},
				Output: OutputConfig{
					Directory: "./dist",
					Format:    []string{"html"},
				},
				Options: OptionsConfig{
					ConcurrentRequests: 5,
				},
			},
			expectError: false,
		},
		{
			name: "invalid granularity",
			config: &Config{
				Auth: AuthConfig{
					GithubToken: "ghp_test123",
				},
				Repositories: []RepositoryConfig{
					{Owner: "testorg", Name: "testrepo"},
				},
				Granularity: []string{"invalid"},
				Output: OutputConfig{
					Directory: "./dist",
					Format:    []string{"html"},
				},
				Options: OptionsConfig{
					ConcurrentRequests: 5,
				},
			},
			expectError: true,
			errorField:  "granularity",
		},
		{
			name: "team without name",
			config: &Config{
				Auth: AuthConfig{
					GithubToken: "ghp_test123",
				},
				Repositories: []RepositoryConfig{
					{Owner: "testorg", Name: "testrepo"},
				},
				Teams: []TeamConfig{
					{Members: []string{"user1"}},
				},
				Granularity: []string{"daily"},
				Output: OutputConfig{
					Directory: "./dist",
					Format:    []string{"html"},
				},
				Options: OptionsConfig{
					ConcurrentRequests: 5,
				},
			},
			expectError: true,
			errorField:  "teams[0].name",
		},
		{
			name: "team without members",
			config: &Config{
				Auth: AuthConfig{
					GithubToken: "ghp_test123",
				},
				Repositories: []RepositoryConfig{
					{Owner: "testorg", Name: "testrepo"},
				},
				Teams: []TeamConfig{
					{Name: "Backend", Members: []string{}},
				},
				Granularity: []string{"daily"},
				Output: OutputConfig{
					Directory: "./dist",
					Format:    []string{"html"},
				},
				Options: OptionsConfig{
					ConcurrentRequests: 5,
				},
			},
			expectError: true,
			errorField:  "teams[0].members",
		},
		// Note: Achievement validation tests removed because achievements are now hardcoded
		// and not user-configurable to prevent manipulation
		{
			name: "missing output directory",
			config: &Config{
				Auth: AuthConfig{
					GithubToken: "ghp_test123",
				},
				Repositories: []RepositoryConfig{
					{Owner: "testorg", Name: "testrepo"},
				},
				Granularity: []string{"daily"},
				Output: OutputConfig{
					Directory: "",
					Format:    []string{"html"},
				},
				Options: OptionsConfig{
					ConcurrentRequests: 5,
				},
			},
			expectError: true,
			errorField:  "output.directory",
		},
		{
			name: "invalid output format",
			config: &Config{
				Auth: AuthConfig{
					GithubToken: "ghp_test123",
				},
				Repositories: []RepositoryConfig{
					{Owner: "testorg", Name: "testrepo"},
				},
				Granularity: []string{"daily"},
				Output: OutputConfig{
					Directory: "./dist",
					Format:    []string{"invalid"},
				},
				Options: OptionsConfig{
					ConcurrentRequests: 5,
				},
			},
			expectError: true,
			errorField:  "output.format",
		},
		{
			name: "cache enabled but no directory",
			config: &Config{
				Auth: AuthConfig{
					GithubToken: "ghp_test123",
				},
				Repositories: []RepositoryConfig{
					{Owner: "testorg", Name: "testrepo"},
				},
				Granularity: []string{"daily"},
				Output: OutputConfig{
					Directory: "./dist",
					Format:    []string{"html"},
				},
				Cache: CacheConfig{
					Enabled:   true,
					Directory: "",
					TTL:       "24h",
				},
				Options: OptionsConfig{
					ConcurrentRequests: 5,
				},
			},
			expectError: true,
			errorField:  "cache.directory",
		},
		{
			name: "invalid cache TTL",
			config: &Config{
				Auth: AuthConfig{
					GithubToken: "ghp_test123",
				},
				Repositories: []RepositoryConfig{
					{Owner: "testorg", Name: "testrepo"},
				},
				Granularity: []string{"daily"},
				Output: OutputConfig{
					Directory: "./dist",
					Format:    []string{"html"},
				},
				Cache: CacheConfig{
					Enabled:   true,
					Directory: "./.cache",
					TTL:       "invalid",
				},
				Options: OptionsConfig{
					ConcurrentRequests: 5,
				},
			},
			expectError: true,
			errorField:  "cache.ttl",
		},
		{
			name: "concurrent requests too low",
			config: &Config{
				Auth: AuthConfig{
					GithubToken: "ghp_test123",
				},
				Repositories: []RepositoryConfig{
					{Owner: "testorg", Name: "testrepo"},
				},
				Granularity: []string{"daily"},
				Output: OutputConfig{
					Directory: "./dist",
					Format:    []string{"html"},
				},
				Options: OptionsConfig{
					ConcurrentRequests: 0,
				},
			},
			expectError: true,
			errorField:  "options.concurrent_requests",
		},
		{
			name: "concurrent requests too high",
			config: &Config{
				Auth: AuthConfig{
					GithubToken: "ghp_test123",
				},
				Repositories: []RepositoryConfig{
					{Owner: "testorg", Name: "testrepo"},
				},
				Granularity: []string{"daily"},
				Output: OutputConfig{
					Directory: "./dist",
					Format:    []string{"html"},
				},
				Options: OptionsConfig{
					ConcurrentRequests: 100,
				},
			},
			expectError: true,
			errorField:  "options.concurrent_requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Validate(tt.config)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorField)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	t.Parallel()

	err := ValidationError{
		Field:   "test.field",
		Message: "test error message",
	}

	assert.Equal(t, "test.field: test error message", err.Error())
}

func TestValidationErrors_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		errs     ValidationErrors
		expected string
	}{
		{
			name:     "empty errors",
			errs:     ValidationErrors{},
			expected: "",
		},
		{
			name: "single error",
			errs: ValidationErrors{
				{Field: "field1", Message: "error1"},
			},
			expected: "field1: error1",
		},
		{
			name: "multiple errors",
			errs: ValidationErrors{
				{Field: "field1", Message: "error1"},
				{Field: "field2", Message: "error2"},
			},
			expected: "field1: error1; field2: error2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.errs.Error())
		})
	}
}
