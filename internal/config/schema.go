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
	Enabled bool         `yaml:"enabled"`
	Points  PointsConfig `yaml:"points"`
}

// GetAchievements returns the hardcoded achievements (not configurable to prevent manipulation)
func (s *ScoringConfig) GetAchievements() []AchievementConfig {
	return defaultAchievements()
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
	OutOfHours      int     `yaml:"out_of_hours"` // Bonus per commit outside 9am-5pm
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
	ConcurrentRequests    int         `yaml:"concurrent_requests"`
	IncludeBots           bool        `yaml:"include_bots"`
	AdditionalBotPatterns []string    `yaml:"additional_bot_patterns"` // User-defined patterns (added to hardcoded defaults)
	CloneDirectory        string      `yaml:"clone_directory"`         // Directory for local git clones
	UseLocalGit           bool        `yaml:"use_local_git"`           // Use local git for commits (faster)
	UserAliases           []UserAlias `yaml:"user_aliases,omitempty"`  // Manual email/name to login mappings
}

// DefaultBotPatterns returns the hardcoded bot patterns that are always applied
// These cannot be overridden by users to ensure consistent bot filtering
func DefaultBotPatterns() []string {
	return []string{
		"*[bot]",            // GitHub App bots: dependabot[bot], renovate[bot], etc.
		"dependabot*",       // Dependabot variants
		"renovate*",         // Renovate bot variants
		"github-actions*",   // GitHub Actions
		"codecov*",          // Codecov bot
		"snyk*",             // Snyk security bot
		"greenkeeper*",      // Greenkeeper (legacy)
		"imgbot*",           // Image optimization bot
		"allcontributors*",  // All Contributors bot
		"semantic-release*", // Semantic release bot
	}
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
				OutOfHours:      2,
			},
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
			ConcurrentRequests:    5,
			IncludeBots:           false,
			AdditionalBotPatterns: []string{}, // Users can add custom patterns here
			CloneDirectory:        "./.repos",
			UseLocalGit:           true, // Default to faster local git analysis
		},
	}
}

