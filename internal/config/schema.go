package config

import "time"

// Config represents the main configuration structure
type Config struct {
	Version       string             `yaml:"version"`
	Auth          AuthConfig         `yaml:"auth"`
	Repositories  []RepositoryConfig `yaml:"repositories"`
	DateRange     DateRangeConfig    `yaml:"date_range"`
	Granularity   []string           `yaml:"granularity"`
	CustomPeriods []CustomPeriod     `yaml:"custom_periods,omitempty"`
	Teams         []TeamConfig       `yaml:"teams,omitempty"`
	Scoring       ScoringConfig      `yaml:"scoring"`
	Output        OutputConfig       `yaml:"output"`
	Cache         CacheConfig        `yaml:"cache"`
	Options       OptionsConfig      `yaml:"options"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// Token-based authentication
	GithubToken string `yaml:"github_token,omitempty"`

	// GitHub App authentication
	GithubApp *GithubAppConfig `yaml:"github_app,omitempty"`
}

// GithubAppConfig holds GitHub App authentication details
type GithubAppConfig struct {
	AppID          int64  `yaml:"app_id"`
	InstallationID int64  `yaml:"installation_id"`
	PrivateKeyPath string `yaml:"private_key_path,omitempty"`
	PrivateKey     string `yaml:"private_key,omitempty"`
}

// RepositoryConfig defines a repository to analyze
type RepositoryConfig struct {
	Owner   string `yaml:"owner"`
	Name    string `yaml:"name,omitempty"`
	Pattern string `yaml:"pattern,omitempty"` // For wildcard matching
}

// DateRangeConfig specifies the analysis time range
type DateRangeConfig struct {
	Start string `yaml:"start,omitempty"` // ISO 8601 format
	End   string `yaml:"end,omitempty"`   // ISO 8601 format
}

// CustomPeriod defines a custom time period for analysis
type CustomPeriod struct {
	Name  string `yaml:"name"`
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

// TeamConfig defines a team and its members
type TeamConfig struct {
	Name    string   `yaml:"name"`
	Members []string `yaml:"members"`
	Color   string   `yaml:"color,omitempty"`
}

// ScoringConfig holds gamification scoring configuration
type ScoringConfig struct {
	Enabled      bool                `yaml:"enabled"`
	Points       PointsConfig        `yaml:"points"`
	Achievements []AchievementConfig `yaml:"achievements,omitempty"`
}

// PointsConfig defines point values for various activities
type PointsConfig struct {
	Commit          int     `yaml:"commit"`
	CommitWithTests int     `yaml:"commit_with_tests"`
	LinesAdded      float64 `yaml:"lines_added"`
	LinesDeleted    float64 `yaml:"lines_deleted"`
	PROpened        int     `yaml:"pr_opened"`
	PRMerged        int     `yaml:"pr_merged"`
	PRReviewed      int     `yaml:"pr_reviewed"`
	ReviewComment   int     `yaml:"review_comment"` // PR review comments (not code comments)
	IssueOpened     int     `yaml:"issue_opened"`
	IssueClosed     int     `yaml:"issue_closed"`
	FastReview1h    int     `yaml:"fast_review_1h"`
	FastReview4h    int     `yaml:"fast_review_4h"`
	FastReview24h   int     `yaml:"fast_review_24h"`
}

// AchievementConfig defines an achievement badge
type AchievementConfig struct {
	ID          string               `yaml:"id"`
	Name        string               `yaml:"name"`
	Description string               `yaml:"description"`
	Icon        string               `yaml:"icon"`
	Condition   AchievementCondition `yaml:"condition"`
}

// AchievementCondition defines when an achievement is earned
type AchievementCondition struct {
	Type      string  `yaml:"type"` // commit_count, pr_count, review_count, avg_review_time, etc.
	Threshold float64 `yaml:"threshold"`
}

// TierFromThreshold returns the tier level (1-11) based on threshold value
// Tiers: 1=1, 2=10, 3=25, 4=50, 5=100, 6=250, 7=500, 8=1000, 9=5000, 10=10000, 11=25000+
func TierFromThreshold(threshold float64) int {
	tiers := []float64{1, 10, 25, 50, 100, 250, 500, 1000, 5000, 10000, 25000}
	for i := len(tiers) - 1; i >= 0; i-- {
		if threshold >= tiers[i] {
			return i + 1
		}
	}
	return 1
}

// OutputConfig specifies output generation settings
type OutputConfig struct {
	Directory string       `yaml:"directory"`
	Format    []string     `yaml:"format"` // html, json
	Deploy    DeployConfig `yaml:"deploy"`
}

// DeployConfig specifies deployment options
type DeployConfig struct {
	GHPages  bool `yaml:"gh_pages"`
	Artifact bool `yaml:"artifact"`
}

// CacheConfig holds caching configuration
type CacheConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Directory string `yaml:"directory"`
	TTL       string `yaml:"ttl"` // Duration string like "24h"
}

// OptionsConfig holds advanced options
type OptionsConfig struct {
	ConcurrentRequests int         `yaml:"concurrent_requests"`
	IncludeBots        bool        `yaml:"include_bots"`
	BotPatterns        []string    `yaml:"bot_patterns"`
	CloneDirectory     string      `yaml:"clone_directory"`        // Directory for local git clones
	UseLocalGit        bool        `yaml:"use_local_git"`          // Use local git for commits (faster)
	UserAliases        []UserAlias `yaml:"user_aliases,omitempty"` // Manual email/name to login mappings
}

// UserAlias maps git emails or names to a GitHub login
type UserAlias struct {
	GithubLogin string   `yaml:"github_login"`     // The canonical GitHub username
	Emails      []string `yaml:"emails,omitempty"` // Git commit emails to map
	Names       []string `yaml:"names,omitempty"`  // Git commit author names to map
}

// ParsedDateRange holds parsed date range values
type ParsedDateRange struct {
	Start *time.Time
	End   *time.Time
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Version:     "1.0",
		Granularity: []string{"daily", "weekly", "monthly"},
		Scoring: ScoringConfig{
			Enabled: true,
			Points: PointsConfig{
				Commit:          10,
				CommitWithTests: 15,
				LinesAdded:      0.1,
				LinesDeleted:    0.05,
				PROpened:        25,
				PRMerged:        50,
				PRReviewed:      30,
				ReviewComment:   5,
				IssueOpened:     15,
				IssueClosed:     20,
				FastReview1h:    50,
				FastReview4h:    25,
				FastReview24h:   10,
			},
			Achievements: defaultAchievements(),
		},
		Output: OutputConfig{
			Directory: "./dist",
			Format:    []string{"html", "json"},
			Deploy: DeployConfig{
				GHPages:  true,
				Artifact: true,
			},
		},
		Cache: CacheConfig{
			Enabled:   true,
			Directory: "./.cache",
			TTL:       "24h",
		},
		Options: OptionsConfig{
			ConcurrentRequests: 5,
			IncludeBots:        false,
			BotPatterns: []string{
				"*[bot]",
				"dependabot*",
				"renovate*",
				"github-actions*",
			},
			CloneDirectory: "./.repos",
			UseLocalGit:    true, // Default to faster local git analysis
		},
	}
}

// defaultAchievements returns the default achievement badges
func defaultAchievements() []AchievementConfig {
	return []AchievementConfig{
		{
			ID:          "first-commit",
			Name:        "First Steps",
			Description: "Made your first commit",
			Icon:        "fa-baby",
			Condition:   AchievementCondition{Type: "commit_count", Threshold: 1},
		},
		{
			ID:          "commit-10",
			Name:        "Getting Started",
			Description: "Made 10 commits",
			Icon:        "fa-seedling",
			Condition:   AchievementCondition{Type: "commit_count", Threshold: 10},
		},
		{
			ID:          "commit-100",
			Name:        "Committed",
			Description: "Made 100 commits",
			Icon:        "fa-fire",
			Condition:   AchievementCondition{Type: "commit_count", Threshold: 100},
		},
		{
			ID:          "commit-500",
			Name:        "Code Machine",
			Description: "Made 500 commits",
			Icon:        "fa-robot",
			Condition:   AchievementCondition{Type: "commit_count", Threshold: 500},
		},
		{
			ID:          "commit-1000",
			Name:        "Code Warrior",
			Description: "Made 1000 commits",
			Icon:        "fa-crown",
			Condition:   AchievementCondition{Type: "commit_count", Threshold: 1000},
		},
		{
			ID:          "pr-opener",
			Name:        "PR Pioneer",
			Description: "Opened your first pull request",
			Icon:        "fa-code-pull-request",
			Condition:   AchievementCondition{Type: "pr_opened_count", Threshold: 1},
		},
		{
			ID:          "pr-10",
			Name:        "Pull Request Pro",
			Description: "Opened 10 pull requests",
			Icon:        "fa-code-branch",
			Condition:   AchievementCondition{Type: "pr_opened_count", Threshold: 10},
		},
		{
			ID:          "pr-50",
			Name:        "Merge Master",
			Description: "Opened 50 pull requests",
			Icon:        "fa-code-merge",
			Condition:   AchievementCondition{Type: "pr_opened_count", Threshold: 50},
		},
		{
			ID:          "reviewer",
			Name:        "Code Reviewer",
			Description: "Reviewed your first pull request",
			Icon:        "fa-magnifying-glass-chart",
			Condition:   AchievementCondition{Type: "review_count", Threshold: 1},
		},
		{
			ID:          "reviewer-25",
			Name:        "Review Regular",
			Description: "Reviewed 25 pull requests",
			Icon:        "fa-eye",
			Condition:   AchievementCondition{Type: "review_count", Threshold: 25},
		},
		{
			ID:          "reviewer-100",
			Name:        "Review Guru",
			Description: "Reviewed 100 pull requests",
			Icon:        "fa-user-graduate",
			Condition:   AchievementCondition{Type: "review_count", Threshold: 100},
		},
		{
			ID:          "speed-demon",
			Name:        "Speed Demon",
			Description: "Average review response under 1 hour",
			Icon:        "fa-bolt",
			Condition:   AchievementCondition{Type: "avg_review_time_hours", Threshold: 1},
		},
		{
			ID:          "quick-responder",
			Name:        "Quick Responder",
			Description: "Average review response under 4 hours",
			Icon:        "fa-clock",
			Condition:   AchievementCondition{Type: "avg_review_time_hours", Threshold: 4},
		},
		{
			ID:          "commentator",
			Name:        "Commentator",
			Description: "Left 50 PR review comments",
			Icon:        "fa-comments",
			Condition:   AchievementCondition{Type: "comment_count", Threshold: 50},
		},
		{
			ID:          "lines-1000",
			Name:        "Thousand Lines",
			Description: "Added 1000 lines of code",
			Icon:        "fa-layer-group",
			Condition:   AchievementCondition{Type: "lines_added", Threshold: 1000},
		},
		{
			ID:          "lines-10000",
			Name:        "Ten Thousand",
			Description: "Added 10000 lines of code",
			Icon:        "fa-mountain",
			Condition:   AchievementCondition{Type: "lines_added", Threshold: 10000},
		},
		{
			ID:          "cleaner",
			Name:        "Code Cleaner",
			Description: "Deleted 1000 lines of code",
			Icon:        "fa-broom",
			Condition:   AchievementCondition{Type: "lines_deleted", Threshold: 1000},
		},
		{
			ID:          "refactorer",
			Name:        "Refactoring Champion",
			Description: "Deleted 10000 lines of code",
			Icon:        "fa-recycle",
			Condition:   AchievementCondition{Type: "lines_deleted", Threshold: 10000},
		},
		{
			ID:          "multi-repo",
			Name:        "Multi-Repo Master",
			Description: "Contributed to 5 repositories",
			Icon:        "fa-folder-tree",
			Condition:   AchievementCondition{Type: "repo_count", Threshold: 5},
		},
		{
			ID:          "team-player",
			Name:        "Team Player",
			Description: "Reviewed PRs from 10 different contributors",
			Icon:        "fa-people-group",
			Condition:   AchievementCondition{Type: "unique_reviewees", Threshold: 10},
		},
		// PR Quality achievements
		{
			ID:          "big-pr",
			Name:        "Heavy Lifter",
			Description: "Merged a PR with 1000+ lines changed",
			Icon:        "fa-weight-hanging",
			Condition:   AchievementCondition{Type: "largest_pr_size", Threshold: 1000},
		},
		{
			ID:          "mega-pr",
			Name:        "Mega Merge",
			Description: "Merged a PR with 5000+ lines changed",
			Icon:        "fa-dumbbell",
			Condition:   AchievementCondition{Type: "largest_pr_size", Threshold: 5000},
		},
		{
			ID:          "small-pr-10",
			Name:        "Small PR Advocate",
			Description: "Merged 10 PRs under 100 lines",
			Icon:        "fa-compress",
			Condition:   AchievementCondition{Type: "small_pr_count", Threshold: 10},
		},
		{
			ID:          "small-pr-50",
			Name:        "Atomic Commits Hero",
			Description: "Merged 50 PRs under 100 lines",
			Icon:        "fa-atom",
			Condition:   AchievementCondition{Type: "small_pr_count", Threshold: 50},
		},
		{
			ID:          "perfect-pr-5",
			Name:        "Clean Code",
			Description: "5 PRs merged without changes requested",
			Icon:        "fa-check-double",
			Condition:   AchievementCondition{Type: "perfect_prs", Threshold: 5},
		},
		{
			ID:          "perfect-pr-25",
			Name:        "Flawless",
			Description: "25 PRs merged without changes requested",
			Icon:        "fa-gem",
			Condition:   AchievementCondition{Type: "perfect_prs", Threshold: 25},
		},
		// Activity pattern achievements
		{
			ID:          "streak-7",
			Name:        "Week Warrior",
			Description: "7 day contribution streak",
			Icon:        "fa-calendar-week",
			Condition:   AchievementCondition{Type: "longest_streak", Threshold: 7},
		},
		{
			ID:          "streak-30",
			Name:        "Month Master",
			Description: "30 day contribution streak",
			Icon:        "fa-calendar-check",
			Condition:   AchievementCondition{Type: "longest_streak", Threshold: 30},
		},
		{
			ID:          "early-bird",
			Name:        "Early Bird",
			Description: "50 commits before 9am",
			Icon:        "fa-sun",
			Condition:   AchievementCondition{Type: "early_bird_count", Threshold: 50},
		},
		{
			ID:          "night-owl",
			Name:        "Night Owl",
			Description: "50 commits after 9pm",
			Icon:        "fa-moon",
			Condition:   AchievementCondition{Type: "night_owl_count", Threshold: 50},
		},
		{
			ID:          "nosferatu",
			Name:        "Nosferatu",
			Description: "25 commits between midnight and 4am",
			Icon:        "fa-skull",
			Condition:   AchievementCondition{Type: "midnight_count", Threshold: 25},
		},
		{
			ID:          "weekend-warrior",
			Name:        "Weekend Warrior",
			Description: "25 weekend commits",
			Icon:        "fa-couch",
			Condition:   AchievementCondition{Type: "weekend_warrior", Threshold: 25},
		},
		{
			ID:          "active-30",
			Name:        "Consistent Contributor",
			Description: "Active on 30 different days",
			Icon:        "fa-chart-line",
			Condition:   AchievementCondition{Type: "active_days", Threshold: 30},
		},
		{
			ID:          "active-100",
			Name:        "Dedicated Developer",
			Description: "Active on 100 different days",
			Icon:        "fa-fire-flame-curved",
			Condition:   AchievementCondition{Type: "active_days", Threshold: 100},
		},
	}
}
