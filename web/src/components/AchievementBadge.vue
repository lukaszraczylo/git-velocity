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

// Extract threshold from achievement ID (e.g., "commit-100" -> 100)
const extractThreshold = (id) => {
  const match = id.match(/(\d+)$/)
  if (match) return parseInt(match[1], 10)
  // Special cases for non-numeric achievements
  if (id === 'first-commit' || id === 'pr-opener' || id === 'reviewer') return 1
  return 50 // Default for special achievements
}

// Achievement definitions matching the Go backend
const achievements = {
  // Commit achievements - Journey from apprentice to legend
  'first-commit': { name: 'Hello World', description: 'Made your first commit', icon: 'fa-baby' },
  'commit-10': { name: 'Seedling', description: 'Made 10 commits', icon: 'fa-seedling' },
  'commit-25': { name: 'Momentum', description: 'Made 25 commits', icon: 'fa-wind' },
  'commit-50': { name: 'Trailblazer', description: 'Made 50 commits', icon: 'fa-hiking' },
  'commit-100': { name: 'Centurion', description: 'Made 100 commits', icon: 'fa-shield-halved' },
  'commit-250': { name: 'Relentless', description: 'Made 250 commits', icon: 'fa-bolt-lightning' },
  'commit-500': { name: 'Unstoppable', description: 'Made 500 commits', icon: 'fa-meteor' },
  'commit-1000': { name: 'Grandmaster', description: 'Made 1000 commits', icon: 'fa-chess-king' },
  'commit-5000': { name: 'Titan', description: 'Made 5000 commits', icon: 'fa-mountain-sun' },
  'commit-10000': { name: 'Immortal', description: 'Made 10000 commits', icon: 'fa-dragon' },
  'commit-25000': { name: 'Ascended', description: 'Made 25000 commits', icon: 'fa-infinity' },

  // PR achievements - The art of collaboration
  'pr-opener': { name: 'First Blood', description: 'Opened your first pull request', icon: 'fa-flag-checkered' },
  'pr-10': { name: 'Collaborator', description: 'Opened 10 pull requests', icon: 'fa-handshake' },
  'pr-25': { name: 'Integrator', description: 'Opened 25 pull requests', icon: 'fa-code-branch' },
  'pr-50': { name: 'Architect', description: 'Opened 50 pull requests', icon: 'fa-building' },
  'pr-100': { name: 'Vanguard', description: 'Opened 100 pull requests', icon: 'fa-rocket' },

  // Review achievements - The guardian path
  'reviewer': { name: 'Watchful Eye', description: 'Reviewed your first pull request', icon: 'fa-eye' },
  'reviewer-10': { name: 'Sentinel', description: 'Reviewed 10 pull requests', icon: 'fa-shield' },
  'reviewer-25': { name: 'Gatekeeper', description: 'Reviewed 25 pull requests', icon: 'fa-dungeon' },
  'reviewer-50': { name: 'Oracle', description: 'Reviewed 50 pull requests', icon: 'fa-hat-wizard' },
  'reviewer-100': { name: 'Sage', description: 'Reviewed 100 pull requests', icon: 'fa-book-skull' },

  // Speed achievements - Time is of the essence
  'speed-demon': { name: 'Lightning Rod', description: 'Average review response under 1 hour', icon: 'fa-bolt' },
  'quick-responder': { name: 'Flash', description: 'Average review response under 4 hours', icon: 'fa-gauge-high' },

  // Comment achievements
  'commentator': { name: 'Wordsmith', description: 'Left 50 PR review comments', icon: 'fa-feather-pointed' },

  // Lines of code achievements - Volume mastery
  'lines-1000': { name: 'Scribe', description: 'Added 1000 lines of code', icon: 'fa-scroll' },
  'lines-10000': { name: 'Novelist', description: 'Added 10000 lines of code', icon: 'fa-book' },
  'lines-100000': { name: 'Encyclopedia', description: 'Added 100000 lines of code', icon: 'fa-landmark' },

  // Deletion achievements - The minimalist way
  'cleaner': { name: 'Pruner', description: 'Deleted 1000 lines of code', icon: 'fa-scissors' },
  'refactorer': { name: 'Surgeon', description: 'Deleted 10000 lines of code', icon: 'fa-syringe' },
  'annihilator': { name: 'Annihilator', description: 'Deleted 100000 lines of code', icon: 'fa-explosion' },

  // Multi-repo achievements - The wanderer
  'multi-repo': { name: 'Nomad', description: 'Contributed to 5 repositories', icon: 'fa-compass' },
  'multi-repo-10': { name: 'Explorer', description: 'Contributed to 10 repositories', icon: 'fa-map' },

  // Team collaboration - Social butterfly
  'team-player': { name: 'Ambassador', description: 'Reviewed PRs from 10 different contributors', icon: 'fa-users' },
  'team-player-25': { name: 'Diplomat', description: 'Reviewed PRs from 25 different contributors', icon: 'fa-globe' },

  // PR size achievements - Go big or go home
  'big-pr': { name: 'Heavyweight', description: 'Merged a PR with 1000+ lines', icon: 'fa-dumbbell' },
  'mega-pr': { name: 'Colossus', description: 'Merged a PR with 5000+ lines', icon: 'fa-monument' },

  // Small PR achievements - Precision strikes
  'small-pr-10': { name: 'Minimalist', description: 'Merged 10 PRs under 100 lines', icon: 'fa-compress' },
  'small-pr-50': { name: 'Atomic', description: 'Merged 50 PRs under 100 lines', icon: 'fa-atom' },

  // Perfect PR achievements - Flawless execution
  'perfect-pr-5': { name: 'Sharpshooter', description: '5 PRs merged without changes requested', icon: 'fa-bullseye' },
  'perfect-pr-25': { name: 'Perfectionist', description: '25 PRs merged without changes requested', icon: 'fa-gem' },
  'perfect-pr-100': { name: 'Immaculate', description: '100 PRs merged without changes requested', icon: 'fa-crown' },

  // Streak achievements - Consistency is key
  'streak-7': { name: 'Hot Streak', description: '7 day contribution streak', icon: 'fa-fire' },
  'streak-30': { name: 'Ironclad', description: '30 day contribution streak', icon: 'fa-link' },
  'streak-90': { name: 'Unbreakable', description: '90 day contribution streak', icon: 'fa-diamond' },

  // Time-based achievements - When you code matters
  'early-bird': { name: 'Dawn Patrol', description: '50 commits before 9am', icon: 'fa-sun' },
  'night-owl': { name: 'Nighthawk', description: '50 commits after 9pm', icon: 'fa-moon' },
  'nosferatu': { name: 'Vampire', description: '25 commits between midnight and 4am', icon: 'fa-ghost' },
  'weekend-warrior': { name: 'No Days Off', description: '25 weekend commits', icon: 'fa-calendar-xmark' },

  // Activity achievements - Showing up matters
  'active-30': { name: 'Reliable', description: 'Active on 30 different days', icon: 'fa-calendar-check' },
  'active-100': { name: 'Stalwart', description: 'Active on 100 different days', icon: 'fa-tower-observation' },
  'active-365': { name: 'Eternal', description: 'Active on 365 different days', icon: 'fa-sun-plant-wilt' }
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