// defaultAchievements returns the hardcoded achievement badges with proper tiers
// Achievements are not user-configurable to prevent manipulation
func defaultAchievements() []AchievementConfig {
	return []AchievementConfig{
		// ===== COMMIT COUNT (Tiers: 1, 10, 50, 100, 500, 1000) =====
		{ID: "commit-1", Name: "First Steps", Description: "Made your first commit", Icon: "fa-baby", Condition: AchievementCondition{Type: "commit_count", Threshold: 1}},
		{ID: "commit-10", Name: "Getting Started", Description: "Made 10 commits", Icon: "fa-seedling", Condition: AchievementCondition{Type: "commit_count", Threshold: 10}},
		{ID: "commit-50", Name: "Contributor", Description: "Made 50 commits", Icon: "fa-code", Condition: AchievementCondition{Type: "commit_count", Threshold: 50}},
		{ID: "commit-100", Name: "Committed", Description: "Made 100 commits", Icon: "fa-fire", Condition: AchievementCondition{Type: "commit_count", Threshold: 100}},
		{ID: "commit-500", Name: "Code Machine", Description: "Made 500 commits", Icon: "fa-robot", Condition: AchievementCondition{Type: "commit_count", Threshold: 500}},
		{ID: "commit-1000", Name: "Code Warrior", Description: "Made 1000 commits", Icon: "fa-crown", Condition: AchievementCondition{Type: "commit_count", Threshold: 1000}},

		// ===== PR OPENED (Tiers: 1, 10, 25, 50, 100, 250) =====
		{ID: "pr-1", Name: "PR Pioneer", Description: "Opened your first pull request", Icon: "fa-code-pull-request", Condition: AchievementCondition{Type: "pr_opened_count", Threshold: 1}},
		{ID: "pr-10", Name: "PR Regular", Description: "Opened 10 pull requests", Icon: "fa-code-branch", Condition: AchievementCondition{Type: "pr_opened_count", Threshold: 10}},
		{ID: "pr-25", Name: "PR Pro", Description: "Opened 25 pull requests", Icon: "fa-code-compare", Condition: AchievementCondition{Type: "pr_opened_count", Threshold: 25}},
		{ID: "pr-50", Name: "Merge Master", Description: "Opened 50 pull requests", Icon: "fa-code-merge", Condition: AchievementCondition{Type: "pr_opened_count", Threshold: 50}},
		{ID: "pr-100", Name: "PR Champion", Description: "Opened 100 pull requests", Icon: "fa-trophy", Condition: AchievementCondition{Type: "pr_opened_count", Threshold: 100}},
		{ID: "pr-250", Name: "PR Legend", Description: "Opened 250 pull requests", Icon: "fa-medal", Condition: AchievementCondition{Type: "pr_opened_count", Threshold: 250}},

		// ===== REVIEWS (Tiers: 1, 10, 25, 50, 100, 250) =====
		{ID: "review-1", Name: "First Review", Description: "Reviewed your first pull request", Icon: "fa-magnifying-glass", Condition: AchievementCondition{Type: "review_count", Threshold: 1}},
		{ID: "review-10", Name: "Reviewer", Description: "Reviewed 10 pull requests", Icon: "fa-eye", Condition: AchievementCondition{Type: "review_count", Threshold: 10}},
		{ID: "review-25", Name: "Review Regular", Description: "Reviewed 25 pull requests", Icon: "fa-glasses", Condition: AchievementCondition{Type: "review_count", Threshold: 25}},
		{ID: "review-50", Name: "Review Expert", Description: "Reviewed 50 pull requests", Icon: "fa-user-check", Condition: AchievementCondition{Type: "review_count", Threshold: 50}},
		{ID: "review-100", Name: "Review Guru", Description: "Reviewed 100 pull requests", Icon: "fa-user-graduate", Condition: AchievementCondition{Type: "review_count", Threshold: 100}},
		{ID: "review-250", Name: "Review Master", Description: "Reviewed 250 pull requests", Icon: "fa-award", Condition: AchievementCondition{Type: "review_count", Threshold: 250}},

		// ===== REVIEW COMMENTS (Tiers: 10, 50, 100, 250, 500) =====
		{ID: "comment-10", Name: "Commentator", Description: "Left 10 PR review comments", Icon: "fa-comment", Condition: AchievementCondition{Type: "comment_count", Threshold: 10}},
		{ID: "comment-50", Name: "Feedback Giver", Description: "Left 50 PR review comments", Icon: "fa-comments", Condition: AchievementCondition{Type: "comment_count", Threshold: 50}},
		{ID: "comment-100", Name: "Code Critic", Description: "Left 100 PR review comments", Icon: "fa-comment-dots", Condition: AchievementCondition{Type: "comment_count", Threshold: 100}},
		{ID: "comment-250", Name: "Feedback Expert", Description: "Left 250 PR review comments", Icon: "fa-message", Condition: AchievementCondition{Type: "comment_count", Threshold: 250}},
		{ID: "comment-500", Name: "Comment Champion", Description: "Left 500 PR review comments", Icon: "fa-scroll", Condition: AchievementCondition{Type: "comment_count", Threshold: 500}},

		// ===== LINES ADDED (Tiers: 100, 1000, 5000, 10000, 50000) =====
		{ID: "lines-added-100", Name: "First Hundred", Description: "Added 100 lines of code", Icon: "fa-plus", Condition: AchievementCondition{Type: "lines_added", Threshold: 100}},
		{ID: "lines-added-1000", Name: "Thousand Lines", Description: "Added 1000 lines of code", Icon: "fa-layer-group", Condition: AchievementCondition{Type: "lines_added", Threshold: 1000}},
		{ID: "lines-added-5000", Name: "Five Thousand", Description: "Added 5000 lines of code", Icon: "fa-cubes", Condition: AchievementCondition{Type: "lines_added", Threshold: 5000}},
		{ID: "lines-added-10000", Name: "Ten Thousand", Description: "Added 10000 lines of code", Icon: "fa-mountain", Condition: AchievementCondition{Type: "lines_added", Threshold: 10000}},
		{ID: "lines-added-50000", Name: "Code Mountain", Description: "Added 50000 lines of code", Icon: "fa-mountain-sun", Condition: AchievementCondition{Type: "lines_added", Threshold: 50000}},

		// ===== LINES DELETED (Tiers: 100, 500, 1000, 5000, 10000) =====
		{ID: "lines-deleted-100", Name: "Tidying Up", Description: "Deleted 100 lines of code", Icon: "fa-eraser", Condition: AchievementCondition{Type: "lines_deleted", Threshold: 100}},
		{ID: "lines-deleted-500", Name: "Spring Cleaning", Description: "Deleted 500 lines of code", Icon: "fa-broom", Condition: AchievementCondition{Type: "lines_deleted", Threshold: 500}},
		{ID: "lines-deleted-1000", Name: "Code Cleaner", Description: "Deleted 1000 lines of code", Icon: "fa-trash-can", Condition: AchievementCondition{Type: "lines_deleted", Threshold: 1000}},
		{ID: "lines-deleted-5000", Name: "Refactoring Hero", Description: "Deleted 5000 lines of code", Icon: "fa-recycle", Condition: AchievementCondition{Type: "lines_deleted", Threshold: 5000}},
		{ID: "lines-deleted-10000", Name: "Deletion Master", Description: "Deleted 10000 lines of code", Icon: "fa-dumpster-fire", Condition: AchievementCondition{Type: "lines_deleted", Threshold: 10000}},

		// ===== REVIEW RESPONSE TIME (Tiers: 24h, 4h, 1h - lower is better) =====
		{ID: "review-time-24h", Name: "Same Day Reviewer", Description: "Average review response under 24 hours", Icon: "fa-clock", Condition: AchievementCondition{Type: "avg_review_time_hours", Threshold: 24}},
		{ID: "review-time-4h", Name: "Quick Responder", Description: "Average review response under 4 hours", Icon: "fa-stopwatch", Condition: AchievementCondition{Type: "avg_review_time_hours", Threshold: 4}},
		{ID: "review-time-1h", Name: "Speed Demon", Description: "Average review response under 1 hour", Icon: "fa-bolt", Condition: AchievementCondition{Type: "avg_review_time_hours", Threshold: 1}},

		// ===== MULTI-REPO (Tiers: 2, 5, 10) =====
		{ID: "repo-2", Name: "Multi-Repo", Description: "Contributed to 2 repositories", Icon: "fa-folder", Condition: AchievementCondition{Type: "repo_count", Threshold: 2}},
		{ID: "repo-5", Name: "Repo Explorer", Description: "Contributed to 5 repositories", Icon: "fa-folder-tree", Condition: AchievementCondition{Type: "repo_count", Threshold: 5}},
		{ID: "repo-10", Name: "Repo Master", Description: "Contributed to 10 repositories", Icon: "fa-network-wired", Condition: AchievementCondition{Type: "repo_count", Threshold: 10}},

		// ===== UNIQUE REVIEWEES (Tiers: 3, 10, 25) =====
		{ID: "reviewees-3", Name: "Helpful Colleague", Description: "Reviewed PRs from 3 different contributors", Icon: "fa-user-group", Condition: AchievementCondition{Type: "unique_reviewees", Threshold: 3}},
		{ID: "reviewees-10", Name: "Team Player", Description: "Reviewed PRs from 10 different contributors", Icon: "fa-people-group", Condition: AchievementCondition{Type: "unique_reviewees", Threshold: 10}},
		{ID: "reviewees-25", Name: "Community Pillar", Description: "Reviewed PRs from 25 different contributors", Icon: "fa-people-roof", Condition: AchievementCondition{Type: "unique_reviewees", Threshold: 25}},

		// ===== PR SIZE - LARGE (Tiers: 500, 1000, 5000) =====
		{ID: "large-pr-500", Name: "Big Change", Description: "Merged a PR with 500+ lines changed", Icon: "fa-expand", Condition: AchievementCondition{Type: "largest_pr_size", Threshold: 500}},
		{ID: "large-pr-1000", Name: "Heavy Lifter", Description: "Merged a PR with 1000+ lines changed", Icon: "fa-weight-hanging", Condition: AchievementCondition{Type: "largest_pr_size", Threshold: 1000}},
		{ID: "large-pr-5000", Name: "Mega Merge", Description: "Merged a PR with 5000+ lines changed", Icon: "fa-dumbbell", Condition: AchievementCondition{Type: "largest_pr_size", Threshold: 5000}},

		// ===== SMALL PRs (Tiers: 5, 10, 25, 50) =====
		{ID: "small-pr-5", Name: "Small Changes", Description: "Merged 5 PRs under 100 lines", Icon: "fa-compress", Condition: AchievementCondition{Type: "small_pr_count", Threshold: 5}},
		{ID: "small-pr-10", Name: "Small PR Advocate", Description: "Merged 10 PRs under 100 lines", Icon: "fa-minimize", Condition: AchievementCondition{Type: "small_pr_count", Threshold: 10}},
		{ID: "small-pr-25", Name: "Atomic Commits", Description: "Merged 25 PRs under 100 lines", Icon: "fa-atom", Condition: AchievementCondition{Type: "small_pr_count", Threshold: 25}},
		{ID: "small-pr-50", Name: "Micro PR Master", Description: "Merged 50 PRs under 100 lines", Icon: "fa-microchip", Condition: AchievementCondition{Type: "small_pr_count", Threshold: 50}},

		// ===== PERFECT PRs (Tiers: 1, 5, 10, 25) =====
		{ID: "perfect-pr-1", Name: "First Try", Description: "1 PR merged without changes requested", Icon: "fa-check", Condition: AchievementCondition{Type: "perfect_prs", Threshold: 1}},
		{ID: "perfect-pr-5", Name: "Clean Code", Description: "5 PRs merged without changes requested", Icon: "fa-check-double", Condition: AchievementCondition{Type: "perfect_prs", Threshold: 5}},
		{ID: "perfect-pr-10", Name: "Quality Author", Description: "10 PRs merged without changes requested", Icon: "fa-circle-check", Condition: AchievementCondition{Type: "perfect_prs", Threshold: 10}},
		{ID: "perfect-pr-25", Name: "Flawless", Description: "25 PRs merged without changes requested", Icon: "fa-gem", Condition: AchievementCondition{Type: "perfect_prs", Threshold: 25}},

		// ===== ACTIVE DAYS (Tiers: 7, 30, 60, 100) =====
		{ID: "active-7", Name: "Week Active", Description: "Active on 7 different days", Icon: "fa-calendar-day", Condition: AchievementCondition{Type: "active_days", Threshold: 7}},
		{ID: "active-30", Name: "Month Active", Description: "Active on 30 different days", Icon: "fa-calendar-week", Condition: AchievementCondition{Type: "active_days", Threshold: 30}},
		{ID: "active-60", Name: "Consistent Contributor", Description: "Active on 60 different days", Icon: "fa-chart-line", Condition: AchievementCondition{Type: "active_days", Threshold: 60}},
		{ID: "active-100", Name: "Dedicated Developer", Description: "Active on 100 different days", Icon: "fa-fire-flame-curved", Condition: AchievementCondition{Type: "active_days", Threshold: 100}},

		// ===== LONGEST STREAK (Tiers: 3, 7, 14, 30) =====
		{ID: "streak-3", Name: "Getting Rolling", Description: "3 day contribution streak", Icon: "fa-forward", Condition: AchievementCondition{Type: "longest_streak", Threshold: 3}},
		{ID: "streak-7", Name: "Week Warrior", Description: "7 day contribution streak", Icon: "fa-calendar-week", Condition: AchievementCondition{Type: "longest_streak", Threshold: 7}},
		{ID: "streak-14", Name: "Two Week Streak", Description: "14 day contribution streak", Icon: "fa-fire", Condition: AchievementCondition{Type: "longest_streak", Threshold: 14}},
		{ID: "streak-30", Name: "Month Master", Description: "30 day contribution streak", Icon: "fa-calendar-check", Condition: AchievementCondition{Type: "longest_streak", Threshold: 30}},

		// ===== WORK WEEK STREAK (Tiers: 3, 5, 10, 20) =====
		{ID: "workweek-3", Name: "Work Week Start", Description: "3 consecutive weekday streak", Icon: "fa-briefcase", Condition: AchievementCondition{Type: "work_week_streak", Threshold: 3}},
		{ID: "workweek-5", Name: "Full Work Week", Description: "5 consecutive weekday streak", Icon: "fa-building", Condition: AchievementCondition{Type: "work_week_streak", Threshold: 5}},
		{ID: "workweek-10", Name: "Two Week Grind", Description: "10 consecutive weekday streak", Icon: "fa-business-time", Condition: AchievementCondition{Type: "work_week_streak", Threshold: 10}},
		{ID: "workweek-20", Name: "Month of Mondays", Description: "20 consecutive weekday streak", Icon: "fa-landmark", Condition: AchievementCondition{Type: "work_week_streak", Threshold: 20}},

		// ===== EARLY BIRD (Tiers: 10, 25, 50, 100) =====
		{ID: "earlybird-10", Name: "Early Riser", Description: "10 commits before 9am", Icon: "fa-mug-hot", Condition: AchievementCondition{Type: "early_bird_count", Threshold: 10}},
		{ID: "earlybird-25", Name: "Morning Person", Description: "25 commits before 9am", Icon: "fa-cloud-sun", Condition: AchievementCondition{Type: "early_bird_count", Threshold: 25}},
		{ID: "earlybird-50", Name: "Early Bird", Description: "50 commits before 9am", Icon: "fa-sun", Condition: AchievementCondition{Type: "early_bird_count", Threshold: 50}},
		{ID: "earlybird-100", Name: "Dawn Warrior", Description: "100 commits before 9am", Icon: "fa-sunrise", Condition: AchievementCondition{Type: "early_bird_count", Threshold: 100}},

		// ===== NIGHT OWL (Tiers: 10, 25, 50, 100) =====
		{ID: "nightowl-10", Name: "Late Worker", Description: "10 commits after 9pm", Icon: "fa-cloud-moon", Condition: AchievementCondition{Type: "night_owl_count", Threshold: 10}},
		{ID: "nightowl-25", Name: "Evening Coder", Description: "25 commits after 9pm", Icon: "fa-moon", Condition: AchievementCondition{Type: "night_owl_count", Threshold: 25}},
		{ID: "nightowl-50", Name: "Night Owl", Description: "50 commits after 9pm", Icon: "fa-star", Condition: AchievementCondition{Type: "night_owl_count", Threshold: 50}},
		{ID: "nightowl-100", Name: "Nocturnal", Description: "100 commits after 9pm", Icon: "fa-star-and-crescent", Condition: AchievementCondition{Type: "night_owl_count", Threshold: 100}},

		// ===== MIDNIGHT CODER (Tiers: 5, 10, 25, 50) =====
		{ID: "midnight-5", Name: "Night Shift", Description: "5 commits between midnight and 4am", Icon: "fa-ghost", Condition: AchievementCondition{Type: "midnight_count", Threshold: 5}},
		{ID: "midnight-10", Name: "Insomniac", Description: "10 commits between midnight and 4am", Icon: "fa-bed", Condition: AchievementCondition{Type: "midnight_count", Threshold: 10}},
		{ID: "midnight-25", Name: "Nosferatu", Description: "25 commits between midnight and 4am", Icon: "fa-skull", Condition: AchievementCondition{Type: "midnight_count", Threshold: 25}},
		{ID: "midnight-50", Name: "Vampire Coder", Description: "50 commits between midnight and 4am", Icon: "fa-skull-crossbones", Condition: AchievementCondition{Type: "midnight_count", Threshold: 50}},

		// ===== WEEKEND WARRIOR (Tiers: 5, 10, 25, 50) =====
		{ID: "weekend-5", Name: "Weekend Work", Description: "5 weekend commits", Icon: "fa-couch", Condition: AchievementCondition{Type: "weekend_warrior", Threshold: 5}},
		{ID: "weekend-10", Name: "Weekend Regular", Description: "10 weekend commits", Icon: "fa-house-laptop", Condition: AchievementCondition{Type: "weekend_warrior", Threshold: 10}},
		{ID: "weekend-25", Name: "Weekend Warrior", Description: "25 weekend commits", Icon: "fa-gamepad", Condition: AchievementCondition{Type: "weekend_warrior", Threshold: 25}},
		{ID: "weekend-50", Name: "No Days Off", Description: "50 weekend commits", Icon: "fa-person-running", Condition: AchievementCondition{Type: "weekend_warrior", Threshold: 50}},

		// ===== OUT OF HOURS (Tiers: 10, 25, 50, 100) =====
		{ID: "ooh-10", Name: "Extra Hours", Description: "10 commits outside 9am-5pm", Icon: "fa-clock-rotate-left", Condition: AchievementCondition{Type: "out_of_hours_count", Threshold: 10}},
		{ID: "ooh-25", Name: "Flexible Schedule", Description: "25 commits outside 9am-5pm", Icon: "fa-user-clock", Condition: AchievementCondition{Type: "out_of_hours_count", Threshold: 25}},
		{ID: "ooh-50", Name: "Off-Hours Hero", Description: "50 commits outside 9am-5pm", Icon: "fa-hourglass-half", Condition: AchievementCondition{Type: "out_of_hours_count", Threshold: 50}},
		{ID: "ooh-100", Name: "Time Bender", Description: "100 commits outside 9am-5pm", Icon: "fa-infinity", Condition: AchievementCondition{Type: "out_of_hours_count", Threshold: 100}},
	}
}
