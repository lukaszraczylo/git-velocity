package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v68/github"
)

// DateFilterResult represents the result of date filtering
type DateFilterResult int

const (
	// DateInclude means the item is within the date range
	DateInclude DateFilterResult = iota
	// DateTooNew means the item is newer than the 'until' date
	DateTooNew
	// DateTooOld means the item is older than the 'since' date
	DateTooOld
)

// FilterByDate checks if a time falls within the specified date range
func FilterByDate(t time.Time, since, until *time.Time) DateFilterResult {
	if until != nil && t.After(*until) {
		return DateTooNew
	}
	if since != nil && t.Before(*since) {
		return DateTooOld
	}
	return DateInclude
}

// PageFetcher is a generic interface for fetching paginated resources
type PageFetcher[T any, R any] interface {
	// Fetch retrieves a page of items
	Fetch(ctx context.Context, page int) (items []T, resp *github.Response, err error)
	// Convert transforms a raw item into the result type
	Convert(item T) R
	// Filter determines if an item should be included based on date range
	// Returns DateInclude to include, DateTooNew/DateTooOld to exclude
	Filter(item T) DateFilterResult
	// ShouldSkip returns true if the item should be skipped entirely (e.g., PRs in issues list)
	ShouldSkip(item T) bool
}

// FetchConfig holds configuration for paginated fetching
type FetchConfig struct {
	// ResourceName is used for progress messages (e.g., "issues", "pull requests")
	ResourceName string
	// EarlyTermination enables stopping when all items on a page are too old
	EarlyTermination bool
	// EarlyTerminationThreshold is the number of consecutive old pages before stopping
	EarlyTerminationThreshold int
	// Quiet suppresses per-page progress messages (useful for sub-fetches like reviews)
	Quiet bool
}

// DefaultFetchConfig returns sensible defaults
func DefaultFetchConfig(resourceName string) FetchConfig {
	return FetchConfig{
		ResourceName:              resourceName,
		EarlyTermination:          true,
		EarlyTerminationThreshold: 2,
	}
}

