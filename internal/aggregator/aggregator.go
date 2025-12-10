package aggregator

import (
	"sort"
	"strings"
	"time"

	"github.com/lukaszraczylo/git-velocity/internal/config"
	"github.com/lukaszraczylo/git-velocity/internal/domain/models"
)

// UserProfile contains GitHub user profile information for deduplication
type UserProfile struct {
	ID        int64  // GitHub user ID
	Login     string // GitHub username
	Name      string // Display name
	Email     string // Public email (may be empty)
	AvatarURL string
}

// Aggregator handles metrics aggregation
type Aggregator struct {
	config       *config.Config
	userProfiles map[string]UserProfile // GitHub login -> profile
}

// New creates a new Aggregator
func New(cfg *config.Config) *Aggregator {
	return &Aggregator{
		config:       cfg,
		userProfiles: make(map[string]UserProfile),
	}
}

// SetUserProfiles sets the user profiles for enhanced deduplication
func (a *Aggregator) SetUserProfiles(profiles map[string]UserProfile) {
	a.userProfiles = profiles
}

// Aggregate processes raw data and produces global metrics
func (a *Aggregator) Aggregate(data *models.RawData, dateRange *config.ParsedDateRange) (*models.GlobalMetrics, error) {
	period := models.Period{
		End:         time.Now(),
		Granularity: "all",
		Label:       "All Time",
	}

	if dateRange.Start != nil {
		period.Start = *dateRange.Start
	}
	if dateRange.End != nil {
		period.End = *dateRange.End
	}

	// Build email-to-login mapping from PRs and reviews (these have real GitHub logins)
	// This helps normalize commit authors to their GitHub usernames
	emailToLogin := buildEmailToLoginMapping(data, a.userProfiles)

	// Build login-to-login mapping for sanitized logins (e.g., lukasz-raczylo -> lukaszraczylo)
	// Also returns verified login info with avatar URLs
	loginToLogin, loginToInfo := buildLoginMapping(data)

	// Build contributor map (global stats across all repos)
	contributorMap := make(map[string]*models.ContributorMetrics)
	repoMap := make(map[string]*models.RepositoryMetrics)

	// Per-repository contributor maps (repo -> login -> metrics)
	repoContributorMap := make(map[string]map[string]*models.ContributorMetrics)

	// Track activity days per contributor for streak calculation
	activityDays := make(map[string]map[string]bool) // login -> set of date strings
	// Per-repo activity days
	repoActivityDays := make(map[string]map[string]map[string]bool) // repo -> login -> set of date strings

	// Helper to get or create per-repo contributor
	getRepoContributor := func(repo, login, name, avatarURL string) *models.ContributorMetrics {
		if repoContributorMap[repo] == nil {
			repoContributorMap[repo] = make(map[string]*models.ContributorMetrics)
		}
		if _, ok := repoContributorMap[repo][login]; !ok {
			repoContributorMap[repo][login] = &models.ContributorMetrics{
				Login:     login,
				Name:      name,
				AvatarURL: avatarURL,
				Period:    period,
			}
		}
		return repoContributorMap[repo][login]
	}

	// Process commits
	for _, commit := range data.Commits {
		login := commit.Author.Login
		if login == "" {
			continue
		}

		// Normalize login using email mapping (prefer GitHub login over git-derived login)
		if mappedLogin, ok := emailToLogin[commit.Author.Email]; ok {
			login = mappedLogin
		}

		// Also check login-to-login mapping for sanitized logins
		if mappedLogin, ok := loginToLogin[login]; ok {
			login = mappedLogin
		}

		// Initialize contributor if needed
		if _, ok := contributorMap[login]; !ok {
			name := commit.Author.Name
			avatarURL := commit.Author.AvatarURL

			// Use verified info if available (has better name/avatar from GitHub API)
			if info, exists := loginToInfo[login]; exists {
				if info.Name != "" {
					name = info.Name
				}
				if info.AvatarURL != "" {
					avatarURL = info.AvatarURL
				}
			}

			// If still no name, use login as display name
			if name == "" {
				name = login
			}

			contributorMap[login] = &models.ContributorMetrics{
				Login:     login,
				Name:      name,
				AvatarURL: avatarURL,
				Period:    period,
			}
		}

		cm := contributorMap[login]
		cm.CommitCount++
		cm.LinesAdded += commit.Additions
		cm.LinesDeleted += commit.Deletions
		cm.FilesChanged += commit.FilesChanged

		// Update per-repo contributor stats
		rcm := getRepoContributor(commit.Repository, login, cm.Name, cm.AvatarURL)
		rcm.CommitCount++
		rcm.LinesAdded += commit.Additions
		rcm.LinesDeleted += commit.Deletions
		rcm.FilesChanged += commit.FilesChanged

		// Track activity patterns based on commit time
		hour := commit.Date.Hour()
		weekday := commit.Date.Weekday()

		// Early bird: commits before 9am
		if hour >= 5 && hour < 9 {
			cm.EarlyBirdCount++
			rcm.EarlyBirdCount++
		}
		// Night owl: commits after 9pm
		if hour >= 21 || hour < 5 {
			cm.NightOwlCount++
			rcm.NightOwlCount++
		}
		// Nosferatu: commits between midnight and 4am
		if hour >= 0 && hour < 4 {
			cm.MidnightCount++
			rcm.MidnightCount++
		}
		// Weekend warrior
		if weekday == time.Saturday || weekday == time.Sunday {
			cm.WeekendWarrior++
			rcm.WeekendWarrior++
		}

		// Track activity days (global)
		if activityDays[login] == nil {
			activityDays[login] = make(map[string]bool)
		}
		dateStr := commit.Date.Format("2006-01-02")
		activityDays[login][dateStr] = true

		// Track activity days (per-repo)
		if repoActivityDays[commit.Repository] == nil {
			repoActivityDays[commit.Repository] = make(map[string]map[string]bool)
		}
		if repoActivityDays[commit.Repository][login] == nil {
			repoActivityDays[commit.Repository][login] = make(map[string]bool)
		}
		repoActivityDays[commit.Repository][login][dateStr] = true

		// Track repository participation
		if !contains(cm.RepositoriesContributed, commit.Repository) {
			cm.RepositoriesContributed = append(cm.RepositoriesContributed, commit.Repository)
		}

		// Update repository metrics
		a.updateRepoMetrics(repoMap, commit.Repository, period)
		rm := repoMap[commit.Repository]
		rm.TotalCommits++
		rm.TotalLinesAdded += commit.Additions
		rm.TotalLinesDeleted += commit.Deletions
	}

	// Calculate active days and streaks for each contributor
	for login, days := range activityDays {
		if cm, ok := contributorMap[login]; ok {
			cm.ActiveDays = len(days)
			cm.LongestStreak, cm.CurrentStreak = calculateStreaks(days)
		}
	}

	// Track PRs with changes requested per contributor
	prChangesRequested := make(map[string]map[int]bool) // login -> set of PR numbers with changes requested

	// Process pull requests
	for _, pr := range data.PullRequests {
		login := pr.Author.Login
		if login == "" {
			continue
		}

		// Initialize contributor if needed
		if _, ok := contributorMap[login]; !ok {
			contributorMap[login] = &models.ContributorMetrics{
				Login:     login,
				Name:      pr.Author.Name,
				AvatarURL: pr.Author.AvatarURL,
				Period:    period,
			}
		}

		cm := contributorMap[login]
		cm.PRsOpened++

		// Get per-repo contributor
		rcm := getRepoContributor(pr.Repository, login, cm.Name, cm.AvatarURL)
		rcm.PRsOpened++

		prSize := pr.Additions + pr.Deletions

		if pr.IsMerged() {
			cm.PRsMerged++
			rcm.PRsMerged++
			if pr.TimeToMerge != nil {
				// Accumulate for average calculation
				cm.AvgTimeToMerge += pr.TimeToMerge.Hours()
				rcm.AvgTimeToMerge += pr.TimeToMerge.Hours()
			}

			// Track largest PR
			if prSize > cm.LargestPRSize {
				cm.LargestPRSize = prSize
			}
			if prSize > rcm.LargestPRSize {
				rcm.LargestPRSize = prSize
			}

			// Track small PRs (under 100 lines - good practice)
			if prSize < 100 {
				cm.SmallPRCount++
				rcm.SmallPRCount++
			}
		} else if pr.State == models.PRStateClosed {
			cm.PRsClosed++
			rcm.PRsClosed++
		}

		// Track repository participation
		if !contains(cm.RepositoriesContributed, pr.Repository) {
			cm.RepositoriesContributed = append(cm.RepositoriesContributed, pr.Repository)
		}

		// Update repository metrics
		a.updateRepoMetrics(repoMap, pr.Repository, period)
		rm := repoMap[pr.Repository]
		rm.TotalPRs++
	}

	// Process reviews
	reviewerReviewees := make(map[string]map[string]bool) // reviewer -> set of reviewees
	for _, review := range data.Reviews {
		login := review.Author.Login
		if login == "" {
			continue
		}

		// Initialize contributor if needed
		if _, ok := contributorMap[login]; !ok {
			contributorMap[login] = &models.ContributorMetrics{
				Login:  login,
				Period: period,
			}
		}

		cm := contributorMap[login]
		cm.ReviewsGiven++
		cm.ReviewComments += review.CommentsCount

		// Get per-repo contributor
		rcm := getRepoContributor(review.Repository, login, cm.Name, cm.AvatarURL)
		rcm.ReviewsGiven++
		rcm.ReviewComments += review.CommentsCount

		if review.IsApproval() {
			cm.ApprovalsGiven++
			rcm.ApprovalsGiven++
		} else if review.RequestsChanges() {
			cm.ChangesRequested++
			rcm.ChangesRequested++

			// Track which PRs had changes requested (for calculating "perfect PRs" for the PR author)
			for _, pr := range data.PullRequests {
				if pr.Number == review.PullRequest && pr.Repository == review.Repository {
					prAuthor := pr.Author.Login
					if prChangesRequested[prAuthor] == nil {
						prChangesRequested[prAuthor] = make(map[int]bool)
					}
					prChangesRequested[prAuthor][pr.Number] = true
					break
				}
			}
		}

		if review.ResponseTime != nil {
			cm.AvgReviewTime += review.ResponseTime.Hours()
			rcm.AvgReviewTime += review.ResponseTime.Hours()
		}

		// Track unique reviewees
		if reviewerReviewees[login] == nil {
			reviewerReviewees[login] = make(map[string]bool)
		}

		// Find PR author (reviewee)
		for _, pr := range data.PullRequests {
			if pr.Number == review.PullRequest && pr.Repository == review.Repository {
				reviewerReviewees[login][pr.Author.Login] = true
				break
			}
		}

		// Update repository metrics
		a.updateRepoMetrics(repoMap, review.Repository, period)
		rm := repoMap[review.Repository]
		rm.TotalReviews++
	}

	// Calculate perfect PRs (merged PRs without changes requested) for each contributor
	for login, cm := range contributorMap {
		changesRequestedPRs := prChangesRequested[login]
		// Count merged PRs that didn't have changes requested
		for _, pr := range data.PullRequests {
			if pr.Author.Login == login && pr.IsMerged() {
				if changesRequestedPRs == nil || !changesRequestedPRs[pr.Number] {
					cm.PerfectPRs++
				}
			}
		}
	}

	// Process issues
	for _, issue := range data.Issues {
		login := issue.Author.Login
		if login == "" {
			continue
		}

		// Initialize contributor if needed
		if _, ok := contributorMap[login]; !ok {
			contributorMap[login] = &models.ContributorMetrics{
				Login:  login,
				Period: period,
			}
		}

		cm := contributorMap[login]
		cm.IssuesOpened++

		if issue.IsClosed() && issue.ClosedBy != nil && issue.ClosedBy.Login == login {
			cm.IssuesClosed++
		}

		// Track repository participation
		if !contains(cm.RepositoriesContributed, issue.Repository) {
			cm.RepositoriesContributed = append(cm.RepositoriesContributed, issue.Repository)
		}
	}

	// Calculate averages and finalize contributor metrics
	for login, cm := range contributorMap {
		// Calculate average time to merge
		if cm.PRsMerged > 0 {
			cm.AvgTimeToMerge = cm.AvgTimeToMerge / float64(cm.PRsMerged)
		}

		// Calculate average review time
		if cm.ReviewsGiven > 0 {
			cm.AvgReviewTime = cm.AvgReviewTime / float64(cm.ReviewsGiven)
		}

		// Calculate average PR size
		if cm.PRsOpened > 0 {
			totalPRLines := 0
			for _, pr := range data.PullRequests {
				if pr.Author.Login == login {
					totalPRLines += pr.TotalChanges()
				}
			}
			cm.AvgPRSize = float64(totalPRLines) / float64(cm.PRsOpened)
		}

		// Set unique reviewees count
		if reviewees, ok := reviewerReviewees[login]; ok {
			cm.UniqueReviewees = len(reviewees)
		}
	}

	// Convert maps to slices
	var contributors []models.ContributorMetrics
	for _, cm := range contributorMap {
		contributors = append(contributors, *cm)
	}

	// Sort contributors by commit count
	sort.Slice(contributors, func(i, j int) bool {
		return contributors[i].CommitCount > contributors[j].CommitCount
	})

	// Calculate per-repo contributor averages and streaks
	for repo, repoContribs := range repoContributorMap {
		// Calculate active days and streaks for per-repo contributors
		if repoDays, ok := repoActivityDays[repo]; ok {
			for login, days := range repoDays {
				if rcm, ok := repoContribs[login]; ok {
					rcm.ActiveDays = len(days)
					rcm.LongestStreak, rcm.CurrentStreak = calculateStreaks(days)
				}
			}
		}

		// Calculate averages for per-repo contributors
		for login, rcm := range repoContribs {
			if rcm.PRsMerged > 0 {
				rcm.AvgTimeToMerge = rcm.AvgTimeToMerge / float64(rcm.PRsMerged)
			}
			if rcm.ReviewsGiven > 0 {
				rcm.AvgReviewTime = rcm.AvgReviewTime / float64(rcm.ReviewsGiven)
			}

			// Calculate average PR size for this repo
			if rcm.PRsOpened > 0 {
				totalPRLines := 0
				for _, pr := range data.PullRequests {
					if pr.Author.Login == login && pr.Repository == repo {
						totalPRLines += pr.TotalChanges()
					}
				}
				rcm.AvgPRSize = float64(totalPRLines) / float64(rcm.PRsOpened)
			}

			// Calculate perfect PRs for this repo
			for _, pr := range data.PullRequests {
				if pr.Author.Login == login && pr.Repository == repo && pr.IsMerged() {
					changesRequestedPRs := prChangesRequested[login]
					if changesRequestedPRs == nil || !changesRequestedPRs[pr.Number] {
						rcm.PerfectPRs++
					}
				}
			}
		}
	}

	var repositories []models.RepositoryMetrics
	for _, rm := range repoMap {
		// Add per-repo contributors (with repo-specific stats)
		if repoContribs, ok := repoContributorMap[rm.FullName]; ok {
			for _, rcm := range repoContribs {
				rm.Contributors = append(rm.Contributors, *rcm)
			}
		}
		// Sort contributors by commit count
		sort.Slice(rm.Contributors, func(i, j int) bool {
			return rm.Contributors[i].CommitCount > rm.Contributors[j].CommitCount
		})
		rm.ActiveContributors = len(rm.Contributors)
		repositories = append(repositories, *rm)
	}

	// Build team metrics
	var teams []models.TeamMetrics
	for _, teamCfg := range a.config.Teams {
		team := models.TeamMetrics{
			Name:    teamCfg.Name,
			Color:   teamCfg.Color,
			Members: teamCfg.Members,
			Period:  period,
		}

		var totalScore int
		for _, member := range teamCfg.Members {
			if cm, ok := contributorMap[member]; ok {
				team.MemberMetrics = append(team.MemberMetrics, *cm)
				totalScore += cm.Score.Total

				// Aggregate team metrics
				team.AggregatedMetrics.CommitCount += cm.CommitCount
				team.AggregatedMetrics.LinesAdded += cm.LinesAdded
				team.AggregatedMetrics.LinesDeleted += cm.LinesDeleted
				team.AggregatedMetrics.PRsOpened += cm.PRsOpened
				team.AggregatedMetrics.PRsMerged += cm.PRsMerged
				team.AggregatedMetrics.ReviewsGiven += cm.ReviewsGiven
			}
		}

		team.TotalScore = totalScore
		if len(team.MemberMetrics) > 0 {
			team.AvgScore = float64(totalScore) / float64(len(team.MemberMetrics))
		}

		teams = append(teams, team)
	}

	// Calculate totals
	var totalCommits, totalPRs, totalReviews, totalLinesAdded, totalLinesDeleted int
	for _, rm := range repositories {
		totalCommits += rm.TotalCommits
		totalPRs += rm.TotalPRs
		totalReviews += rm.TotalReviews
		totalLinesAdded += rm.TotalLinesAdded
		totalLinesDeleted += rm.TotalLinesDeleted
	}

	// Build velocity timeline (weekly aggregation)
	velocityTimeline := buildVelocityTimeline(data, period, a.config.Scoring)

	return &models.GlobalMetrics{
		Period:            period,
		Repositories:      repositories,
		Teams:             teams,
		TotalContributors: len(contributors),
		TotalCommits:      totalCommits,
		TotalPRs:          totalPRs,
		TotalReviews:      totalReviews,
		TotalLinesAdded:   totalLinesAdded,
		TotalLinesDeleted: totalLinesDeleted,
		VelocityTimeline:  velocityTimeline,
	}, nil
}

