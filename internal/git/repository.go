package git

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/lukaszraczylo/git-velocity/internal/diff"
	"github.com/lukaszraczylo/git-velocity/internal/domain/models"
)

// commitProgressBar handles terminal progress display for commit iteration
type commitProgressBar struct {
	progress progress.Model
	label    string
	current  int
	out      io.Writer
}

func newCommitProgressBar(label string) *commitProgressBar {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)
	return &commitProgressBar{
		progress: p,
		label:    label,
		current:  0,
		out:      os.Stderr,
	}
}

func (p *commitProgressBar) update(count int) {
	p.current = count

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// Use a spinner-like display since we don't know total
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinChar := spinner[count%len(spinner)]

	fmt.Fprintf(p.out, "\r%s %s %s",
		labelStyle.Render(p.label),
		spinChar,
		countStyle.Render(fmt.Sprintf("%d commits", p.current)),
	)
}

func (p *commitProgressBar) done(total int) {
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	fmt.Fprintf(p.out, "\r%s %s %s\n",
		labelStyle.Render(p.label),
		p.progress.ViewAs(1.0),
		countStyle.Render(fmt.Sprintf("%d commits", total)),
	)
}

// ProgressCallback is called to report progress during git operations
type ProgressCallback func(message string)

// Repository manages local git repository operations using go-git
type Repository struct {
	baseDir  string
	progress ProgressCallback
}

// NewRepository creates a new repository manager
func NewRepository(baseDir string) (*Repository, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &Repository{
		baseDir:  baseDir,
		progress: func(string) {}, // no-op by default
	}, nil
}

// SetProgressCallback sets the callback function for progress reporting
func (r *Repository) SetProgressCallback(cb ProgressCallback) {
	if cb != nil {
		r.progress = cb
	}
}

// repoPath returns the local path for a repository
func (r *Repository) repoPath(owner, name string) string {
	return filepath.Join(r.baseDir, owner, name)
}

// CloneOptions contains options for cloning a repository
type CloneOptions struct {
	// Depth limits the clone to the specified number of commits (0 = full clone)
	Depth int
}

// EnsureClonedWithOptions ensures a repository is cloned with specific options
func (r *Repository) EnsureClonedWithOptions(ctx context.Context, owner, name, token string, opts *CloneOptions) error {
	repoPath := r.repoPath(owner, name)

	// Check if already cloned
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// Repository exists, fetch latest
		r.progress(fmt.Sprintf("      Updating local clone of %s/%s...", owner, name))
		return r.fetch(ctx, repoPath, token)
	}

	// Clone the repository
	if opts != nil && opts.Depth > 0 {
		r.progress(fmt.Sprintf("      Shallow cloning %s/%s (depth: %d)...", owner, name, opts.Depth))
	} else {
		r.progress(fmt.Sprintf("      Cloning %s/%s...", owner, name))
	}
	return r.clone(ctx, owner, name, token, repoPath, opts)
}

// clone clones a repository using go-git
func (r *Repository) clone(ctx context.Context, owner, name, token, destPath string, opts *CloneOptions) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(destPath), 0750); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, name)

	cloneOpts := &git.CloneOptions{
		URL:      cloneURL,
		Progress: nil, // Could add progress writer here
	}

	// Apply shallow clone depth if provided
	if opts != nil && opts.Depth > 0 {
		cloneOpts.Depth = opts.Depth
	}

	// Add authentication if token provided
	if token != "" {
		cloneOpts.Auth = &http.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}
	}

	_, err := git.PlainCloneContext(ctx, destPath, false, cloneOpts)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

// fetch fetches latest changes from remote using go-git
func (r *Repository) fetch(ctx context.Context, repoPath, token string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	fetchOpts := &git.FetchOptions{
		RemoteName: "origin",
		Force:      true,
		Prune:      true,
		RefSpecs:   []config.RefSpec{"+refs/*:refs/*"},
	}

	// Add authentication if token provided
	if token != "" {
		fetchOpts.Auth = &http.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}
	}

	err = repo.FetchContext(ctx, fetchOpts)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	return nil
}

