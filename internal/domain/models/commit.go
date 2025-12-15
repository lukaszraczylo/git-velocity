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
