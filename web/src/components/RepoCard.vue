<script setup>
import { RouterLink } from 'vue-router'
import Card from './Card.vue'
import { formatNumber } from '../composables/formatters'

defineProps({
  repo: {
    type: Object,
    required: true
  }
})
</script>

<template>
  <RouterLink
    :to="`/repos/${repo.owner}/${repo.name}`"
    class="block group"
  >
    <Card hover>
      <div class="flex items-center justify-between mb-4">
        <h3 class="font-semibold text-white group-hover:text-primary-500 transition">
          {{ repo.name }}
        </h3>
        <i class="fas fa-arrow-right text-gray-400 group-hover:text-primary-500 transition"></i>
      </div>
      <p class="text-sm text-gray-400 mb-4">{{ repo.owner }}/{{ repo.name }}</p>

      <div class="grid grid-cols-3 gap-4 text-center">
        <div>
          <div class="text-lg font-semibold text-white">
            {{ formatNumber(repo.total_commits) }}
          </div>
          <div class="text-xs text-gray-400">Commits</div>
        </div>
        <div>
          <div class="text-lg font-semibold text-white">
            {{ formatNumber(repo.total_prs) }}
          </div>
          <div class="text-xs text-gray-400">PRs</div>
        </div>
        <div>
          <div class="text-lg font-semibold text-white">
            {{ repo.active_contributors }}
          </div>
          <div class="text-xs text-gray-400">Contributors</div>
        </div>
      </div>
    </Card>
  </RouterLink>
</template>
