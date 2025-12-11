package models

import "time"

// Commit represents a Git commit
type Commit struct {
	SHA          string    `json:"sha"`
	Message      string    `json:"message"`
	Author       Author    `json:"author"`
	Committer    Author    `json:"committer"`
	Date         time.Time `json:"date"`
	Additions    int       `json:"additions"`
	Deletions    int       `json:"deletions"`
	FilesChanged int       `json:"files_changed"`
	Repository   string    `json:"repository"` // owner/repo format
	URL          string    `json:"url"`

	// Meaningful line counts (excludes comments and whitespace)
	MeaningfulAdditions int `json:"meaningful_additions"`
	MeaningfulDeletions int `json:"meaningful_deletions"`

	// Comment line counts
	CommentAdditions int `json:"comment_additions"`
	CommentDeletions int `json:"comment_deletions"`

	// Derived fields
	HasTests bool `json:"has_tests"`
}

// TotalChanges returns the total lines changed (additions + deletions)
func (c *Commit) TotalChanges() int {
	return c.Additions + c.Deletions
}

// ShortSHA returns the first 7 characters of the SHA
func (c *Commit) ShortSHA() string {
	if len(c.SHA) >= 7 {
		return c.SHA[:7]
	}
	return c.SHA
}

// ShortMessage returns the first line of the commit message
func (c *Commit) ShortMessage() string {
	for i, r := range c.Message {
		if r == '\n' {
			return c.Message[:i]
		}
	}
	return c.Message
}
