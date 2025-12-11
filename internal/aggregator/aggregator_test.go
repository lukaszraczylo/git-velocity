package aggregator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukaszraczylo/git-velocity/internal/config"
	"github.com/lukaszraczylo/git-velocity/internal/domain/models"
)

func TestNew(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}
	agg := New(cfg)

	assert.NotNil(t, agg)
	assert.Equal(t, cfg, agg.config)
}

func TestAggregator_AggregateEmptyData(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	data := &models.RawData{}
	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.TotalContributors)
	assert.Equal(t, 0, metrics.TotalCommits)
	assert.Equal(t, 0, metrics.TotalPRs)
	assert.Equal(t, 0, metrics.TotalReviews)
}

func TestAggregator_AggregateCommits(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	data := &models.RawData{
		Commits: []models.Commit{
			{
				SHA:          "abc123",
				Message:      "Test commit",
				Author:       models.Author{Login: "user1", Name: "User One"},
				Date:         time.Now(),
				Additions:    100,
				Deletions:    50,
				FilesChanged: 5,
				Repository:   "owner/repo",
			},
			{
				SHA:          "def456",
				Message:      "Another commit",
				Author:       models.Author{Login: "user1", Name: "User One"},
				Date:         time.Now(),
				Additions:    200,
				Deletions:    75,
				FilesChanged: 3,
				Repository:   "owner/repo",
			},
			{
				SHA:          "ghi789",
				Message:      "User2 commit",
				Author:       models.Author{Login: "user2", Name: "User Two"},
				Date:         time.Now(),
				Additions:    50,
				Deletions:    25,
				FilesChanged: 2,
				Repository:   "owner/repo",
			},
		},
	}

	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	assert.Equal(t, 2, metrics.TotalContributors)
	assert.Equal(t, 3, metrics.TotalCommits)
	assert.Equal(t, 350, metrics.TotalLinesAdded)
	assert.Equal(t, 150, metrics.TotalLinesDeleted)

	// Check repository metrics
	require.Len(t, metrics.Repositories, 1)
	repo := metrics.Repositories[0]
	assert.Equal(t, "owner", repo.Owner)
	assert.Equal(t, "repo", repo.Name)
	assert.Equal(t, 3, repo.TotalCommits)
}

func TestAggregator_AggregatePullRequests(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	mergedAt := time.Now()
	data := &models.RawData{
		PullRequests: []models.PullRequest{
			{
				Number:     1,
				Title:      "PR 1",
				State:      models.PRStateMerged,
				Author:     models.Author{Login: "user1", Name: "User One"},
				Repository: "owner/repo",
				CreatedAt:  time.Now().Add(-time.Hour),
				MergedAt:   &mergedAt,
				Additions:  100,
				Deletions:  50,
			},
			{
				Number:     2,
				Title:      "PR 2",
				State:      models.PRStateOpen,
				Author:     models.Author{Login: "user2", Name: "User Two"},
				Repository: "owner/repo",
				CreatedAt:  time.Now().Add(-30 * time.Minute),
				Additions:  200,
				Deletions:  75,
			},
		},
	}

	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	assert.Equal(t, 2, metrics.TotalContributors)
	assert.Equal(t, 2, metrics.TotalPRs)

	// Check repository metrics
	require.Len(t, metrics.Repositories, 1)
	repo := metrics.Repositories[0]
	assert.Equal(t, 2, repo.TotalPRs)
}

func TestAggregator_AggregateReviews(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	data := &models.RawData{
		PullRequests: []models.PullRequest{
			{
				Number:     1,
				Title:      "PR 1",
				State:      models.PRStateOpen,
				Author:     models.Author{Login: "user1"},
				Repository: "owner/repo",
				CreatedAt:  time.Now(),
			},
		},
		Reviews: []models.Review{
			{
				ID:          1,
				PullRequest: 1,
				Repository:  "owner/repo",
				Author:      models.Author{Login: "reviewer1"},
				State:       models.ReviewApproved,
				SubmittedAt: time.Now(),
			},
			{
				ID:          2,
				PullRequest: 1,
				Repository:  "owner/repo",
				Author:      models.Author{Login: "reviewer2"},
				State:       models.ReviewChangesRequested,
				SubmittedAt: time.Now(),
			},
		},
	}

	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	assert.Equal(t, 3, metrics.TotalContributors) // user1, reviewer1, reviewer2
	assert.Equal(t, 2, metrics.TotalReviews)
}

