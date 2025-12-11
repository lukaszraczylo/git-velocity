<script setup>
defineProps({
  achievementId: { type: String, required: true },
  size: { type: String, default: 'md' }, // sm, md, lg
  showLabel: { type: Boolean, default: false }
})

// Tier colors based on threshold (1, 10, 25, 50, 100, 250, 500, 1000, 5000, 10000, 25000+)
const tierGradients = {
  1: 'from-stone-400 to-stone-500',      // Bronze - tier 1
  2: 'from-green-400 to-emerald-500',    // Green - tier 10
  3: 'from-blue-400 to-indigo-500',      // Blue - tier 25
  4: 'from-purple-400 to-violet-500',    // Purple - tier 50
  5: 'from-yellow-400 to-amber-500',     // Gold - tier 100
  6: 'from-orange-400 to-red-500',       // Orange - tier 250
  7: 'from-red-500 to-rose-600',         // Red - tier 500
  8: 'from-pink-500 to-fuchsia-600',     // Pink - tier 1000
  9: 'from-cyan-400 to-teal-500',        // Cyan - tier 5000
  10: 'from-emerald-400 to-cyan-500',    // Emerald - tier 10000
  11: 'from-violet-500 to-purple-600',   // Legendary - tier 25000+
}

// Get tier from threshold number
const getTierFromThreshold = (threshold) => {
  const tiers = [1, 10, 25, 50, 100, 250, 500, 1000, 5000, 10000, 25000]
  for (let i = tiers.length - 1; i >= 0; i--) {
    if (threshold >= tiers[i]) return i + 1
  }
  return 1
}

// Extract threshold from achievement ID (e.g., "commit-100" -> 100, "docs-del-50" -> 50)
const extractThreshold = (id) => {
  const match = id.match(/(\d+)$/)
  if (match) return parseInt(match[1], 10)
  return 50 // Default for special achievements
}

