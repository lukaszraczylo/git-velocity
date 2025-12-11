package models

import "time"

// Period represents a time period for metrics aggregation
type Period struct {
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Granularity string    `json:"granularity"` // daily, weekly, monthly, custom
	Label       string    `json:"label"`       // e.g., "Week 42", "December 2024", "Q1 2024"
}

// ContributorMetrics holds aggregated metrics for a single contributor
type ContributorMetrics struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Period    Period `json:"period"`

	// Commit metrics
	CommitCount  int `json:"commit_count"`
	LinesAdded   int `json:"lines_added"`
	LinesDeleted int `json:"lines_deleted"`
	FilesChanged int `json:"files_changed"`

	// Meaningful line counts (excludes comments and whitespace)
	MeaningfulLinesAdded   int `json:"meaningful_lines_added"`
	MeaningfulLinesDeleted int `json:"meaningful_lines_deleted"`

	// Comment and documentation line counts
	CommentLinesAdded   int `json:"comment_lines_added"`
	CommentLinesDeleted int `json:"comment_lines_deleted"`

	// PR metrics
	PRsOpened      int     `json:"prs_opened"`
	PRsMerged      int     `json:"prs_merged"`
	PRsClosed      int     `json:"prs_closed"`
	AvgPRSize      float64 `json:"avg_pr_size"`
	AvgTimeToMerge float64 `json:"avg_time_to_merge_hours"`
	LargestPRSize  int     `json:"largest_pr_size"` // Biggest single PR by lines changed
	SmallPRCount   int     `json:"small_pr_count"`  // PRs under 100 lines (good practice)
	PerfectPRs     int     `json:"perfect_prs"`     // PRs merged without changes requested

	// Review metrics
	ReviewsGiven     int     `json:"reviews_given"`
	ReviewComments   int     `json:"review_comments"`
	ApprovalsGiven   int     `json:"approvals_given"`
	ChangesRequested int     `json:"changes_requested"`
	AvgReviewTime    float64 `json:"avg_review_time_hours"`

	// Issue metrics
	IssuesOpened  int `json:"issues_opened"`
	IssuesClosed  int `json:"issues_closed"`
	IssueComments int `json:"issue_comments"`

	// Activity patterns
	ActiveDays      int `json:"active_days"`        // Unique days with activity
	CurrentStreak   int `json:"current_streak"`     // Current consecutive days
	LongestStreak   int `json:"longest_streak"`     // Longest consecutive days
	WorkWeekStreak  int `json:"work_week_streak"`   // Longest consecutive weekdays (Mon-Fri, weekends don't break streak)
	EarlyBirdCount  int `json:"early_bird_count"`   // Commits before 9am
	NightOwlCount   int `json:"night_owl_count"`    // Commits after 9pm
	MidnightCount   int `json:"midnight_count"`     // Commits between midnight and 4am
	WeekendWarrior  int `json:"weekend_warrior"`    // Weekend commits
	OutOfHoursCount int `json:"out_of_hours_count"` // Commits outside 9am-5pm

	// Repository participation
	RepositoriesContributed []string `json:"repositories_contributed,omitempty"`
	UniqueReviewees         int      `json:"unique_reviewees"`

	// Scoring
	Score        Score    `json:"score"`
	Achievements []string `json:"achievements"` // Achievement IDs
}

// Score holds the calculated score and breakdown
type Score struct {
	Total          int            `json:"total"`
	Breakdown      ScoreBreakdown `json:"breakdown"`
	Rank           int            `json:"rank"`
	PercentileRank float64        `json:"percentile_rank"`
}

// ScoreBreakdown shows how the score was calculated
type ScoreBreakdown struct {
	Commits       int `json:"commits"`
	PRs           int `json:"prs"`
	Reviews       int `json:"reviews"`
	Comments      int `json:"comments"` // PR review comments (not code comments)
	ResponseBonus int `json:"response_bonus"`
	LineChanges   int `json:"line_changes"`
	OutOfHours    int `json:"out_of_hours"` // Bonus for out-of-hours commits
}

// RepositoryMetrics holds aggregated metrics for a single repository
type RepositoryMetrics struct {
	Owner              string               `json:"owner"`
	Name               string               `json:"name"`
	FullName           string               `json:"full_name"` // owner/name
	Period             Period               `json:"period"`
	Contributors       []ContributorMetrics `json:"contributors"`
	TotalCommits       int                  `json:"total_commits"`
	TotalPRs           int                  `json:"total_prs"`
	TotalReviews       int                  `json:"total_reviews"`
	ActiveContributors int                  `json:"active_contributors"`
	TotalLinesAdded    int                  `json:"total_lines_added"`
	TotalLinesDeleted  int                  `json:"total_lines_deleted"`

	// Meaningful line counts (excludes comments and whitespace)
	TotalMeaningfulLinesAdded   int `json:"total_meaningful_lines_added"`
	TotalMeaningfulLinesDeleted int `json:"total_meaningful_lines_deleted"`
}

