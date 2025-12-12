<script setup>
import { ref, inject, computed } from 'vue'
import { RouterLink } from 'vue-router'
import Card from '../components/Card.vue'
import PageHeader from '../components/PageHeader.vue'
import DataTable from '../components/DataTable.vue'
import ContributorRow from '../components/ContributorRow.vue'
import RankBadge from '../components/RankBadge.vue'
import Avatar from '../components/Avatar.vue'
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
  { key: 'achievements', label: 'Achievements', align: 'left', headerClass: 'hidden md:table-cell' },
  { key: 'team', label: 'Team', align: 'left', headerClass: 'hidden xl:table-cell' },
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

    <!-- Search and Leaderboard -->
    <section class="py-4 sm:py-8 px-4">
      <div class="container mx-auto max-w-5xl">
        <!-- Search Input -->
        <div class="mb-4 sm:mb-6">
          <div class="relative">
            <i class="fas fa-search absolute left-3 top-1/2 -translate-y-1/2 text-gray-500"></i>
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search contributors..."
              class="w-full pl-10 pr-10 py-2.5 rounded-lg border border-gray-700 bg-gray-800 text-gray-100 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition text-sm sm:text-base"
            />
            <button
              v-if="searchQuery"
              @click="searchQuery = ''"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-200"
            >
              <i class="fas fa-times"></i>
            </button>
          </div>
          <p v-if="searchQuery && leaderboard.length !== allContributors.length" class="mt-2 text-sm text-gray-400">
            Showing {{ leaderboard.length }} of {{ allContributors.length }} contributors
          </p>
        </div>

        <!-- Mobile Card Layout -->
        <div class="md:hidden space-y-3">
          <RouterLink
            v-for="item in leaderboard"
            :key="item.login"
            :to="{ name: 'contributor', params: { login: item.login } }"
            class="block"
          >
            <Card hover class="!p-4">
              <div class="flex items-center gap-3">
                <!-- Rank -->
                <RankBadge :rank="item.rank" size="sm" />

                <!-- Avatar -->
                <Avatar :src="item.avatar_url" :name="item.login" size="md" />

                <!-- Info -->
                <div class="flex-1 min-w-0">
                  <div class="font-semibold text-white truncate">
                    {{ item.name || item.login }}
                  </div>
                  <div class="text-xs text-gray-400 truncate">
                    @{{ item.login }}
                  </div>
                </div>

                <!-- Score -->
                <div class="text-right">
                  <div class="text-lg font-bold bg-gradient-to-r from-primary-400 to-accent-400 bg-clip-text text-transparent">
                    {{ formatNumber(item.score) }}
                  </div>
                  <div class="text-xs text-gray-400">pts</div>
                </div>
              </div>

              <!-- Achievements row -->
              <div v-if="item.achievements?.length" class="mt-3 pt-3 border-t border-gray-700">
                <div class="flex flex-wrap gap-1.5">
                  <AchievementBadge
                    v-for="achievement in getHighestTierAchievements(item.achievements).slice(0, 6)"
                    :key="achievement"
                    :achievement-id="achievement"
                    size="sm"
                  />
                  <span
                    v-if="getHighestTierAchievements(item.achievements).length > 6"
                    class="inline-flex items-center justify-center px-2 h-7 rounded-lg bg-gray-700 text-gray-300 text-xs font-bold"
                  >
                    +{{ getHighestTierAchievements(item.achievements).length - 6 }}
                  </span>
                </div>
              </div>
            </Card>
          </RouterLink>

          <!-- Empty State -->
          <div v-if="!leaderboard.length" class="text-center py-12">
            <i class="fas fa-users text-4xl text-gray-500 mb-4"></i>
            <p class="text-gray-400">No contributors found</p>
          </div>
        </div>

        <!-- Desktop Table Layout -->
        <div class="hidden md:block">
          <DataTable
            :columns="tableColumns"
            :items="leaderboard"
            empty-icon="fas fa-users"
            empty-message="No contributors found"
            row-class="hover:bg-gray-800/30 transition group"
          >
            <template #rank="{ item }">
              <RankBadge :rank="item.rank" />
            </template>

            <template #contributor="{ item }">
              <ContributorRow :contributor="item" show-github-link />
            </template>

            <template #achievements="{ item }">
              <td class="hidden md:table-cell">
                <div class="flex flex-wrap gap-1.5 max-w-[280px]">
                  <AchievementBadge
                    v-for="achievement in getHighestTierAchievements(item.achievements)"
                    :key="achievement"
                    :achievement-id="achievement"
                    size="sm"
                  />
                  <span v-if="!(item.achievements || []).length" class="text-gray-400 text-sm">-</span>
                </div>
              </td>
            </template>

            <template #team="{ item }">
              <td class="hidden xl:table-cell">
                <span
                  v-if="item.team"
                  class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-900/30 text-purple-300"
                >
                  {{ item.team }}
                </span>
                <span v-else class="text-gray-400">-</span>
              </td>
            </template>

            <template #score="{ item }">
              <span class="text-lg font-bold bg-gradient-to-r from-primary-400 to-accent-400 bg-clip-text text-transparent">
                {{ formatNumber(item.score) }}
              </span>
            </template>
          </DataTable>
        </div>
      </div>
    </section>
  </div>
</template>