func (a *Aggregator) updateRepoMetrics(repoMap map[string]*models.RepositoryMetrics, fullName string, period models.Period) {
	if _, ok := repoMap[fullName]; !ok {
		owner, name := parseRepoName(fullName)
		repoMap[fullName] = &models.RepositoryMetrics{
			Owner:    owner,
			Name:     name,
			FullName: fullName,
			Period:   period,
		}
	}
}

func parseRepoName(fullName string) (owner, name string) {
	for i, c := range fullName {
		if c == '/' {
			return fullName[:i], fullName[i+1:]
		}
	}
	return fullName, ""
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// normalizeForComparison normalizes a string for fuzzy comparison
// by lowercasing and removing spaces, hyphens, underscores, dots, and digits
func normalizeForComparison(s string) string {
	var result []rune
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' {
			result = append(result, r)
		}
	}
	return string(result)
}

// buildEmailToLoginMapping creates mappings to normalize authors to GitHub logins
// Strategy:
// 1. Build map of GitHub user ID -> login from PR/review data
// 2. Build map of email -> login from user profiles (fetched from GitHub API)
// 3. Parse GitHub noreply emails (ID+username@users.noreply.github.com) and map via ID
// 4. For each email, collect all author names used with that email
// 5. If ANY name used with an email matches a verified login (case-insensitive), map that email to that login
// 6. Map remaining emails by author name matching
func buildEmailToLoginMapping(data *models.RawData, userProfiles map[string]UserProfile) map[string]string {
	mapping := make(map[string]string)

	// Build map of GitHub user ID -> login info from PR/review data
	idToLogin := make(map[int64]string)
	verifiedLogins := make(map[string]string) // lowercase -> original case
	for _, pr := range data.PullRequests {
		if pr.Author.Login != "" {
			verifiedLogins[strings.ToLower(pr.Author.Login)] = pr.Author.Login
			if pr.Author.ID != 0 {
				idToLogin[pr.Author.ID] = pr.Author.Login
			}
		}
	}
	for _, review := range data.Reviews {
		if review.Author.Login != "" {
			if _, exists := verifiedLogins[strings.ToLower(review.Author.Login)]; !exists {
				verifiedLogins[strings.ToLower(review.Author.Login)] = review.Author.Login
			}
			if review.Author.ID != 0 {
				if _, exists := idToLogin[review.Author.ID]; !exists {
					idToLogin[review.Author.ID] = review.Author.Login
				}
			}
		}
	}

	// Build email -> login mapping from user profiles (public emails from GitHub profiles)
	// This is the most reliable way to match users who have different emails
	profileEmailToLogin := make(map[string]string)
	profileNameToLogin := make(map[string]string)
	for _, profile := range userProfiles {
		if profile.Email != "" {
			profileEmailToLogin[strings.ToLower(profile.Email)] = profile.Login
		}
		// Also map by ID from profile
		if profile.ID != 0 {
			idToLogin[profile.ID] = profile.Login
		}
		// Map by name (for fuzzy matching later)
		if profile.Name != "" {
			profileNameToLogin[strings.ToLower(profile.Name)] = profile.Login
		}
	}

	// First pass: handle GitHub noreply emails via user ID (most reliable)
	// Format: ID+username@users.noreply.github.com
	for _, commit := range data.Commits {
		email := commit.Author.Email
		if email == "" || !strings.Contains(email, "@users.noreply.github.com") {
			continue
		}

		localPart := strings.Split(email, "@")[0]
		var idStr, loginFromEmail string
		if idx := strings.Index(localPart, "+"); idx != -1 {
			idStr = localPart[:idx]
			loginFromEmail = localPart[idx+1:]
		} else {
			// Could be just numeric ID
			idStr = localPart
		}

		// Try to parse numeric ID
		var id int64
		for _, c := range idStr {
			if c >= '0' && c <= '9' {
				id = id*10 + int64(c-'0')
			} else {
				id = 0
				break
			}
		}

		// Map via ID first (most reliable)
		if id != 0 {
			if login, ok := idToLogin[id]; ok {
				mapping[email] = login
				continue
			}
		}

		// Fallback to username from email
		if loginFromEmail != "" {
			mapping[email] = loginFromEmail
		}
	}

	// Second pass: Check commit emails against profile emails (from GitHub API)
	// This handles cases where users have multiple emails (org, personal, etc.)
	for _, commit := range data.Commits {
		email := commit.Author.Email
		if email == "" || mapping[email] != "" {
			continue
		}

		// Check if this email matches any profile's public email
		emailLower := strings.ToLower(email)
		if login, ok := profileEmailToLogin[emailLower]; ok {
			mapping[email] = login
			continue
		}

		// Also check by name against profile names
		if commit.Author.Name != "" {
			nameLower := strings.ToLower(commit.Author.Name)
			if login, ok := profileNameToLogin[nameLower]; ok {
				mapping[email] = login
			}
		}
	}

	// Build email -> set of author names/logins used with that email
	emailToNames := make(map[string]map[string]bool)
	for _, commit := range data.Commits {
		email := commit.Author.Email
		if email == "" {
			continue
		}
		if emailToNames[email] == nil {
			emailToNames[email] = make(map[string]bool)
		}
		if commit.Author.Name != "" {
			emailToNames[email][commit.Author.Name] = true
		}
		if commit.Author.Login != "" {
			emailToNames[email][commit.Author.Login] = true
		}
	}

	// For each email not yet mapped, check if ANY name matches a verified login
	for email, names := range emailToNames {
		if mapping[email] != "" {
			continue
		}
		for name := range names {
			// Clean up name (remove quotes, trim)
			nameLower := strings.ToLower(strings.Trim(name, "\"' "))
			if verifiedLogin, ok := verifiedLogins[nameLower]; ok {
				mapping[email] = verifiedLogin
				break
			}
		}

		// Still not mapped? Try fuzzy matching by normalizing name (removing spaces, hyphens)
		if mapping[email] == "" {
			for name := range names {
				// Normalize: lowercase, remove spaces, hyphens, underscores
				normalized := normalizeForComparison(name)
				for verifiedLower, verifiedLogin := range verifiedLogins {
					if normalized == normalizeForComparison(verifiedLower) {
						mapping[email] = verifiedLogin
						break
					}
				}
				if mapping[email] != "" {
					break
				}
			}
		}

		// Still not mapped? Try extracting email username for matching
		if mapping[email] == "" {
			emailLower := strings.ToLower(email)
			if idx := strings.Index(emailLower, "@"); idx > 0 {
				emailUser := emailLower[:idx]
				// Remove common suffixes like numbers
				emailUserNorm := normalizeForComparison(emailUser)
				for verifiedLower, verifiedLogin := range verifiedLogins {
					verifiedNorm := normalizeForComparison(verifiedLower)
					// Check if email username is similar to verified login
					if emailUserNorm == verifiedNorm || strings.HasPrefix(emailUserNorm, verifiedNorm) || strings.HasPrefix(verifiedNorm, emailUserNorm) {
						mapping[email] = verifiedLogin
						break
					}
				}
			}
		}
	}

	// Build name-to-login mapping for remaining matches
	nameToLogin := make(map[string]string)
	for _, pr := range data.PullRequests {
		if pr.Author.Login != "" {
			if pr.Author.Name != "" {
				nameToLogin[strings.ToLower(pr.Author.Name)] = pr.Author.Login
			}
			nameToLogin[strings.ToLower(pr.Author.Login)] = pr.Author.Login
		}
	}
	for _, review := range data.Reviews {
		if review.Author.Login != "" {
			if review.Author.Name != "" {
				if _, exists := nameToLogin[strings.ToLower(review.Author.Name)]; !exists {
					nameToLogin[strings.ToLower(review.Author.Name)] = review.Author.Login
				}
			}
			if _, exists := nameToLogin[strings.ToLower(review.Author.Login)]; !exists {
				nameToLogin[strings.ToLower(review.Author.Login)] = review.Author.Login
			}
		}
	}

	// Also add name mappings from GitHub noreply emails
	for _, commit := range data.Commits {
		if mapping[commit.Author.Email] != "" && commit.Author.Name != "" {
			nameToLogin[strings.ToLower(commit.Author.Name)] = mapping[commit.Author.Email]
		}
	}

	// Final pass: map remaining emails by author name
	for _, commit := range data.Commits {
		email := commit.Author.Email
		if email == "" || mapping[email] != "" {
			continue
		}

		// Try to find by name (case-insensitive)
		if login, ok := nameToLogin[strings.ToLower(commit.Author.Name)]; ok {
			mapping[email] = login
		}
	}

	return mapping
}