// TeamMetrics holds aggregated metrics for a team
type TeamMetrics struct {
	Name              string               `json:"name"`
	Color             string               `json:"color"`
	Members           []string             `json:"members"`
	Period            Period               `json:"period"`
	AggregatedMetrics ContributorMetrics   `json:"aggregated_metrics"`
	MemberMetrics     []ContributorMetrics `json:"member_metrics"`
	TotalScore        int                  `json:"total_score"`
	AvgScore          float64              `json:"avg_score"`
}

// GlobalMetrics holds metrics aggregated across all repositories
type GlobalMetrics struct {
	Period       Period               `json:"period"`
	Repositories []RepositoryMetrics  `json:"repositories"`
	Contributors []ContributorMetrics `json:"contributors"` // Aggregated across all repos
	Teams        []TeamMetrics        `json:"teams"`
	Leaderboard  []LeaderboardEntry   `json:"leaderboard"`
	TopAchievers map[string]string    `json:"top_achievers"` // category -> login

	// Summary stats
	TotalContributors int `json:"total_contributors"`
	TotalCommits      int `json:"total_commits"`
	TotalPRs          int `json:"total_prs"`
	TotalReviews      int `json:"total_reviews"`
	TotalLinesAdded   int `json:"total_lines_added"`
	TotalLinesDeleted int `json:"total_lines_deleted"`

	// Meaningful line counts (excludes comments and whitespace)
	TotalMeaningfulLinesAdded   int `json:"total_meaningful_lines_added"`
	TotalMeaningfulLinesDeleted int `json:"total_meaningful_lines_deleted"`

	// Velocity timeline (weekly granularity)
	VelocityTimeline *VelocityTimeline `json:"velocity_timeline,omitempty"`
}

// VelocityTimeline holds weekly velocity data for trend visualization
type VelocityTimeline struct {
	Labels []string                 `json:"labels"` // Week labels (e.g., "Dec 2", "Dec 9")
	Series []VelocityTimelineSeries `json:"series"` // Data series (commits, PRs, reviews, score)
}

// VelocityTimelineSeries represents a single data series in the velocity timeline
type VelocityTimelineSeries struct {
	Name  string    `json:"name"`  // Series name (e.g., "Commits", "PRs", "Score")
	Color string    `json:"color"` // Series color
	Data  []float64 `json:"data"`  // Values for each week
}

// LeaderboardEntry represents a single entry in the leaderboard
type LeaderboardEntry struct {
	Rank         int      `json:"rank"`
	Login        string   `json:"login"`
	Name         string   `json:"name"`
	AvatarURL    string   `json:"avatar_url"`
	Score        int      `json:"score"`
	Team         string   `json:"team,omitempty"`
	TopCategory  string   `json:"top_category,omitempty"` // What they're best at
	Achievements []string `json:"achievements,omitempty"` // Achievement IDs earned
}

// TimeSeriesPoint represents a single data point in a time series
type TimeSeriesPoint struct {
	Date  time.Time `json:"date"`
	Label string    `json:"label"`
	Value float64   `json:"value"`
}

// TimeSeries represents a series of data points over time
type TimeSeries struct {
	Name   string            `json:"name"`
	Color  string            `json:"color,omitempty"`
	Points []TimeSeriesPoint `json:"points"`
}

// ChartData holds data formatted for charts
type ChartData struct {
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Type        string       `json:"type"` // line, bar, pie, doughnut
	Labels      []string     `json:"labels"`
	Series      []TimeSeries `json:"series"`
}

// DashboardData holds all data needed for the dashboard
type DashboardData struct {
	GeneratedAt   time.Time       `json:"generated_at"`
	Period        Period          `json:"period"`
	GlobalMetrics GlobalMetrics   `json:"global_metrics"`
	Charts        []ChartData     `json:"charts"`
	Achievements  []Achievement   `json:"achievements"`
	Configuration DashboardConfig `json:"configuration"`
}

// DashboardConfig holds UI configuration
type DashboardConfig struct {
	Title            string   `json:"title"`
	Description      string   `json:"description,omitempty"`
	Repositories     []string `json:"repositories"`
	Teams            []string `json:"teams,omitempty"`
	Granularities    []string `json:"granularities"`
	ScoringEnabled   bool     `json:"scoring_enabled"`
	ShowAchievements bool     `json:"show_achievements"`
}

// Achievement represents an earned achievement badge
type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	EarnedBy    string `json:"earned_by"` // Login of user who earned it
	EarnedAt    string `json:"earned_at"` // When it was earned (period label)
}
