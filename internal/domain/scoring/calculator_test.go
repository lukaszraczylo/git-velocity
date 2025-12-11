package scoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukaszraczylo/git-velocity/internal/config"
	"github.com/lukaszraczylo/git-velocity/internal/domain/models"
)

func TestNewCalculator(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}
	calc := NewCalculator(cfg)

	assert.NotNil(t, calc)
	assert.Equal(t, cfg, calc.config)
}

func TestCalculator_ScoringDisabled(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Scoring.Enabled = false
	calc := NewCalculator(cfg)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				Contributors: []models.ContributorMetrics{
					{Login: "user1", CommitCount: 100},
				},
			},
		},
	}

	result := calc.Calculate(metrics)

	// Should return unchanged metrics when scoring is disabled
	assert.Equal(t, 0, result.Repositories[0].Contributors[0].Score.Total)
	assert.Empty(t, result.Leaderboard)
}

func TestCalculator_BasicScoring(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Scoring.Enabled = true
	cfg.Scoring.Points = config.PointsConfig{
		Commit:        10,
		PROpened:      25,
		PRMerged:      50,
		PRReviewed:    30,
		ReviewComment: 5,
		LinesAdded:    0.1,
		LinesDeleted:  0.05,
	}
	calc := NewCalculator(cfg)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				FullName: "owner/repo",
				Contributors: []models.ContributorMetrics{
					{
						Login:                   "user1",
						Name:                    "User One",
						CommitCount:             10,
						LinesAdded:              1000,
						LinesDeleted:            500,
						PRsOpened:               5,
						PRsMerged:               3,
						ReviewsGiven:            8,
						ReviewComments:          20,
						RepositoriesContributed: []string{"owner/repo"},
					},
				},
			},
		},
	}

	result := calc.Calculate(metrics)

	require.Len(t, result.Leaderboard, 1)
	entry := result.Leaderboard[0]
	assert.Equal(t, "user1", entry.Login)
	assert.Equal(t, 1, entry.Rank)

	// Verify score breakdown:
	// Commits: 10 * 10 = 100
	// Lines: 1000 * 0.1 + 500 * 0.05 = 100 + 25 = 125
	// PRs: 5 * 25 + 3 * 50 = 125 + 150 = 275
	// Reviews: 8 * 30 + 20 * 5 = 240 + 100 = 340
	// Total: 100 + 125 + 275 + 340 = 840
	assert.Equal(t, 840, entry.Score)
}

func TestCalculator_FastReviewBonus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		avgReviewTime  float64
		expectedBonus  int
		expectedPoints config.PointsConfig
	}{
		{
			name:          "1 hour review gets 1h bonus",
			avgReviewTime: 0.5,
			expectedBonus: 50,
			expectedPoints: config.PointsConfig{
				FastReview1h:  50,
				FastReview4h:  30,
				FastReview24h: 10,
			},
		},
		{
			name:          "3 hour review gets 4h bonus",
			avgReviewTime: 3.0,
			expectedBonus: 30,
			expectedPoints: config.PointsConfig{
				FastReview1h:  50,
				FastReview4h:  30,
				FastReview24h: 10,
			},
		},
		{
			name:          "12 hour review gets 24h bonus",
			avgReviewTime: 12.0,
			expectedBonus: 10,
			expectedPoints: config.PointsConfig{
				FastReview1h:  50,
				FastReview4h:  30,
				FastReview24h: 10,
			},
		},
		{
			name:          "48 hour review gets no bonus",
			avgReviewTime: 48.0,
			expectedBonus: 0,
			expectedPoints: config.PointsConfig{
				FastReview1h:  50,
				FastReview4h:  30,
				FastReview24h: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.DefaultConfig()
			cfg.Scoring.Enabled = true
			cfg.Scoring.Points = tt.expectedPoints
			calc := NewCalculator(cfg)

			metrics := &models.GlobalMetrics{
				Repositories: []models.RepositoryMetrics{
					{
						FullName: "owner/repo",
						Contributors: []models.ContributorMetrics{
							{
								Login:                   "user1",
								ReviewsGiven:            5,
								AvgReviewTime:           tt.avgReviewTime,
								RepositoriesContributed: []string{"owner/repo"},
							},
						},
					},
				},
			}

			result := calc.Calculate(metrics)

			require.Len(t, result.Leaderboard, 1)
			// Get the contributor from the repository to check breakdown
			contributor := result.Repositories[0].Contributors[0]
			assert.Equal(t, tt.expectedBonus, contributor.Score.Breakdown.ResponseBonus)
		})
	}
}

