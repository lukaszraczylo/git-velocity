package scoring

import (
	"sort"

	"github.com/lukaszraczylo/git-velocity/internal/config"
	"github.com/lukaszraczylo/git-velocity/internal/domain/models"
)

// Calculator handles score and achievement calculations
type Calculator struct {
	config *config.Config
}

// NewCalculator creates a new scoring calculator
func NewCalculator(cfg *config.Config) *Calculator {
	return &Calculator{config: cfg}
}

// Calculate computes scores and achievements for all metrics
func (c *Calculator) Calculate(metrics *models.GlobalMetrics) *models.GlobalMetrics {
	if !c.config.Scoring.Enabled {
		return metrics
	}

	// Collect all contributor metrics across repositories
	contributorMap := make(map[string]*models.ContributorMetrics)

	for _, repo := range metrics.Repositories {
		for i := range repo.Contributors {
			login := repo.Contributors[i].Login
			if _, ok := contributorMap[login]; !ok {
				// Copy the contributor metrics
				cm := repo.Contributors[i]
				contributorMap[login] = &cm
			} else {
				// Aggregate metrics from multiple repos
				existing := contributorMap[login]
				cm := repo.Contributors[i]
				existing.CommitCount += cm.CommitCount
				existing.LinesAdded += cm.LinesAdded
				existing.LinesDeleted += cm.LinesDeleted
				existing.MeaningfulLinesAdded += cm.MeaningfulLinesAdded
				existing.MeaningfulLinesDeleted += cm.MeaningfulLinesDeleted
				existing.CommentLinesAdded += cm.CommentLinesAdded
				existing.CommentLinesDeleted += cm.CommentLinesDeleted
				existing.PRsOpened += cm.PRsOpened
				existing.PRsMerged += cm.PRsMerged
				existing.ReviewsGiven += cm.ReviewsGiven
				existing.ReviewComments += cm.ReviewComments
				// Combine unique repositories
				for _, r := range cm.RepositoriesContributed {
					if !contains(existing.RepositoriesContributed, r) {
						existing.RepositoriesContributed = append(existing.RepositoriesContributed, r)
					}
				}
			}
		}
	}

	// Calculate scores for each contributor
	for _, cm := range contributorMap {
		cm.Score = c.calculateScore(cm)
		// Check achievements
		cm.Achievements = c.checkAchievements(cm)
	}

	// Convert to slice and sort by score
	var contributors []models.ContributorMetrics
	for _, cm := range contributorMap {
		contributors = append(contributors, *cm)
	}

	sort.Slice(contributors, func(i, j int) bool {
		return contributors[i].Score.Total > contributors[j].Score.Total
	})

	// Assign ranks
	for i := range contributors {
		contributors[i].Score.Rank = i + 1
		contributors[i].Score.PercentileRank = float64(len(contributors)-i) / float64(len(contributors)) * 100
	}

	// Build leaderboard
	leaderboard := make([]models.LeaderboardEntry, len(contributors))
	topAchievers := make(map[string]string)

	for i, cm := range contributors {
		// Find team for user
		team := ""
		if teamCfg := c.config.GetTeamForUser(cm.Login); teamCfg != nil {
			team = teamCfg.Name
		}

		// Determine top category
		topCategory := c.determineTopCategory(&cm)

		leaderboard[i] = models.LeaderboardEntry{
			Rank:         i + 1,
			Login:        cm.Login,
			Name:         cm.Name,
			AvatarURL:    cm.AvatarURL,
			Score:        cm.Score.Total,
			Team:         team,
			TopCategory:  topCategory,
			Achievements: cm.Achievements,
		}

		// Track top achievers
		if i == 0 {
			topAchievers["overall"] = cm.Login
		}
	}

	// Find top achievers in each category
	c.findTopAchievers(contributors, topAchievers)

	// Update the metrics
	metrics.Leaderboard = leaderboard
	metrics.TopAchievers = topAchievers
	metrics.Contributors = contributors // Update global contributors with scored data

	// Calculate per-repository scores (based on repo-specific metrics, not global)
	for i := range metrics.Repositories {
		for j := range metrics.Repositories[i].Contributors {
			repoContrib := &metrics.Repositories[i].Contributors[j]
			repoContrib.Score = c.calculateScore(repoContrib)
			// Achievements are based on repo-specific activity
			repoContrib.Achievements = c.checkAchievements(repoContrib)
		}
		// Re-sort by score after calculation
		sort.Slice(metrics.Repositories[i].Contributors, func(a, b int) bool {
			return metrics.Repositories[i].Contributors[a].Score.Total > metrics.Repositories[i].Contributors[b].Score.Total
		})
	}

	// Update team scores
	for i := range metrics.Teams {
		var totalScore int
		for j := range metrics.Teams[i].MemberMetrics {
			login := metrics.Teams[i].MemberMetrics[j].Login
			if cm, ok := contributorMap[login]; ok {
				metrics.Teams[i].MemberMetrics[j].Score = cm.Score
				metrics.Teams[i].MemberMetrics[j].Achievements = cm.Achievements
				totalScore += cm.Score.Total
			}
		}
		metrics.Teams[i].TotalScore = totalScore
		if len(metrics.Teams[i].MemberMetrics) > 0 {
			metrics.Teams[i].AvgScore = float64(totalScore) / float64(len(metrics.Teams[i].MemberMetrics))
		}
	}

	return metrics
}

