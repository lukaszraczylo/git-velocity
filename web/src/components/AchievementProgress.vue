<script setup>
import { computed } from 'vue'
import { formatNumber } from '../composables/formatters'

const props = defineProps({
  contributor: { type: Object, required: true },
  showEarned: { type: Boolean, default: false },
  maxDisplay: { type: Number, default: 6 }
})

// Achievement tier thresholds
const tiers = [1, 10, 25, 50, 100, 250, 500, 1000, 5000, 10000, 25000]

// Tier gradient colors
const tierGradients = {
  1: 'from-stone-400 to-stone-500',
  2: 'from-green-400 to-emerald-500',
  3: 'from-blue-400 to-indigo-500',
  4: 'from-purple-400 to-violet-500',
  5: 'from-yellow-400 to-amber-500',
  6: 'from-orange-400 to-red-500',
  7: 'from-red-500 to-rose-600',
  8: 'from-pink-500 to-fuchsia-600',
  9: 'from-cyan-400 to-teal-500',
  10: 'from-emerald-400 to-cyan-500',
  11: 'from-violet-500 to-purple-600',
}

// Progress bar colors based on tier
const tierProgressColors = {
  1: 'bg-stone-500',
  2: 'bg-green-500',
  3: 'bg-blue-500',
  4: 'bg-purple-500',
  5: 'bg-yellow-500',
  6: 'bg-orange-500',
  7: 'bg-red-500',
  8: 'bg-pink-500',
  9: 'bg-cyan-500',
  10: 'bg-emerald-500',
  11: 'bg-violet-500',
}

// Achievement definitions with progress tracking
const achievementTypes = [
  {
    category: 'Commits',
    icon: 'fa-code-commit',
    iconColor: 'text-green-500',
    getValue: (c) => c.commit_count || 0,
    achievements: [
      { id: 'first-commit', threshold: 1, name: 'First Steps' },
      { id: 'commit-10', threshold: 10, name: 'Getting Started' },
      { id: 'commit-25', threshold: 25, name: 'Warming Up' },
      { id: 'commit-50', threshold: 50, name: 'On A Roll' },
      { id: 'commit-100', threshold: 100, name: 'Committed' },
      { id: 'commit-250', threshold: 250, name: 'Dedicated' },
      { id: 'commit-500', threshold: 500, name: 'Code Machine' },
      { id: 'commit-1000', threshold: 1000, name: 'Code Warrior' },
      { id: 'commit-5000', threshold: 5000, name: 'Legendary' },
      { id: 'commit-10000', threshold: 10000, name: 'Mythical' },
      { id: 'commit-25000', threshold: 25000, name: 'Transcendent' },
    ]
  },
  {
    category: 'Pull Requests',
    icon: 'fa-code-pull-request',
    iconColor: 'text-blue-500',
    getValue: (c) => c.prs_opened || 0,
    achievements: [
      { id: 'pr-opener', threshold: 1, name: 'PR Pioneer' },
      { id: 'pr-10', threshold: 10, name: 'Pull Request Pro' },
      { id: 'pr-25', threshold: 25, name: 'PR Regular' },
      { id: 'pr-50', threshold: 50, name: 'Merge Master' },
      { id: 'pr-100', threshold: 100, name: 'PR Champion' },
    ]
  },
  {
    category: 'Reviews',
    icon: 'fa-eye',
    iconColor: 'text-purple-500',
    getValue: (c) => c.reviews_given || 0,
    achievements: [
      { id: 'reviewer', threshold: 1, name: 'Code Reviewer' },
      { id: 'reviewer-10', threshold: 10, name: 'Review Starter' },
      { id: 'reviewer-25', threshold: 25, name: 'Review Regular' },
      { id: 'reviewer-50', threshold: 50, name: 'Review Expert' },
      { id: 'reviewer-100', threshold: 100, name: 'Review Guru' },
    ]
  },
  {
    category: 'Lines Added',
    icon: 'fa-plus',
    iconColor: 'text-emerald-500',
    getValue: (c) => c.lines_added || 0,
    achievements: [
      { id: 'lines-1000', threshold: 1000, name: 'Thousand Lines' },
      { id: 'lines-10000', threshold: 10000, name: 'Ten Thousand' },
    ]
  },
  {
    category: 'Lines Deleted',
    icon: 'fa-minus',
    iconColor: 'text-red-500',
    getValue: (c) => c.lines_deleted || 0,
    achievements: [
      { id: 'cleaner', threshold: 1000, name: 'Code Cleaner' },
      { id: 'refactorer', threshold: 10000, name: 'Refactoring Champion' },
    ]
  },
  {
    category: 'Small PRs',
    icon: 'fa-compress',
    iconColor: 'text-cyan-500',
    getValue: (c) => c.small_pr_count || 0,
    achievements: [
      { id: 'small-pr-10', threshold: 10, name: 'Small PR Advocate' },
      { id: 'small-pr-50', threshold: 50, name: 'Atomic Commits Hero' },
    ]
  },
  {
    category: 'Perfect PRs',
    icon: 'fa-gem',
    iconColor: 'text-pink-500',
    getValue: (c) => c.perfect_prs || 0,
    achievements: [
      { id: 'perfect-pr-5', threshold: 5, name: 'Clean Code' },
      { id: 'perfect-pr-25', threshold: 25, name: 'Flawless' },
    ]
  },
  {
    category: 'Active Days',
    icon: 'fa-calendar-check',
    iconColor: 'text-orange-500',
    getValue: (c) => c.active_days || 0,
    achievements: [
      { id: 'active-30', threshold: 30, name: 'Consistent Contributor' },
      { id: 'active-100', threshold: 100, name: 'Dedicated Developer' },
    ]
  },
  {
    category: 'Streak',
    icon: 'fa-fire',
    iconColor: 'text-amber-500',
    getValue: (c) => c.longest_streak || 0,
    achievements: [
      { id: 'streak-7', threshold: 7, name: 'Week Warrior' },
      { id: 'streak-30', threshold: 30, name: 'Month Master' },
    ]
  },
]