func TestCalculator_MultipleContributorsRanking(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Scoring.Enabled = true
	cfg.Scoring.Points = config.PointsConfig{
		Commit: 10,
	}
	calc := NewCalculator(cfg)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				FullName: "owner/repo",
				Contributors: []models.ContributorMetrics{
					{
						Login:                   "user1",
						CommitCount:             100,
						RepositoriesContributed: []string{"owner/repo"},
					},
					{
						Login:                   "user2",
						CommitCount:             50,
						RepositoriesContributed: []string{"owner/repo"},
					},
					{
						Login:                   "user3",
						CommitCount:             200,
						RepositoriesContributed: []string{"owner/repo"},
					},
				},
			},
		},
	}

	result := calc.Calculate(metrics)

	require.Len(t, result.Leaderboard, 3)

	// Should be sorted by score (highest first)
	assert.Equal(t, "user3", result.Leaderboard[0].Login)
	assert.Equal(t, 1, result.Leaderboard[0].Rank)
	assert.Equal(t, 2000, result.Leaderboard[0].Score)

	assert.Equal(t, "user1", result.Leaderboard[1].Login)
	assert.Equal(t, 2, result.Leaderboard[1].Rank)
	assert.Equal(t, 1000, result.Leaderboard[1].Score)

	assert.Equal(t, "user2", result.Leaderboard[2].Login)
	assert.Equal(t, 3, result.Leaderboard[2].Rank)
	assert.Equal(t, 500, result.Leaderboard[2].Score)
}

func TestCalculator_PercentileRank(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Scoring.Enabled = true
	cfg.Scoring.Points = config.PointsConfig{Commit: 10}
	calc := NewCalculator(cfg)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				FullName: "owner/repo",
				Contributors: []models.ContributorMetrics{
					{Login: "user1", CommitCount: 100, RepositoriesContributed: []string{"owner/repo"}},
					{Login: "user2", CommitCount: 80, RepositoriesContributed: []string{"owner/repo"}},
					{Login: "user3", CommitCount: 60, RepositoriesContributed: []string{"owner/repo"}},
					{Login: "user4", CommitCount: 40, RepositoriesContributed: []string{"owner/repo"}},
				},
			},
		},
	}

	result := calc.Calculate(metrics)

	require.Len(t, result.Leaderboard, 4)

	// Leaderboard should be sorted by score (highest first)
	// user1: 100 commits * 10 = 1000, rank 1
	// user2: 80 commits * 10 = 800, rank 2
	// user3: 60 commits * 10 = 600, rank 3
	// user4: 40 commits * 10 = 400, rank 4
	assert.Equal(t, "user1", result.Leaderboard[0].Login)
	assert.Equal(t, 1, result.Leaderboard[0].Rank)
	assert.Equal(t, 1000, result.Leaderboard[0].Score)

	assert.Equal(t, "user2", result.Leaderboard[1].Login)
	assert.Equal(t, 2, result.Leaderboard[1].Rank)
	assert.Equal(t, 800, result.Leaderboard[1].Score)

	assert.Equal(t, "user3", result.Leaderboard[2].Login)
	assert.Equal(t, 3, result.Leaderboard[2].Rank)
	assert.Equal(t, 600, result.Leaderboard[2].Score)

	assert.Equal(t, "user4", result.Leaderboard[3].Login)
	assert.Equal(t, 4, result.Leaderboard[3].Rank)
	assert.Equal(t, 400, result.Leaderboard[3].Score)
}