// calculateScore computes the score for a contributor based on their metrics
func (c *Calculator) calculateScore(cm *models.ContributorMetrics) models.Score {
	points := c.config.Scoring.Points
	breakdown := models.ScoreBreakdown{}

	// Commit points
	breakdown.Commits = cm.CommitCount * points.Commit

	// Line change points - use meaningful lines if configured, otherwise raw counts
	linesAdded := cm.LinesAdded
	linesDeleted := cm.LinesDeleted
	if points.UseMeaningfulLines {
		linesAdded = cm.MeaningfulLinesAdded
		linesDeleted = cm.MeaningfulLinesDeleted
	}
	breakdown.LineChanges = int(float64(linesAdded)*points.LinesAdded +
		float64(linesDeleted)*points.LinesDeleted)

	// PR points
	breakdown.PRs = cm.PRsOpened*points.PROpened + cm.PRsMerged*points.PRMerged

	// Review points (PR reviews)
	breakdown.Reviews = cm.ReviewsGiven * points.PRReviewed

	// Comment points (PR review comments)
	breakdown.Comments = cm.ReviewComments * points.ReviewComment

	// Response time bonus
	if cm.ReviewsGiven > 0 && cm.AvgReviewTime > 0 {
		if cm.AvgReviewTime <= 1 {
			breakdown.ResponseBonus = points.FastReview1h
		} else if cm.AvgReviewTime <= 4 {
			breakdown.ResponseBonus = points.FastReview4h
		} else if cm.AvgReviewTime <= 24 {
			breakdown.ResponseBonus = points.FastReview24h
		}
	}

	// Out of hours bonus (commits outside 9am-5pm)
	breakdown.OutOfHours = cm.OutOfHoursCount * points.OutOfHours

	// Calculate total
	total := breakdown.Commits + breakdown.LineChanges + breakdown.PRs +
		breakdown.Reviews + breakdown.ResponseBonus + breakdown.Comments + breakdown.OutOfHours

	return models.Score{
		Total:     total,
		Breakdown: breakdown,
	}
}