// Get tier number from threshold
const getTier = (threshold) => {
  for (let i = tiers.length - 1; i >= 0; i--) {
    if (threshold >= tiers[i]) return i + 1
  }
  return 1
}

// Find all tiers for a category to show progression
const getTiersForCategory = (achievements) => {
  return achievements.map(a => ({
    threshold: a.threshold,
    name: a.name,
    tier: getTier(a.threshold)
  }))
}

// Calculate progress for each achievement type
const progressItems = computed(() => {
  const earnedSet = new Set(props.contributor.achievements || [])
  const results = []

  for (const type of achievementTypes) {
    const currentValue = type.getValue(props.contributor)

    // Find the FIRST achievement where currentValue < threshold (true next target)
    // Also track all earned achievements
    let targetAchievement = null
    let lastEarned = null
    const allTiers = getTiersForCategory(type.achievements)

    for (const ach of type.achievements) {
      if (currentValue >= ach.threshold) {
        // User has reached this threshold (should be earned)
        lastEarned = ach
      } else if (!targetAchievement) {
        // First achievement they haven't reached yet
        targetAchievement = ach
      }
    }

    // Skip if no target (all thresholds exceeded)
    if (!targetAchievement) continue

    // Calculate progress from last threshold to next
    const previousThreshold = lastEarned ? lastEarned.threshold : 0
    const progressRange = targetAchievement.threshold - previousThreshold
    const currentProgress = currentValue - previousThreshold
    const progress = Math.min(100, Math.max(0, Math.round((currentProgress / progressRange) * 100)))
    const tier = getTier(targetAchievement.threshold)

    // Find current tier position and total tiers
    const currentTierIndex = allTiers.findIndex(t => t.threshold === targetAchievement.threshold)
    const totalTiers = allTiers.length

    results.push({
      category: type.category,
      icon: type.icon,
      iconColor: type.iconColor,
      currentValue,
      target: targetAchievement.threshold,
      name: targetAchievement.name,
      id: targetAchievement.id,
      progress,
      tier,
      tierIndex: currentTierIndex + 1,
      totalTiers,
      allTiers,
      gradient: tierGradients[tier],
      progressColor: tierProgressColors[tier],
      isClose: progress >= 75,
      remaining: targetAchievement.threshold - currentValue,
      isEarned: earnedSet.has(targetAchievement.id),
    })
  }

  // Sort by progress descending (closest to next tier first - highest % complete)
  results.sort((a, b) => b.progress - a.progress)

  return results
})

