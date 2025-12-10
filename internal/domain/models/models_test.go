package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuthor_DisplayName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		author   Author
		expected string
	}{
		{
			name:     "prefers name over login",
			author:   Author{Login: "johndoe", Name: "John Doe", Email: "john@example.com"},
			expected: "John Doe",
		},
		{
			name:     "falls back to login",
			author:   Author{Login: "johndoe", Email: "john@example.com"},
			expected: "johndoe",
		},
		{
			name:     "falls back to email",
			author:   Author{Email: "john@example.com"},
			expected: "john@example.com",
		},
		{
			name:     "returns Unknown when empty",
			author:   Author{},
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.author.DisplayName())
		})
	}
}

func TestCommit_TotalChanges(t *testing.T) {
	t.Parallel()

	commit := Commit{Additions: 100, Deletions: 50}
	assert.Equal(t, 150, commit.TotalChanges())
}

func TestCommit_ShortSHA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sha      string
		expected string
	}{
		{
			name:     "full SHA",
			sha:      "abc123456789def",
			expected: "abc1234",
		},
		{
			name:     "short SHA",
			sha:      "abc",
			expected: "abc",
		},
		{
			name:     "exactly 7 chars",
			sha:      "abc1234",
			expected: "abc1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			commit := Commit{SHA: tt.sha}
			assert.Equal(t, tt.expected, commit.ShortSHA())
		})
	}
}

func TestCommit_ShortMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "single line",
			message:  "Fix bug in login",
			expected: "Fix bug in login",
		},
		{
			name:     "multiline",
			message:  "Fix bug in login\n\nThis fixes the issue where users couldn't log in.",
			expected: "Fix bug in login",
		},
		{
			name:     "empty",
			message:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			commit := Commit{Message: tt.message}
			assert.Equal(t, tt.expected, commit.ShortMessage())
		})
	}
}

func TestPullRequest_IsMerged(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name     string
		pr       PullRequest
		expected bool
	}{
		{
			name:     "merged state",
			pr:       PullRequest{State: PRStateMerged},
			expected: true,
		},
		{
			name:     "has merged_at",
			pr:       PullRequest{State: PRStateClosed, MergedAt: &now},
			expected: true,
		},
		{
			name:     "open PR",
			pr:       PullRequest{State: PRStateOpen},
			expected: false,
		},
		{
			name:     "closed without merge",
			pr:       PullRequest{State: PRStateClosed},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.pr.IsMerged())
		})
	}
}

func TestPullRequest_Size(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		additions int
		deletions int
		expected  PRSize
	}{
		{
			name:      "xs",
			additions: 5,
			deletions: 3,
			expected:  PRSizeXS,
		},
		{
			name:      "s",
			additions: 30,
			deletions: 15,
			expected:  PRSizeS,
		},
		{
			name:      "m",
			additions: 100,
			deletions: 50,
			expected:  PRSizeM,
		},
		{
			name:      "l",
			additions: 300,
			deletions: 100,
			expected:  PRSizeL,
		},
		{
			name:      "xl",
			additions: 400,
			deletions: 200,
			expected:  PRSizeXL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pr := PullRequest{Additions: tt.additions, Deletions: tt.deletions}
			assert.Equal(t, tt.expected, pr.Size())
		})
	}
}

func TestPullRequest_CalculateTimeToMerge(t *testing.T) {
	t.Parallel()

	t.Run("returns duration when merged", func(t *testing.T) {
		t.Parallel()

		created := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
		merged := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
		pr := PullRequest{CreatedAt: created, MergedAt: &merged}

		result := pr.CalculateTimeToMerge()
		assert.NotNil(t, result)
		assert.Equal(t, 4*time.Hour, *result)
	})

	t.Run("returns nil when not merged", func(t *testing.T) {
		t.Parallel()

		pr := PullRequest{CreatedAt: time.Now()}
		assert.Nil(t, pr.CalculateTimeToMerge())
	})
}

func TestPullRequest_CalculateTimeToFirstReview(t *testing.T) {
	t.Parallel()

	t.Run("returns duration to first review", func(t *testing.T) {
		t.Parallel()

		created := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
		review1 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		review2 := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)

		pr := PullRequest{
			CreatedAt: created,
			Reviews: []Review{
				{SubmittedAt: review2},
				{SubmittedAt: review1}, // Earlier review
			},
		}

		result := pr.CalculateTimeToFirstReview()
		assert.NotNil(t, result)
		assert.Equal(t, 2*time.Hour, *result)
	})

	t.Run("returns nil when no reviews", func(t *testing.T) {
		t.Parallel()

		pr := PullRequest{CreatedAt: time.Now(), Reviews: []Review{}}
		assert.Nil(t, pr.CalculateTimeToFirstReview())
	})
}

func TestReview_IsApproval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    ReviewState
		expected bool
	}{
		{name: "approved", state: ReviewApproved, expected: true},
		{name: "changes requested", state: ReviewChangesRequested, expected: false},
		{name: "commented", state: ReviewCommented, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := Review{State: tt.state}
			assert.Equal(t, tt.expected, r.IsApproval())
		})
	}
}

func TestReview_RequestsChanges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    ReviewState
		expected bool
	}{
		{name: "approved", state: ReviewApproved, expected: false},
		{name: "changes requested", state: ReviewChangesRequested, expected: true},
		{name: "commented", state: ReviewCommented, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := Review{State: tt.state}
			assert.Equal(t, tt.expected, r.RequestsChanges())
		})
	}
}

func TestReview_IsSubstantive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		review   Review
		expected bool
	}{
		{
			name:     "has body",
			review:   Review{Body: "Good work!"},
			expected: true,
		},
		{
			name:     "has comments",
			review:   Review{CommentsCount: 3},
			expected: true,
		},
		{
			name:     "requests changes",
			review:   Review{State: ReviewChangesRequested},
			expected: true,
		},
		{
			name:     "empty approval",
			review:   Review{State: ReviewApproved},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.review.IsSubstantive())
		})
	}
}

func TestIssue_IsClosed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    IssueState
		expected bool
	}{
		{name: "open", state: IssueStateOpen, expected: false},
		{name: "closed", state: IssueStateClosed, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			issue := Issue{State: tt.state}
			assert.Equal(t, tt.expected, issue.IsClosed())
		})
	}
}

func TestIssue_CalculateTimeToClose(t *testing.T) {
	t.Parallel()

	t.Run("returns duration when closed", func(t *testing.T) {
		t.Parallel()

		created := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
		closed := time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC)
		issue := Issue{CreatedAt: created, ClosedAt: &closed}

		result := issue.CalculateTimeToClose()
		assert.NotNil(t, result)
		assert.Equal(t, 48*time.Hour, *result)
	})

	t.Run("returns nil when not closed", func(t *testing.T) {
		t.Parallel()

		issue := Issue{CreatedAt: time.Now()}
		assert.Nil(t, issue.CalculateTimeToClose())
	})
}

func TestPullRequest_TotalChanges(t *testing.T) {
	t.Parallel()

	pr := PullRequest{Additions: 200, Deletions: 100}
	assert.Equal(t, 300, pr.TotalChanges())
}
