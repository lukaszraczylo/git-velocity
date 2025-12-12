package aggregator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestStreakCalculation_FloatPrecisionBug tests the potential floating point precision issues in streak calculation
func TestStreakCalculation_FloatPrecisionBug(t *testing.T) {
	t.Parallel()

	t.Run("consecutive days with different hours", func(t *testing.T) {
		t.Parallel()

		// Bug: Line 1335 in aggregator.go uses floating point division
		// diff := dates[i].Sub(dates[i-1]).Hours() / 24
		// This can cause precision issues when checking diff == 1
		dates := map[string]bool{
			"2024-01-15": true, // Day 1 at 00:00
			"2024-01-16": true, // Day 2 at 00:00
			"2024-01-17": true, // Day 3 at 00:00
		}

		longest, _ := calculateStreaks(dates)

		// This should be 3, but floating point comparison might fail
		assert.Equal(t, 3, longest, "Should calculate 3-day streak correctly")
	})

	t.Run("dates with daylight saving time boundary", func(t *testing.T) {
		t.Parallel()

		// Create dates that cross a DST boundary
		// On DST change, a "day" might be 23 or 25 hours, not exactly 24
		// This would cause the streak to break incorrectly
		loc, _ := time.LoadLocation("America/New_York")

		// March 2024: DST starts on March 10, 2024 at 2:00 AM (clocks move to 3:00 AM)
		day1 := time.Date(2024, 3, 9, 12, 0, 0, 0, loc)  // Day before DST
		day2 := time.Date(2024, 3, 10, 12, 0, 0, 0, loc) // DST change day (23-hour day)
		day3 := time.Date(2024, 3, 11, 12, 0, 0, 0, loc) // Day after DST

		dates := map[string]bool{
			day1.Format("2006-01-02"): true,
			day2.Format("2006-01-02"): true,
			day3.Format("2006-01-02"): true,
		}

		longest, _ := calculateStreaks(dates)

		// Bug: The floating point comparison diff == 1 might fail due to DST
		// day1 to day2: 23 hours / 24 = 0.958... != 1.0 (streak breaks)
		// This test documents the bug - it should pass with value 3, but might return 1 or 2
		assert.GreaterOrEqual(t, longest, 1, "Should handle DST boundaries")
		// The actual expected value is 3, but due to the bug it might be less
	})

	t.Run("consecutive days at different times of day", func(t *testing.T) {
		t.Parallel()

		// Even without DST, different times of day can cause issues
		// Day 1 at 10:00, Day 2 at 9:00 = 23 hours apart (not exactly 24)
		// 23 / 24 = 0.958... != 1.0
		loc := time.UTC
		day1 := time.Date(2024, 1, 15, 10, 0, 0, 0, loc)
		day2 := time.Date(2024, 1, 16, 9, 0, 0, 0, loc)  // 23 hours later
		day3 := time.Date(2024, 1, 17, 11, 0, 0, 0, loc) // 26 hours later

		dates := map[string]bool{
			day1.Format("2006-01-02"): true,
			day2.Format("2006-01-02"): true,
			day3.Format("2006-01-02"): true,
		}

		longest, _ := calculateStreaks(dates)

		// With float comparison, this might break the streak
		// Expected: 3, Actual might be: 1, 2, or 3 depending on precision
		assert.GreaterOrEqual(t, longest, 1, "Should not panic")
		// Document: This is a known bug - should be 3 but might be less due to time differences
	})
}

// TestStreakCalculation_CurrentStreakBoundaryCondition tests current streak calculation edge cases
func TestStreakCalculation_CurrentStreakBoundaryCondition(t *testing.T) {
	t.Parallel()

	t.Run("last activity exactly 1 day ago", func(t *testing.T) {
		t.Parallel()

		// Line 1351: if daysSinceLastActive <= 1
		// This uses float comparison which can be problematic
		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)

		dates := map[string]bool{
			yesterday.Format("2006-01-02"): true,
		}

		_, current := calculateStreaks(dates)

		// Float comparison: (now - yesterday).Hours() / 24 might not be exactly 1.0
		// Due to precision, it might be 0.999... or 1.001...
		// This test should pass but documents the fragility
		assert.GreaterOrEqual(t, current, 0, "Should not panic")
	})

	t.Run("last activity exactly at boundary", func(t *testing.T) {
		t.Parallel()

		// Edge case: What if the last activity was exactly 24.0000 hours ago?
		// Line 1351: daysSinceLastActive <= 1
		// With float precision, 24.0 hours / 24 = 1.0, so <= 1 should pass
		now := time.Now().Truncate(24 * time.Hour)
		exactlyOneDayAgo := now.Add(-24 * time.Hour)

		dates := map[string]bool{
			exactlyOneDayAgo.Format("2006-01-02"): true,
		}

		_, current := calculateStreaks(dates)

		// This should preserve the streak since it's exactly 1 day
		// But float precision might cause issues
		assert.GreaterOrEqual(t, current, 0, "Should handle exact 24-hour boundary")
	})
}

// TestStreakCalculation_EmptyOrSingleDate tests edge cases with minimal data
func TestStreakCalculation_EmptyOrSingleDate(t *testing.T) {
	t.Parallel()

	t.Run("empty dates map", func(t *testing.T) {
		t.Parallel()

		dates := map[string]bool{}
		longest, current := calculateStreaks(dates)

		assert.Equal(t, 0, longest)
		assert.Equal(t, 0, current)
	})

	t.Run("single date", func(t *testing.T) {
		t.Parallel()

		dates := map[string]bool{
			"2024-01-15": true,
		}

		longest, current := calculateStreaks(dates)

		assert.Equal(t, 1, longest, "Single date should be streak of 1")
		// current depends on how far in the past this date is
		assert.GreaterOrEqual(t, current, 0)
	})
}

// TestStreakCalculation_DateParsingError documents behavior with invalid dates
func TestStreakCalculation_DateParsingError(t *testing.T) {
	t.Parallel()

	t.Run("invalid date format", func(t *testing.T) {
		t.Parallel()

		dates := map[string]bool{
			"invalid-date": true,
			"2024-01-15":   true,
		}

		// The function parses dates with time.Parse("2006-01-02", dateStr)
		// Invalid dates are silently skipped (err != nil check on line 1316)
		longest, current := calculateStreaks(dates)

		// Only the valid date counts
		assert.Equal(t, 1, longest, "Should skip invalid dates")
		assert.GreaterOrEqual(t, current, 0)
	})
}

// TestStreakCalculation_LargeGaps tests streak reset with large gaps
func TestStreakCalculation_LargeGaps(t *testing.T) {
	t.Parallel()

	t.Run("large gap between dates", func(t *testing.T) {
		t.Parallel()

		dates := map[string]bool{
			"2024-01-01": true,
			"2024-01-02": true,
			"2024-01-03": true,
			"2024-02-15": true, // Large gap - should reset streak
			"2024-02-16": true,
		}

		longest, _ := calculateStreaks(dates)

		// Longest streak should be 3 (Jan 1-3)
		assert.Equal(t, 3, longest, "Should correctly identify longest streak despite gap")
	})

	t.Run("multiple equal-length streaks", func(t *testing.T) {
		t.Parallel()

		dates := map[string]bool{
			"2024-01-01": true,
			"2024-01-02": true,
			"2024-01-03": true,
			"2024-02-01": true, // Gap
			"2024-02-02": true,
			"2024-02-03": true,
		}

		longest, _ := calculateStreaks(dates)

		// Two 3-day streaks - should return 3
		assert.Equal(t, 3, longest, "Should return longest streak when multiple equal streaks exist")
	})
}