func TestCalculator_Achievements(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Scoring.Enabled = true
	// Achievements are now hardcoded, no need to set them
	calc := NewCalculator(cfg)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				FullName: "owner/repo",
				Contributors: []models.ContributorMetrics{
					{
						Login:                   "user1",
						CommitCount:             15,  // Should earn commit-1, commit-10
						PRsOpened:               6,   // Should earn pr-1
						ReviewsGiven:            5,   // Should earn review-1
						AvgReviewTime:           0.5, // Should earn review-time-1h, review-time-4h, review-time-24h
						RepositoriesContributed: []string{"owner/repo"},
					},
				},
			},
		},
	}

	result := calc.Calculate(metrics)

	contributor := result.Repositories[0].Contributors[0]
	// Should have hardcoded achievements based on thresholds
	assert.Contains(t, contributor.Achievements, "commit-1")
	assert.Contains(t, contributor.Achievements, "commit-10")
	assert.Contains(t, contributor.Achievements, "pr-1")
	assert.Contains(t, contributor.Achievements, "review-1")
	assert.Contains(t, contributor.Achievements, "review-time-1h") // 0.5h < 1h threshold
	// Should NOT have commit-50 (only 15 commits)
	assert.NotContains(t, contributor.Achievements, "commit-50")
	// Should NOT have review-10 (only 5 reviews)
	assert.NotContains(t, contributor.Achievements, "review-10")
}

func TestCalculator_AllAchievementTypes(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Scoring.Enabled = true
	// Achievements are now hardcoded
	calc := NewCalculator(cfg)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				FullName: "owner/repo1",
				Contributors: []models.ContributorMetrics{
					{
						Login:                   "user1",
						CommitCount:             15,
						PRsOpened:               6,
						PRsMerged:               4,
						ReviewsGiven:            10,
						ReviewComments:          25,
						LinesAdded:              1500,
						LinesDeleted:            600,
						AvgReviewTime:           1.5,
						UniqueReviewees:         7,
						RepositoriesContributed: []string{"owner/repo1", "owner/repo2"},
					},
				},
			},
		},
	}

	result := calc.Calculate(metrics)

	contributor := result.Repositories[0].Contributors[0]
	// Should have various hardcoded achievements based on thresholds
	// Check some key achievements are earned
	assert.Contains(t, contributor.Achievements, "commit-1")
	assert.Contains(t, contributor.Achievements, "commit-10")
	assert.Contains(t, contributor.Achievements, "pr-1")
	assert.Contains(t, contributor.Achievements, "review-1")
	assert.Contains(t, contributor.Achievements, "review-10")
	assert.Contains(t, contributor.Achievements, "comment-10")
	assert.Contains(t, contributor.Achievements, "lines-added-100")
	assert.Contains(t, contributor.Achievements, "lines-added-1000")
	assert.Contains(t, contributor.Achievements, "lines-deleted-100")
	assert.Contains(t, contributor.Achievements, "lines-deleted-500")
	assert.Contains(t, contributor.Achievements, "review-time-4h") // 1.5h < 4h
	assert.Contains(t, contributor.Achievements, "repo-2")         // 2 repos
	assert.Contains(t, contributor.Achievements, "reviewees-3")    // 7 reviewees >= 3
	// Should have earned multiple achievements (more than 10)
	assert.Greater(t, len(contributor.Achievements), 10)
}

func TestCalculator_TopAchievers(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Scoring.Enabled = true
	cfg.Scoring.Points = config.PointsConfig{
		Commit:     10,
		PROpened:   25,
		PRReviewed: 30,
	}
	calc := NewCalculator(cfg)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				FullName: "owner/repo",
				Contributors: []models.ContributorMetrics{
					{
						Login:                   "committer",
						CommitCount:             100,
						PRsOpened:               5,
						ReviewsGiven:            2,
						RepositoriesContributed: []string{"owner/repo"},
					},
					{
						Login:                   "pr-author",
						CommitCount:             10,
						PRsOpened:               50,
						ReviewsGiven:            3,
						RepositoriesContributed: []string{"owner/repo"},
					},
					{
						Login:                   "reviewer",
						CommitCount:             5,
						PRsOpened:               2,
						ReviewsGiven:            100,
						RepositoriesContributed: []string{"owner/repo"},
					},
				},
			},
		},
	}

	result := calc.Calculate(metrics)

	assert.Equal(t, "committer", result.TopAchievers["commits"])
	assert.Equal(t, "pr-author", result.TopAchievers["pull_requests"])
	assert.Equal(t, "reviewer", result.TopAchievers["reviews"])
	// Overall top achiever has highest score
	assert.NotEmpty(t, result.TopAchievers["overall"])
}

