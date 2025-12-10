package site

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	json "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukaszraczylo/git-velocity/internal/config"
	"github.com/lukaszraczylo/git-velocity/internal/domain/models"
)

func TestNewGenerator(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	gen, err := NewGenerator("/tmp/output", cfg)

	require.NoError(t, err)
	assert.NotNil(t, gen)
	assert.Equal(t, "/tmp/output", gen.outputDir)
	assert.Equal(t, cfg, gen.config)
}

func TestGenerator_GenerateCreatesOutputDir(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "new-output")

	cfg := config.DefaultConfig()
	gen, err := NewGenerator(outputDir, cfg)
	require.NoError(t, err)

	metrics := &models.GlobalMetrics{
		Period: models.Period{
			Start: time.Now().Add(-24 * time.Hour),
			End:   time.Now(),
		},
	}

	err = gen.Generate(metrics)
	require.NoError(t, err)

	// Verify output directory was created
	info, err := os.Stat(outputDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestGenerator_GenerateCreatesDataDir(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.DefaultConfig()
	gen, err := NewGenerator(tempDir, cfg)
	require.NoError(t, err)

	metrics := &models.GlobalMetrics{}

	err = gen.Generate(metrics)
	require.NoError(t, err)

	// Verify data directory was created
	dataDir := filepath.Join(tempDir, "data")
	info, err := os.Stat(dataDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestGenerator_GenerateGlobalJSON(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.DefaultConfig()
	gen, err := NewGenerator(tempDir, cfg)
	require.NoError(t, err)

	metrics := &models.GlobalMetrics{
		TotalContributors: 5,
		TotalCommits:      100,
		TotalPRs:          50,
		TotalReviews:      75,
		TotalLinesAdded:   10000,
		TotalLinesDeleted: 5000,
	}

	err = gen.Generate(metrics)
	require.NoError(t, err)

	// Read and verify global.json
	globalPath := filepath.Join(tempDir, "data", "global.json")
	data, err := os.ReadFile(globalPath)
	require.NoError(t, err)

	var result struct {
		TotalContributors int       `json:"total_contributors"`
		TotalCommits      int       `json:"total_commits"`
		TotalPRs          int       `json:"total_prs"`
		TotalReviews      int       `json:"total_reviews"`
		TotalLinesAdded   int       `json:"total_lines_added"`
		TotalLinesDeleted int       `json:"total_lines_deleted"`
		GeneratedAt       time.Time `json:"generated_at"`
	}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, 5, result.TotalContributors)
	assert.Equal(t, 100, result.TotalCommits)
	assert.Equal(t, 50, result.TotalPRs)
	assert.Equal(t, 75, result.TotalReviews)
	assert.Equal(t, 10000, result.TotalLinesAdded)
	assert.Equal(t, 5000, result.TotalLinesDeleted)
	assert.False(t, result.GeneratedAt.IsZero())
}

func TestGenerator_GenerateLeaderboardJSON(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.DefaultConfig()
	gen, err := NewGenerator(tempDir, cfg)
	require.NoError(t, err)

	metrics := &models.GlobalMetrics{
		Leaderboard: []models.LeaderboardEntry{
			{Rank: 1, Login: "user1", Score: 1000},
			{Rank: 2, Login: "user2", Score: 800},
			{Rank: 3, Login: "user3", Score: 600},
		},
	}

	err = gen.Generate(metrics)
	require.NoError(t, err)

	// Read and verify leaderboard.json
	leaderboardPath := filepath.Join(tempDir, "data", "leaderboard.json")
	data, err := os.ReadFile(leaderboardPath)
	require.NoError(t, err)

	var result []models.LeaderboardEntry
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	require.Len(t, result, 3)
	assert.Equal(t, "user1", result[0].Login)
	assert.Equal(t, 1000, result[0].Score)
	assert.Equal(t, "user2", result[1].Login)
	assert.Equal(t, 800, result[1].Score)
}

func TestGenerator_GenerateRepositoryJSON(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.DefaultConfig()
	gen, err := NewGenerator(tempDir, cfg)
	require.NoError(t, err)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				Owner:        "myorg",
				Name:         "myrepo",
				TotalCommits: 42,
				TotalPRs:     10,
			},
		},
	}

	err = gen.Generate(metrics)
	require.NoError(t, err)

	// Read and verify repository metrics
	repoPath := filepath.Join(tempDir, "data", "repos", "myorg", "myrepo", "metrics.json")
	data, err := os.ReadFile(repoPath)
	require.NoError(t, err)

	var result models.RepositoryMetrics
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "myorg", result.Owner)
	assert.Equal(t, "myrepo", result.Name)
	assert.Equal(t, 42, result.TotalCommits)
	assert.Equal(t, 10, result.TotalPRs)
}