func TestAggregator_AggregateIssues(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	closedAt := time.Now()
	data := &models.RawData{
		// Need a commit to create the repository
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Author:     models.Author{Login: "user1"},
				Repository: "owner/repo",
			},
		},
		Issues: []models.Issue{
			{
				Number:     1,
				Title:      "Issue 1",
				State:      models.IssueStateOpen,
				Author:     models.Author{Login: "user1"},
				Repository: "owner/repo",
				CreatedAt:  time.Now(),
			},
			{
				Number:     2,
				Title:      "Issue 2",
				State:      models.IssueStateClosed,
				Author:     models.Author{Login: "user1"},
				Repository: "owner/repo",
				CreatedAt:  time.Now().Add(-time.Hour),
				ClosedAt:   &closedAt,
				ClosedBy:   &models.Author{Login: "user1"},
			},
		},
	}

	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	assert.Equal(t, 1, metrics.TotalContributors)

	// Find user1 in repository contributors
	require.Len(t, metrics.Repositories, 1)
	repo := metrics.Repositories[0]
	require.Len(t, repo.Contributors, 1)
	assert.Equal(t, 2, repo.Contributors[0].IssuesOpened)
	assert.Equal(t, 1, repo.Contributors[0].IssuesClosed)
}

func TestAggregator_AggregateTeams(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Teams = []config.TeamConfig{
		{
			Name:    "Backend Team",
			Members: []string{"user1", "user2"},
			Color:   "#ff0000",
		},
	}
	agg := New(cfg)

	data := &models.RawData{
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Author:     models.Author{Login: "user1"},
				Repository: "owner/repo",
				Additions:  100,
				Deletions:  50,
			},
			{
				SHA:        "def456",
				Author:     models.Author{Login: "user2"},
				Repository: "owner/repo",
				Additions:  200,
				Deletions:  75,
			},
		},
	}

	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	require.Len(t, metrics.Teams, 1)
	team := metrics.Teams[0]
	assert.Equal(t, "Backend Team", team.Name)
	assert.Equal(t, "#ff0000", team.Color)
	assert.Len(t, team.MemberMetrics, 2)
	assert.Equal(t, 2, team.AggregatedMetrics.CommitCount)
}

func TestAggregator_DateRange(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	data := &models.RawData{}
	dateRange := &config.ParsedDateRange{
		Start: &start,
		End:   &end,
	}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	assert.Equal(t, start, metrics.Period.Start)
	assert.Equal(t, end, metrics.Period.End)
}

func TestAggregator_MultipleRepositories(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	data := &models.RawData{
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Author:     models.Author{Login: "user1"},
				Repository: "owner/repo1",
				Additions:  100,
			},
			{
				SHA:        "def456",
				Author:     models.Author{Login: "user1"},
				Repository: "owner/repo2",
				Additions:  200,
			},
			{
				SHA:        "ghi789",
				Author:     models.Author{Login: "user2"},
				Repository: "owner/repo1",
				Additions:  50,
			},
		},
	}

	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	assert.Equal(t, 2, metrics.TotalContributors)
	assert.Len(t, metrics.Repositories, 2)
}

func TestContains(t *testing.T) {
	t.Parallel()

	slice := []string{"a", "b", "c"}

	assert.True(t, contains(slice, "a"))
	assert.True(t, contains(slice, "b"))
	assert.True(t, contains(slice, "c"))
	assert.False(t, contains(slice, "d"))
	assert.False(t, contains([]string{}, "a"))
}

func TestParseRepoName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		fullName      string
		expectedOwner string
		expectedName  string
	}{
		{"owner/repo", "owner", "repo"},
		{"org/project-name", "org", "project-name"},
		{"user/repo-with-dashes", "user", "repo-with-dashes"},
		{"single", "single", ""},
	}

	for _, tt := range tests {
		owner, name := parseRepoName(tt.fullName)
		assert.Equal(t, tt.expectedOwner, owner, "owner mismatch for %s", tt.fullName)
		assert.Equal(t, tt.expectedName, name, "name mismatch for %s", tt.fullName)
	}
}

