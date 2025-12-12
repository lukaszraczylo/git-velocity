package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lukaszraczylo/git-velocity/internal/aggregator"
	"github.com/lukaszraczylo/git-velocity/internal/config"
	"github.com/lukaszraczylo/git-velocity/internal/domain/models"
	"github.com/lukaszraczylo/git-velocity/internal/domain/scoring"
	"github.com/lukaszraczylo/git-velocity/internal/generator/site"
	"github.com/lukaszraczylo/git-velocity/internal/git"
	"github.com/lukaszraczylo/git-velocity/internal/github"
)

// App is the main application orchestrator
type App struct {
	config    *config.Config
	outputDir string
	verbose   bool
	client    *github.Client
	gitRepo   *git.Repository
}

// New creates a new application instance
func New(configPath, outputDir string, verbose bool) (*App, error) {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return &App{
		config:    cfg,
		outputDir: outputDir,
		verbose:   verbose,
	}, nil
}

// Run executes the main application workflow
func (a *App) Run(ctx context.Context) error {
	startTime := time.Now()
	a.log("Starting Git Velocity analysis...")

	// Initialize GitHub client
	a.log("Initializing GitHub client...")
	client, err := github.NewClient(ctx, a.config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}
	a.client = client

	// Set up progress callback
	client.SetProgressCallback(func(msg string) {
		a.log("%s", msg)
	})

	// Initialize local git repository manager (always used for accurate commit data)
	a.log("Initializing local git repository manager...")
	gitRepo, err := git.NewRepository(a.config.Options.CloneDirectory)
	if err != nil {
		return fmt.Errorf("failed to create git repository manager: %w", err)
	}
	gitRepo.SetProgressCallback(func(msg string) {
		a.log("%s", msg)
	})
	a.gitRepo = gitRepo

	// Parse date range
	dateRange, err := a.config.GetParsedDateRange()
	if err != nil {
		return fmt.Errorf("failed to parse date range: %w", err)
	}

	// Collect data from all repositories
	a.log("Fetching data from repositories...")
	rawData, err := a.collectData(ctx, dateRange)
	if err != nil {
		return fmt.Errorf("failed to collect data: %w", err)
	}

	a.log("Collected %d commits, %d PRs, %d reviews, %d issues",
		len(rawData.Commits), len(rawData.PullRequests), len(rawData.Reviews), len(rawData.Issues))

	// Fetch user profiles for better deduplication
	// This gets public emails and names from GitHub profiles to help match commit authors
	a.log("Fetching user profiles for deduplication...")
	userProfiles, err := a.fetchUserProfiles(ctx, rawData)
	if err != nil {
		a.log("Warning: failed to fetch some user profiles: %v", err)
		// Continue anyway, deduplication will still work with other methods
	}
	a.log("Fetched %d user profiles", len(userProfiles))

	// Aggregate metrics
	a.log("Aggregating metrics...")
	agg := aggregator.New(a.config)
	agg.SetUserProfiles(userProfiles)
	globalMetrics, err := agg.Aggregate(rawData, dateRange)
	if err != nil {
		return fmt.Errorf("failed to aggregate metrics: %w", err)
	}

	// Calculate scores
	if a.config.Scoring.Enabled {
		a.log("Calculating scores and achievements...")
		scorer := scoring.NewCalculator(a.config)
		globalMetrics = scorer.Calculate(globalMetrics)
	}

	// Generate the site
	a.log("Generating static site...")
	gen, err := site.NewGenerator(a.outputDir, a.config)
	if err != nil {
		return fmt.Errorf("failed to create site generator: %w", err)
	}

	if err := gen.Generate(globalMetrics); err != nil {
		return fmt.Errorf("failed to generate site: %w", err)
	}

	duration := time.Since(startTime)
	a.log("Analysis complete! Dashboard generated in %s", a.outputDir)
	a.log("Total time: %s", duration.Round(time.Millisecond))

	return nil
}

