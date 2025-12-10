<script setup>
import { RouterLink } from 'vue-router'
import Avatar from './Avatar.vue'
import { formatNumber, slugify } from '../composables/formatters'

defineProps({
  team: {
    type: Object,
    required: true
  }
})
</script>

<template>
  <RouterLink
    :to="`/teams/${slugify(team.name)}`"
    class="card hover:shadow-lg transition group"
  >
    <div class="flex items-center justify-between mb-4">
      <h3 class="font-semibold text-gray-800 dark:text-white group-hover:text-primary-500 transition">
        {{ team.name }}
      </h3>
      <span
        class="w-3 h-3 rounded-full"
        :style="{ backgroundColor: team.color || '#8b5cf6' }"
      ></span>
    </div>

    <div class="flex items-center space-x-2 mb-4">
      <template v-for="(member, i) in (team.members || []).slice(0, 5)" :key="member">
        <Avatar :name="member" size="sm" />
      </template>
      <span
        v-if="(team.members?.length || 0) > 5"
        class="w-8 h-8 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center text-gray-600 dark:text-gray-300 text-xs font-bold"
      >
        +{{ team.members.length - 5 }}
      </span>
    </div>

    <div class="grid grid-cols-2 gap-4 text-center">
      <div>
        <div class="text-lg font-semibold gradient-text">
          {{ formatNumber(team.total_score) }}
        </div>
        <div class="text-xs text-gray-500 dark:text-gray-400">Total Score</div>
      </div>
      <div>
        <div class="text-lg font-semibold text-gray-800 dark:text-white">
          {{ team.members?.length || 0 }}
        </div>
        <div class="text-xs text-gray-500 dark:text-gray-400">Members</div>
      </div>
    </div>
  </RouterLink>
</template>
