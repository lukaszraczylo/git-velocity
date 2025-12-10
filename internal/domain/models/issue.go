package models

import "time"

// IssueState represents the state of an issue
type IssueState string

const (
	IssueStateOpen   IssueState = "open"
	IssueStateClosed IssueState = "closed"
)

// Issue represents a GitHub issue
type Issue struct {
	Number     int        `json:"number"`
	Title      string     `json:"title"`
	State      IssueState `json:"state"`
	Author     Author     `json:"author"`
	Repository string     `json:"repository"` // owner/repo format
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ClosedAt   *time.Time `json:"closed_at,omitempty"`
	ClosedBy   *Author    `json:"closed_by,omitempty"`
	Comments   int        `json:"comments"`
	Labels     []string   `json:"labels,omitempty"`
	URL        string     `json:"url"`

	// Derived fields
	TimeToClose *time.Duration `json:"time_to_close,omitempty"`
}

// IsClosed returns true if the issue is closed
func (i *Issue) IsClosed() bool {
	return i.State == IssueStateClosed
}

// CalculateTimeToClose calculates the time from issue creation to close
func (i *Issue) CalculateTimeToClose() *time.Duration {
	if i.ClosedAt == nil {
		return nil
	}
	d := i.ClosedAt.Sub(i.CreatedAt)
	return &d
}

// IssueComment represents a comment on an issue
type IssueComment struct {
	ID         int64     `json:"id"`
	Issue      int       `json:"issue"`
	Repository string    `json:"repository"`
	Author     Author    `json:"author"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
}