func (a *App) collectData(ctx context.Context, dateRange *config.ParsedDateRange) (*models.RawData, error) {
	data := &models.RawData{}

	for _, repo := range a.config.Repositories {
		if repo.Pattern != "" {
			// Pattern-based repository selection (e.g., "org/*")
			repos, err := a.client.ListOrgRepos(ctx, repo.Owner, repo.Pattern)
			if err != nil {
				return nil, fmt.Errorf("failed to list repos for %s/%s: %w", repo.Owner, repo.Pattern, err)
			}

			for _, r := range repos {
				if err := a.collectRepoData(ctx, repo.Owner, r, dateRange, data); err != nil {
					a.log("Warning: failed to collect data for %s/%s: %v", repo.Owner, r, err)
					// Continue with other repos
				}
			}
		} else {
			// Single repository
			if err := a.collectRepoData(ctx, repo.Owner, repo.Name, dateRange, data); err != nil {
				return nil, fmt.Errorf("failed to collect data for %s/%s: %w", repo.Owner, repo.Name, err)
			}
		}
	}

	return data, nil
}

func (a *App) collectRepoData(ctx context.Context, owner, name string, dateRange *config.ParsedDateRange, data *models.RawData) error {
	repoName := fmt.Sprintf("%s/%s", owner, name)
	a.log("  Fetching data from %s...", repoName)

	// Clone/update repository locally (required for accurate commit data)
	token := a.config.Auth.GithubToken

	// Determine clone options (shallow clone if enabled)
	var cloneOpts *git.CloneOptions
	if a.config.Options.ShallowClone && dateRange.Start != nil {
		// Get commit count since start date to determine shallow clone depth
		commitCount, countErr := a.client.GetCommitCountSince(ctx, owner, name, *dateRange.Start)
		if countErr != nil {
			a.log("    Warning: failed to get commit count for shallow clone: %v", countErr)
			// Proceed with full clone
		} else if commitCount > 0 {
			// Add buffer for safety margin
			depth := commitCount + a.config.Options.ShallowCloneBuffer
			cloneOpts = &git.CloneOptions{Depth: depth}
			a.log("    Using shallow clone (depth: %d = %d commits + %d buffer)", depth, commitCount, a.config.Options.ShallowCloneBuffer)
		}
	}

	if err := a.gitRepo.EnsureClonedWithOptions(ctx, owner, name, token, cloneOpts); err != nil {
		return fmt.Errorf("failed to clone repository %s: %w", repoName, err)
	}

	// Fetch commits from local git clone
	commits, err := a.gitRepo.FetchCommits(ctx, owner, name, dateRange.Start, dateRange.End)
	if err != nil {
		return fmt.Errorf("failed to fetch commits: %w", err)
	}

	// Filter out bots
	for _, c := range commits {
		if !a.config.IsBot(c.Author.Login) {
			data.Commits = append(data.Commits, c)
		}
	}

	// Fetch pull requests and reviews
	// Use GraphQL if available (much fewer API calls), otherwise fall back to REST
	if a.client.HasGraphQL() {
		prs, reviews, err := a.client.FetchPRsWithReviewsGraphQL(ctx, owner, name, dateRange.Start, dateRange.End)
		if err != nil {
			a.log("    Warning: GraphQL fetch failed, falling back to REST: %v", err)
			// Fall back to REST
			prs, reviews, err = a.fetchPRsAndReviewsREST(ctx, owner, name, dateRange, data)
			if err != nil {
				return err
			}
		}

		// Filter out bots
		for _, pr := range prs {
			if !a.config.IsBot(pr.Author.Login) {
				data.PullRequests = append(data.PullRequests, pr)
			}
		}
		for _, r := range reviews {
			if !a.config.IsBot(r.Author.Login) {
				data.Reviews = append(data.Reviews, r)
			}
		}
	} else {
		// Use REST API
		prs, reviews, err := a.fetchPRsAndReviewsREST(ctx, owner, name, dateRange, data)
		if err != nil {
			return err
		}
		// Filter out bots and add to data
		for _, pr := range prs {
			if !a.config.IsBot(pr.Author.Login) {
				data.PullRequests = append(data.PullRequests, pr)
			}
		}
		for _, r := range reviews {
			if !a.config.IsBot(r.Author.Login) {
				data.Reviews = append(data.Reviews, r)
			}
		}
	}

	// Fetch issues and comments
	// Use GraphQL if available (much fewer API calls), otherwise fall back to REST
	if a.client.HasGraphQL() {
		issues, comments, err := a.client.FetchIssuesWithCommentsGraphQL(ctx, owner, name, dateRange.Start, dateRange.End)
		if err != nil {
			a.log("    Warning: GraphQL fetch failed, falling back to REST: %v", err)
			// Fall back to REST
			if err := a.fetchIssuesAndCommentsREST(ctx, owner, name, dateRange, data); err != nil {
				return err
			}
		} else {

			// Filter out bots
			for _, issue := range issues {
				if !a.config.IsBot(issue.Author.Login) {
					data.Issues = append(data.Issues, issue)
				}
			}
			for _, comment := range comments {
				if !a.config.IsBot(comment.Author.Login) {
					data.IssueComments = append(data.IssueComments, comment)
				}
			}
		}
	} else {
		// Use REST API
		if err := a.fetchIssuesAndCommentsREST(ctx, owner, name, dateRange, data); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) log(format string, args ...interface{}) {
	if a.verbose {
		log.Printf(format, args...)
	} else {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// fetchUserProfiles collects unique GitHub logins from PR/review data and fetches their profiles
// The profiles contain public emails and names that help with commit author deduplication
func (a *App) fetchUserProfiles(ctx context.Context, data *models.RawData) (map[string]aggregator.UserProfile, error) {
	// Collect unique logins from PRs and reviews
	loginSet := make(map[string]bool)
	for _, pr := range data.PullRequests {
		if pr.Author.Login != "" {
			loginSet[pr.Author.Login] = true
		}
	}
	for _, review := range data.Reviews {
		if review.Author.Login != "" {
			loginSet[review.Author.Login] = true
		}
	}

	// Convert to slice
	logins := make([]string, 0, len(loginSet))
	for login := range loginSet {
		logins = append(logins, login)
	}

	if len(logins) == 0 {
		return make(map[string]aggregator.UserProfile), nil
	}

	// Fetch profiles from GitHub (uses cache)
	ghProfiles, err := a.client.FetchUserProfiles(ctx, logins)
	if err != nil {
		return nil, err
	}

	// Convert to aggregator.UserProfile
	profiles := make(map[string]aggregator.UserProfile)
	for login, p := range ghProfiles {
		profiles[login] = aggregator.UserProfile{
			ID:        p.ID,
			Login:     p.Login,
			Name:      p.Name,
			Email:     p.Email,
			AvatarURL: p.AvatarURL,
		}
	}

	return profiles, nil
}

// fetchPRsAndReviewsREST fetches PRs and reviews using the REST API (fallback when GraphQL fails)
func (a *App) fetchPRsAndReviewsREST(ctx context.Context, owner, name string, dateRange *config.ParsedDateRange, data *models.RawData) ([]models.PullRequest, []models.Review, error) {
	prs, err := a.client.FetchPullRequests(ctx, owner, name, dateRange.Start, dateRange.End)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch pull requests: %w", err)
	}
	a.log("    Found %d pull requests", len(prs))

	// Fetch reviews for each PR
	var reviews []models.Review
	for _, pr := range prs {
		prReviews, err := a.client.FetchReviews(ctx, owner, name, pr.Number)
		if err != nil {
			a.log("    Warning: failed to fetch reviews for PR #%d: %v", pr.Number, err)
			continue
		}
		reviews = append(reviews, prReviews...)
	}
	a.log("    Found %d reviews (REST)", len(reviews))

	return prs, reviews, nil
}

// fetchIssuesAndCommentsREST fetches issues and comments using the REST API (fallback when GraphQL fails)
func (a *App) fetchIssuesAndCommentsREST(ctx context.Context, owner, name string, dateRange *config.ParsedDateRange, data *models.RawData) error {
	issues, err := a.client.FetchIssues(ctx, owner, name, dateRange.Start, dateRange.End)
	if err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}
	a.log("    Found %d issues", len(issues))

	// Filter out bots and add to data
	for _, issue := range issues {
		if !a.config.IsBot(issue.Author.Login) {
			data.Issues = append(data.Issues, issue)
		}
	}

	// Fetch all comments for the repository within date range
	comments, err := a.client.FetchIssueComments(ctx, owner, name, dateRange.Start, dateRange.End)
	if err != nil {
		a.log("    Warning: failed to fetch issue comments: %v", err)
	} else {
		for _, comment := range comments {
			if !a.config.IsBot(comment.Author.Login) {
				data.IssueComments = append(data.IssueComments, comment)
			}
		}
		a.log("    Found %d issue comments (REST)", len(comments))
	}

	return nil
}