func TestCalculator_TeamScoring(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Scoring.Enabled = true
	cfg.Scoring.Points = config.PointsConfig{Commit: 10}
	cfg.Teams = []config.TeamConfig{
		{
			Name:    "Backend Team",
			Members: []string{"user1", "user2"},
			Color:   "#ff0000",
		},
	}
	calc := NewCalculator(cfg)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				FullName: "owner/repo",
				Contributors: []models.ContributorMetrics{
					{Login: "user1", CommitCount: 50, RepositoriesContributed: []string{"owner/repo"}},
					{Login: "user2", CommitCount: 30, RepositoriesContributed: []string{"owner/repo"}},
				},
			},
		},
		Teams: []models.TeamMetrics{
			{
				Name:    "Backend Team",
				Members: []string{"user1", "user2"},
				MemberMetrics: []models.ContributorMetrics{
					{Login: "user1"},
					{Login: "user2"},
				},
			},
		},
	}

	result := calc.Calculate(metrics)

	require.Len(t, result.Teams, 1)
	team := result.Teams[0]
	// Total: 500 + 300 = 800
	assert.Equal(t, 800, team.TotalScore)
	// Avg: 800 / 2 = 400
	assert.Equal(t, 400.0, team.AvgScore)

	// Check individual member scores
	assert.Equal(t, 500, team.MemberMetrics[0].Score.Total)
	assert.Equal(t, 300, team.MemberMetrics[1].Score.Total)
}

func TestCalculator_TeamInLeaderboard(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Scoring.Enabled = true
	cfg.Scoring.Points = config.PointsConfig{Commit: 10}
	cfg.Teams = []config.TeamConfig{
		{
			Name:    "Backend Team",
			Members: []string{"user1"},
		},
	}
	calc := NewCalculator(cfg)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				FullName: "owner/repo",
				Contributors: []models.ContributorMetrics{
					{Login: "user1", CommitCount: 50, RepositoriesContributed: []string{"owner/repo"}},
					{Login: "user2", CommitCount: 30, RepositoriesContributed: []string{"owner/repo"}},
				},
			},
		},
	}

	result := calc.Calculate(metrics)

	// user1 should have team name in leaderboard
	assert.Equal(t, "Backend Team", result.Leaderboard[0].Team)
	// user2 should not have a team
	assert.Empty(t, result.Leaderboard[1].Team)
}

func TestCalculator_DetermineTopCategory(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Scoring.Enabled = true
	calc := NewCalculator(cfg)

	tests := []struct {
		name             string
		contributor      models.ContributorMetrics
		expectedCategory string
	}{
		{
			name: "Top committer",
			contributor: models.ContributorMetrics{
				CommitCount:    100,
				PRsOpened:      10,
				ReviewsGiven:   5,
				ReviewComments: 20,
			},
			expectedCategory: "Commits",
		},
		{
			name: "Top PR author",
			contributor: models.ContributorMetrics{
				CommitCount:    10,
				PRsOpened:      100,
				ReviewsGiven:   5,
				ReviewComments: 20,
			},
			expectedCategory: "PRs",
		},
		{
			name: "Top reviewer",
			contributor: models.ContributorMetrics{
				CommitCount:    10,
				PRsOpened:      5,
				ReviewsGiven:   100,
				ReviewComments: 20,
			},
			expectedCategory: "Reviews",
		},
		{
			name: "Top commenter",
			contributor: models.ContributorMetrics{
				CommitCount:    10,
				PRsOpened:      5,
				ReviewsGiven:   20,
				ReviewComments: 100,
			},
			expectedCategory: "Comments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := calc.determineTopCategory(&tt.contributor)
			assert.Equal(t, tt.expectedCategory, result)
		})
	}
}

