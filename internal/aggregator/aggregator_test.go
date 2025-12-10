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
