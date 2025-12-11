<script setup>
import { ref, computed, onMounted, watch, inject } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import LoadingState from '../components/LoadingState.vue'
import ErrorState from '../components/ErrorState.vue'
import StatCard from '../components/StatCard.vue'
import Avatar from '../components/Avatar.vue'
import AchievementBadge from '../components/AchievementBadge.vue'
import AchievementProgress from '../components/AchievementProgress.vue'
import SectionHeader from '../components/SectionHeader.vue'
import GithubLink from '../components/GithubLink.vue'
import { formatNumber, formatPercent, formatDuration } from '../composables/formatters'
import { getHighestTierAchievements } from '../composables/achievements'

const route = useRoute()
const globalData = inject('globalData')
const contributor = ref(null)
const loading = ref(true)
const error = ref(null)

const breadcrumbs = computed(() => [
  { label: 'Dashboard', to: '/' },
  { label: 'Contributors' },
  { label: contributor.value?.login || route.params.login }
])

async function loadContributor() {
  loading.value = true
  error.value = null

  const login = route.params.login

  try {
    const response = await fetch(`./data/contributors/${login}.json`)
    if (response.ok) {
      const data = await response.json()

      const leaderboard = globalData.value?.leaderboard || []
      const leaderboardEntry = leaderboard.find(e => e.login === login)

      if (leaderboardEntry) {
        data.score = {
          total: leaderboardEntry.score,
          rank: leaderboardEntry.rank,
          breakdown: data.score?.breakdown
        }
        data.achievements = leaderboardEntry.achievements
      }

      contributor.value = data
    } else {
      const leaderboard = globalData.value?.leaderboard || []
      let found = leaderboard.find(e => e.login === login)

      if (!found) {
        const repos = globalData.value?.repositories || []
        for (const repo of repos) {
          const c = repo.contributors?.find(c => c.login === login)
          if (c) {
            found = c
            break
          }
        }
      }

      if (found) {
        contributor.value = found
      } else {
        error.value = 'Contributor not found'
      }
    }
  } catch (e) {
    error.value = `Failed to load contributor: ${e.message}`
  }

  loading.value = false
}

onMounted(loadContributor)
watch(() => route.params, loadContributor)
watch(globalData, loadContributor)
</script>

