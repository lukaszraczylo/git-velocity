package models

// RawData holds the raw collected data from GitHub
type RawData struct {
	Commits       []Commit
	PullRequests  []PullRequest
	Reviews       []Review
	Issues        []Issue
	IssueComments []IssueComment
}