func TestSetUserProfiles(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	profiles := map[string]UserProfile{
		"user1": {Login: "user1", Email: "user1@example.com", Name: "User One", ID: 12345},
		"user2": {Login: "user2", Email: "user2@example.com", Name: "User Two", ID: 67890},
	}

	agg.SetUserProfiles(profiles)
	assert.Equal(t, profiles, agg.userProfiles)
}

func TestNormalizeForComparison(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"John Doe", "johndoe"},
		{"john-doe", "johndoe"},
		{"john_doe", "johndoe"},
		{"john.doe", "johndoe"},
		{"JOHN DOE", "johndoe"},
		{"John123Doe", "johndoe"},
		{"123", ""},
		{"", ""},
		{"ABC xyz 123", "abcxyz"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeForComparison(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildEmailToLoginMapping_NoReplyEmails(t *testing.T) {
	t.Parallel()

	data := &models.RawData{
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Author:     models.Author{Login: "", Email: "12345+johndoe@users.noreply.github.com", Name: "John Doe"},
				Repository: "owner/repo",
			},
		},
		PullRequests: []models.PullRequest{
			{
				Number: 1,
				Author: models.Author{Login: "johndoe", ID: 12345},
			},
		},
	}

	mapping := buildEmailToLoginMapping(data, nil)

	// Should map via the ID
	assert.Equal(t, "johndoe", mapping["12345+johndoe@users.noreply.github.com"])
}

func TestBuildEmailToLoginMapping_ProfileEmails(t *testing.T) {
	t.Parallel()

	data := &models.RawData{
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Author:     models.Author{Login: "", Email: "john@company.com", Name: "John Doe"},
				Repository: "owner/repo",
			},
		},
	}

	profiles := map[string]UserProfile{
		"johndoe": {Login: "johndoe", Email: "john@company.com", Name: "John Doe", ID: 12345},
	}

	mapping := buildEmailToLoginMapping(data, profiles)

	// Should map via profile email
	assert.Equal(t, "johndoe", mapping["john@company.com"])
}

func TestBuildEmailToLoginMapping_NameMatching(t *testing.T) {
	t.Parallel()

	data := &models.RawData{
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Author:     models.Author{Login: "", Email: "john@somewhere.com", Name: "John Doe"},
				Repository: "owner/repo",
			},
		},
		PullRequests: []models.PullRequest{
			{
				Number: 1,
				Author: models.Author{Login: "johndoe", Name: "John Doe"},
			},
		},
	}

	mapping := buildEmailToLoginMapping(data, nil)

	// Should map via name matching
	assert.Equal(t, "johndoe", mapping["john@somewhere.com"])
}

func TestCalculateWorkWeekStreak(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		dates          map[string]bool
		expectedStreak int
	}{
		{
			name:           "empty dates",
			dates:          map[string]bool{},
			expectedStreak: 0,
		},
		{
			name: "single weekday",
			dates: map[string]bool{
				"2024-01-08": true, // Monday
			},
			expectedStreak: 1,
		},
		{
			name: "consecutive weekdays",
			dates: map[string]bool{
				"2024-01-08": true, // Monday
				"2024-01-09": true, // Tuesday
				"2024-01-10": true, // Wednesday
			},
			expectedStreak: 3,
		},
		{
			name: "weekdays with weekend gap",
			dates: map[string]bool{
				"2024-01-12": true, // Friday
				"2024-01-15": true, // Monday
				"2024-01-16": true, // Tuesday
			},
			expectedStreak: 3, // Weekend doesn't break streak
		},
		{
			name: "broken streak on weekday",
			dates: map[string]bool{
				"2024-01-08": true, // Monday
				"2024-01-10": true, // Wednesday (skipped Tuesday)
			},
			expectedStreak: 1,
		},
		{
			name: "weekend only",
			dates: map[string]bool{
				"2024-01-13": true, // Saturday
				"2024-01-14": true, // Sunday
			},
			expectedStreak: 0, // Weekends don't count
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateWorkWeekStreak(tt.dates)
			assert.Equal(t, tt.expectedStreak, result)
		})
	}
}

