package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		envVars     map[string]string
		expectError bool
		validate    func(t *testing.T, cfg *Config)
	}{
		{
			name: "valid config with token",
			configYAML: `
version: "1.0"
auth:
  github_token: "ghp_test123"
repositories:
  - owner: "testorg"
    name: "testrepo"
`,
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "1.0", cfg.Version)
				assert.Equal(t, "ghp_test123", cfg.Auth.GithubToken)
				assert.Len(t, cfg.Repositories, 1)
				assert.Equal(t, "testorg", cfg.Repositories[0].Owner)
				assert.Equal(t, "testrepo", cfg.Repositories[0].Name)
			},
		},
		{
			name: "config with env var substitution",
			configYAML: `
version: "1.0"
auth:
  github_token: "${TEST_GITHUB_TOKEN_LOAD}"
repositories:
  - owner: "testorg"
    name: "testrepo"
`,
			envVars: map[string]string{
				"TEST_GITHUB_TOKEN_LOAD": "ghp_from_env",
			},
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "ghp_from_env", cfg.Auth.GithubToken)
			},
		},
		{
			name: "config with date range",
			configYAML: `
version: "1.0"
auth:
  github_token: "ghp_test123"
repositories:
  - owner: "testorg"
    name: "testrepo"
date_range:
  start: "2024-01-01"
  end: "2024-12-31"
`,
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				dateRange, err := cfg.GetParsedDateRange()
				require.NoError(t, err)
				assert.NotNil(t, dateRange.Start)
				assert.NotNil(t, dateRange.End)
				assert.Equal(t, 2024, dateRange.Start.Year())
				assert.Equal(t, time.January, dateRange.Start.Month())
				assert.Equal(t, 1, dateRange.Start.Day())
			},
		},
		{
			name: "config with teams",
			configYAML: `
version: "1.0"
auth:
  github_token: "ghp_test123"
repositories:
  - owner: "testorg"
    name: "testrepo"
teams:
  - name: "Backend"
    members:
      - "user1"
      - "user2"
    color: "#3b82f6"
  - name: "Frontend"
    members:
      - "user3"
`,
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Len(t, cfg.Teams, 2)
				assert.Equal(t, "Backend", cfg.Teams[0].Name)
				assert.Contains(t, cfg.Teams[0].Members, "user1")
				assert.Equal(t, "#3b82f6", cfg.Teams[0].Color)
			},
		},
		{
			name: "config with custom scoring",
			configYAML: `
version: "1.0"
auth:
  github_token: "ghp_test123"
repositories:
  - owner: "testorg"
    name: "testrepo"
scoring:
  enabled: true
  points:
    commit: 20
    pr_merged: 100
`,
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.Scoring.Enabled)
				assert.Equal(t, 20, cfg.Scoring.Points.Commit)
				assert.Equal(t, 100, cfg.Scoring.Points.PRMerged)
			},
		},
		{
			name: "config with github app",
			configYAML: `
version: "1.0"
auth:
  github_app:
    app_id: 12345
    installation_id: 67890
    private_key: "test-key-content"
repositories:
  - owner: "testorg"
    name: "testrepo"
`,
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.HasGithubApp())
				assert.Equal(t, int64(12345), cfg.Auth.GithubApp.AppID)
				assert.Equal(t, int64(67890), cfg.Auth.GithubApp.InstallationID)
			},
		},
		{
			name: "invalid config - no auth",
			configYAML: `
version: "1.0"
repositories:
  - owner: "testorg"
    name: "testrepo"
`,
			expectError: true,
		},
		{
			name: "invalid config - no repositories",
			configYAML: `
version: "1.0"
auth:
  github_token: "ghp_test123"
`,
			expectError: true,
		},
		{
			name: "invalid config - invalid date format",
			configYAML: `
version: "1.0"
auth:
  github_token: "ghp_test123"
repositories:
  - owner: "testorg"
    name: "testrepo"
date_range:
  start: "not-a-date"
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables (sequential test due to env var usage)
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			// Create temp config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configPath, []byte(tt.configYAML), 0644)
			require.NoError(t, err)

			// Load config
			cfg, err := Load(configPath)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestExpandEnvVars(t *testing.T) {
	// Note: Tests that use t.Setenv cannot use t.Parallel in subtests
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "simple substitution",
			input:    "token: ${TEST_TOKEN_SIMPLE}",
			envVars:  map[string]string{"TEST_TOKEN_SIMPLE": "secret123"},
			expected: "token: secret123",
		},
		{
			name:     "multiple substitutions",
			input:    "user: ${TEST_USER_MULTI}, pass: ${TEST_PASS_MULTI}",
			envVars:  map[string]string{"TEST_USER_MULTI": "admin", "TEST_PASS_MULTI": "123"},
			expected: "user: admin, pass: 123",
		},
		{
			name:     "missing env var returns empty",
			input:    "token: ${TEST_MISSING_VAR_12345}",
			envVars:  map[string]string{},
			expected: "token: ",
		},
		{
			name:     "no substitution needed",
			input:    "token: plaintext",
			envVars:  map[string]string{},
			expected: "token: plaintext",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Sequential test due to env var usage
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			result := expandEnvVars(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_GetParsedDateRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		dateRange   DateRangeConfig
		expectError bool
		validate    func(t *testing.T, result *ParsedDateRange)
	}{
		{
			name: "valid date range",
			dateRange: DateRangeConfig{
				Start: "2024-01-01",
				End:   "2024-12-31",
			},
			expectError: false,
			validate: func(t *testing.T, result *ParsedDateRange) {
				assert.NotNil(t, result.Start)
				assert.NotNil(t, result.End)
				assert.Equal(t, 2024, result.Start.Year())
				assert.Equal(t, time.January, result.Start.Month())
				assert.Equal(t, 2024, result.End.Year())
				assert.Equal(t, time.December, result.End.Month())
			},
		},
		{
			name: "only start date",
			dateRange: DateRangeConfig{
				Start: "2024-06-15",
			},
			expectError: false,
			validate: func(t *testing.T, result *ParsedDateRange) {
				assert.NotNil(t, result.Start)
				assert.NotNil(t, result.End) // Should default to now
				assert.Equal(t, 2024, result.Start.Year())
				assert.Equal(t, time.June, result.Start.Month())
			},
		},
		{
			name:        "empty date range defaults to now",
			dateRange:   DateRangeConfig{},
			expectError: false,
			validate: func(t *testing.T, result *ParsedDateRange) {
				assert.Nil(t, result.Start)
				assert.NotNil(t, result.End)
			},
		},
		{
			name: "invalid start date",
			dateRange: DateRangeConfig{
				Start: "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid end date",
			dateRange: DateRangeConfig{
				Start: "2024-01-01",
				End:   "invalid",
			},
			expectError: true,
		},
		{
			name: "relative date - 90 days ago",
			dateRange: DateRangeConfig{
				Start: "-90d",
			},
			expectError: false,
			validate: func(t *testing.T, result *ParsedDateRange) {
				assert.NotNil(t, result.Start)
				assert.NotNil(t, result.End)
				// Start should be approximately 90 days ago
				expected := time.Now().AddDate(0, 0, -90)
				assert.Equal(t, expected.Year(), result.Start.Year())
				assert.Equal(t, expected.Month(), result.Start.Month())
				assert.Equal(t, expected.Day(), result.Start.Day())
			},
		},
		{
			name: "relative date - 2 weeks ago",
			dateRange: DateRangeConfig{
				Start: "-2w",
			},
			expectError: false,
			validate: func(t *testing.T, result *ParsedDateRange) {
				assert.NotNil(t, result.Start)
				expected := time.Now().AddDate(0, 0, -14)
				assert.Equal(t, expected.Year(), result.Start.Year())
				assert.Equal(t, expected.Month(), result.Start.Month())
				assert.Equal(t, expected.Day(), result.Start.Day())
			},
		},
		{
			name: "relative date - 3 months ago",
			dateRange: DateRangeConfig{
				Start: "-3m",
			},
			expectError: false,
			validate: func(t *testing.T, result *ParsedDateRange) {
				assert.NotNil(t, result.Start)
				expected := time.Now().AddDate(0, -3, 0)
				assert.Equal(t, expected.Year(), result.Start.Year())
				assert.Equal(t, expected.Month(), result.Start.Month())
			},
		},
		{
			name: "relative date - 1 year ago",
			dateRange: DateRangeConfig{
				Start: "-1y",
			},
			expectError: false,
			validate: func(t *testing.T, result *ParsedDateRange) {
				assert.NotNil(t, result.Start)
				expected := time.Now().AddDate(-1, 0, 0)
				assert.Equal(t, expected.Year(), result.Start.Year())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &Config{DateRange: tt.dateRange}
			result, err := cfg.GetParsedDateRange()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestConfig_GetCacheTTL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ttl         string
		expected    time.Duration
		expectError bool
	}{
		{
			name:     "24 hours",
			ttl:      "24h",
			expected: 24 * time.Hour,
		},
		{
			name:     "1 hour",
			ttl:      "1h",
			expected: 1 * time.Hour,
		},
		{
			name:     "30 minutes",
			ttl:      "30m",
			expected: 30 * time.Minute,
		},
		{
			name:     "empty defaults to 24h",
			ttl:      "",
			expected: 24 * time.Hour,
		},
		{
			name:        "invalid duration",
			ttl:         "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &Config{Cache: CacheConfig{TTL: tt.ttl}}
			result, err := cfg.GetCacheTTL()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_HasGithubToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{
			name:     "has token",
			token:    "ghp_test123",
			expected: true,
		},
		{
			name:     "empty token",
			token:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &Config{Auth: AuthConfig{GithubToken: tt.token}}
			assert.Equal(t, tt.expected, cfg.HasGithubToken())
		})
	}
}

func TestConfig_HasGithubApp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		appCfg   *GithubAppConfig
		expected bool
	}{
		{
			name: "valid github app config",
			appCfg: &GithubAppConfig{
				AppID:          12345,
				InstallationID: 67890,
				PrivateKey:     "key-content",
			},
			expected: true,
		},
		{
			name: "valid github app config with path",
			appCfg: &GithubAppConfig{
				AppID:          12345,
				InstallationID: 67890,
				PrivateKeyPath: "/path/to/key.pem",
			},
			expected: true,
		},
		{
			name:     "nil github app config",
			appCfg:   nil,
			expected: false,
		},
		{
			name: "missing app id",
			appCfg: &GithubAppConfig{
				InstallationID: 67890,
				PrivateKey:     "key-content",
			},
			expected: false,
		},
		{
			name: "missing installation id",
			appCfg: &GithubAppConfig{
				AppID:      12345,
				PrivateKey: "key-content",
			},
			expected: false,
		},
		{
			name: "missing private key",
			appCfg: &GithubAppConfig{
				AppID:          12345,
				InstallationID: 67890,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &Config{Auth: AuthConfig{GithubApp: tt.appCfg}}
			assert.Equal(t, tt.expected, cfg.HasGithubApp())
		})
	}
}

func TestConfig_GetTeamForUser(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Teams: []TeamConfig{
			{
				Name:    "Backend",
				Members: []string{"alice", "bob"},
				Color:   "#blue",
			},
			{
				Name:    "Frontend",
				Members: []string{"charlie", "dave"},
				Color:   "#green",
			},
		},
	}

	tests := []struct {
		name         string
		username     string
		expectedTeam string
		expectNil    bool
	}{
		{
			name:         "user in first team",
			username:     "alice",
			expectedTeam: "Backend",
		},
		{
			name:         "user in second team",
			username:     "charlie",
			expectedTeam: "Frontend",
		},
		{
			name:         "case insensitive match",
			username:     "ALICE",
			expectedTeam: "Backend",
		},
		{
			name:      "user not in any team",
			username:  "unknown",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			team := cfg.GetTeamForUser(tt.username)
			if tt.expectNil {
				assert.Nil(t, team)
			} else {
				require.NotNil(t, team)
				assert.Equal(t, tt.expectedTeam, team.Name)
			}
		})
	}
}

func TestConfig_IsBot(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Options: OptionsConfig{
			IncludeBots: false,
			BotPatterns: []string{
				"*[bot]",
				"dependabot*",
				"renovate*",
				"github-actions*",
			},
		},
	}

	tests := []struct {
		name     string
		username string
		expected bool
	}{
		{
			name:     "bot suffix pattern",
			username: "my-app[bot]",
			expected: true,
		},
		{
			name:     "dependabot prefix pattern",
			username: "dependabot-preview",
			expected: true,
		},
		{
			name:     "renovate prefix pattern",
			username: "renovate[bot]",
			expected: true,
		},
		{
			name:     "github-actions prefix pattern",
			username: "github-actions[bot]",
			expected: true,
		},
		{
			name:     "regular user",
			username: "alice",
			expected: false,
		},
		{
			name:     "user with bot in name",
			username: "robotics-engineer",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := cfg.IsBot(tt.username)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_IsBot_IncludeBots(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Options: OptionsConfig{
			IncludeBots: true,
			BotPatterns: []string{"*[bot]"},
		},
	}

	// When IncludeBots is true, nothing should be considered a bot
	assert.False(t, cfg.IsBot("my-app[bot]"))
	assert.False(t, cfg.IsBot("dependabot"))
}

func TestMatchPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		s        string
		pattern  string
		expected bool
	}{
		{
			name:     "exact match",
			s:        "hello",
			pattern:  "hello",
			expected: true,
		},
		{
			name:     "prefix match",
			s:        "hello-world",
			pattern:  "hello*",
			expected: true,
		},
		{
			name:     "suffix match",
			s:        "hello-world",
			pattern:  "*world",
			expected: true,
		},
		{
			name:     "contains match",
			s:        "hello-world-test",
			pattern:  "*world*",
			expected: true,
		},
		{
			name:     "no match",
			s:        "hello",
			pattern:  "world",
			expected: false,
		},
		{
			name:     "prefix no match",
			s:        "hello",
			pattern:  "world*",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := matchPattern(tt.s, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_GetCustomPeriods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		customPeriods []CustomPeriod
		expectError   bool
		validate      func(t *testing.T, periods []ParsedCustomPeriod)
	}{
		{
			name: "valid custom periods",
			customPeriods: []CustomPeriod{
				{Name: "Q1", Start: "2024-01-01", End: "2024-03-31"},
				{Name: "Q2", Start: "2024-04-01", End: "2024-06-30"},
			},
			expectError: false,
			validate: func(t *testing.T, periods []ParsedCustomPeriod) {
				assert.Len(t, periods, 2)
				assert.Equal(t, "Q1", periods[0].Name)
				assert.Equal(t, time.January, periods[0].Start.Month())
				assert.Equal(t, time.March, periods[0].End.Month())
			},
		},
		{
			name:          "empty custom periods",
			customPeriods: []CustomPeriod{},
			expectError:   false,
			validate: func(t *testing.T, periods []ParsedCustomPeriod) {
				assert.Empty(t, periods)
			},
		},
		{
			name: "invalid start date",
			customPeriods: []CustomPeriod{
				{Name: "Bad", Start: "invalid", End: "2024-03-31"},
			},
			expectError: true,
		},
		{
			name: "invalid end date",
			customPeriods: []CustomPeriod{
				{Name: "Bad", Start: "2024-01-01", End: "invalid"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &Config{CustomPeriods: tt.customPeriods}
			periods, err := cfg.GetCustomPeriods()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, periods)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.Equal(t, "1.0", cfg.Version)
	assert.Contains(t, cfg.Granularity, "daily")
	assert.Contains(t, cfg.Granularity, "weekly")
	assert.Contains(t, cfg.Granularity, "monthly")
	assert.True(t, cfg.Scoring.Enabled)
	assert.Equal(t, 10, cfg.Scoring.Points.Commit)
	assert.Equal(t, 50, cfg.Scoring.Points.PRMerged)
	assert.NotEmpty(t, cfg.Scoring.GetAchievements())
	assert.Equal(t, "./dist", cfg.Output.Directory)
	assert.True(t, cfg.Cache.Enabled)
	assert.Equal(t, "./.cache", cfg.Cache.Directory)
	assert.Equal(t, "24h", cfg.Cache.TTL)
	assert.Equal(t, 5, cfg.Options.ConcurrentRequests)
	assert.False(t, cfg.Options.IncludeBots)
}

func TestConfig_GetGithubAppPrivateKey(t *testing.T) {
	t.Parallel()

	t.Run("returns inline key", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Auth: AuthConfig{
				GithubApp: &GithubAppConfig{
					PrivateKey: "inline-key-content",
				},
			},
		}

		key, err := cfg.GetGithubAppPrivateKey()
		require.NoError(t, err)
		assert.Equal(t, []byte("inline-key-content"), key)
	})

	t.Run("returns key from file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		keyPath := filepath.Join(tmpDir, "key.pem")
		err := os.WriteFile(keyPath, []byte("file-key-content"), 0600)
		require.NoError(t, err)

		cfg := &Config{
			Auth: AuthConfig{
				GithubApp: &GithubAppConfig{
					PrivateKeyPath: keyPath,
				},
			},
		}

		key, err := cfg.GetGithubAppPrivateKey()
		require.NoError(t, err)
		assert.Equal(t, []byte("file-key-content"), key)
	})

	t.Run("error when no github app configured", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		_, err := cfg.GetGithubAppPrivateKey()
		assert.Error(t, err)
	})

	t.Run("error when no key configured", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Auth: AuthConfig{
				GithubApp: &GithubAppConfig{
					AppID:          12345,
					InstallationID: 67890,
				},
			},
		}

		_, err := cfg.GetGithubAppPrivateKey()
		assert.Error(t, err)
	})
}

func TestLoad_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := Load("/nonexistent/path/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)
	require.NoError(t, err)

	_, err = Load(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}
