package github

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v68/github"

	"github.com/lukaszraczylo/git-velocity/internal/config"
	"github.com/lukaszraczylo/git-velocity/internal/diff"
	"github.com/lukaszraczylo/git-velocity/internal/domain/models"
	"github.com/lukaszraczylo/git-velocity/internal/github/cache"
)

// ProgressCallback is called to report progress during API operations
type ProgressCallback func(message string)

// RetryConfig holds retry settings
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
	}
}

// Client wraps the GitHub API client with rate limiting and caching
type Client struct {
	gh       *github.Client
	config   *config.Config
	cache    cache.Cache
	retry    RetryConfig
	progress ProgressCallback
}

// NewClient creates a new GitHub client with the appropriate authentication
func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	var gh *github.Client

	// Determine authentication method
	if cfg.HasGithubToken() {
		gh = github.NewClient(nil).WithAuthToken(cfg.Auth.GithubToken)
	} else if cfg.HasGithubApp() {
		// GitHub App authentication
		privateKey, err := cfg.GetGithubAppPrivateKey()
		if err != nil {
			return nil, fmt.Errorf("failed to get GitHub App private key: %w", err)
		}

		itr, err := ghinstallation.New(
			http.DefaultTransport,
			cfg.Auth.GithubApp.AppID,
			cfg.Auth.GithubApp.InstallationID,
			privateKey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub App transport: %w", err)
		}

		gh = github.NewClient(&http.Client{Transport: itr})
	} else {
		return nil, fmt.Errorf("no authentication method configured")
	}

	// Initialize cache
	var c cache.Cache
	if cfg.Cache.Enabled {
		ttl, err := cfg.GetCacheTTL()
		if err != nil {
			return nil, fmt.Errorf("failed to parse cache TTL: %w", err)
		}
		c, err = cache.NewFileCache(cfg.Cache.Directory, ttl)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize cache: %w", err)
		}
	} else {
		c = cache.NewNoopCache()
	}

	return &Client{
		gh:       gh,
		config:   cfg,
		cache:    c,
		retry:    DefaultRetryConfig(),
		progress: func(string) {}, // no-op by default
	}, nil
}

// SetProgressCallback sets the callback function for progress reporting
func (c *Client) SetProgressCallback(cb ProgressCallback) {
	if cb != nil {
		c.progress = cb
	}
}

// SetRetryConfig sets the retry configuration
func (c *Client) SetRetryConfig(rc RetryConfig) {
	c.retry = rc
}

// retryWithBackoff executes a function with retry logic
// - For rate limit errors: waits until the limit resets (no retry count limit)
// - For network/transient errors: uses exponential backoff with MaxRetries limit
func (c *Client) retryWithBackoff(ctx context.Context, operation string, fn func() error) error {
	var lastErr error
	backoff := c.retry.InitialBackoff
	networkRetries := 0

	for {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Check if error is retryable at all
		if !isRetryableError(lastErr) {
			return lastErr
		}

		c.progress(fmt.Sprintf("      %s failed: %v", operation, lastErr))

		// Determine wait strategy based on error type
		if resetTime := getRateLimitResetTime(lastErr); resetTime != nil {
			// Rate limit error - wait until reset, no retry count limit
			waitDuration := time.Until(*resetTime) + time.Second // Add 1s buffer
			if waitDuration < 0 {
				waitDuration = time.Second
			}
			c.progress(fmt.Sprintf("      Rate limit hit. Waiting until %s (%s)...", resetTime.Format("15:04:05"), waitDuration.Round(time.Second)))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitDuration):
			}
			// Reset network retry counter after successful rate limit wait
			networkRetries = 0
			backoff = c.retry.InitialBackoff
		} else {
			// Network/transient error - use exponential backoff with retry limit
			networkRetries++
			if networkRetries > c.retry.MaxRetries {
				return fmt.Errorf("%s failed after %d retries: %w", operation, c.retry.MaxRetries, lastErr)
			}

			c.progress(fmt.Sprintf("      Retry %d/%d for %s (waiting %s)...", networkRetries, c.retry.MaxRetries, operation, backoff))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}

			backoff *= 2
			if backoff > c.retry.MaxBackoff {
				backoff = c.retry.MaxBackoff
			}
		}
	}
}