func TestCalculateWorkWeekStreak_LongestStreak(t *testing.T) {
	t.Parallel()

	// Multiple streaks - should return longest
	dates := map[string]bool{
		"2024-01-08": true, // Monday
		"2024-01-09": true, // Tuesday
		"2024-01-15": true, // Monday (gap - breaks streak)
		"2024-01-16": true, // Tuesday
		"2024-01-17": true, // Wednesday
		"2024-01-18": true, // Thursday
		"2024-01-19": true, // Friday
		"2024-01-22": true, // Monday (weekend doesn't break)
	}

	result := calculateWorkWeekStreak(dates)
	assert.Equal(t, 6, result) // Mon-Fri + Mon = 6 weekdays in a row
}

func TestAggregator_OutOfHoursTracking(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	data := &models.RawData{
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Author:     models.Author{Login: "user1"},
				Date:       time.Date(2024, 1, 15, 7, 0, 0, 0, time.UTC), // 7am - before 9am
				Repository: "owner/repo",
			},
			{
				SHA:        "def456",
				Author:     models.Author{Login: "user1"},
				Date:       time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), // 10am - work hours
				Repository: "owner/repo",
			},
			{
				SHA:        "ghi789",
				Author:     models.Author{Login: "user1"},
				Date:       time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC), // 6pm - after 5pm
				Repository: "owner/repo",
			},
		},
	}

	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	require.Len(t, metrics.Repositories, 1)
	require.Len(t, metrics.Repositories[0].Contributors, 1)
	contrib := metrics.Repositories[0].Contributors[0]
	assert.Equal(t, 2, contrib.OutOfHoursCount) // 7am and 6pm are out of hours
}

func TestAggregator_WorkWeekStreakTracking(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	data := &models.RawData{
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Author:     models.Author{Login: "user1"},
				Date:       time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC), // Monday
				Repository: "owner/repo",
			},
			{
				SHA:        "def456",
				Author:     models.Author{Login: "user1"},
				Date:       time.Date(2024, 1, 9, 10, 0, 0, 0, time.UTC), // Tuesday
				Repository: "owner/repo",
			},
			{
				SHA:        "ghi789",
				Author:     models.Author{Login: "user1"},
				Date:       time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC), // Wednesday
				Repository: "owner/repo",
			},
		},
	}

	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	require.Len(t, metrics.Repositories, 1)
	require.Len(t, metrics.Repositories[0].Contributors, 1)
	contrib := metrics.Repositories[0].Contributors[0]
	assert.Equal(t, 3, contrib.WorkWeekStreak)
}

// Note: Bot filtering tests removed - bot filtering happens in app.go before data reaches aggregator
// The aggregator receives already filtered data

func TestAggregator_EarlyBirdTracking(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	data := &models.RawData{
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Author:     models.Author{Login: "user1"},
				Date:       time.Date(2024, 1, 15, 6, 0, 0, 0, time.UTC), // 6am
				Repository: "owner/repo",
			},
			{
				SHA:        "def456",
				Author:     models.Author{Login: "user1"},
				Date:       time.Date(2024, 1, 16, 8, 30, 0, 0, time.UTC), // 8:30am
				Repository: "owner/repo",
			},
		},
	}

	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	require.Len(t, metrics.Repositories, 1)
	require.Len(t, metrics.Repositories[0].Contributors, 1)
	contrib := metrics.Repositories[0].Contributors[0]
	assert.Equal(t, 2, contrib.EarlyBirdCount) // Both before 9am
}

func TestAggregator_NightOwlTracking(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	data := &models.RawData{
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Author:     models.Author{Login: "user1"},
				Date:       time.Date(2024, 1, 15, 21, 0, 0, 0, time.UTC), // 9pm
				Repository: "owner/repo",
			},
			{
				SHA:        "def456",
				Author:     models.Author{Login: "user1"},
				Date:       time.Date(2024, 1, 16, 23, 30, 0, 0, time.UTC), // 11:30pm
				Repository: "owner/repo",
			},
		},
	}

	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	require.Len(t, metrics.Repositories, 1)
	require.Len(t, metrics.Repositories[0].Contributors, 1)
	contrib := metrics.Repositories[0].Contributors[0]
	assert.Equal(t, 2, contrib.NightOwlCount) // Both after 9pm
}

