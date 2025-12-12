<script setup>
import { ref, inject, computed } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import DataTable from '../components/DataTable.vue'
import ContributorRow from '../components/ContributorRow.vue'
import RankBadge from '../components/RankBadge.vue'
import AchievementBadge from '../components/AchievementBadge.vue'
import { formatNumber } from '../composables/formatters'
import { getHighestTierAchievements } from '../composables/achievements'

const globalData = inject('globalData')
const searchQuery = ref('')

const allContributors = computed(() => globalData.value?.leaderboard || [])

const leaderboard = computed(() => {
  if (!searchQuery.value.trim()) return allContributors.value

  const query = searchQuery.value.toLowerCase().trim()
  return allContributors.value.filter(contributor => {
    const name = (contributor.name || '').toLowerCase()
    const login = (contributor.login || '').toLowerCase()
    return name.includes(query) || login.includes(query)
  })
})

const tableColumns = [
  { key: 'rank', label: 'Rank', align: 'left' },
  { key: 'contributor', label: 'Contributor', align: 'left' },
  { key: 'achievements', label: 'Achievements', align: 'left' },
  { key: 'team', label: 'Team', align: 'left', headerClass: 'hidden md:table-cell' },
  { key: 'category', label: 'Best At', align: 'left', headerClass: 'hidden sm:table-cell' },
  { key: 'score', label: 'Score', align: 'right' }
]

const categoryIcon = (category) => {
  const icons = {
    'Commits': 'fas fa-code-commit text-green-500',
    'PRs': 'fas fa-code-pull-request text-blue-500',
    'Reviews': 'fas fa-eye text-purple-500',
    'Comments': 'fas fa-comment text-orange-500'
  }
  return icons[category] || ''
}
</script>

<template>
  <div>
    <PageHeader
      title="Leaderboard"
      subtitle="Top contributors ranked by their velocity score"
      icon="fas fa-trophy"
      icon-color="text-yellow-500"
      centered
    />

    <!-- Search and Leaderboard Table -->
    <section class="py-8 px-4">
      <div class="container mx-auto max-w-5xl">
        <!-- Search Input -->
        <div class="mb-6">
          <div class="relative max-w-md">
            <i class="fas fa-search absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"></i>
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search by name or username..."
              class="w-full pl-10 pr-4 py-2.5 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition"
            />
            <button
              v-if="searchQuery"
              @click="searchQuery = ''"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
            >
              <i class="fas fa-times"></i>
            </button>
          </div>
          <p v-if="searchQuery && leaderboard.length !== allContributors.length" class="mt-2 text-sm text-gray-500 dark:text-gray-400">
            Showing {{ leaderboard.length }} of {{ allContributors.length }} contributors
          </p>
        </div>

        <DataTable
          :columns="tableColumns"
          :items="leaderboard"
          empty-icon="fas fa-users"
          empty-message="No contributors found"
          row-class="hover:bg-gray-50 dark:hover:bg-gray-800/30 transition group"
        >
          <template #rank="{ item }">
            <RankBadge :rank="item.rank" />
          </template>

          <template #contributor="{ item }">
            <ContributorRow :contributor="item" show-github-link />
          </template>

          <template #achievements="{ item }">
            <div class="flex flex-wrap gap-1.5 max-w-[280px]">
              <AchievementBadge
                v-for="achievement in getHighestTierAchievements(item.achievements)"
                :key="achievement"
                :achievement-id="achievement"
                size="sm"
              />
              <span v-if="!(item.achievements || []).length" class="text-gray-400 text-sm">-</span>
            </div>
          </template>

          <template #team="{ item }">
            <td class="hidden md:table-cell">
              <span
                v-if="item.team"
                class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 dark:bg-purple-900/30 text-purple-800 dark:text-purple-300"
              >
                {{ item.team }}
              </span>
              <span v-else class="text-gray-400">-</span>
            </td>
          </template>

          <template #category="{ item }">
            <td class="hidden sm:table-cell">
              <span v-if="item.top_category" class="text-sm text-gray-600 dark:text-gray-300">
                <i :class="categoryIcon(item.top_category)" class="mr-1"></i>
                {{ item.top_category }}
              </span>
              <span v-else class="text-gray-400">-</span>
            </td>
          </template>

          <template #score="{ item }">
            <span class="text-lg font-bold gradient-text">
              {{ formatNumber(item.score) }}
            </span>
          </template>
        </DataTable>
      </div>
    </section>
  </div>
</template>