func TestGenerator_GenerateMultipleRepositories(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.DefaultConfig()
	gen, err := NewGenerator(tempDir, cfg)
	require.NoError(t, err)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{Owner: "org1", Name: "repo1", TotalCommits: 100},
			{Owner: "org1", Name: "repo2", TotalCommits: 200},
			{Owner: "org2", Name: "repo3", TotalCommits: 300},
		},
	}

	err = gen.Generate(metrics)
	require.NoError(t, err)

	// Verify all repository files exist
	_, err = os.Stat(filepath.Join(tempDir, "data", "repos", "org1", "repo1", "metrics.json"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "data", "repos", "org1", "repo2", "metrics.json"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "data", "repos", "org2", "repo3", "metrics.json"))
	assert.NoError(t, err)
}

func TestGenerator_GenerateTeamJSON(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.DefaultConfig()
	gen, err := NewGenerator(tempDir, cfg)
	require.NoError(t, err)

	metrics := &models.GlobalMetrics{
		Teams: []models.TeamMetrics{
			{
				Name:       "Backend Team",
				Color:      "#ff0000",
				Members:    []string{"user1", "user2"},
				TotalScore: 1500,
			},
		},
	}

	err = gen.Generate(metrics)
	require.NoError(t, err)

	// Read and verify team JSON (slugified name)
	teamPath := filepath.Join(tempDir, "data", "teams", "backend-team.json")
	data, err := os.ReadFile(teamPath)
	require.NoError(t, err)

	var result models.TeamMetrics
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "Backend Team", result.Name)
	assert.Equal(t, "#ff0000", result.Color)
	assert.Equal(t, 1500, result.TotalScore)
	assert.Len(t, result.Members, 2)
}

func TestGenerator_GenerateContributorJSON(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.DefaultConfig()
	gen, err := NewGenerator(tempDir, cfg)
	require.NoError(t, err)

	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				Contributors: []models.ContributorMetrics{
					{
						Login:       "john-doe",
						Name:        "John Doe",
						CommitCount: 50,
						PRsOpened:   10,
					},
				},
			},
		},
	}

	err = gen.Generate(metrics)
	require.NoError(t, err)

	// Read and verify contributor JSON
	contributorPath := filepath.Join(tempDir, "data", "contributors", "john-doe.json")
	data, err := os.ReadFile(contributorPath)
	require.NoError(t, err)

	var result models.ContributorMetrics
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "john-doe", result.Login)
	assert.Equal(t, "John Doe", result.Name)
	assert.Equal(t, 50, result.CommitCount)
	assert.Equal(t, 10, result.PRsOpened)
}

func TestGenerator_ContributorDeduplication(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.DefaultConfig()
	gen, err := NewGenerator(tempDir, cfg)
	require.NoError(t, err)

	// Same contributor in multiple repos
	metrics := &models.GlobalMetrics{
		Repositories: []models.RepositoryMetrics{
			{
				Contributors: []models.ContributorMetrics{
					{Login: "user1", CommitCount: 50},
				},
			},
			{
				Contributors: []models.ContributorMetrics{
					{Login: "user1", CommitCount: 75}, // Same user, different count
				},
			},
		},
	}

	err = gen.Generate(metrics)
	require.NoError(t, err)

	// Should only have one contributor file (first one seen)
	contributorPath := filepath.Join(tempDir, "data", "contributors", "user1.json")
	data, err := os.ReadFile(contributorPath)
	require.NoError(t, err)

	var result models.ContributorMetrics
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Should be the first one (50 commits)
	assert.Equal(t, 50, result.CommitCount)
}

func TestGenerator_NoTeamsDoesNotCreateTeamDir(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.DefaultConfig()
	gen, err := NewGenerator(tempDir, cfg)
	require.NoError(t, err)

	metrics := &models.GlobalMetrics{
		Teams: []models.TeamMetrics{}, // Empty teams
	}

	err = gen.Generate(metrics)
	require.NoError(t, err)

	// Team directory should not exist
	teamDir := filepath.Join(tempDir, "data", "teams")
	_, err = os.Stat(teamDir)
	assert.True(t, os.IsNotExist(err))
}