func TestAggregator_MidnightTracking(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	data := &models.RawData{
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Author:     models.Author{Login: "user1"},
				Date:       time.Date(2024, 1, 15, 0, 30, 0, 0, time.UTC), // 12:30am
				Repository: "owner/repo",
			},
			{
				SHA:        "def456",
				Author:     models.Author{Login: "user1"},
				Date:       time.Date(2024, 1, 16, 3, 0, 0, 0, time.UTC), // 3am
				Repository: "owner/repo",
			},
		},
	}

	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	require.Len(t, metrics.Repositories, 1)
	require.Len(t, metrics.Repositories[0].Contributors, 1)
	contrib := metrics.Repositories[0].Contributors[0]
	assert.Equal(t, 2, contrib.MidnightCount) // Both between 0-4am
}

func TestAggregator_WeekendWarriorTracking(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	data := &models.RawData{
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Author:     models.Author{Login: "user1"},
				Date:       time.Date(2024, 1, 13, 10, 0, 0, 0, time.UTC), // Saturday
				Repository: "owner/repo",
			},
			{
				SHA:        "def456",
				Author:     models.Author{Login: "user1"},
				Date:       time.Date(2024, 1, 14, 15, 0, 0, 0, time.UTC), // Sunday
				Repository: "owner/repo",
			},
			{
				SHA:        "ghi789",
				Author:     models.Author{Login: "user1"},
				Date:       time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), // Monday (not weekend)
				Repository: "owner/repo",
			},
		},
	}

	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	require.Len(t, metrics.Repositories, 1)
	require.Len(t, metrics.Repositories[0].Contributors, 1)
	contrib := metrics.Repositories[0].Contributors[0]
	assert.Equal(t, 2, contrib.WeekendWarrior) // Saturday and Sunday only
}

func TestAggregator_MultiRepoContributions(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	data := &models.RawData{
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Author:     models.Author{Login: "user1"},
				Repository: "owner/repo1",
			},
			{
				SHA:        "def456",
				Author:     models.Author{Login: "user1"},
				Repository: "owner/repo2",
			},
			{
				SHA:        "ghi789",
				Author:     models.Author{Login: "user1"},
				Repository: "owner/repo3",
			},
		},
	}

	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	// MultiRepoCount is tracked in the global leaderboard entries, not repo contributors
	// The leaderboard entry should show 3 repos for user1
	require.Len(t, metrics.Repositories, 3)
	assert.Equal(t, 1, metrics.TotalContributors)
}

func TestBuildEmailToLoginMapping_EmptyData(t *testing.T) {
	t.Parallel()

	data := &models.RawData{}
	mapping := buildEmailToLoginMapping(data, nil)
	assert.Empty(t, mapping)
}

func TestBuildEmailToLoginMapping_NoReplyEmailWithoutID(t *testing.T) {
	t.Parallel()

	// When the email is just "username@users.noreply.github.com" (without ID+),
	// the mapping only happens if there's a matching PR author (via name matching later)
	// The direct extraction only works for "ID+username@" format
	data := &models.RawData{
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Author:     models.Author{Login: "", Email: "johndoe@users.noreply.github.com", Name: "John Doe"},
				Repository: "owner/repo",
			},
		},
		// Add a PR to enable name matching
		PullRequests: []models.PullRequest{
			{
				Number: 1,
				Author: models.Author{Login: "johndoe", Name: "John Doe"},
			},
		},
	}

	mapping := buildEmailToLoginMapping(data, nil)
	// Should map via name matching since there's a PR author with the same name
	assert.Equal(t, "johndoe", mapping["johndoe@users.noreply.github.com"])
}