// FetchCommits retrieves commits from the local repository using go-git
func (r *Repository) FetchCommits(ctx context.Context, owner, name string, since, until *time.Time) ([]models.Commit, error) {
	repoPath := r.repoPath(owner, name)

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get all references to iterate all branches
	refs, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %w", err)
	}

	// Collect all commit hashes from all branches
	seenCommits := make(map[plumbing.Hash]bool)
	var commits []models.Commit
	testPatterns := []string{"_test.go", ".test.", ".spec.", "/tests/", "/test/", "__tests__"}

	// Progress bar for commit iteration
	pbar := newCommitProgressBar("      Iterating commits:")
	processedCount := 0

	// Hard cutoff: 1 week before start date - stop iterating entirely past this point
	var hardCutoff *time.Time
	if since != nil {
		cutoff := since.AddDate(0, 0, -7)
		hardCutoff = &cutoff
	}

	// errStopIteration is used to signal early termination (not a real error)
	var errStopIteration = fmt.Errorf("stop iteration")

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		// Skip non-branch references
		if !ref.Name().IsBranch() && !ref.Name().IsRemote() && !ref.Name().IsTag() {
			return nil
		}

		// Get commit iterator for this reference
		commitIter, err := repo.Log(&git.LogOptions{
			From:  ref.Hash(),
			Order: git.LogOrderCommitterTime,
			All:   false,
		})
		if err != nil {
			// Skip refs that don't point to commits
			return nil
		}

		consecutiveOld := 0
		err = commitIter.ForEach(func(c *object.Commit) error {
			// Check context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Skip already seen commits
			if seenCommits[c.Hash] {
				return nil
			}
			seenCommits[c.Hash] = true
			processedCount++

			// Update progress every 10 commits to avoid too much I/O
			if processedCount%10 == 0 {
				pbar.update(processedCount)
			}

			commitTime := c.Author.When

			// Hard cutoff - stop entirely if past this date
			if hardCutoff != nil && commitTime.Before(*hardCutoff) {
				return errStopIteration
			}

			// Filter by date range
			if since != nil && commitTime.Before(*since) {
				consecutiveOld++
				// Early termination: if we've seen 100 consecutive old commits, stop this branch
				if consecutiveOld >= 100 {
					return errStopIteration
				}
				return nil
			}
			consecutiveOld = 0 // Reset counter when we find a valid commit

			if until != nil && commitTime.After(*until) {
				return nil
			}

			// Get file stats for this commit
			stats := r.getCommitStats(c, testPatterns)

			// Extract login from email
			authorLogin := extractLoginFromEmail(c.Author.Email, c.Author.Name)
			committerLogin := extractLoginFromEmail(c.Committer.Email, c.Committer.Name)

			commit := models.Commit{
				SHA:     c.Hash.String(),
				Message: strings.Split(c.Message, "\n")[0], // First line only
				Author: models.Author{
					Login: authorLogin,
					Name:  c.Author.Name,
					Email: c.Author.Email,
				},
				Committer: models.Author{
					Login: committerLogin,
					Name:  c.Committer.Name,
					Email: c.Committer.Email,
				},
				Date:                commitTime,
				Additions:           stats.Additions,
				Deletions:           stats.Deletions,
				MeaningfulAdditions: stats.MeaningfulAdditions,
				MeaningfulDeletions: stats.MeaningfulDeletions,
				CommentAdditions:    stats.CommentAdditions,
				CommentDeletions:    stats.CommentDeletions,
				FilesChanged:        stats.FilesChanged,
				Repository:          fmt.Sprintf("%s/%s", owner, name),
				URL:                 fmt.Sprintf("https://github.com/%s/%s/commit/%s", owner, name, c.Hash.String()),
				HasTests:            stats.HasTests,
			}

			commits = append(commits, commit)
			return nil
		})

		// Handle expected termination conditions
		if err == errStopIteration {
			return nil // Not an error, just early termination for this branch
		}

		// Handle shallow clone boundary - "object not found" means we've reached
		// the edge of the shallow clone history, which is expected behavior
		if err != nil && isShallowBoundaryError(err) {
			err = nil // Treat as normal end of history
		}

		return err
	})

	// Complete progress bar
	pbar.done(len(commits))

	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}

