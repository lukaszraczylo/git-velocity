package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/lukaszraczylo/git-velocity/internal/domain/models"
)

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

// EnsureCloned ensures a repository is cloned and up to date
func (r *Repository) EnsureCloned(ctx context.Context, owner, name, token string) error {
	repoPath := r.repoPath(owner, name)

	// Check if already cloned
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// Repository exists, fetch latest
		r.progress(fmt.Sprintf("      Updating local clone of %s/%s...", owner, name))
		return r.fetch(ctx, repoPath, token)
	}

	// Clone the repository
	r.progress(fmt.Sprintf("      Cloning %s/%s...", owner, name))
	return r.clone(ctx, owner, name, token, repoPath)
}

// clone clones a repository using go-git
func (r *Repository) clone(ctx context.Context, owner, name, token, destPath string) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(destPath), 0750); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, name)

	cloneOpts := &git.CloneOptions{
		URL:      cloneURL,
		Progress: nil, // Could add progress writer here
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

// isCommentLine checks if a line is a code comment (should not count as contribution)
func isCommentLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true // Empty lines don't count
	}

	// Common comment patterns across languages
	commentPrefixes := []string{
		"//",     // C, C++, Java, Go, JS, etc.
		"#",      // Python, Ruby, Shell, YAML
		"/*",     // C-style block comment start
		"*/",     // C-style block comment end
		"*",      // C-style block comment continuation
		"<!--",   // HTML/XML comment
		"-->",    // HTML/XML comment end
		"--",     // SQL, Lua, Haskell
		";",      // Assembly, Lisp, INI files
		"'",      // VB comment
		"\"\"\"", // Python docstring
		"'''",    // Python docstring
	}

	for _, prefix := range commentPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}

	return false
}

// isDocumentationFile checks if a file is documentation-only
func isDocumentationFile(filename string) bool {
	// Documentation file extensions and patterns
	docPatterns := []string{
		".md", ".markdown", ".rst", ".txt", ".adoc",
		"README", "CHANGELOG", "LICENSE", "CONTRIBUTING",
		"docs/", "documentation/", "/doc/",
	}

	lowerFilename := strings.ToLower(filename)
	for _, pattern := range docPatterns {
		if strings.Contains(lowerFilename, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// FetchCommits retrieves commits from the local repository using go-git
func (r *Repository) FetchCommits(ctx context.Context, owner, name string, since, until *time.Time) ([]models.Commit, error) {
	repoPath := r.repoPath(owner, name)

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	r.progress("      Iterating commits with go-git...")

	// Get all references to iterate all branches
	refs, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %w", err)
	}

	// Collect all commit hashes from all branches
	seenCommits := make(map[plumbing.Hash]bool)
	var commits []models.Commit
	testPatterns := []string{"_test.go", ".test.", ".spec.", "/tests/", "/test/", "__tests__"}

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

			commitTime := c.Author.When

			// Filter by date range
			if since != nil && commitTime.Before(*since) {
				return nil
			}
			if until != nil && commitTime.After(*until) {
				return nil
			}

			// Get file stats for this commit
			additions, deletions, filesChanged, hasTests := r.getCommitStats(c, testPatterns)

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
				Date:         commitTime,
				Additions:    additions,
				Deletions:    deletions,
				FilesChanged: filesChanged,
				Repository:   fmt.Sprintf("%s/%s", owner, name),
				URL:          fmt.Sprintf("https://github.com/%s/%s/commit/%s", owner, name, c.Hash.String()),
				HasTests:     hasTests,
			}

			commits = append(commits, commit)
			return nil
		})

		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	r.progress(fmt.Sprintf("      Found %d commits", len(commits)))

	return commits, nil
}

// getCommitStats calculates additions, deletions, files changed for a commit
func (r *Repository) getCommitStats(c *object.Commit, testPatterns []string) (additions, deletions, filesChanged int, hasTests bool) {
	// Get parent commit for diff
	parentIter := c.Parents()
	parent, err := parentIter.Next()

	var parentTree *object.Tree
	if err == nil {
		parentTree, _ = parent.Tree()
	}

	currentTree, err := c.Tree()
	if err != nil {
		return 0, 0, 0, false
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
		return 0, 0, 0, false
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
		if isDocumentationFile(filePath) {
			continue
		}

		// Count unique files
		if !filesSet[filePath] {
			filesSet[filePath] = true
			filesChanged++

			// Check for test files
			for _, pattern := range testPatterns {
				if strings.Contains(filePath, pattern) {
					hasTests = true
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
						if !isCommentLine(line) {
							additions++
						}
					}
				case 2: // Delete
					for _, line := range lines {
						if !isCommentLine(line) {
							deletions++
						}
					}
				}
			}
		}
	}

	return additions, deletions, filesChanged, hasTests
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

// GetAuthorMappings fetches author login mappings
// This helps map commit authors to GitHub usernames
func (r *Repository) GetAuthorMappings(ctx context.Context, owner, name string) (map[string]string, error) {
	repoPath := r.repoPath(owner, name)

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	mappings := make(map[string]string)

	// Iterate all commits to collect author mappings
	commitIter, err := repo.Log(&git.LogOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	err = commitIter.ForEach(func(c *object.Commit) error {
		if _, exists := mappings[c.Author.Email]; !exists {
			mappings[c.Author.Email] = extractLoginFromEmail(c.Author.Email, c.Author.Name)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return mappings, nil
}

// Cleanup removes the local clone of a repository
func (r *Repository) Cleanup(owner, name string) error {
	repoPath := r.repoPath(owner, name)
	return os.RemoveAll(repoPath)
}

// CleanupAll removes all local clones
func (r *Repository) CleanupAll() error {
	return os.RemoveAll(r.baseDir)
}
