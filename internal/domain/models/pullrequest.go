package models

import "time"

// PRState represents the state of a pull request
type PRState string

const (
	PRStateOpen   PRState = "open"
	PRStateClosed PRState = "closed"
	PRStateMerged PRState = "merged"
)

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number       int        `json:"number"`
	Title        string     `json:"title"`
	State        PRState    `json:"state"`
	Author       Author     `json:"author"`
	Repository   string     `json:"repository"`  // owner/repo format
	BaseBranch   string     `json:"base_branch"` // Target branch (e.g., main, master)
	HeadBranch   string     `json:"head_branch"` // Source branch
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	MergedAt     *time.Time `json:"merged_at,omitempty"`
	ClosedAt     *time.Time `json:"closed_at,omitempty"`
	Additions    int        `json:"additions"`
	Deletions    int        `json:"deletions"`
	FilesChanged int        `json:"files_changed"`
	CommitCount  int        `json:"commit_count"`
	Comments     int        `json:"comments"`
	Reviews      []Review   `json:"reviews,omitempty"`
	URL          string     `json:"url"`

	// Derived fields
	TimeToMerge       *time.Duration `json:"time_to_merge,omitempty"`
	TimeToFirstReview *time.Duration `json:"time_to_first_review,omitempty"`
}

// IsMerged returns true if the PR has been merged
func (pr *PullRequest) IsMerged() bool {
	return pr.State == PRStateMerged || pr.MergedAt != nil
}

// TotalChanges returns the total lines changed (additions + deletions)
func (pr *PullRequest) TotalChanges() int {
	return pr.Additions + pr.Deletions
}

// CalculateTimeToMerge calculates the time from PR creation to merge
func (pr *PullRequest) CalculateTimeToMerge() *time.Duration {
	if pr.MergedAt == nil {
		return nil
	}
	d := pr.MergedAt.Sub(pr.CreatedAt)
	return &d
}

// CalculateTimeToFirstReview calculates the time from PR creation to first review
func (pr *PullRequest) CalculateTimeToFirstReview() *time.Duration {
	if len(pr.Reviews) == 0 {
		return nil
	}

	var firstReview *time.Time
	for _, review := range pr.Reviews {
		if firstReview == nil || review.SubmittedAt.Before(*firstReview) {
			t := review.SubmittedAt
			firstReview = &t
		}
	}

	if firstReview == nil {
		return nil
	}

	d := firstReview.Sub(pr.CreatedAt)
	return &d
}

// PRSize represents the size category of a pull request
type PRSize string

const (
	PRSizeXS PRSize = "xs" // < 10 lines
	PRSizeS  PRSize = "s"  // 10-50 lines
	PRSizeM  PRSize = "m"  // 50-200 lines
	PRSizeL  PRSize = "l"  // 200-500 lines
	PRSizeXL PRSize = "xl" // > 500 lines
)

// Size returns the size category of the PR based on total changes
func (pr *PullRequest) Size() PRSize {
	total := pr.TotalChanges()
	switch {
	case total < 10:
		return PRSizeXS
	case total < 50:
		return PRSizeS
	case total < 200:
		return PRSizeM
	case total < 500:
		return PRSizeL
	default:
		return PRSizeXL
	}
}
