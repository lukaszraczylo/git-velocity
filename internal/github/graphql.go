package github

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/lukaszraczylo/git-velocity/internal/domain/models"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// progressBar handles terminal progress display
type progressBar struct {
	progress progress.Model
	label    string
	total    int
	current  int
	out      io.Writer
}

func newProgressBar(label string, total int) *progressBar {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)
	return &progressBar{
		progress: p,
		label:    label,
		total:    total,
		current:  0,
		out:      os.Stderr,
	}
}

func (p *progressBar) update(fetched int) {
	p.current = fetched
	// Guard against division by zero
	var percent float64
	if p.total > 0 {
		percent = float64(p.current) / float64(p.total)
		if percent > 1.0 {
			percent = 1.0
		}
	} else {
		percent = 0.0
	}

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	fmt.Fprintf(p.out, "\r%s %s %s",
		labelStyle.Render(p.label),
		p.progress.ViewAs(percent),
		countStyle.Render(fmt.Sprintf("%d/%d", p.current, p.total)),
	)
}

func (p *progressBar) done() {
	p.update(p.total)
	fmt.Fprintln(p.out)
}

// GraphQLClient wraps the githubv4 client for GitHub API
type GraphQLClient struct {
	client *githubv4.Client
}

// NewGraphQLClient creates a new GraphQL client for GitHub
func NewGraphQLClient(token string) *GraphQLClient {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	return &GraphQLClient{
		client: client,
	}
}

// PageInfo contains pagination info from GraphQL responses
type PageInfo struct {
	HasNextPage bool
	EndCursor   githubv4.String
}

// PageResult represents a page of results from GraphQL
type PageResult[T any] struct {
	TotalCount int
	PageInfo   PageInfo
	Nodes      []T
}

// GQLFetchConfig configures the generic paginated fetcher for GraphQL
type GQLFetchConfig[Q any, T any, R any] struct {
	Label         string
	Query         *Q
	GetPageResult func(q *Q) PageResult[T]
	// ProcessNode returns items, whether this node is "old" (outside date range),
	// and whether to hard stop immediately (past cutoff date)
	ProcessNode func(node T, repo string) (items []R, isOld bool, hardStop bool)
	// ConsecutiveOldPagesToStop controls early termination (default: 2)
	ConsecutiveOldPagesToStop int
}

