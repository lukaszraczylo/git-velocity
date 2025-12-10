<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import LoadingState from '../components/LoadingState.vue'
import ErrorState from '../components/ErrorState.vue'
import StatCard from '../components/StatCard.vue'
import DataTable from '../components/DataTable.vue'
import ContributorRow from '../components/ContributorRow.vue'
import SectionHeader from '../components/SectionHeader.vue'
import GithubLink from '../components/GithubLink.vue'
import { formatNumber } from '../composables/formatters'

const route = useRoute()
const repository = ref(null)
const loading = ref(true)
const error = ref(null)

const breadcrumbs = computed(() => [
  { label: 'Dashboard', to: '/' },
  { label: 'Repositories' },
  { label: repository.value?.name || route.params.name }
])

const tableColumns = [
  { key: 'contributor', label: 'Contributor', align: 'left' },
  { key: 'commits', label: 'Commits', align: 'center' },
  { key: 'prs', label: 'PRs', align: 'center' },
  { key: 'reviews', label: 'Reviews', align: 'center' },
  { key: 'lines', label: 'Lines +/-', align: 'center' },
  { key: 'score', label: 'Score', align: 'right' }
]

async function loadRepository() {
  loading.value = true
  error.value = null

  try {
    const response = await fetch(`./data/repos/${route.params.owner}/${route.params.name}/metrics.json`)
    if (!response.ok) throw new Error('Repository not found')
    repository.value = await response.json()
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(loadRepository)
watch(() => route.params, loadRepository)
</script>

<template>
  <div>
    <LoadingState v-if="loading" message="Loading repository..." />
    <ErrorState v-else-if="error" :message="error" />

    <template v-else-if="repository">
      <PageHeader
        :title="repository.name"
        icon="fas fa-code-branch"
        icon-color="text-accent-500"
        :breadcrumbs="breadcrumbs"
      >
        <template #subtitle>
          <GithubLink :url="`https://github.com/${repository.owner}/${repository.name}`">
            {{ repository.owner }}/{{ repository.name }}
          </GithubLink>
        </template>
      </PageHeader>

      <!-- Stats -->
      <section class="py-8 px-4">
        <div class="container mx-auto">
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard
              :value="repository.total_commits"
              label="Commits"
              icon="fas fa-code-commit"
              icon-color="text-green-500"
            />
            <StatCard
              :value="repository.total_prs"
              label="Pull Requests"
              icon="fas fa-code-pull-request"
              icon-color="text-blue-500"
            />
            <StatCard
              :value="repository.total_reviews"
              label="Reviews"
              icon="fas fa-eye"
              icon-color="text-purple-500"
            />
            <StatCard
              :value="repository.active_contributors"
              label="Contributors"
              icon="fas fa-users"
              icon-color="text-orange-500"
            />
          </div>
        </div>
      </section>

      <!-- Contributors -->
      <section class="py-8 px-4">
        <div class="container mx-auto">
          <SectionHeader title="Contributors" icon="fas fa-users" icon-color="text-blue-500" />

          <DataTable
            :columns="tableColumns"
            :items="repository.contributors"
            empty-icon="fas fa-users"
            empty-message="No contributors found"
            row-class="hover:bg-gray-50 dark:hover:bg-gray-800/30 transition group"
          >
            <template #contributor="{ item }">
              <ContributorRow :contributor="item" />
            </template>
            <template #commits="{ item }">
              <span class="text-gray-800 dark:text-white">{{ formatNumber(item.commit_count) }}</span>
            </template>
            <template #prs="{ item }">
              <span class="text-gray-800 dark:text-white">{{ formatNumber(item.prs_opened) }}</span>
            </template>
            <template #reviews="{ item }">
              <span class="text-gray-800 dark:text-white">{{ formatNumber(item.reviews_given) }}</span>
            </template>
            <template #lines="{ item }">
              <span class="text-green-500">+{{ formatNumber(item.lines_added) }}</span>
              <span class="text-gray-400 mx-1">/</span>
              <span class="text-red-500">-{{ formatNumber(item.lines_deleted) }}</span>
            </template>
            <template #score="{ item }">
              <span class="text-lg font-bold gradient-text">
                {{ formatNumber(item.score?.total || 0) }}
              </span>
            </template>
          </DataTable>
        </div>
      </section>
    </template>
  </div>
</template>