<template>
  <div>
    <LoadingState v-if="loading" message="Loading contributor..." />
    <ErrorState v-else-if="error" :message="error" />

    <template v-else-if="contributor">
      <!-- Profile Header -->
      <header class="py-12 px-4">
        <div class="container mx-auto">
          <PageHeader :breadcrumbs="breadcrumbs" :title="''" />

          <div class="flex flex-col md:flex-row items-center md:items-start space-y-4 md:space-y-0 md:space-x-8">
            <Avatar
              :src="contributor.avatar_url"
              :name="contributor.login"
              size="2xl"
              class="shadow-modern"
            />

            <div class="text-center md:text-left">
              <h1 class="text-4xl font-bold gradient-text">
                {{ contributor.name || contributor.login }}
              </h1>
              <p class="text-xl text-gray-500 dark:text-gray-400 mt-1">
                <GithubLink :url="`https://github.com/${contributor.login}`">
                  @{{ contributor.login }}
                </GithubLink>
              </p>

              <div class="flex items-center justify-center md:justify-start space-x-4 mt-4">
                <div class="score-card rounded-lg px-4 py-2">
                  <span class="text-sm text-gray-500 dark:text-gray-400">Score:</span>
                  <span class="text-2xl font-bold gradient-text ml-2">
                    {{ formatNumber(contributor.score?.total || contributor.score || 0) }}
                  </span>
                </div>
                <div v-if="contributor.score?.rank" class="text-sm text-gray-500 dark:text-gray-400">
                  Rank #{{ contributor.score.rank }}
                  <span v-if="contributor.score?.percentile_rank">
                    (Top {{ formatPercent(contributor.score.percentile_rank) }})
                  </span>
                </div>
              </div>

              <div v-if="contributor.achievements?.length" class="mt-6 flex flex-wrap justify-center md:justify-start gap-3">
                <AchievementBadge
                  v-for="achievement in getHighestTierAchievements(contributor.achievements)"
                  :key="achievement"
                  :achievement-id="achievement"
                  size="lg"
                  show-label
                />
              </div>
            </div>
          </div>
        </div>
      </header>

      <!-- Stats Grid -->
      <section class="py-8 px-4">
        <div class="container mx-auto">
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard
              :value="contributor.commit_count || 0"
              label="Commits"
              icon="fas fa-code-commit"
              icon-color="text-green-500"
            />
            <StatCard
              :value="contributor.prs_opened || 0"
              label="PRs Opened"
              icon="fas fa-code-pull-request"
              icon-color="text-blue-500"
            />
            <StatCard
              :value="contributor.prs_merged || 0"
              label="PRs Merged"
              icon="fas fa-code-merge"
              icon-color="text-purple-500"
            />
            <StatCard
              :value="contributor.reviews_given || 0"
              label="Reviews Given"
              icon="fas fa-eye"
              icon-color="text-orange-500"
            />
          </div>
        </div>
      </section>

      <!-- Detailed Stats -->
      <section class="py-8 px-4">
        <div class="container mx-auto">
          <div class="grid md:grid-cols-2 gap-6">
            <!-- Code Stats -->
            <div class="card">
              <h3 class="text-lg font-semibold text-gray-800 dark:text-white mb-4">
                <i class="fas fa-code text-green-500 mr-2"></i>Code Contributions
              </h3>

              <div class="space-y-4">
                <div class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Lines Added</span>
                  <span class="text-green-500 font-semibold">
                    +{{ formatNumber(contributor.lines_added || 0) }}
                  </span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Lines Deleted</span>
                  <span class="text-red-500 font-semibold">
                    -{{ formatNumber(contributor.lines_deleted || 0) }}
                  </span>
                </div>
                <div v-if="contributor.meaningful_lines_added !== undefined" class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Meaningful Lines Added</span>
                  <span class="text-emerald-500 font-semibold">
                    +{{ formatNumber(contributor.meaningful_lines_added || 0) }}
                  </span>
                </div>
                <div v-if="contributor.meaningful_lines_deleted !== undefined" class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Meaningful Lines Deleted</span>
                  <span class="text-rose-500 font-semibold">
                    -{{ formatNumber(contributor.meaningful_lines_deleted || 0) }}
                  </span>
                </div>
                <div v-if="contributor.comment_lines_added !== undefined" class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Comment Lines Added</span>
                  <span class="text-cyan-500 font-semibold">
                    +{{ formatNumber(contributor.comment_lines_added || 0) }}
                  </span>
                </div>
                <div v-if="contributor.comment_lines_deleted !== undefined" class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Comment Lines Deleted</span>
                  <span class="text-amber-500 font-semibold">
                    -{{ formatNumber(contributor.comment_lines_deleted || 0) }}
                  </span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Files Changed</span>
                  <span class="text-gray-800 dark:text-white font-semibold">
                    {{ formatNumber(contributor.files_changed || 0) }}
                  </span>
                </div>
                <div v-if="contributor.avg_pr_size" class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Avg PR Size</span>
                  <span class="text-gray-800 dark:text-white font-semibold">
                    {{ formatNumber(Math.round(contributor.avg_pr_size)) }} lines
                  </span>
                </div>
              </div>
            </div>

            <!-- Review Stats -->
            <div class="card">
              <h3 class="text-lg font-semibold text-gray-800 dark:text-white mb-4">
                <i class="fas fa-comments text-purple-500 mr-2"></i>Review Activity
              </h3>

              <div class="space-y-4">
                <div class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Reviews Given</span>
                  <span class="text-gray-800 dark:text-white font-semibold">
                    {{ formatNumber(contributor.reviews_given || 0) }}
                  </span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Approvals</span>
                  <span class="text-green-500 font-semibold">
                    {{ formatNumber(contributor.approvals_given || 0) }}
                  </span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Changes Requested</span>
                  <span class="text-orange-500 font-semibold">
                    {{ formatNumber(contributor.changes_requested || 0) }}
                  </span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Review Comments</span>
                  <span class="text-gray-800 dark:text-white font-semibold">
                    {{ formatNumber(contributor.review_comments || 0) }}
                  </span>
                </div>
                <div v-if="contributor.avg_review_time_hours" class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Avg Review Time</span>
                  <span class="text-gray-800 dark:text-white font-semibold">
                    {{ formatDuration(contributor.avg_review_time_hours) }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Issue Stats -->
            <div v-if="contributor.issues_opened || contributor.issues_closed || contributor.issue_comments || contributor.issue_references_in_commits" class="card">
              <h3 class="text-lg font-semibold text-gray-800 dark:text-white mb-4">
                <i class="fas fa-bug text-red-500 mr-2"></i>Issue Activity
              </h3>

              <div class="space-y-4">
                <div class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Issues Opened</span>
                  <span class="text-red-500 font-semibold">
                    {{ formatNumber(contributor.issues_opened || 0) }}
                  </span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Issues Closed</span>
                  <span class="text-green-500 font-semibold">
                    {{ formatNumber(contributor.issues_closed || 0) }}
                  </span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Issue Comments</span>
                  <span class="text-blue-500 font-semibold">
                    {{ formatNumber(contributor.issue_comments || 0) }}
                  </span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-300">Issue References in Commits</span>
                  <span class="text-purple-500 font-semibold">
                    {{ formatNumber(contributor.issue_references_in_commits || 0) }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Score Breakdown -->
      <section v-if="contributor.score?.breakdown" class="py-8 px-4">
        <div class="container mx-auto">
          <div class="card">
            <h3 class="text-lg font-semibold text-gray-800 dark:text-white mb-4">
              <i class="fas fa-chart-pie gradient-text mr-2"></i>Score Breakdown
            </h3>

            <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-8 gap-4">
              <div class="text-center p-4 rounded-lg bg-gray-50 dark:bg-gray-800/50">
                <div class="text-2xl font-bold text-green-500">
                  {{ formatNumber(contributor.score.breakdown.commits || 0) }}
                </div>
                <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">Commits</div>
                <div class="text-xs text-gray-400 dark:text-gray-500">{{ contributor.commit_count || 0 }} × 10 pts</div>
              </div>
              <div class="text-center p-4 rounded-lg bg-gray-50 dark:bg-gray-800/50">
                <div class="text-2xl font-bold text-blue-500">
                  {{ formatNumber(contributor.score.breakdown.prs || 0) }}
                </div>
                <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">PRs</div>
                <div class="text-xs text-gray-400 dark:text-gray-500">{{ contributor.prs_opened || 0 }} opened + {{ contributor.prs_merged || 0 }} merged</div>
              </div>
              <div class="text-center p-4 rounded-lg bg-gray-50 dark:bg-gray-800/50">
                <div class="text-2xl font-bold text-purple-500">
                  {{ formatNumber(contributor.score.breakdown.reviews || 0) }}
                </div>
                <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">Reviews</div>
                <div class="text-xs text-gray-400 dark:text-gray-500">{{ contributor.reviews_given || 0 }} × 30 pts</div>
              </div>
              <div class="text-center p-4 rounded-lg bg-gray-50 dark:bg-gray-800/50">
                <div class="text-2xl font-bold text-pink-500">
                  {{ formatNumber(contributor.score.breakdown.comments || 0) }}
                </div>
                <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">Comments</div>
                <div class="text-xs text-gray-400 dark:text-gray-500">{{ contributor.review_comments || 0 }} × 5 pts</div>
              </div>
              <div class="text-center p-4 rounded-lg bg-gray-50 dark:bg-gray-800/50">
                <div class="text-2xl font-bold text-red-500">
                  {{ formatNumber(contributor.score.breakdown.issues || 0) }}
                </div>
                <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">Issues</div>
                <div class="text-xs text-gray-400 dark:text-gray-500">opened, closed, comments, refs</div>
              </div>
              <div class="text-center p-4 rounded-lg bg-gray-50 dark:bg-gray-800/50">
                <div class="text-2xl font-bold text-orange-500">
                  {{ formatNumber(contributor.score.breakdown.line_changes || 0) }}
                </div>
                <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">Line Changes</div>
                <div class="text-xs text-gray-400 dark:text-gray-500">meaningful lines × 0.1 pts</div>
              </div>
              <div class="text-center p-4 rounded-lg bg-gray-50 dark:bg-gray-800/50">
                <div class="text-2xl font-bold text-yellow-500">
                  {{ formatNumber(contributor.score.breakdown.response_bonus || 0) }}
                </div>
                <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">Response Bonus</div>
                <div class="text-xs text-gray-400 dark:text-gray-500">fast review bonus</div>
              </div>
              <div class="text-center p-4 rounded-lg bg-gray-50 dark:bg-gray-800/50">
                <div class="text-2xl font-bold text-indigo-500">
                  {{ formatNumber(contributor.score.breakdown.out_of_hours || 0) }}
                </div>
                <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">Out of Hours</div>
                <div class="text-xs text-gray-400 dark:text-gray-500">{{ contributor.out_of_hours_count || 0 }} × 2 pts</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Achievement Progress Section -->
      <section class="py-8 px-4">
        <div class="container mx-auto">
          <div class="grid md:grid-cols-2 gap-6">
            <!-- Earned Achievements -->
            <div v-if="contributor.achievements?.length" class="card">
              <div class="flex items-center justify-between mb-6">
                <h3 class="text-lg font-semibold text-gray-800 dark:text-white">
                  <i class="fas fa-award gradient-text mr-2"></i>Achievements Earned
                </h3>
                <span class="px-2.5 py-1 rounded-full bg-gradient-to-r from-yellow-400 to-amber-500 text-white text-sm font-bold shadow-md">
                  {{ contributor.achievements.length }}
                </span>
              </div>

              <div class="grid grid-cols-4 sm:grid-cols-5 gap-3">
                <div
                  v-for="achievement in contributor.achievements"
                  :key="achievement"
                  class="flex flex-col items-center p-2 rounded-xl bg-gray-50 dark:bg-gray-800/50 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                >
                  <AchievementBadge
                    :achievement-id="achievement"
                    size="md"
                    show-label
                  />
                </div>
              </div>
            </div>

            <!-- Progress to Next Achievements -->
            <div class="card">
              <h3 class="text-lg font-semibold text-gray-800 dark:text-white mb-6">
                <i class="fas fa-chart-line text-primary-500 mr-2"></i>Next Achievements
              </h3>

              <AchievementProgress
                :contributor="contributor"
                :max-display="6"
              />
            </div>
          </div>
        </div>
      </section>

      <!-- Repositories Contributed -->
      <section v-if="contributor.repositories_contributed?.length" class="py-8 px-4">
        <div class="container mx-auto">
          <SectionHeader
            :title="`Contributed to ${contributor.repositories_contributed.length} Repositories`"
            icon="fas fa-folder-tree"
            icon-color="text-blue-500"
          />

          <div class="flex flex-wrap gap-2">
            <RouterLink
              v-for="repo in contributor.repositories_contributed"
              :key="repo"
              :to="`/repos/${repo}`"
              class="inline-flex items-center px-3 py-1.5 rounded-full text-sm bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-primary-100 dark:hover:bg-primary-900/30 hover:text-primary-700 dark:hover:text-primary-300 transition-colors"
            >
              <i class="fas fa-code-branch text-gray-400 mr-2"></i>
              {{ repo }}
            </RouterLink>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>
