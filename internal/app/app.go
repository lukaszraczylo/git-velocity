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

	// Initialize local git repository manager if using local git
	if a.config.Options.UseLocalGit {
		a.log("Initializing local git repository manager...")
		gitRepo, err := git.NewRepository(a.config.Options.CloneDirectory)
		if err != nil {
			return fmt.Errorf("failed to create git repository manager: %w", err)
		}
		gitRepo.SetProgressCallback(func(msg string) {
			a.log("%s", msg)
		})
		a.gitRepo = gitRepo
	}

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

	// Fetch commits - use local git if enabled (much faster)
	var commits []models.Commit
	var err error

	if a.gitRepo != nil {
		// Clone/update repository locally
		token := a.config.Auth.GithubToken
		cloneErr := a.gitRepo.EnsureCloned(ctx, owner, name, token)
		if cloneErr != nil {
			a.log("    Warning: failed to clone repository locally, falling back to API: %v", cloneErr)
			// Fallback to API
			commits, err = a.client.FetchCommits(ctx, owner, name, dateRange.Start, dateRange.End)
		} else {
			// Use local git for commits
			commits, err = a.gitRepo.FetchCommits(ctx, owner, name, dateRange.Start, dateRange.End)
		}
	} else {
		// Use API for commits
		commits, err = a.client.FetchCommits(ctx, owner, name, dateRange.Start, dateRange.End)
	}

	if err != nil {
		return fmt.Errorf("failed to fetch commits: %w", err)
	}
	a.log("    Found %d commits", len(commits))

	// Filter out bots
	for _, c := range commits {
		if !a.config.IsBot(c.Author.Login) {
			data.Commits = append(data.Commits, c)
		}
	}

	// Fetch pull requests
	prs, err := a.client.FetchPullRequests(ctx, owner, name, dateRange.Start, dateRange.End)
	if err != nil {
		return fmt.Errorf("failed to fetch pull requests: %w", err)
	}
	a.log("    Found %d pull requests", len(prs))

	for _, pr := range prs {
		if !a.config.IsBot(pr.Author.Login) {
			data.PullRequests = append(data.PullRequests, pr)
		}
	}

	// Fetch reviews in parallel for all PRs (already filtered by FetchPullRequests)
	if len(prs) > 0 {
		a.log("    Fetching reviews for %d PRs in parallel...", len(prs))

		type reviewResult struct {
			reviews []models.Review
			err     error
		}

		// Use worker pool to limit concurrent requests
		concurrency := a.config.Options.ConcurrentRequests
		if concurrency <= 0 {
			concurrency = 5
		}

		results := make(chan reviewResult, len(prs))
		sem := make(chan struct{}, concurrency)

		for _, pr := range prs {
			go func(prNum int) {
				sem <- struct{}{}        // Acquire
				defer func() { <-sem }() // Release

				reviews, err := a.client.FetchReviews(ctx, owner, name, prNum)
				results <- reviewResult{reviews: reviews, err: err}
			}(pr.Number)
		}

		// Collect results
		reviewCount := 0
		for i := 0; i < len(prs); i++ {
			result := <-results
			if result.err != nil {
				continue
			}
			for _, r := range result.reviews {
				if !a.config.IsBot(r.Author.Login) {
					data.Reviews = append(data.Reviews, r)
					reviewCount++
				}
			}
		}
		a.log("    Found %d reviews across %d PRs", reviewCount, len(prs))
	}

	// Fetch issues
	issues, err := a.client.FetchIssues(ctx, owner, name, dateRange.Start, dateRange.End)
	if err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}
	a.log("    Found %d issues", len(issues))

	for _, issue := range issues {
		if !a.config.IsBot(issue.Author.Login) {
			data.Issues = append(data.Issues, issue)
		}
	}

	// Fetch issue comments
	issueComments, err := a.client.FetchIssueComments(ctx, owner, name, dateRange.Start, dateRange.End)
	if err != nil {
		return fmt.Errorf("failed to fetch issue comments: %w", err)
	}
	a.log("    Found %d issue comments", len(issueComments))

	for _, comment := range issueComments {
		if !a.config.IsBot(comment.Author.Login) {
			data.IssueComments = append(data.IssueComments, comment)
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