// commitStats holds the statistics for a commit
type commitStats struct {
	Additions           int
	Deletions           int
	MeaningfulAdditions int
	MeaningfulDeletions int
	CommentAdditions    int
	CommentDeletions    int
	FilesChanged        int
	HasTests            bool
}

// getCommitStats calculates additions, deletions, files changed for a commit
func (r *Repository) getCommitStats(c *object.Commit, testPatterns []string) commitStats {
	stats := commitStats{}

	// Get parent commit for diff
	parentIter := c.Parents()
	parent, err := parentIter.Next()

	var parentTree *object.Tree
	if err == nil {
		parentTree, _ = parent.Tree()
	}

	currentTree, err := c.Tree()
	if err != nil {
		return stats
	}

	// Get changes between parent and current
	var changes object.Changes
	if parentTree != nil {
		changes, err = parentTree.Diff(currentTree)
	} else {
		// Initial commit - all files are additions
		changes, err = object.DiffTree(nil, currentTree)
	}

	if err != nil {
		return stats
	}

	filesSet := make(map[string]bool)

	for _, change := range changes {
		// Get the file path
		var filePath string
		if change.To.Name != "" {
			filePath = change.To.Name
		} else if change.From.Name != "" {
			filePath = change.From.Name
		}

		// Skip documentation files
		if diff.IsDocumentationFile(filePath) {
			continue
		}

		// Count unique files
		if !filesSet[filePath] {
			filesSet[filePath] = true
			stats.FilesChanged++

			// Check for test files
			for _, pattern := range testPatterns {
				if strings.Contains(filePath, pattern) {
					stats.HasTests = true
					break
				}
			}
		}

		// Get patch to count lines
		patch, err := change.Patch()
		if err != nil {
			continue
		}

		for _, filePatch := range patch.FilePatches() {
			for _, chunk := range filePatch.Chunks() {
				content := chunk.Content()
				lines := strings.Split(content, "\n")

				switch chunk.Type() {
				case 1: // Add
					for _, line := range lines {
						stats.Additions++
						if diff.IsMeaningfulLine(line) {
							stats.MeaningfulAdditions++
						} else if diff.IsCommentLine(line) && !diff.IsWhitespaceLine(line) {
							stats.CommentAdditions++
						}
					}
				case 2: // Delete
					for _, line := range lines {
						stats.Deletions++
						if diff.IsMeaningfulLine(line) {
							stats.MeaningfulDeletions++
						} else if diff.IsCommentLine(line) && !diff.IsWhitespaceLine(line) {
							stats.CommentDeletions++
						}
					}
				}
			}
		}
	}

	return stats
}

// isShallowBoundaryError checks if an error indicates we've hit the shallow clone boundary
func isShallowBoundaryError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// go-git returns "object not found" when trying to access commits beyond shallow depth
	return strings.Contains(errStr, "object not found")
}

// extractLoginFromEmail tries to extract GitHub login from email
func extractLoginFromEmail(email, fallbackName string) string {
	// Pattern: 12345678+username@users.noreply.github.com
	// or: username@users.noreply.github.com
	if strings.Contains(email, "@users.noreply.github.com") {
		localPart := strings.Split(email, "@")[0]
		// Remove numeric prefix if present (e.g., "12345678+username")
		if idx := strings.Index(localPart, "+"); idx != -1 {
			return localPart[idx+1:]
		}
		return localPart
	}

	// Fallback: use sanitized name as login
	login := strings.ToLower(fallbackName)
	login = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(login, "-")
	return login
}