func (c *Calculator) checkAchievements(cm *models.ContributorMetrics) []string {
	// Collect ALL earned achievements (including all tiers)
	var achievements []string

	for _, ach := range c.config.Scoring.GetAchievements() {
		earned := false

		switch ach.Condition.Type {
		case "commit_count":
			earned = float64(cm.CommitCount) >= ach.Condition.Threshold
		case "pr_opened_count":
			earned = float64(cm.PRsOpened) >= ach.Condition.Threshold
		case "pr_merged_count":
			earned = float64(cm.PRsMerged) >= ach.Condition.Threshold
		case "review_count":
			earned = float64(cm.ReviewsGiven) >= ach.Condition.Threshold
		case "comment_count":
			earned = float64(cm.ReviewComments) >= ach.Condition.Threshold
		case "lines_added":
			earned = float64(cm.LinesAdded) >= ach.Condition.Threshold
		case "lines_deleted":
			earned = float64(cm.LinesDeleted) >= ach.Condition.Threshold
		case "avg_review_time_hours":
			// For avg review time, lower is better, so lower threshold = harder achievement
			if cm.AvgReviewTime > 0 && cm.AvgReviewTime <= ach.Condition.Threshold {
				earned = true
			}
		case "repo_count":
			earned = float64(len(cm.RepositoriesContributed)) >= ach.Condition.Threshold
		case "unique_reviewees":
			earned = float64(cm.UniqueReviewees) >= ach.Condition.Threshold
		// New PR quality metrics
		case "largest_pr_size":
			earned = float64(cm.LargestPRSize) >= ach.Condition.Threshold
		case "small_pr_count":
			earned = float64(cm.SmallPRCount) >= ach.Condition.Threshold
		case "perfect_prs":
			earned = float64(cm.PerfectPRs) >= ach.Condition.Threshold
		// Activity pattern metrics
		case "active_days":
			earned = float64(cm.ActiveDays) >= ach.Condition.Threshold
		case "longest_streak":
			earned = float64(cm.LongestStreak) >= ach.Condition.Threshold
		case "early_bird_count":
			earned = float64(cm.EarlyBirdCount) >= ach.Condition.Threshold
		case "night_owl_count":
			earned = float64(cm.NightOwlCount) >= ach.Condition.Threshold
		case "midnight_count":
			earned = float64(cm.MidnightCount) >= ach.Condition.Threshold
		case "weekend_warrior":
			earned = float64(cm.WeekendWarrior) >= ach.Condition.Threshold
		case "out_of_hours_count":
			earned = float64(cm.OutOfHoursCount) >= ach.Condition.Threshold
		case "work_week_streak":
			earned = float64(cm.WorkWeekStreak) >= ach.Condition.Threshold
		// Documentation & comments
		case "comment_lines_added":
			earned = float64(cm.CommentLinesAdded) >= ach.Condition.Threshold
		case "comment_lines_deleted":
			earned = float64(cm.CommentLinesDeleted) >= ach.Condition.Threshold
		}

		if earned {
			achievements = append(achievements, ach.ID)
		}
	}

	return achievements
}

func (c *Calculator) determineTopCategory(cm *models.ContributorMetrics) string {
	// Determine what the contributor is best at
	categories := map[string]int{
		"Commits":  cm.CommitCount,
		"PRs":      cm.PRsOpened,
		"Reviews":  cm.ReviewsGiven,
		"Comments": cm.ReviewComments,
	}

	topCategory := ""
	topValue := 0

	for category, value := range categories {
		if value > topValue {
			topValue = value
			topCategory = category
		}
	}

	return topCategory
}

func (c *Calculator) findTopAchievers(contributors []models.ContributorMetrics, topAchievers map[string]string) {
	var topCommitter, topReviewer, topPRAuthor string
	var maxCommits, maxReviews, maxPRs int

	for _, cm := range contributors {
		if cm.CommitCount > maxCommits {
			maxCommits = cm.CommitCount
			topCommitter = cm.Login
		}
		if cm.ReviewsGiven > maxReviews {
			maxReviews = cm.ReviewsGiven
			topReviewer = cm.Login
		}
		if cm.PRsOpened > maxPRs {
			maxPRs = cm.PRsOpened
			topPRAuthor = cm.Login
		}
	}

	if topCommitter != "" {
		topAchievers["commits"] = topCommitter
	}
	if topReviewer != "" {
		topAchievers["reviews"] = topReviewer
	}
	if topPRAuthor != "" {
		topAchievers["pull_requests"] = topPRAuthor
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