// Achievement definitions matching the Go backend (internal/config/schema.go)
const achievements = {
  // ===== COMMIT COUNT (Tiers: 1, 10, 50, 100, 500, 1000) =====
  'commit-1': { name: 'First Steps', description: 'Made your first commit', icon: 'fa-baby' },
  'commit-10': { name: 'Getting Started', description: 'Made 10 commits', icon: 'fa-seedling' },
  'commit-50': { name: 'Contributor', description: 'Made 50 commits', icon: 'fa-code' },
  'commit-100': { name: 'Committed', description: 'Made 100 commits', icon: 'fa-fire' },
  'commit-500': { name: 'Code Machine', description: 'Made 500 commits', icon: 'fa-robot' },
  'commit-1000': { name: 'Code Warrior', description: 'Made 1000 commits', icon: 'fa-crown' },

  // ===== PR OPENED (Tiers: 1, 10, 25, 50, 100, 250) =====
  'pr-1': { name: 'PR Pioneer', description: 'Opened your first pull request', icon: 'fa-code-pull-request' },
  'pr-10': { name: 'PR Regular', description: 'Opened 10 pull requests', icon: 'fa-code-branch' },
  'pr-25': { name: 'PR Pro', description: 'Opened 25 pull requests', icon: 'fa-code-compare' },
  'pr-50': { name: 'Merge Master', description: 'Opened 50 pull requests', icon: 'fa-code-merge' },
  'pr-100': { name: 'PR Champion', description: 'Opened 100 pull requests', icon: 'fa-trophy' },
  'pr-250': { name: 'PR Legend', description: 'Opened 250 pull requests', icon: 'fa-medal' },

  // ===== REVIEWS (Tiers: 1, 10, 25, 50, 100, 250) =====
  'review-1': { name: 'First Review', description: 'Reviewed your first pull request', icon: 'fa-magnifying-glass' },
  'review-10': { name: 'Reviewer', description: 'Reviewed 10 pull requests', icon: 'fa-eye' },
  'review-25': { name: 'Review Regular', description: 'Reviewed 25 pull requests', icon: 'fa-glasses' },
  'review-50': { name: 'Review Expert', description: 'Reviewed 50 pull requests', icon: 'fa-user-check' },
  'review-100': { name: 'Review Guru', description: 'Reviewed 100 pull requests', icon: 'fa-user-graduate' },
  'review-250': { name: 'Review Master', description: 'Reviewed 250 pull requests', icon: 'fa-award' },

  // ===== REVIEW COMMENTS (Tiers: 10, 50, 100, 250, 500) =====
  'comment-10': { name: 'Commentator', description: 'Left 10 PR review comments', icon: 'fa-comment' },
  'comment-50': { name: 'Feedback Giver', description: 'Left 50 PR review comments', icon: 'fa-comments' },
  'comment-100': { name: 'Code Critic', description: 'Left 100 PR review comments', icon: 'fa-comment-dots' },
  'comment-250': { name: 'Feedback Expert', description: 'Left 250 PR review comments', icon: 'fa-message' },
  'comment-500': { name: 'Comment Champion', description: 'Left 500 PR review comments', icon: 'fa-scroll' },

  // ===== LINES ADDED (Tiers: 100, 1000, 5000, 10000, 50000) =====
  'lines-added-100': { name: 'First Hundred', description: 'Added 100 lines of code', icon: 'fa-plus' },
  'lines-added-1000': { name: 'Thousand Lines', description: 'Added 1000 lines of code', icon: 'fa-layer-group' },
  'lines-added-5000': { name: 'Five Thousand', description: 'Added 5000 lines of code', icon: 'fa-cubes' },
  'lines-added-10000': { name: 'Ten Thousand', description: 'Added 10000 lines of code', icon: 'fa-mountain' },
  'lines-added-50000': { name: 'Code Mountain', description: 'Added 50000 lines of code', icon: 'fa-mountain-sun' },

  // ===== LINES DELETED (Tiers: 100, 500, 1000, 5000, 10000) =====
  'lines-deleted-100': { name: 'Tidying Up', description: 'Deleted 100 lines of code', icon: 'fa-eraser' },
  'lines-deleted-500': { name: 'Spring Cleaning', description: 'Deleted 500 lines of code', icon: 'fa-broom' },
  'lines-deleted-1000': { name: 'Code Cleaner', description: 'Deleted 1000 lines of code', icon: 'fa-trash-can' },
  'lines-deleted-5000': { name: 'Refactoring Hero', description: 'Deleted 5000 lines of code', icon: 'fa-recycle' },
  'lines-deleted-10000': { name: 'Deletion Master', description: 'Deleted 10000 lines of code', icon: 'fa-dumpster-fire' },

  // ===== REVIEW RESPONSE TIME (Tiers: 24h, 4h, 1h) =====
  'review-time-24h': { name: 'Same Day Reviewer', description: 'Average review response under 24 hours', icon: 'fa-clock' },
  'review-time-4h': { name: 'Quick Responder', description: 'Average review response under 4 hours', icon: 'fa-stopwatch' },
  'review-time-1h': { name: 'Speed Demon', description: 'Average review response under 1 hour', icon: 'fa-bolt' },

  // ===== MULTI-REPO (Tiers: 2, 5, 10) =====
  'repo-2': { name: 'Multi-Repo', description: 'Contributed to 2 repositories', icon: 'fa-folder' },
  'repo-5': { name: 'Repo Explorer', description: 'Contributed to 5 repositories', icon: 'fa-folder-tree' },
  'repo-10': { name: 'Repo Master', description: 'Contributed to 10 repositories', icon: 'fa-network-wired' },

  // ===== UNIQUE REVIEWEES (Tiers: 3, 10, 25) =====
  'reviewees-3': { name: 'Helpful Colleague', description: 'Reviewed PRs from 3 different contributors', icon: 'fa-user-group' },
  'reviewees-10': { name: 'Team Player', description: 'Reviewed PRs from 10 different contributors', icon: 'fa-people-group' },
  'reviewees-25': { name: 'Community Pillar', description: 'Reviewed PRs from 25 different contributors', icon: 'fa-people-roof' },

  // ===== PR SIZE - LARGE (Tiers: 500, 1000, 5000) =====
  'large-pr-500': { name: 'Big Change', description: 'Merged a PR with 500+ lines changed', icon: 'fa-expand' },
  'large-pr-1000': { name: 'Heavy Lifter', description: 'Merged a PR with 1000+ lines changed', icon: 'fa-weight-hanging' },
  'large-pr-5000': { name: 'Mega Merge', description: 'Merged a PR with 5000+ lines changed', icon: 'fa-dumbbell' },

  // ===== SMALL PRs (Tiers: 5, 10, 25, 50) =====
  'small-pr-5': { name: 'Small Changes', description: 'Merged 5 PRs under 100 lines', icon: 'fa-compress' },
  'small-pr-10': { name: 'Small PR Advocate', description: 'Merged 10 PRs under 100 lines', icon: 'fa-minimize' },
  'small-pr-25': { name: 'Atomic Commits', description: 'Merged 25 PRs under 100 lines', icon: 'fa-atom' },
  'small-pr-50': { name: 'Micro PR Master', description: 'Merged 50 PRs under 100 lines', icon: 'fa-microchip' },

  // ===== PERFECT PRs (Tiers: 1, 5, 10, 25) =====
  'perfect-pr-1': { name: 'First Try', description: '1 PR merged without changes requested', icon: 'fa-check' },
  'perfect-pr-5': { name: 'Clean Code', description: '5 PRs merged without changes requested', icon: 'fa-check-double' },
  'perfect-pr-10': { name: 'Quality Author', description: '10 PRs merged without changes requested', icon: 'fa-circle-check' },
  'perfect-pr-25': { name: 'Flawless', description: '25 PRs merged without changes requested', icon: 'fa-gem' },

  // ===== ACTIVE DAYS (Tiers: 7, 30, 60, 100) =====
  'active-7': { name: 'Week Active', description: 'Active on 7 different days', icon: 'fa-calendar-day' },
  'active-30': { name: 'Month Active', description: 'Active on 30 different days', icon: 'fa-calendar-week' },
  'active-60': { name: 'Consistent Contributor', description: 'Active on 60 different days', icon: 'fa-chart-line' },
  'active-100': { name: 'Dedicated Developer', description: 'Active on 100 different days', icon: 'fa-fire-flame-curved' },

  // ===== LONGEST STREAK (Tiers: 3, 7, 14, 30) =====
  'streak-3': { name: 'Getting Rolling', description: '3 day contribution streak', icon: 'fa-forward' },
  'streak-7': { name: 'Week Warrior', description: '7 day contribution streak', icon: 'fa-calendar-week' },
  'streak-14': { name: 'Two Week Streak', description: '14 day contribution streak', icon: 'fa-fire' },
  'streak-30': { name: 'Month Master', description: '30 day contribution streak', icon: 'fa-calendar-check' },

  // ===== WORK WEEK STREAK (Tiers: 3, 5, 10, 20) =====
  'workweek-3': { name: 'Work Week Start', description: '3 consecutive weekday streak', icon: 'fa-briefcase' },
  'workweek-5': { name: 'Full Work Week', description: '5 consecutive weekday streak', icon: 'fa-building' },
  'workweek-10': { name: 'Two Week Grind', description: '10 consecutive weekday streak', icon: 'fa-business-time' },
  'workweek-20': { name: 'Month of Mondays', description: '20 consecutive weekday streak', icon: 'fa-landmark' },

  // ===== EARLY BIRD (Tiers: 10, 25, 50, 100) =====
  'earlybird-10': { name: 'Early Riser', description: '10 commits before 9am', icon: 'fa-mug-hot' },
  'earlybird-25': { name: 'Morning Person', description: '25 commits before 9am', icon: 'fa-cloud-sun' },
  'earlybird-50': { name: 'Early Bird', description: '50 commits before 9am', icon: 'fa-sun' },
  'earlybird-100': { name: 'Dawn Warrior', description: '100 commits before 9am', icon: 'fa-sunrise' },

  // ===== NIGHT OWL (Tiers: 10, 25, 50, 100) =====
  'nightowl-10': { name: 'Late Worker', description: '10 commits after 9pm', icon: 'fa-cloud-moon' },
  'nightowl-25': { name: 'Evening Coder', description: '25 commits after 9pm', icon: 'fa-moon' },
  'nightowl-50': { name: 'Night Owl', description: '50 commits after 9pm', icon: 'fa-star' },
  'nightowl-100': { name: 'Nocturnal', description: '100 commits after 9pm', icon: 'fa-star-and-crescent' },

  // ===== MIDNIGHT CODER (Tiers: 5, 10, 25, 50) =====
  'midnight-5': { name: 'Night Shift', description: '5 commits between midnight and 4am', icon: 'fa-ghost' },
  'midnight-10': { name: 'Insomniac', description: '10 commits between midnight and 4am', icon: 'fa-bed' },
  'midnight-25': { name: 'Nosferatu', description: '25 commits between midnight and 4am', icon: 'fa-skull' },
  'midnight-50': { name: 'Vampire Coder', description: '50 commits between midnight and 4am', icon: 'fa-skull-crossbones' },

  // ===== WEEKEND WARRIOR (Tiers: 5, 10, 25, 50) =====
  'weekend-5': { name: 'Weekend Work', description: '5 weekend commits', icon: 'fa-couch' },
  'weekend-10': { name: 'Weekend Regular', description: '10 weekend commits', icon: 'fa-house-laptop' },
  'weekend-25': { name: 'Weekend Warrior', description: '25 weekend commits', icon: 'fa-gamepad' },
  'weekend-50': { name: 'No Days Off', description: '50 weekend commits', icon: 'fa-person-running' },

  // ===== OUT OF HOURS (Tiers: 10, 25, 50, 100) =====
  'ooh-10': { name: 'Extra Hours', description: '10 commits outside 9am-5pm', icon: 'fa-clock-rotate-left' },
  'ooh-25': { name: 'Flexible Schedule', description: '25 commits outside 9am-5pm', icon: 'fa-user-clock' },
  'ooh-50': { name: 'Off-Hours Hero', description: '50 commits outside 9am-5pm', icon: 'fa-hourglass-half' },
  'ooh-100': { name: 'Time Bender', description: '100 commits outside 9am-5pm', icon: 'fa-infinity' },

  // ===== DOCUMENTATION & COMMENTS ADDED (Tiers: 100, 500, 1000, 2500, 5000) =====
  'docs-100': { name: 'Documenter', description: 'Added 100 lines of comments/docs', icon: 'fa-file-lines' },
  'docs-500': { name: 'Technical Writer', description: 'Added 500 lines of comments/docs', icon: 'fa-book' },
  'docs-1000': { name: 'Documentation Hero', description: 'Added 1000 lines of comments/docs', icon: 'fa-book-open' },
  'docs-2500': { name: 'Knowledge Keeper', description: 'Added 2500 lines of comments/docs', icon: 'fa-scroll' },
  'docs-5000': { name: 'Code Historian', description: 'Added 5000 lines of comments/docs', icon: 'fa-landmark' },

  // ===== COMMENT CLEANUP (Tiers: 50, 200, 500, 1000, 2500) =====
  'docs-del-50': { name: 'Comment Trimmer', description: 'Removed 50 lines of outdated comments', icon: 'fa-scissors' },
  'docs-del-200': { name: 'Cleanup Crew', description: 'Removed 200 lines of outdated comments', icon: 'fa-broom' },
  'docs-del-500': { name: 'Dead Code Hunter', description: 'Removed 500 lines of outdated comments', icon: 'fa-skull-crossbones' },
  'docs-del-1000': { name: 'Comment Surgeon', description: 'Removed 1000 lines of outdated comments', icon: 'fa-user-doctor' },
  'docs-del-2500': { name: 'Noise Eliminator', description: 'Removed 2500 lines of outdated comments', icon: 'fa-volume-xmark' },

  // ===== ISSUES OPENED (Tiers: 1, 5, 10, 25, 50) =====
  'issue-1': { name: 'Bug Hunter', description: 'Opened your first issue', icon: 'fa-bug' },
  'issue-5': { name: 'Issue Reporter', description: 'Opened 5 issues', icon: 'fa-flag' },
  'issue-10': { name: 'Quality Advocate', description: 'Opened 10 issues', icon: 'fa-clipboard-list' },
  'issue-25': { name: 'Issue Expert', description: 'Opened 25 issues', icon: 'fa-list-check' },
  'issue-50': { name: 'Issue Champion', description: 'Opened 50 issues', icon: 'fa-bullhorn' },

  // ===== ISSUES CLOSED (Tiers: 1, 5, 10, 25, 50) =====
  'issue-close-1': { name: 'Problem Solver', description: 'Closed your first issue', icon: 'fa-circle-check' },
  'issue-close-5': { name: 'Bug Squasher', description: 'Closed 5 issues', icon: 'fa-bug-slash' },
  'issue-close-10': { name: 'Issue Resolver', description: 'Closed 10 issues', icon: 'fa-check-double' },
  'issue-close-25': { name: 'Closure Expert', description: 'Closed 25 issues', icon: 'fa-square-check' },
  'issue-close-50': { name: 'Issue Terminator', description: 'Closed 50 issues', icon: 'fa-crosshairs' },

  // ===== ISSUE COMMENTS (Tiers: 5, 10, 25, 50, 100) =====
  'issue-comment-5': { name: 'Issue Commenter', description: 'Left 5 issue comments', icon: 'fa-comment' },
  'issue-comment-10': { name: 'Discussion Starter', description: 'Left 10 issue comments', icon: 'fa-comments' },
  'issue-comment-25': { name: 'Issue Collaborator', description: 'Left 25 issue comments', icon: 'fa-people-arrows' },
  'issue-comment-50': { name: 'Community Voice', description: 'Left 50 issue comments', icon: 'fa-bullhorn' },
  'issue-comment-100': { name: 'Issue Guru', description: 'Left 100 issue comments', icon: 'fa-graduation-cap' },

  // ===== ISSUE REFERENCES IN COMMITS (Tiers: 5, 10, 25, 50, 100) =====
  'issue-ref-5': { name: 'Issue Linker', description: 'Referenced issues in 5 commits', icon: 'fa-link' },
  'issue-ref-10': { name: 'Commit Connector', description: 'Referenced issues in 10 commits', icon: 'fa-diagram-project' },
  'issue-ref-25': { name: 'Traceability Pro', description: 'Referenced issues in 25 commits', icon: 'fa-sitemap' },
  'issue-ref-50': { name: 'Issue Tracker', description: 'Referenced issues in 50 commits', icon: 'fa-chart-gantt' },
  'issue-ref-100': { name: 'Traceability Master', description: 'Referenced issues in 100 commits', icon: 'fa-network-wired' },
}