// fetchGQLPaginated is a generic paginated fetcher for GraphQL queries
func fetchGQLPaginated[Q any, T any, R any](
	ctx context.Context,
	client *githubv4.Client,
	owner, repo string,
	config GQLFetchConfig[Q, T, R],
) ([]R, error) {
	var allResults []R

	variables := map[string]interface{}{
		"owner":  githubv4.String(owner),
		"repo":   githubv4.String(repo),
		"cursor": (*githubv4.String)(nil),
	}

	var pbar *progressBar
	fetched := 0
	repoFullName := fmt.Sprintf("%s/%s", owner, repo)
	consecutiveOldPages := 0
	pagesToStop := config.ConsecutiveOldPagesToStop
	if pagesToStop == 0 {
		pagesToStop = 2 // default
	}

	for {
		// Retry logic for transient errors
		var queryErr error
		for retries := 0; retries < 3; retries++ {
			queryErr = client.Query(ctx, config.Query, variables)
			if queryErr == nil {
				break
			}
			// Check if error is retryable
			if !isGQLRetryableError(queryErr) {
				break
			}
			// Wait before retry with exponential backoff
			backoff := time.Duration(1<<retries) * time.Second
			fmt.Fprintf(os.Stderr, "\r      GraphQL retry %d/3 (waiting %s): %v\n", retries+1, backoff, queryErr)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
		if queryErr != nil {
			return nil, fmt.Errorf("graphql query failed: %w", queryErr)
		}

		page := config.GetPageResult(config.Query)

		// Initialize progress bar on first query
		if pbar == nil && page.TotalCount > 0 {
			pbar = newProgressBar(config.Label, page.TotalCount)
		}

		oldInPage := 0
		totalInPage := 0
		shouldHardStop := false
		for _, node := range page.Nodes {
			fetched++
			totalInPage++
			items, isOld, hardStop := config.ProcessNode(node, repoFullName)
			allResults = append(allResults, items...)
			if isOld {
				oldInPage++
			}
			if hardStop {
				shouldHardStop = true
				break
			}
		}

		if pbar != nil {
			pbar.update(fetched)
		}

		// Hard stop takes priority (past cutoff date)
		if shouldHardStop {
			if pbar != nil {
				pbar.done()
			}
			break
		}

		// Track consecutive pages where all items are old
		if totalInPage > 0 && oldInPage == totalInPage {
			consecutiveOldPages++
		} else {
			consecutiveOldPages = 0
		}

		// Stop if we've seen enough consecutive old pages or no more pages
		if consecutiveOldPages >= pagesToStop || !page.PageInfo.HasNextPage {
			if pbar != nil {
				pbar.done()
			}
			break
		}
		variables["cursor"] = githubv4.NewString(page.PageInfo.EndCursor)
	}

	return allResults, nil
}

// Query structs for PRs with reviews
type gqlPRQuery struct {
	Repository struct {
		PullRequests struct {
			TotalCount int
			PageInfo   PageInfo
			Nodes      []gqlPRNode
		} `graphql:"pullRequests(first: 100, after: $cursor, states: [OPEN, MERGED, CLOSED], orderBy: {field: UPDATED_AT, direction: DESC})"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

type gqlPRNode struct {
	Number       int
	Title        string
	State        string
	Merged       bool
	Additions    int
	Deletions    int
	ChangedFiles int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	MergedAt     *time.Time
	ClosedAt     *time.Time
	BaseRefName  string
	HeadRefName  string
	URL          string
	Commits      struct{ TotalCount int }
	Author       gqlActor
	Reviews      struct {
		TotalCount int
		Nodes      []gqlReviewNode
		PageInfo   PageInfo
	} `graphql:"reviews(first: 100)"`
}

type gqlActor struct {
	Login     string
	AvatarURL string `graphql:"avatarUrl"`
}

type gqlReviewNode struct {
	ID          string `graphql:"id"`
	Author      gqlActor
	State       string
	SubmittedAt *time.Time
	Body        string
	Comments    struct{ TotalCount int } `graphql:"comments"`
}

// Query struct for issues with comments
type gqlIssueQuery struct {
	Repository struct {
		Issues struct {
			TotalCount int
			PageInfo   PageInfo
			Nodes      []gqlIssueNode
		} `graphql:"issues(first: 100, after: $cursor, orderBy: {field: CREATED_AT, direction: DESC})"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

type gqlIssueNode struct {
	Number    int
	Title     string
	State     string
	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt  *time.Time
	URL       string
	Author    gqlActor
	Labels    struct {
		Nodes []struct{ Name string }
	} `graphql:"labels(first: 10)"`
	Comments struct {
		TotalCount int
		Nodes      []gqlCommentNode
		PageInfo   PageInfo
	} `graphql:"comments(first: 100)"`
}

type gqlCommentNode struct {
	ID        string `graphql:"id"`
	Author    gqlActor
	Body      string
	CreatedAt time.Time
}

// prWithReviews bundles a PR with its reviews for the generic fetcher
type prWithReviews struct {
	PR      models.PullRequest
	Reviews []models.Review
}

// FetchPRsWithReviews fetches pull requests with their reviews using GraphQL
func (g *GraphQLClient) FetchPRsWithReviews(ctx context.Context, owner, repo string, since, until *time.Time) ([]models.PullRequest, []models.Review, error) {
	var query gqlPRQuery

	// Hard cutoff: 1 week before start date - stop fetching entirely past this point
	var hardCutoff *time.Time
	if since != nil {
		cutoff := since.AddDate(0, 0, -7)
		hardCutoff = &cutoff
	}

	results, err := fetchGQLPaginated(ctx, g.client, owner, repo, GQLFetchConfig[gqlPRQuery, gqlPRNode, prWithReviews]{
		Label:                     "      Fetching PRs:",
		Query:                     &query,
		ConsecutiveOldPagesToStop: 2,
		GetPageResult: func(q *gqlPRQuery) PageResult[gqlPRNode] {
			return PageResult[gqlPRNode]{
				TotalCount: q.Repository.PullRequests.TotalCount,
				PageInfo:   q.Repository.PullRequests.PageInfo,
				Nodes:      q.Repository.PullRequests.Nodes,
			}
		},
		ProcessNode: func(node gqlPRNode, repoName string) ([]prWithReviews, bool, bool) {
			// Determine the relevant date for filtering:
			// - For merged PRs: use MergedAt
			// - For closed PRs: use ClosedAt
			// - For open PRs: use CreatedAt (they're still active)
			var relevantDate time.Time
			if node.MergedAt != nil {
				relevantDate = *node.MergedAt
			} else if node.ClosedAt != nil {
				relevantDate = *node.ClosedAt
			} else {
				relevantDate = node.CreatedAt
			}

			// Hard cutoff check - stop entirely if past this date
			if hardCutoff != nil && relevantDate.Before(*hardCutoff) {
				return nil, true, true // Hard stop
			}

			// Check date range - skip if outside range
			if until != nil && relevantDate.After(*until) {
				return nil, false, false // Too new, not "old"
			}
			if since != nil && relevantDate.Before(*since) {
				return nil, true, false // Too old - signal for early termination tracking
			}

			// Convert PR
			pr := convertPRNode(node, repoName)

			// Convert reviews
			var reviews []models.Review
			for _, r := range node.Reviews.Nodes {
				reviews = append(reviews, convertReviewNode(r, repoName, node.Number))
			}

			return []prWithReviews{{PR: pr, Reviews: reviews}}, false, false
		},
	})
	if err != nil {
		return nil, nil, err
	}

	// Flatten results
	var prs []models.PullRequest
	var reviews []models.Review
	for _, r := range results {
		prs = append(prs, r.PR)
		reviews = append(reviews, r.Reviews...)
	}

	return prs, reviews, nil
}

// issueWithComments bundles an issue with its comments for the generic fetcher
type issueWithComments struct {
	Issue    models.Issue
	Comments []models.IssueComment
}

// FetchIssuesWithComments fetches issues with their comments using GraphQL
func (g *GraphQLClient) FetchIssuesWithComments(ctx context.Context, owner, repo string, since, until *time.Time) ([]models.Issue, []models.IssueComment, error) {
	var query gqlIssueQuery

	// Hard cutoff: 1 week before start date - stop fetching entirely past this point
	var hardCutoff *time.Time
	if since != nil {
		cutoff := since.AddDate(0, 0, -7)
		hardCutoff = &cutoff
	}

	results, err := fetchGQLPaginated(ctx, g.client, owner, repo, GQLFetchConfig[gqlIssueQuery, gqlIssueNode, issueWithComments]{
		Label:                     "      Fetching issues:",
		Query:                     &query,
		ConsecutiveOldPagesToStop: 2,
		GetPageResult: func(q *gqlIssueQuery) PageResult[gqlIssueNode] {
			return PageResult[gqlIssueNode]{
				TotalCount: q.Repository.Issues.TotalCount,
				PageInfo:   q.Repository.Issues.PageInfo,
				Nodes:      q.Repository.Issues.Nodes,
			}
		},
		ProcessNode: func(node gqlIssueNode, repoName string) ([]issueWithComments, bool, bool) {
			// Hard cutoff check - stop entirely if past this date
			if hardCutoff != nil && node.CreatedAt.Before(*hardCutoff) {
				return nil, true, true // Hard stop
			}

			// Check date range
			if until != nil && node.CreatedAt.After(*until) {
				return nil, false, false // Too new, not "old"
			}
			if since != nil && node.CreatedAt.Before(*since) {
				return nil, true, false // Too old - signal for early termination tracking
			}

			// Convert issue
			issue := convertIssueNode(node, repoName)

			// Convert comments within date range
			var comments []models.IssueComment
			for _, c := range node.Comments.Nodes {
				if until != nil && c.CreatedAt.After(*until) {
					continue
				}
				if since != nil && c.CreatedAt.Before(*since) {
					continue
				}
				comments = append(comments, convertCommentNode(c, repoName, node.Number))
			}

			return []issueWithComments{{Issue: issue, Comments: comments}}, false, false
		},
	})
	if err != nil {
		return nil, nil, err
	}

	// Flatten results
	var issues []models.Issue
	var comments []models.IssueComment
	for _, r := range results {
		issues = append(issues, r.Issue)
		comments = append(comments, r.Comments...)
	}

	return issues, comments, nil
}

// Conversion helpers

func convertActor(a gqlActor) models.Author {
	return models.Author{
		Login:     a.Login,
		AvatarURL: a.AvatarURL,
	}
}

func convertPRNode(node gqlPRNode, repoName string) models.PullRequest {
	state := models.PRStateOpen
	if node.Merged {
		state = models.PRStateMerged
	} else if node.State == "CLOSED" {
		state = models.PRStateClosed
	}

	return models.PullRequest{
		Number:       node.Number,
		Title:        node.Title,
		State:        state,
		Author:       convertActor(node.Author),
		Repository:   repoName,
		BaseBranch:   node.BaseRefName,
		HeadBranch:   node.HeadRefName,
		CreatedAt:    node.CreatedAt,
		UpdatedAt:    node.UpdatedAt,
		MergedAt:     node.MergedAt,
		ClosedAt:     node.ClosedAt,
		Additions:    node.Additions,
		Deletions:    node.Deletions,
		FilesChanged: node.ChangedFiles,
		CommitCount:  node.Commits.TotalCount,
		Comments:     node.Reviews.TotalCount,
		URL:          node.URL,
	}
}

func convertReviewNode(node gqlReviewNode, repoName string, prNumber int) models.Review {
	var submittedAt time.Time
	if node.SubmittedAt != nil {
		submittedAt = *node.SubmittedAt
	}

	return models.Review{
		PullRequest:   prNumber,
		Repository:    repoName,
		Author:        convertActor(node.Author),
		State:         models.ReviewState(node.State),
		SubmittedAt:   submittedAt,
		Body:          node.Body,
		CommentsCount: node.Comments.TotalCount,
	}
}

func convertIssueNode(node gqlIssueNode, repoName string) models.Issue {
	state := models.IssueStateOpen
	if node.State == "CLOSED" {
		state = models.IssueStateClosed
	}

	var labels []string
	for _, l := range node.Labels.Nodes {
		labels = append(labels, l.Name)
	}

	return models.Issue{
		Number:     node.Number,
		Title:      node.Title,
		State:      state,
		Author:     convertActor(node.Author),
		Repository: repoName,
		CreatedAt:  node.CreatedAt,
		UpdatedAt:  node.UpdatedAt,
		ClosedAt:   node.ClosedAt,
		Comments:   node.Comments.TotalCount,
		Labels:     labels,
		URL:        node.URL,
	}
}

func convertCommentNode(node gqlCommentNode, repoName string, issueNumber int) models.IssueComment {
	return models.IssueComment{
		Issue:      issueNumber,
		Repository: repoName,
		Author:     convertActor(node.Author),
		Body:       node.Body,
		CreatedAt:  node.CreatedAt,
	}
}

// isGQLRetryableError checks if a GraphQL error is transient and should be retried
func isGQLRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"stream error",
		"cancel",
		"eof",
		"connection reset",
		"connection refused",
		"timeout",
		"temporary failure",
		"broken pipe",
		"502",
		"503",
		"504",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}
