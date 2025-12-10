<script setup>
import { inject, computed } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import DataTable from '../components/DataTable.vue'
import ContributorRow from '../components/ContributorRow.vue'
import RankBadge from '../components/RankBadge.vue'
import AchievementBadge from '../components/AchievementBadge.vue'
import { formatNumber } from '../composables/formatters'

const globalData = inject('globalData')
const leaderboard = computed(() => globalData.value?.leaderboard || [])

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

    <!-- Leaderboard Table -->
    <section class="py-8 px-4">
      <div class="container mx-auto max-w-5xl">
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
            <div class="flex flex-wrap gap-1.5 max-w-[180px]">
              <AchievementBadge
                v-for="achievement in (item.achievements || []).slice(0, 6)"
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