func TestCalculator_MultipleRepositories(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Scoring.Enabled = true
	cfg.Scoring.Points = config.PointsConfig{Commit: 10}
	calc := NewCalculator(cfg)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				FullName: "owner/repo1",
				Contributors: []models.ContributorMetrics{
					{Login: "user1", CommitCount: 50, RepositoriesContributed: []string{"owner/repo1"}},
				},
			},
			{
				FullName: "owner/repo2",
				Contributors: []models.ContributorMetrics{
					{Login: "user1", CommitCount: 30, RepositoriesContributed: []string{"owner/repo2"}},
				},
			},
		},
	}

	result := calc.Calculate(metrics)

	// Should aggregate commits from both repos
	require.Len(t, result.Leaderboard, 1)
	// 50 + 30 = 80 commits * 10 = 800
	assert.Equal(t, 800, result.Leaderboard[0].Score)

	// Per-repo scores should reflect repo-specific metrics, not global
	// Repo1 has 50 commits * 10 = 500
	contributor := result.Repositories[0].Contributors[0]
	assert.Equal(t, 500, contributor.Score.Total)
}

func TestCalculator_EmptyMetrics(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Scoring.Enabled = true
	calc := NewCalculator(cfg)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{},
	}

	result := calc.Calculate(metrics)

	assert.Empty(t, result.Leaderboard)
	assert.Empty(t, result.TopAchievers)
}

func TestCalculator_NoReviewsNoBonus(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Scoring.Enabled = true
	cfg.Scoring.Points = config.PointsConfig{
		FastReview1h: 50,
	}
	calc := NewCalculator(cfg)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				FullName: "owner/repo",
				Contributors: []models.ContributorMetrics{
					{
						Login:                   "user1",
						ReviewsGiven:            0,
						AvgReviewTime:           0.5, // Fast but no reviews
						RepositoriesContributed: []string{"owner/repo"},
					},
				},
			},
		},
	}

	result := calc.Calculate(metrics)

	contributor := result.Repositories[0].Contributors[0]
	// Should not get bonus if no reviews given
	assert.Equal(t, 0, contributor.Score.Breakdown.ResponseBonus)
}

func TestCalculator_OutOfHoursScoring(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Scoring.Enabled = true
	cfg.Scoring.Points = config.PointsConfig{
		Commit:     10,
		OutOfHours: 5, // 5 points per out-of-hours commit
	}
	calc := NewCalculator(cfg)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				FullName: "owner/repo",
				Contributors: []models.ContributorMetrics{
					{
						Login:                   "night-owl",
						CommitCount:             10,
						OutOfHoursCount:         8, // 8 commits outside 9am-5pm
						RepositoriesContributed: []string{"owner/repo"},
					},
				},
			},
		},
	}

	result := calc.Calculate(metrics)

	contributor := result.Repositories[0].Contributors[0]
	// Commits: 10 * 10 = 100
	// OutOfHours: 8 * 5 = 40
	// Total: 140
	assert.Equal(t, 100, contributor.Score.Breakdown.Commits)
	assert.Equal(t, 40, contributor.Score.Breakdown.OutOfHours)
	assert.Equal(t, 140, contributor.Score.Total)
}

func TestCalculator_WorkWeekStreakAchievement(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Scoring.Enabled = true
	// Achievements are now hardcoded
	calc := NewCalculator(cfg)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				FullName: "owner/repo",
				Contributors: []models.ContributorMetrics{
					{
						Login:                   "consistent-worker",
						CommitCount:             20,
						WorkWeekStreak:          5, // 5-day work week streak
						RepositoriesContributed: []string{"owner/repo"},
					},
				},
			},
		},
	}

	result := calc.Calculate(metrics)

	contributor := result.Repositories[0].Contributors[0]
	// Should have earned work week streak achievements for 3 and 5 days
	assert.Contains(t, contributor.Achievements, "workweek-3")
	assert.Contains(t, contributor.Achievements, "workweek-5")
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