// FetchAllPages fetches all pages of a resource with caching, filtering, and early termination
func FetchAllPages[T any, R any](
	ctx context.Context,
	c *Client,
	cacheKey string,
	config FetchConfig,
	fetcher PageFetcher[T, R],
) ([]R, error) {
	// Check cache first (skip if no cache key provided)
	if cacheKey != "" {
		if cached, ok := c.cache.Get(cacheKey); ok {
			if results, ok := cached.([]R); ok {
				c.progress(fmt.Sprintf("      Using cached %s data", config.ResourceName))
				return results, nil
			}
		}
	}

	var allResults []R
	page := 1
	consecutiveOldPages := 0

	for {
		items, resp, err := fetcher.Fetch(ctx, page)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", config.ResourceName, err)
		}

		// Safety check for nil response
		if resp == nil {
			break
		}

		if !config.Quiet {
			c.progress(fmt.Sprintf("      Fetching %s page %d (%d %s so far)...",
				config.ResourceName, page, len(allResults), config.ResourceName))
		}

		oldInPage := 0
		totalEligible := 0

		for _, item := range items {
			// Skip items that should be filtered out entirely (e.g., PRs in issues API)
			if fetcher.ShouldSkip(item) {
				continue
			}

			totalEligible++

			// Apply date filtering
			switch fetcher.Filter(item) {
			case DateTooNew:
				continue
			case DateTooOld:
				oldInPage++
				continue
			case DateInclude:
				allResults = append(allResults, fetcher.Convert(item))
			}
		}

		// Early termination logic
		if config.EarlyTermination && totalEligible > 0 && oldInPage == totalEligible {
			consecutiveOldPages++
			if consecutiveOldPages >= config.EarlyTerminationThreshold {
				if !config.Quiet {
					c.progress(fmt.Sprintf("      Reached %s older than date range, stopping early (page %d)",
						config.ResourceName, page))
				}
				break
			}
		} else {
			consecutiveOldPages = 0
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	// Cache results (skip if no cache key provided)
	if cacheKey != "" {
		c.cache.Set(cacheKey, allResults)
	}

	return allResults, nil
}

// SimpleFetcher is a helper for creating simple fetchers without date filtering
type SimpleFetcher[T any, R any] struct {
	FetchFn   func(ctx context.Context, page int) ([]T, *github.Response, error)
	ConvertFn func(item T) R
}

func (f *SimpleFetcher[T, R]) Fetch(ctx context.Context, page int) ([]T, *github.Response, error) {
	return f.FetchFn(ctx, page)
}

func (f *SimpleFetcher[T, R]) Convert(item T) R {
	return f.ConvertFn(item)
}

func (f *SimpleFetcher[T, R]) Filter(item T) DateFilterResult {
	return DateInclude // No filtering
}

func (f *SimpleFetcher[T, R]) ShouldSkip(item T) bool {
	return false
}

// DateFilteredFetcher extends SimpleFetcher with date filtering
type DateFilteredFetcher[T any, R any] struct {
	FetchFn   func(ctx context.Context, page int) ([]T, *github.Response, error)
	ConvertFn func(item T) R
	GetDateFn func(item T) time.Time
	SkipFn    func(item T) bool
	Since     *time.Time
	Until     *time.Time
}

func (f *DateFilteredFetcher[T, R]) Fetch(ctx context.Context, page int) ([]T, *github.Response, error) {
	return f.FetchFn(ctx, page)
}

func (f *DateFilteredFetcher[T, R]) Convert(item T) R {
	return f.ConvertFn(item)
}

func (f *DateFilteredFetcher[T, R]) Filter(item T) DateFilterResult {
	return FilterByDate(f.GetDateFn(item), f.Since, f.Until)
}

func (f *DateFilteredFetcher[T, R]) ShouldSkip(item T) bool {
	if f.SkipFn != nil {
		return f.SkipFn(item)
	}
	return false
}

// WithRetry wraps a fetch function with retry logic
func (c *Client) WithRetry(ctx context.Context, operation string, fn func() error) error {
	return c.retryWithBackoff(ctx, operation, fn)
}

// EnrichingFetcher extends DateFilteredFetcher with per-item enrichment
// This is useful when you need to fetch additional details for each item (e.g., commit details)
type EnrichingFetcher[T any, R any] struct {
	FetchFn   func(ctx context.Context, page int) ([]T, *github.Response, error)
	EnrichFn  func(ctx context.Context, item T) (R, error) // Enriches and converts in one step
	GetDateFn func(item T) time.Time
	SkipFn    func(item T) bool
	Since     *time.Time
	Until     *time.Time
}

func (f *EnrichingFetcher[T, R]) Fetch(ctx context.Context, page int) ([]T, *github.Response, error) {
	return f.FetchFn(ctx, page)
}

func (f *EnrichingFetcher[T, R]) Convert(item T) R {
	// This won't be used - FetchAllPagesWithEnrichment handles enrichment
	var zero R
	return zero
}

func (f *EnrichingFetcher[T, R]) Filter(item T) DateFilterResult {
	return FilterByDate(f.GetDateFn(item), f.Since, f.Until)
}

func (f *EnrichingFetcher[T, R]) ShouldSkip(item T) bool {
	if f.SkipFn != nil {
		return f.SkipFn(item)
	}
	return false
}

// FetchAllPagesWithEnrichment is like FetchAllPages but calls EnrichFn for each item
// This is useful when you need to make additional API calls per item (e.g., fetching commit details)
func FetchAllPagesWithEnrichment[T any, R any](
	ctx context.Context,
	c *Client,
	cacheKey string,
	config FetchConfig,
	fetcher *EnrichingFetcher[T, R],
	progressEvery int, // Report progress every N items (0 = disabled)
) ([]R, error) {
	// Check cache first
	if cacheKey != "" {
		if cached, ok := c.cache.Get(cacheKey); ok {
			if results, ok := cached.([]R); ok {
				c.progress(fmt.Sprintf("      Using cached %s data", config.ResourceName))
				return results, nil
			}
		}
	}

	var allResults []R
	page := 1

	for {
		items, resp, err := fetcher.Fetch(ctx, page)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", config.ResourceName, err)
		}

		// Safety check for nil response
		if resp == nil {
			break
		}

		if !config.Quiet {
			c.progress(fmt.Sprintf("      Fetching %s page %d (%d %s so far)...",
				config.ResourceName, page, len(allResults), config.ResourceName))
		}

		itemsInPage := 0
		for i, item := range items {
			// Skip items that should be filtered out entirely
			if fetcher.ShouldSkip(item) {
				continue
			}

			// Apply date filtering
			if fetcher.Filter(item) != DateInclude {
				continue
			}

			// Enrich the item (this may make additional API calls)
			enriched, err := fetcher.EnrichFn(ctx, item)
			if err != nil {
				c.progress(fmt.Sprintf("      Warning: failed to enrich item: %v", err))
				continue
			}

			allResults = append(allResults, enriched)
			itemsInPage++

			// Progress reporting
			if progressEvery > 0 && (i+1)%progressEvery == 0 {
				c.progress(fmt.Sprintf("      Processing item %d/%d on page %d...", i+1, len(items), page))
			}
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	// Cache results
	if cacheKey != "" {
		c.cache.Set(cacheKey, allResults)
	}

	return allResults, nil
}