const getAchievement = (id) => {
  const base = achievements[id] || { name: id, description: '', icon: 'fa-medal' }
  const threshold = extractThreshold(id)
  const tier = getTierFromThreshold(threshold)
  const gradient = tierGradients[tier] || 'from-gray-400 to-gray-500'
  return { ...base, gradient, tier, threshold }
}

const sizeClasses = {
  sm: { wrapper: 'w-9 h-9', icon: 'text-sm', radius: 'rounded-lg' },
  md: { wrapper: 'w-11 h-11', icon: 'text-base', radius: 'rounded-xl' },
  lg: { wrapper: 'w-14 h-14', icon: 'text-lg', radius: 'rounded-xl' }
}
</script>

<template>
  <div class="inline-flex flex-col items-center gap-2">
    <!-- Badge -->
    <div
      class="relative group/badge"
      :title="getAchievement(achievementId).name"
    >
      <!-- Badge square with rounded corners -->
      <div
        class="flex items-center justify-center bg-gradient-to-br shadow-lg hover:scale-105 hover:shadow-xl transition-all duration-200 cursor-pointer"
        :class="[
          sizeClasses[size].wrapper,
          sizeClasses[size].radius,
          getAchievement(achievementId).gradient
        ]"
      >
        <i
          class="fas text-white drop-shadow-sm"
          :class="[getAchievement(achievementId).icon, sizeClasses[size].icon]"
        ></i>
      </div>

      <!-- Tooltip -->
      <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-3 px-3 py-2 bg-gray-900 dark:bg-gray-800 text-white text-xs rounded-xl opacity-0 group-hover/badge:opacity-100 transition-all duration-200 pointer-events-none whitespace-nowrap z-50 shadow-xl border border-white/10">
        <div class="font-bold text-sm">{{ getAchievement(achievementId).name }}</div>
        <div class="text-gray-300 text-[11px] mt-0.5">{{ getAchievement(achievementId).description }}</div>
        <div class="absolute top-full left-1/2 -translate-x-1/2 border-[6px] border-transparent border-t-gray-900 dark:border-t-gray-800"></div>
      </div>
    </div>

    <!-- Label (optional) - no truncation -->
    <span
      v-if="showLabel"
      class="text-[11px] font-medium text-gray-600 dark:text-gray-400 text-center leading-tight"
      style="max-width: 72px; word-wrap: break-word;"
    >
      {{ getAchievement(achievementId).name }}
    </span>
  </div>
</template>