// Get count of remaining achievements (all unearned across all types)
const remainingCount = computed(() => {
  const earnedSet = new Set(props.contributor.achievements || [])
  let totalUnearned = 0

  for (const type of achievementTypes) {
    const currentValue = type.getValue(props.contributor)
    for (const ach of type.achievements) {
      // Count achievements where user hasn't reached the threshold
      if (currentValue < ach.threshold) {
        totalUnearned++
      }
    }
  }
  return Math.max(0, totalUnearned - props.maxDisplay)
})
</script>

<template>
  <div class="space-y-3">
    <div
      v-for="item in progressItems"
      :key="item.id"
      class="bg-gray-50 dark:bg-gray-800/50 rounded-xl p-4 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
    >
      <div class="flex items-start justify-between mb-3">
        <div class="flex items-center space-x-3">
          <div
            class="w-10 h-10 rounded-lg bg-gradient-to-br flex items-center justify-center shadow-md"
            :class="item.gradient"
          >
            <i class="fas text-white text-sm" :class="item.icon"></i>
          </div>
          <div>
            <div class="text-sm font-semibold text-gray-800 dark:text-white">
              {{ item.name }}
            </div>
            <div class="flex items-center space-x-2 text-xs text-gray-500 dark:text-gray-400">
              <span>{{ item.category }}</span>
              <span class="text-gray-300 dark:text-gray-600">•</span>
              <span class="font-medium">Tier {{ item.tierIndex }}/{{ item.totalTiers }}</span>
            </div>
          </div>
        </div>
        <div class="text-right">
          <div class="text-sm font-bold" :class="item.isClose ? 'text-green-500' : 'text-gray-700 dark:text-gray-200'">
            {{ formatNumber(item.currentValue) }}
            <span class="text-gray-400 dark:text-gray-500 font-normal">/</span>
            <span class="text-gray-500 dark:text-gray-400 font-medium">{{ formatNumber(item.target) }}</span>
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            {{ item.remaining > 0 ? `${formatNumber(item.remaining)} to go` : 'Ready to claim!' }}
          </div>
        </div>
      </div>

      <!-- Progress Bar -->
      <div class="h-2.5 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
        <div
          class="h-full rounded-full transition-all duration-500 ease-out"
          :class="item.progressColor"
          :style="{ width: `${item.progress}%` }"
        ></div>
      </div>

      <!-- Progress percentage and tier markers -->
      <div class="flex items-center justify-between mt-1.5">
        <div class="flex items-center space-x-1">
          <span
            v-for="(t, idx) in item.allTiers.slice(0, 5)"
            :key="t.threshold"
            class="w-1.5 h-1.5 rounded-full"
            :class="idx < item.tierIndex ? 'bg-green-500' : 'bg-gray-300 dark:bg-gray-600'"
            :title="`Tier ${idx + 1}: ${t.name} (${formatNumber(t.threshold)})`"
          ></span>
          <span v-if="item.totalTiers > 5" class="text-[10px] text-gray-400">+{{ item.totalTiers - 5 }}</span>
        </div>
        <span
          class="text-xs font-semibold"
          :class="item.isClose ? 'text-green-500' : 'text-gray-400 dark:text-gray-500'"
        >
          {{ item.progress }}%
        </span>
      </div>
    </div>

    <!-- Show more indicator -->
    <div v-if="remainingCount > 0" class="text-center text-xs text-gray-500 dark:text-gray-400 pt-2">
      +{{ remainingCount }} more achievements to unlock
    </div>

    <!-- Empty state -->
    <div v-if="!progressItems.length" class="text-center py-8 text-gray-500 dark:text-gray-400">
      <div class="w-16 h-16 mx-auto mb-3 rounded-2xl bg-gradient-to-br from-yellow-400 to-amber-500 flex items-center justify-center shadow-lg">
        <i class="fas fa-trophy text-2xl text-white"></i>
      </div>
      <p class="font-medium text-gray-700 dark:text-gray-300">All achievements unlocked!</p>
      <p class="text-sm mt-1">You're a legend!</p>
    </div>
  </div>
</template>