// getRateLimitResetTime extracts the reset time from rate limit errors
func getRateLimitResetTime(err error) *time.Time {
	if err == nil {
		return nil
	}

	var rateLimitErr *github.RateLimitError
	if errors.As(err, &rateLimitErr) && rateLimitErr.Rate.Reset.Time.After(time.Now()) {
		t := rateLimitErr.Rate.Reset.Time
		return &t
	}

	var abuseErr *github.AbuseRateLimitError
	if errors.As(err, &abuseErr) && abuseErr.RetryAfter != nil {
		t := time.Now().Add(*abuseErr.RetryAfter)
		return &t
	}

	return nil
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Network errors (timeout only - Temporary() is deprecated)
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	// GitHub rate limit errors
	var rateLimitErr *github.RateLimitError
	if errors.As(err, &rateLimitErr) {
		return true
	}

	// GitHub abuse rate limit errors
	var abuseErr *github.AbuseRateLimitError
	if errors.As(err, &abuseErr) {
		return true
	}

	// Check error message for common transient errors
	errStr := err.Error()
	retryableMessages := []string{
		"connection reset",
		"connection refused",
		"timeout",
		"temporary failure",
		"server error",
		"502",
		"503",
		"504",
	}
	for _, msg := range retryableMessages {
		if strings.Contains(strings.ToLower(errStr), msg) {
			return true
		}
	}

	return false
}

