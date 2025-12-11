// Achievement category mappings and utilities

// Define achievement categories and their tier ordering (highest tier last)
const achievementCategories = {
  // Commits
  'commit': ['commit-1', 'commit-10', 'commit-50', 'commit-100', 'commit-500', 'commit-1000'],
  // PRs opened
  'pr': ['pr-1', 'pr-10', 'pr-25', 'pr-50', 'pr-100', 'pr-250'],
  // Reviews
  'review': ['review-1', 'review-10', 'review-25', 'review-50', 'review-100', 'review-250'],
  // Review comments
  'comment': ['comment-10', 'comment-50', 'comment-100', 'comment-250', 'comment-500'],
  // Lines added
  'lines-added': ['lines-added-100', 'lines-added-1000', 'lines-added-5000', 'lines-added-10000', 'lines-added-50000'],
  // Lines deleted
  'lines-deleted': ['lines-deleted-100', 'lines-deleted-500', 'lines-deleted-1000', 'lines-deleted-5000', 'lines-deleted-10000'],
  // Review time
  'review-time': ['review-time-24h', 'review-time-4h', 'review-time-1h'],
  // Multi-repo
  'repo': ['repo-2', 'repo-5', 'repo-10'],
  // Unique reviewees
  'reviewees': ['reviewees-3', 'reviewees-10', 'reviewees-25'],
  // Large PRs
  'large-pr': ['large-pr-500', 'large-pr-1000', 'large-pr-5000'],
  // Small PRs
  'small-pr': ['small-pr-5', 'small-pr-10', 'small-pr-25', 'small-pr-50'],
  // Perfect PRs
  'perfect-pr': ['perfect-pr-1', 'perfect-pr-5', 'perfect-pr-10', 'perfect-pr-25'],
  // Active days
  'active': ['active-7', 'active-30', 'active-60', 'active-100'],
  // Streaks
  'streak': ['streak-3', 'streak-7', 'streak-14', 'streak-30'],
  // Work week streaks
  'workweek': ['workweek-3', 'workweek-5', 'workweek-10', 'workweek-20'],
  // Early bird
  'earlybird': ['earlybird-10', 'earlybird-25', 'earlybird-50', 'earlybird-100'],
  // Night owl
  'nightowl': ['nightowl-10', 'nightowl-25', 'nightowl-50', 'nightowl-100'],
  // Midnight coder
  'midnight': ['midnight-5', 'midnight-10', 'midnight-25', 'midnight-50'],
  // Weekend warrior
  'weekend': ['weekend-5', 'weekend-10', 'weekend-25', 'weekend-50'],
  // Out of hours
  'ooh': ['ooh-10', 'ooh-25', 'ooh-50', 'ooh-100'],
  // Documentation added
  'docs': ['docs-100', 'docs-500', 'docs-1000', 'docs-2500', 'docs-5000'],
  // Documentation deleted
  'docs-del': ['docs-del-50', 'docs-del-200', 'docs-del-500', 'docs-del-1000', 'docs-del-2500'],
  // Issues opened
  'issue': ['issue-1', 'issue-5', 'issue-10', 'issue-25', 'issue-50'],
  // Issues closed
  'issue-close': ['issue-close-1', 'issue-close-5', 'issue-close-10', 'issue-close-25', 'issue-close-50'],
  // Issue comments
  'issue-comment': ['issue-comment-5', 'issue-comment-10', 'issue-comment-25', 'issue-comment-50', 'issue-comment-100'],
  // Issue references in commits
  'issue-ref': ['issue-ref-5', 'issue-ref-10', 'issue-ref-25', 'issue-ref-50', 'issue-ref-100'],
}

// Get the category for an achievement ID
export function getAchievementCategory(achievementId) {
  for (const [category, tiers] of Object.entries(achievementCategories)) {
    if (tiers.includes(achievementId)) {
      return category
    }
  }
  return null
}

// Get the tier index within a category (higher = better)
export function getAchievementTier(achievementId) {
  const category = getAchievementCategory(achievementId)
  if (!category) return -1
  return achievementCategories[category].indexOf(achievementId)
}

/**
 * Filter achievements to show only the highest tier in each category
 * @param {string[]} achievements - Array of achievement IDs
 * @returns {string[]} - Filtered array with only highest tier per category
 */
export function getHighestTierAchievements(achievements) {
  if (!achievements || !achievements.length) return []

  // Group achievements by category, keeping only the highest tier
  const highestByCategory = {}

  for (const achievementId of achievements) {
    const category = getAchievementCategory(achievementId)
    if (!category) {
      // Unknown achievement, keep it
      highestByCategory[achievementId] = { id: achievementId, tier: -1 }
      continue
    }

    const tier = getAchievementTier(achievementId)

    if (!highestByCategory[category] || tier > highestByCategory[category].tier) {
      highestByCategory[category] = { id: achievementId, tier }
    }
  }

  // Return just the achievement IDs, sorted by tier (highest first)
  return Object.values(highestByCategory)
    .sort((a, b) => b.tier - a.tier)
    .map(item => item.id)
}

/**
 * Get a priority score for sorting achievements (higher = more impressive)
 * Categories are weighted to show most impressive achievements first
 */
const categoryPriority = {
  'commit': 10,
  'pr': 9,
  'review': 8,
  'lines-added': 7,
  'perfect-pr': 6,
  'issue': 5.5,
  'issue-close': 5.4,
  'streak': 5,
  'active': 4,
  'issue-ref': 3.5,
  'issue-comment': 3.2,
  'review-time': 3,
  'docs': 2,
}

export function getAchievementPriority(achievementId) {
  const category = getAchievementCategory(achievementId)
  const basePriority = categoryPriority[category] || 0
  const tier = getAchievementTier(achievementId)
  // Combine category priority with tier (tier adds 0.1 per level)
  return basePriority + (tier * 0.1)
}

/**
 * Get highest tier achievements, sorted by importance
 * @param {string[]} achievements - Array of achievement IDs
 * @param {number} limit - Maximum number to return
 * @returns {string[]} - Filtered and sorted array
 */
export function getTopAchievements(achievements, limit = 6) {
  const highest = getHighestTierAchievements(achievements)

  // Sort by priority (most impressive first)
  highest.sort((a, b) => getAchievementPriority(b) - getAchievementPriority(a))

  return highest.slice(0, limit)
}
