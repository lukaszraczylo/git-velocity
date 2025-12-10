package models

import "time"

// ReviewState represents the state of a review
type ReviewState string

const (
	ReviewApproved         ReviewState = "APPROVED"
	ReviewChangesRequested ReviewState = "CHANGES_REQUESTED"
	ReviewCommented        ReviewState = "COMMENTED"
	ReviewPending          ReviewState = "PENDING"
	ReviewDismissed        ReviewState = "DISMISSED"
)

// Review represents a GitHub pull request review
type Review struct {
	ID            int64       `json:"id"`
	PullRequest   int         `json:"pull_request"`
	Repository    string      `json:"repository"` // owner/repo format
	Author        Author      `json:"author"`
	State         ReviewState `json:"state"`
	SubmittedAt   time.Time   `json:"submitted_at"`
	Body          string      `json:"body,omitempty"`
	CommentsCount int         `json:"comments_count"`

	// Derived fields
	ResponseTime *time.Duration `json:"response_time,omitempty"` // Time from PR creation or review request to review
}

// IsApproval returns true if the review is an approval
func (r *Review) IsApproval() bool {
	return r.State == ReviewApproved
}

// RequestsChanges returns true if the review requests changes
func (r *Review) RequestsChanges() bool {
	return r.State == ReviewChangesRequested
}

// IsSubstantive returns true if the review has meaningful content (not just a simple approval)
func (r *Review) IsSubstantive() bool {
	return r.Body != "" || r.CommentsCount > 0 || r.State == ReviewChangesRequested
}

// ReviewComment represents a comment on a pull request review
type ReviewComment struct {
	ID          int64     `json:"id"`
	ReviewID    int64     `json:"review_id"`
	PullRequest int       `json:"pull_request"`
	Repository  string    `json:"repository"`
	Author      Author    `json:"author"`
	Body        string    `json:"body"`
	Path        string    `json:"path,omitempty"`
	Line        int       `json:"line,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
