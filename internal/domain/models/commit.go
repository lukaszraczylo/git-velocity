package models

import "time"

// Commit represents a Git commit
type Commit struct {
	SHA           string    `json:"sha"`
	Message       string    `json:"message"`
	Author        Author    `json:"author"`
	Committer     Author    `json:"committer"`
	Date          time.Time `json:"date"`
	Additions     int       `json:"additions"`
	Deletions     int       `json:"deletions"`
	FilesChanged  int       `json:"files_changed"`
	FilesModified []string  `json:"files_modified,omitempty"` // List of file paths modified in this commit
	Repository    string    `json:"repository"`               // owner/repo format
	URL           string    `json:"url"`

	// Meaningful line counts (excludes comments and whitespace)
	MeaningfulAdditions int `json:"meaningful_additions"`
	MeaningfulDeletions int `json:"meaningful_deletions"`

	// Comment line counts (all types of comments)
	CommentAdditions int `json:"comment_additions"`
	CommentDeletions int `json:"comment_deletions"`

	// Documentation comment counts (JSDoc, Rust doc comments, docstrings, etc.)
	DocCommentAdditions int `json:"doc_comment_additions"`
	DocCommentDeletions int `json:"doc_comment_deletions"`

	// Commented-out code counts (code that was commented rather than deleted)
	CommentedCodeAdditions int `json:"commented_code_additions"`
	CommentedCodeDeletions int `json:"commented_code_deletions"`

	// Derived fields
	HasTests bool `json:"has_tests"`
}