// loginInfo stores verified GitHub login info
type loginInfo struct {
	Login     string
	Name      string
	AvatarURL string
}

// buildLoginMapping converts potentially sanitized logins to real GitHub logins
// using known mappings from PR/review data, and returns avatar URLs
func buildLoginMapping(data *models.RawData) (map[string]string, map[string]loginInfo) {
	loginMapping := make(map[string]string)
	nameToLoginInfo := make(map[string]loginInfo)
	loginToInfo := make(map[string]loginInfo)
	idToLoginInfo := make(map[int64]loginInfo) // Map GitHub user ID to login info

	// Collect verified GitHub logins from PRs and reviews
	for _, pr := range data.PullRequests {
		if pr.Author.Login != "" {
			info := loginInfo{
				Login:     pr.Author.Login,
				Name:      pr.Author.Name,
				AvatarURL: pr.Author.AvatarURL,
			}
			loginToInfo[pr.Author.Login] = info
			if pr.Author.ID != 0 {
				idToLoginInfo[pr.Author.ID] = info
			}
			if pr.Author.Name != "" {
				nameToLoginInfo[strings.ToLower(pr.Author.Name)] = info
			}
		}
	}
	for _, review := range data.Reviews {
		if review.Author.Login != "" {
			// Only set if not already set (PRs have higher priority)
			if _, exists := loginToInfo[review.Author.Login]; !exists {
				info := loginInfo{
					Login:     review.Author.Login,
					Name:      review.Author.Name,
					AvatarURL: review.Author.AvatarURL,
				}
				loginToInfo[review.Author.Login] = info
				if review.Author.ID != 0 {
					if _, exists := idToLoginInfo[review.Author.ID]; !exists {
						idToLoginInfo[review.Author.ID] = info
					}
				}
				if review.Author.Name != "" {
					if _, exists := nameToLoginInfo[strings.ToLower(review.Author.Name)]; !exists {
						nameToLoginInfo[strings.ToLower(review.Author.Name)] = info
					}
				}
			}
		}
	}

	// Build email-to-verifiedLogin mapping from commits with noreply emails
	// This helps link personal commits to verified GitHub users
	emailToVerified := make(map[string]string)
	for _, commit := range data.Commits {
		email := commit.Author.Email
		if email == "" || !strings.Contains(email, "@users.noreply.github.com") {
			continue
		}
		localPart := strings.Split(email, "@")[0]
		var login string
		if idx := strings.Index(localPart, "+"); idx != -1 {
			login = localPart[idx+1:]
		} else {
			login = localPart
		}
		if login != "" {
			// Map this author's name to verified login
			if commit.Author.Name != "" {
				nameToLoginInfo[strings.ToLower(commit.Author.Name)] = loginInfo{Login: login}
			}
		}
	}
	_ = emailToVerified // suppress unused warning

	// Build a name-to-commit-login map from commits (for reverse lookup)
	// This helps map PR logins (no name) back to commit logins (has name)
	commitNameToLogin := make(map[string]string)
	for _, commit := range data.Commits {
		if commit.Author.Name != "" && commit.Author.Login != "" {
			nameLower := strings.ToLower(commit.Author.Name)
			// Only set if not already a verified login
			if _, isVerified := loginToInfo[commit.Author.Login]; !isVerified {
				if existing, exists := commitNameToLogin[nameLower]; !exists || len(commit.Author.Login) < len(existing) {
					commitNameToLogin[nameLower] = commit.Author.Login
				}
			}
		}
	}

	// For each commit, check if its login can be mapped to a verified login
	for _, commit := range data.Commits {
		commitLogin := commit.Author.Login
		if commitLogin == "" {
			continue
		}

		// If the commit login already matches a verified login, skip
		if _, exists := loginToInfo[commitLogin]; exists {
			continue
		}

		// Already mapped?
		if _, exists := loginMapping[commitLogin]; exists {
			continue
		}

		// Strategy 1 (BEST): Try to map via GitHub user ID from noreply email
		// Format: ID+username@users.noreply.github.com or just ID@users.noreply.github.com
		if commit.Author.Email != "" && strings.Contains(commit.Author.Email, "@users.noreply.github.com") {
			localPart := strings.Split(commit.Author.Email, "@")[0]
			// Try to extract numeric ID from start of local part
			var idStr string
			if idx := strings.Index(localPart, "+"); idx != -1 {
				idStr = localPart[:idx]
			} else {
				// Might be just the ID without username
				idStr = localPart
			}

			// Parse ID and look up
			var id int64
			for _, c := range idStr {
				if c >= '0' && c <= '9' {
					id = id*10 + int64(c-'0')
				} else {
					id = 0
					break
				}
			}

			if id != 0 {
				if info, ok := idToLoginInfo[id]; ok {
					if commitLogin != info.Login {
						loginMapping[commitLogin] = info.Login
						continue
					}
				}
			}
		}

		// Strategy 2: Try to map via author name
		if commit.Author.Name != "" {
			if info, ok := nameToLoginInfo[strings.ToLower(commit.Author.Name)]; ok {
				if commitLogin != info.Login {
					loginMapping[commitLogin] = info.Login
					continue
				}
			}
		}

		// Strategy 3: Check if commitLogin is a sanitized version of any verified login
		// e.g., "lukasz-raczylo" might be sanitized from "lukaszraczylo"
		// Compare by removing hyphens and lowercasing
		sanitizedCommit := strings.ToLower(strings.ReplaceAll(commitLogin, "-", ""))
		for verifiedLogin := range loginToInfo {
			sanitizedVerified := strings.ToLower(strings.ReplaceAll(verifiedLogin, "-", ""))
			if sanitizedCommit == sanitizedVerified && commitLogin != verifiedLogin {
				loginMapping[commitLogin] = verifiedLogin
				break
			}
		}
	}

	// Strategy 4: For each commit name, find if a different commit login (hyphenated)
	// can be mapped to the verified login via sanitized comparison
	// This catches cases missed by the main loop
	for _, commitLogin := range commitNameToLogin {
		if _, exists := loginToInfo[commitLogin]; exists {
			// This commit login is already verified, skip
			continue
		}
		if _, exists := loginMapping[commitLogin]; exists {
			// Already mapped
			continue
		}

		// Check if removing hyphens matches a verified login
		sanitizedCommit := strings.ToLower(strings.ReplaceAll(commitLogin, "-", ""))
		for verifiedLogin := range loginToInfo {
			sanitizedVerified := strings.ToLower(strings.ReplaceAll(verifiedLogin, "-", ""))
			if sanitizedCommit == sanitizedVerified && commitLogin != verifiedLogin {
				loginMapping[commitLogin] = verifiedLogin
				break
			}
		}
	}

	return loginMapping, loginToInfo
}