func TestSlugify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"Backend Team", "backend-team"},
		{"Frontend_Team", "frontend-team"},
		{"UPPER CASE", "upper-case"},
		{"already-slug", "already-slug"},
		{"Multiple   Spaces", "multiple---spaces"},
		{"Mixed_And Spaced", "mixed-and-spaced"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			result := slugify(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriteJSON(t *testing.T) {
	tempDir := t.TempDir()

	testData := map[string]interface{}{
		"key":    "value",
		"number": 42,
		"nested": map[string]string{
			"inner": "data",
		},
	}

	path := filepath.Join(tempDir, "test.json")
	err := writeJSON(path, testData)
	require.NoError(t, err)

	// Verify file was created and is valid JSON
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "value", result["key"])
	assert.Equal(t, float64(42), result["number"]) // JSON numbers are float64
}

func TestWriteJSON_Indented(t *testing.T) {
	tempDir := t.TempDir()

	testData := map[string]string{"key": "value"}
	path := filepath.Join(tempDir, "test.json")

	err := writeJSON(path, testData)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	// Should be formatted with indentation
	assert.Contains(t, string(data), "\n")
	assert.Contains(t, string(data), "  ") // 2-space indent
}

func TestWriteJSON_ErrorOnInvalidPath(t *testing.T) {
	// Try to write to a path that doesn't exist
	path := "/nonexistent/directory/test.json"
	err := writeJSON(path, "data")
	assert.Error(t, err)
}

func TestGenerator_GenerateWithFullMetrics(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.DefaultConfig()
	gen, err := NewGenerator(tempDir, cfg)
	require.NoError(t, err)

	// Create comprehensive metrics
	metrics := &models.GlobalMetrics{
		Period: models.Period{
			Start:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:         time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			Granularity: "monthly",
			Label:       "2024",
		},
		TotalContributors: 10,
		TotalCommits:      500,
		TotalPRs:          100,
		TotalReviews:      200,
		TotalLinesAdded:   50000,
		TotalLinesDeleted: 25000,
		Repositories: []models.RepositoryMetrics{
			{
				Owner:              "org",
				Name:               "repo1",
				TotalCommits:       300,
				TotalPRs:           60,
				ActiveContributors: 5,
				Contributors: []models.ContributorMetrics{
					{Login: "alice", Name: "Alice", CommitCount: 100},
					{Login: "bob", Name: "Bob", CommitCount: 200},
				},
			},
			{
				Owner:              "org",
				Name:               "repo2",
				TotalCommits:       200,
				TotalPRs:           40,
				ActiveContributors: 5,
				Contributors: []models.ContributorMetrics{
					{Login: "alice", Name: "Alice", CommitCount: 50},
					{Login: "charlie", Name: "Charlie", CommitCount: 150},
				},
			},
		},
		Teams: []models.TeamMetrics{
			{
				Name:       "Core Team",
				Members:    []string{"alice", "bob"},
				TotalScore: 5000,
			},
		},
		Leaderboard: []models.LeaderboardEntry{
			{Rank: 1, Login: "alice", Score: 3000},
			{Rank: 2, Login: "bob", Score: 2000},
			{Rank: 3, Login: "charlie", Score: 1500},
		},
	}

	err = gen.Generate(metrics)
	require.NoError(t, err)

	// Verify all expected files exist
	expectedPaths := []string{
		filepath.Join(tempDir, "data", "global.json"),
		filepath.Join(tempDir, "data", "leaderboard.json"),
		filepath.Join(tempDir, "data", "repos", "org", "repo1", "metrics.json"),
		filepath.Join(tempDir, "data", "repos", "org", "repo2", "metrics.json"),
		filepath.Join(tempDir, "data", "teams", "core-team.json"),
		filepath.Join(tempDir, "data", "contributors", "alice.json"),
		filepath.Join(tempDir, "data", "contributors", "bob.json"),
		filepath.Join(tempDir, "data", "contributors", "charlie.json"),
	}

	for _, path := range expectedPaths {
		_, err := os.Stat(path)
		assert.NoError(t, err, "Expected file to exist: %s", path)
	}
}