// ListOrgRepos lists repositories in an organization matching a pattern
func (c *Client) ListOrgRepos(ctx context.Context, org, pattern string) ([]string, error) {
	var allRepos []string

	opts := &github.RepositoryListByOrgOptions{
		Type: "all",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := c.gh.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list org repos: %w", err)
		}

		for _, repo := range repos {
			name := repo.GetName()
			if matchPattern(name, pattern) {
				allRepos = append(allRepos, name)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// FetchCommits fetches commits from a repository within a date range
func (c *Client) FetchCommits(ctx context.Context, owner, repo string, since, until *time.Time) ([]models.Commit, error) {
	cacheKey := fmt.Sprintf("commits:%s/%s:%v:%v", owner, repo, since, until)

	// Check cache
	if cached, ok := c.cache.Get(cacheKey); ok {
		if commits, ok := cached.([]models.Commit); ok {
			c.progress("      Using cached commits data")
			return commits, nil
		}
	}

	var allCommits []models.Commit

	opts := &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	if since != nil {
		opts.Since = *since
	}
	if until != nil {
		opts.Until = *until
	}

	page := 1
	for {
		var commits []*github.RepositoryCommit
		var resp *github.Response

		err := c.retryWithBackoff(ctx, "list commits", func() error {
			var err error
			commits, resp, err = c.gh.Repositories.ListCommits(ctx, owner, repo, opts)
			return err
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list commits: %w", err)
		}

		c.progress(fmt.Sprintf("      Fetching commits page %d (%d commits so far)...", page, len(allCommits)))

		for i, commit := range commits {
			// Fetch detailed commit info for stats
			var detailed *github.RepositoryCommit
			err := c.retryWithBackoff(ctx, fmt.Sprintf("get commit %s", commit.GetSHA()[:7]), func() error {
				var err error
				detailed, _, err = c.gh.Repositories.GetCommit(ctx, owner, repo, commit.GetSHA(), nil)
				return err
			})
			if err != nil {
				// Log and continue - we can still use basic info
				c.progress(fmt.Sprintf("      Warning: failed to get commit details for %s: %v", commit.GetSHA()[:7], err))
				continue
			}

			mc := convertCommit(detailed, owner, repo)
			allCommits = append(allCommits, mc)

			// Progress every 10 commits
			if (i+1)%10 == 0 {
				c.progress(fmt.Sprintf("      Processing commit %d/%d on page %d...", i+1, len(commits), page))
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
		page++
	}

	// Cache results
	c.cache.Set(cacheKey, allCommits)

	return allCommits, nil
}

// mainBranches are the branches we consider as "main" branches
var mainBranches = []string{"main", "master", "develop", "dev"}

// FetchPullRequests fetches pull requests from a repository
// Fetches PRs targeting main branches, filters by merge date
func (c *Client) FetchPullRequests(ctx context.Context, owner, repo string, since, until *time.Time) ([]models.PullRequest, error) {
	cacheKey := fmt.Sprintf("prs:%s/%s:%v:%v", owner, repo, since, until)

	// Check cache
	if cached, ok := c.cache.Get(cacheKey); ok {
		if prs, ok := cached.([]models.PullRequest); ok {
			c.progress("      Using cached pull requests data")
			return prs, nil
		}
	}

	var allPRs []models.PullRequest

	// Fetch PRs for each main branch separately (API supports base filter)
	for _, baseBranch := range mainBranches {
		prs, err := c.fetchPRsForBranch(ctx, owner, repo, baseBranch, since, until)
		if err != nil {
			// Branch might not exist, skip
			continue
		}
		allPRs = append(allPRs, prs...)
	}

	c.progress(fmt.Sprintf("      Found %d merged PRs to main branches in date range", len(allPRs)))

	// Cache results
	c.cache.Set(cacheKey, allPRs)

	return allPRs, nil
}

// fetchPRsForBranch fetches merged PRs for a specific base branch
func (c *Client) fetchPRsForBranch(ctx context.Context, owner, repo, baseBranch string, since, until *time.Time) ([]models.PullRequest, error) {
	var branchPRs []models.PullRequest

	opts := &github.PullRequestListOptions{
		State:     "closed",
		Base:      baseBranch, // Filter by base branch at API level
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	page := 1
	consecutiveOldPages := 0

	for {
		var prs []*github.PullRequest
		var resp *github.Response

		err := c.retryWithBackoff(ctx, "list pull requests", func() error {
			var err error
			prs, resp, err = c.gh.PullRequests.List(ctx, owner, repo, opts)
			return err
		})
		if err != nil {
			return branchPRs, err
		}

		if page == 1 && len(prs) > 0 {
			c.progress(fmt.Sprintf("      Fetching PRs for branch '%s'...", baseBranch))
		}

		matchedInPage := 0
		oldInPage := 0

		for _, pr := range prs {
			// Only consider merged PRs (check MergedAt since Merged field isn't in list response)
			if pr.MergedAt == nil {
				continue
			}

			// Use merge date for filtering
			mergedAt := pr.MergedAt.Time

			// Skip items newer than our range
			if until != nil && mergedAt.After(*until) {
				continue
			}

			// If older than our range, track it
			if since != nil && mergedAt.Before(*since) {
				oldInPage++
				continue
			}

			mpr := convertPullRequest(pr, owner, repo)
			branchPRs = append(branchPRs, mpr)
			matchedInPage++
		}

		// Early termination: if we got a page with only old PRs (or empty), increment counter
		if matchedInPage == 0 && oldInPage > 0 {
			consecutiveOldPages++
			// Stop after 2 consecutive pages of only old PRs
			if consecutiveOldPages >= 2 {
				break
			}
		} else {
			consecutiveOldPages = 0
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
		page++
	}

	return branchPRs, nil
}

// FetchReviews fetches reviews for a specific pull request
func (c *Client) FetchReviews(ctx context.Context, owner, repo string, prNumber int) ([]models.Review, error) {
	cacheKey := fmt.Sprintf("reviews:%s/%s:%d", owner, repo, prNumber)

	// Check cache
	if cached, ok := c.cache.Get(cacheKey); ok {
		if reviews, ok := cached.([]models.Review); ok {
			return reviews, nil
		}
	}

	var allReviews []models.Review

	opts := &github.ListOptions{PerPage: 100}

	for {
		var reviews []*github.PullRequestReview
		var resp *github.Response

		err := c.retryWithBackoff(ctx, fmt.Sprintf("list reviews for PR #%d", prNumber), func() error {
			var err error
			reviews, resp, err = c.gh.PullRequests.ListReviews(ctx, owner, repo, prNumber, opts)
			return err
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list reviews: %w", err)
		}

		for _, review := range reviews {
			mr := convertReview(review, owner, repo, prNumber)
			allReviews = append(allReviews, mr)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Cache results
	c.cache.Set(cacheKey, allReviews)

	return allReviews, nil
}

// FetchIssues fetches issues from a repository
// Uses early termination when sorted by date - stops when items are outside date range
func (c *Client) FetchIssues(ctx context.Context, owner, repo string, since, until *time.Time) ([]models.Issue, error) {
	cacheKey := fmt.Sprintf("issues:%s/%s:%v:%v", owner, repo, since, until)

	// Check cache
	if cached, ok := c.cache.Get(cacheKey); ok {
		if issues, ok := cached.([]models.Issue); ok {
			c.progress("      Using cached issues data")
			return issues, nil
		}
	}

	var allIssues []models.Issue

	// Sort by created date descending - newest first
	// This allows us to stop early when we hit items older than our date range
	opts := &github.IssueListByRepoOptions{
		State:     "all",
		Sort:      "created",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	// Note: GitHub Issues API has a 'since' parameter but it filters by update time, not created time
	// So we use our own filtering with early termination for better control

	page := 1
	reachedOldItems := false

	for {
		var issues []*github.Issue
		var resp *github.Response

		err := c.retryWithBackoff(ctx, "list issues", func() error {
			var err error
			issues, resp, err = c.gh.Issues.ListByRepo(ctx, owner, repo, opts)
			return err
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list issues: %w", err)
		}

		c.progress(fmt.Sprintf("      Fetching issues page %d (%d issues so far)...", page, len(allIssues)))

		oldItemsInPage := 0
		totalNonPRItems := 0

		for _, issue := range issues {
			// Skip pull requests (they appear in issues API)
			if issue.PullRequestLinks != nil {
				continue
			}

			totalNonPRItems++
			createdAt := issue.GetCreatedAt().Time

			// Skip items newer than our range (when until is specified)
			if until != nil && createdAt.After(*until) {
				continue
			}

			// If we've gone past our date range (older than since), count it
			if since != nil && createdAt.Before(*since) {
				oldItemsInPage++
				continue
			}

			mi := convertIssue(issue, owner, repo)
			allIssues = append(allIssues, mi)
		}

		// If all non-PR items in this page are older than our range, we can stop
		// (since results are sorted by created date descending)
		if oldItemsInPage == totalNonPRItems && totalNonPRItems > 0 {
			c.progress(fmt.Sprintf("      Reached issues older than date range, stopping early (page %d)", page))
			reachedOldItems = true
			break
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
		page++
	}

	if !reachedOldItems && page > 1 {
		c.progress(fmt.Sprintf("      Fetched all %d pages of issues", page))
	}

	// Cache results
	c.cache.Set(cacheKey, allIssues)

	return allIssues, nil
}

// UserProfile contains GitHub user profile information useful for deduplication
type UserProfile struct {
	ID        int64  // GitHub user ID
	Login     string // GitHub username
	Name      string // Display name
	Email     string // Public email (may be empty)
	AvatarURL string
}

// FetchUserProfiles fetches GitHub profiles for a list of logins
// This is useful for deduplication by getting user IDs, names, and public emails
func (c *Client) FetchUserProfiles(ctx context.Context, logins []string) (map[string]UserProfile, error) {
	profiles := make(map[string]UserProfile)

	// Use semaphore to limit concurrent requests
	sem := make(chan struct{}, 10)
	results := make(chan struct {
		login   string
		profile UserProfile
		err     error
	}, len(logins))

	for _, login := range logins {
		go func(login string) {
			sem <- struct{}{}
			defer func() { <-sem }()

			cacheKey := fmt.Sprintf("user_profile_%s", login)
			if cached, ok := c.cache.Get(cacheKey); ok {
				if profile, ok := cached.(UserProfile); ok {
					results <- struct {
						login   string
						profile UserProfile
						err     error
					}{login, profile, nil}
					return
				}
			}

			var profile UserProfile
			err := c.retryWithBackoff(ctx, "fetch user profile", func() error {
				user, _, err := c.gh.Users.Get(ctx, login)
				if err != nil {
					return err
				}
				profile = UserProfile{
					ID:        user.GetID(),
					Login:     user.GetLogin(),
					Name:      user.GetName(),
					Email:     user.GetEmail(),
					AvatarURL: user.GetAvatarURL(),
				}
				return nil
			})

			if err == nil {
				c.cache.Set(cacheKey, profile)
			}
			results <- struct {
				login   string
				profile UserProfile
				err     error
			}{login, profile, err}
		}(login)
	}

	// Collect results
	for range logins {
		r := <-results
		if r.err == nil {
			profiles[r.login] = r.profile
		}
	}

	return profiles, nil
}

// Helper functions

func convertCommit(c *github.RepositoryCommit, owner, repo string) models.Commit {
	var author models.Author
	if c.Author != nil {
		author = models.Author{
			Login:     c.Author.GetLogin(),
			AvatarURL: c.Author.GetAvatarURL(),
		}
	}
	if c.Commit != nil && c.Commit.Author != nil {
		author.Name = c.Commit.Author.GetName()
		author.Email = c.Commit.Author.GetEmail()
	}

	var committer models.Author
	if c.Committer != nil {
		committer = models.Author{
			Login:     c.Committer.GetLogin(),
			AvatarURL: c.Committer.GetAvatarURL(),
		}
	}
	if c.Commit != nil && c.Commit.Committer != nil {
		committer.Name = c.Commit.Committer.GetName()
		committer.Email = c.Commit.Committer.GetEmail()
	}

	var date time.Time
	if c.Commit != nil && c.Commit.Author != nil {
		date = c.Commit.Author.GetDate().Time
	}

	var additions, deletions, filesChanged int
	if c.Stats != nil {
		additions = c.Stats.GetAdditions()
		deletions = c.Stats.GetDeletions()
	}
	filesChanged = len(c.Files)

	// Detect if commit includes tests and calculate meaningful/comment line counts
	hasTests := false
	var meaningfulAdditions, meaningfulDeletions int
	var commentAdditions, commentDeletions int

	for _, f := range c.Files {
		filename := f.GetFilename()

		// Check for test files
		if strings.Contains(filename, "_test.go") ||
			strings.Contains(filename, ".test.") ||
			strings.Contains(filename, ".spec.") ||
			strings.Contains(filename, "/tests/") ||
			strings.Contains(filename, "/test/") ||
			strings.Contains(filename, "__tests__") {
			hasTests = true
		}

		// Skip documentation files for meaningful line calculation
		if diff.IsDocumentationFile(filename) {
			continue
		}

		// Analyze file patch to get meaningful and comment line counts
		patch := f.GetPatch()
		if patch != "" {
			stats := diff.AnalyzePatch(patch)
			meaningfulAdditions += stats.MeaningfulAdditions
			meaningfulDeletions += stats.MeaningfulDeletions
			commentAdditions += stats.CommentAdditions
			commentDeletions += stats.CommentDeletions
		}
	}

	message := ""
	if c.Commit != nil {
		message = c.Commit.GetMessage()
	}

	return models.Commit{
		SHA:                 c.GetSHA(),
		Message:             message,
		Author:              author,
		Committer:           committer,
		Date:                date,
		Additions:           additions,
		Deletions:           deletions,
		MeaningfulAdditions: meaningfulAdditions,
		MeaningfulDeletions: meaningfulDeletions,
		CommentAdditions:    commentAdditions,
		CommentDeletions:    commentDeletions,
		FilesChanged:        filesChanged,
		Repository:          fmt.Sprintf("%s/%s", owner, repo),
		URL:                 c.GetHTMLURL(),
		HasTests:            hasTests,
	}
}

func convertPullRequest(pr *github.PullRequest, owner, repo string) models.PullRequest {
	var author models.Author
	if pr.User != nil {
		author = models.Author{
			ID:        pr.User.GetID(),
			Login:     pr.User.GetLogin(),
			Name:      pr.User.GetName(),
			AvatarURL: pr.User.GetAvatarURL(),
		}
	}

	state := models.PRStateOpen
	if pr.GetMerged() {
		state = models.PRStateMerged
	} else if pr.GetState() == "closed" {
		state = models.PRStateClosed
	}

	var mergedAt, closedAt *time.Time
	if pr.MergedAt != nil {
		t := pr.MergedAt.Time
		mergedAt = &t
	}
	if pr.ClosedAt != nil {
		t := pr.ClosedAt.Time
		closedAt = &t
	}

	var baseBranch, headBranch string
	if pr.Base != nil {
		baseBranch = pr.Base.GetRef()
	}
	if pr.Head != nil {
		headBranch = pr.Head.GetRef()
	}

	return models.PullRequest{
		Number:       pr.GetNumber(),
		Title:        pr.GetTitle(),
		State:        state,
		Author:       author,
		Repository:   fmt.Sprintf("%s/%s", owner, repo),
		BaseBranch:   baseBranch,
		HeadBranch:   headBranch,
		CreatedAt:    pr.GetCreatedAt().Time,
		UpdatedAt:    pr.GetUpdatedAt().Time,
		MergedAt:     mergedAt,
		ClosedAt:     closedAt,
		Additions:    pr.GetAdditions(),
		Deletions:    pr.GetDeletions(),
		FilesChanged: pr.GetChangedFiles(),
		CommitCount:  pr.GetCommits(),
		Comments:     pr.GetComments() + pr.GetReviewComments(),
		URL:          pr.GetHTMLURL(),
	}
}

func convertReview(r *github.PullRequestReview, owner, repo string, prNumber int) models.Review {
	var author models.Author
	if r.User != nil {
		author = models.Author{
			ID:        r.User.GetID(),
			Login:     r.User.GetLogin(),
			Name:      r.User.GetName(),
			AvatarURL: r.User.GetAvatarURL(),
		}
	}

	state := models.ReviewState(r.GetState())

	submittedAt := time.Time{}
	if r.SubmittedAt != nil {
		submittedAt = r.SubmittedAt.Time
	}

	return models.Review{
		ID:          r.GetID(),
		PullRequest: prNumber,
		Repository:  fmt.Sprintf("%s/%s", owner, repo),
		Author:      author,
		State:       state,
		SubmittedAt: submittedAt,
		Body:        r.GetBody(),
	}
}

func convertIssue(i *github.Issue, owner, repo string) models.Issue {
	var author models.Author
	if i.User != nil {
		author = models.Author{
			Login:     i.User.GetLogin(),
			Name:      i.User.GetName(),
			AvatarURL: i.User.GetAvatarURL(),
		}
	}

	state := models.IssueStateOpen
	if i.GetState() == "closed" {
		state = models.IssueStateClosed
	}

	var closedAt *time.Time
	var closedBy *models.Author
	if i.ClosedAt != nil {
		t := i.ClosedAt.Time
		closedAt = &t
	}
	if i.ClosedBy != nil {
		cb := models.Author{
			Login:     i.ClosedBy.GetLogin(),
			AvatarURL: i.ClosedBy.GetAvatarURL(),
		}
		closedBy = &cb
	}

	var labels []string
	for _, l := range i.Labels {
		labels = append(labels, l.GetName())
	}

	return models.Issue{
		Number:     i.GetNumber(),
		Title:      i.GetTitle(),
		State:      state,
		Author:     author,
		Repository: fmt.Sprintf("%s/%s", owner, repo),
		CreatedAt:  i.GetCreatedAt().Time,
		UpdatedAt:  i.GetUpdatedAt().Time,
		ClosedAt:   closedAt,
		ClosedBy:   closedBy,
		Comments:   i.GetComments(),
		Labels:     labels,
		URL:        i.GetHTMLURL(),
	}
}

// matchPattern performs simple glob-style pattern matching
func matchPattern(s, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Handle exact match
	if !strings.Contains(pattern, "*") {
		return s == pattern
	}

	// Handle prefix match (pattern*)
	if strings.HasSuffix(pattern, "*") && !strings.HasPrefix(pattern, "*") {
		return strings.HasPrefix(s, strings.TrimSuffix(pattern, "*"))
	}

	// Handle suffix match (*pattern)
	if strings.HasPrefix(pattern, "*") && !strings.HasSuffix(pattern, "*") {
		return strings.HasSuffix(s, strings.TrimPrefix(pattern, "*"))
	}

	// Handle contains match (*pattern*)
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		inner := strings.TrimPrefix(strings.TrimSuffix(pattern, "*"), "*")
		return strings.Contains(s, inner)
	}

	return false
}