// buildVelocityTimeline creates weekly aggregated velocity data for trend visualization
func buildVelocityTimeline(data *models.RawData, period models.Period, scoringConfig config.ScoringConfig) *models.VelocityTimeline {
	// Determine date range
	start := period.Start
	end := period.End

	// Ensure we have valid dates
	if start.IsZero() {
		// Default to 90 days ago
		start = time.Now().AddDate(0, 0, -90)
	}
	if end.IsZero() {
		end = time.Now()
	}

	// Calculate week boundaries (start from Monday of the first week)
	// Go back to the Monday of the start week
	weekday := int(start.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	weekStart := start.AddDate(0, 0, -(weekday - 1))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())

	// Build list of weeks
	var weeks []time.Time
	for w := weekStart; w.Before(end) || w.Equal(end); w = w.AddDate(0, 0, 7) {
		weeks = append(weeks, w)
	}

	if len(weeks) == 0 {
		return nil
	}

	// Initialize counters for each week
	weekCommits := make([]float64, len(weeks))
	weekPRs := make([]float64, len(weeks))
	weekReviews := make([]float64, len(weeks))
	weekScore := make([]float64, len(weeks))

	// Helper to find week index for a date
	findWeekIndex := func(t time.Time) int {
		for i := len(weeks) - 1; i >= 0; i-- {
			if !t.Before(weeks[i]) {
				return i
			}
		}
		return 0
	}

	// Get scoring points from config (defaults are in PointsConfig struct)
	pointsCommit := scoringConfig.Points.Commit
	pointsPROpened := scoringConfig.Points.PROpened
	pointsPRMerged := scoringConfig.Points.PRMerged
	pointsReview := scoringConfig.Points.PRReviewed

	// Use defaults if zero
	if pointsCommit == 0 {
		pointsCommit = 10
	}
	if pointsPROpened == 0 {
		pointsPROpened = 25
	}
	if pointsPRMerged == 0 {
		pointsPRMerged = 50
	}
	if pointsReview == 0 {
		pointsReview = 30
	}

	// Aggregate commits by week
	for _, commit := range data.Commits {
		if commit.Date.Before(start) || commit.Date.After(end) {
			continue
		}
		idx := findWeekIndex(commit.Date)
		if idx >= 0 && idx < len(weeks) {
			weekCommits[idx]++
			weekScore[idx] += float64(pointsCommit)
		}
	}

	// Aggregate PRs by week (use merged date if available, otherwise created date)
	for _, pr := range data.PullRequests {
		prDate := pr.CreatedAt
		if pr.MergedAt != nil {
			prDate = *pr.MergedAt
		}
		if prDate.Before(start) || prDate.After(end) {
			continue
		}
		idx := findWeekIndex(prDate)
		if idx >= 0 && idx < len(weeks) {
			weekPRs[idx]++
			if pr.IsMerged() {
				weekScore[idx] += float64(pointsPRMerged)
			} else {
				weekScore[idx] += float64(pointsPROpened)
			}
		}
	}

	// Aggregate reviews by week
	for _, review := range data.Reviews {
		if review.SubmittedAt.Before(start) || review.SubmittedAt.After(end) {
			continue
		}
		idx := findWeekIndex(review.SubmittedAt)
		if idx >= 0 && idx < len(weeks) {
			weekReviews[idx]++
			weekScore[idx] += float64(pointsReview)
		}
	}

	// Build labels (format: "Jan 2")
	labels := make([]string, len(weeks))
	for i, w := range weeks {
		labels[i] = w.Format("Jan 2")
	}

	return &models.VelocityTimeline{
		Labels: labels,
		Series: []models.VelocityTimelineSeries{
			{Name: "Commits", Color: "#10b981", Data: weekCommits},
			{Name: "PRs", Color: "#3b82f6", Data: weekPRs},
			{Name: "Reviews", Color: "#8b5cf6", Data: weekReviews},
			{Name: "Score", Color: "#f59e0b", Data: weekScore},
		},
	}
}

// calculateStreaks calculates the longest and current streak of consecutive days
func calculateStreaks(days map[string]bool) (longest, current int) {
	if len(days) == 0 {
		return 0, 0
	}

	// Convert to sorted slice of dates
	dates := make([]time.Time, 0, len(days))
	for dateStr := range days {
		t, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			dates = append(dates, t)
		}
	}

	if len(dates) == 0 {
		return 0, 0
	}

	// Sort dates
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	// Calculate streaks
	longest = 1
	current = 1
	streak := 1

	for i := 1; i < len(dates); i++ {
		diff := dates[i].Sub(dates[i-1]).Hours() / 24
		if diff == 1 {
			streak++
			if streak > longest {
				longest = streak
			}
		} else {
			streak = 1
		}
	}

	// Check if current streak is still active (last activity was today or yesterday)
	today := time.Now().Truncate(24 * time.Hour)
	lastActive := dates[len(dates)-1]
	daysSinceLastActive := today.Sub(lastActive).Hours() / 24

	if daysSinceLastActive <= 1 {
		current = streak
	} else {
		current = 0
	}

	return longest, current
}