func TestCountIssueReferences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		message  string
		expected int
	}{
		{
			name:     "no references",
			message:  "Just a regular commit message",
			expected: 0,
		},
		{
			name:     "fixes issue",
			message:  "fixes #123",
			expected: 1,
		},
		{
			name:     "Fixes issue uppercase",
			message:  "Fixes #456",
			expected: 1,
		},
		{
			name:     "closes issue",
			message:  "closes #789",
			expected: 1,
		},
		{
			name:     "resolves issue",
			message:  "resolves #101",
			expected: 1,
		},
		{
			name:     "refs issue",
			message:  "refs #202",
			expected: 1,
		},
		{
			name:     "ref issue",
			message:  "ref #303",
			expected: 1,
		},
		{
			name:     "multiple fixes",
			message:  "fixes #1, fixes #2, fixes #3",
			expected: 3,
		},
		{
			name:     "mixed keywords",
			message:  "fixes #1 and closes #2",
			expected: 2,
		},
		{
			name:     "standalone issue reference",
			message:  "Related to #123",
			expected: 1,
		},
		{
			name:     "multiple standalone references",
			message:  "See #1 and #2 for context",
			expected: 2,
		},
		{
			name:     "fix with extra whitespace",
			message:  "fix  #123",
			expected: 1,
		},
		{
			name:     "closed past tense",
			message:  "closed #123",
			expected: 1,
		},
		{
			name:     "fixed past tense",
			message:  "fixed #456",
			expected: 1,
		},
		{
			name:     "resolved past tense",
			message:  "resolved #789",
			expected: 1,
		},
		{
			name:     "close without s",
			message:  "close #123",
			expected: 1,
		},
		{
			name:     "fix without es",
			message:  "fix #456",
			expected: 1,
		},
		{
			name:     "resolve without s",
			message:  "resolve #789",
			expected: 1,
		},
		{
			name:     "hash without number",
			message:  "This is about # something",
			expected: 0,
		},
		{
			name:     "complex commit message",
			message:  "feat: Add new feature\n\nThis implements the feature requested in #123.\nAlso fixes #456 and closes #789.",
			expected: 3,
		},
		{
			name:     "PR style reference",
			message:  "Merge pull request #100 from feature-branch",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countIssueReferences(tt.message)
			assert.Equal(t, tt.expected, result, "message: %s", tt.message)
		})
	}
}

func TestAggregator_IssueComments(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	data := &models.RawData{
		// Need a commit to create the repository
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Author:     models.Author{Login: "user1"},
				Repository: "owner/repo",
			},
		},
		IssueComments: []models.IssueComment{
			{
				ID:         1,
				Issue:      1,
				Repository: "owner/repo",
				Author:     models.Author{Login: "user1"},
				CreatedAt:  time.Now(),
			},
			{
				ID:         2,
				Issue:      1,
				Repository: "owner/repo",
				Author:     models.Author{Login: "user1"},
				CreatedAt:  time.Now(),
			},
			{
				ID:         3,
				Issue:      2,
				Repository: "owner/repo",
				Author:     models.Author{Login: "user2"},
				CreatedAt:  time.Now(),
			},
		},
	}

	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	// Check that issue comments are counted
	require.Len(t, metrics.Repositories, 1)
	repo := metrics.Repositories[0]

	// Find user1 and user2
	var user1, user2 *models.ContributorMetrics
	for i := range repo.Contributors {
		if repo.Contributors[i].Login == "user1" {
			user1 = &repo.Contributors[i]
		}
		if repo.Contributors[i].Login == "user2" {
			user2 = &repo.Contributors[i]
		}
	}

	require.NotNil(t, user1)
	assert.Equal(t, 2, user1.IssueComments) // user1 has 2 comments

	require.NotNil(t, user2)
	assert.Equal(t, 1, user2.IssueComments) // user2 has 1 comment
}

func TestAggregator_IssueReferencesInCommits(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	agg := New(cfg)

	data := &models.RawData{
		Commits: []models.Commit{
			{
				SHA:        "abc123",
				Message:    "fixes #1 and closes #2",
				Author:     models.Author{Login: "user1"},
				Repository: "owner/repo",
			},
			{
				SHA:        "def456",
				Message:    "Regular commit without issue refs",
				Author:     models.Author{Login: "user1"},
				Repository: "owner/repo",
			},
			{
				SHA:        "ghi789",
				Message:    "resolves #3",
				Author:     models.Author{Login: "user2"},
				Repository: "owner/repo",
			},
		},
	}

	dateRange := &config.ParsedDateRange{}

	metrics, err := agg.Aggregate(data, dateRange)
	require.NoError(t, err)

	require.Len(t, metrics.Repositories, 1)
	repo := metrics.Repositories[0]

	// Find user1 and user2
	var user1, user2 *models.ContributorMetrics
	for i := range repo.Contributors {
		if repo.Contributors[i].Login == "user1" {
			user1 = &repo.Contributors[i]
		}
		if repo.Contributors[i].Login == "user2" {
			user2 = &repo.Contributors[i]
		}
	}

	require.NotNil(t, user1)
	assert.Equal(t, 2, user1.IssueReferencesInCommits) // user1 has 2 issue references (fixes #1, closes #2)

	require.NotNil(t, user2)
	assert.Equal(t, 1, user2.IssueReferencesInCommits) // user2 has 1 issue reference (resolves #3)
}
