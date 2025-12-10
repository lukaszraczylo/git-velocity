<script setup>
import { ref, computed, onMounted, watch, inject } from 'vue'
import { useRoute } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import LoadingState from '../components/LoadingState.vue'
import ErrorState from '../components/ErrorState.vue'
import StatCard from '../components/StatCard.vue'
import MemberCard from '../components/MemberCard.vue'
import SectionHeader from '../components/SectionHeader.vue'
import { slugify } from '../composables/formatters'

const route = useRoute()
const globalData = inject('globalData')
const team = ref(null)
const loading = ref(true)
const error = ref(null)

const breadcrumbs = computed(() => [
  { label: 'Dashboard', to: '/' },
  { label: 'Teams' },
  { label: team.value?.name || route.params.slug }
])

function loadTeam() {
  loading.value = true
  error.value = null

  const teams = globalData.value?.teams || []
  const found = teams.find(t => slugify(t.name) === route.params.slug)

  if (found) {
    team.value = found
  } else {
    error.value = 'Team not found'
  }

  loading.value = false
}

onMounted(loadTeam)
watch(() => route.params, loadTeam)
watch(globalData, loadTeam)
</script>

<template>
  <div>
    <LoadingState v-if="loading" message="Loading team..." />
    <ErrorState v-else-if="error" :message="error" />

    <template v-else-if="team">
      <PageHeader
        :title="team.name"
        :breadcrumbs="breadcrumbs"
        :subtitle="`${team.members?.length || 0} team members`"
      >
        <template #prefix>
          <div
            class="w-4 h-4 rounded-full mr-4"
            :style="{ backgroundColor: team.color || '#8b5cf6' }"
          ></div>
        </template>
      </PageHeader>

      <!-- Team Stats -->
      <section class="py-8 px-4">
        <div class="container mx-auto">
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard
              :value="team.total_score"
              label="Total Score"
              icon="fas fa-star"
              icon-color="text-yellow-500"
            />
            <StatCard
              :value="team.aggregated_metrics?.commit_count || 0"
              label="Commits"
              icon="fas fa-code-commit"
              icon-color="text-green-500"
            />
            <StatCard
              :value="team.aggregated_metrics?.prs_merged || 0"
              label="PRs Merged"
              icon="fas fa-code-merge"
              icon-color="text-purple-500"
            />
            <StatCard
              :value="team.aggregated_metrics?.reviews_given || 0"
              label="Reviews"
              icon="fas fa-eye"
              icon-color="text-blue-500"
            />
          </div>
        </div>
      </section>

      <!-- Team Members -->
      <section class="py-8 px-4">
        <div class="container mx-auto">
          <SectionHeader title="Team Members" icon="fas fa-users" icon-color="text-blue-500" />

          <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            <MemberCard
              v-for="member in team.member_metrics"
              :key="member.login"
              :member="member"
            />
          </div>
        </div>
      </section>
    </template>
  </div>
</template>
