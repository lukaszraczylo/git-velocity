<script setup>
import { RouterLink } from 'vue-router'
import Card from './Card.vue'
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
    class="block group"
  >
    <Card hover>
      <div class="flex items-center justify-between mb-4">
        <h3 class="font-semibold text-white group-hover:text-primary-500 transition">
          {{ team.name }}
        </h3>
        <span
          class="w-3 h-3 rounded-full"
          :style="{ backgroundColor: team.color || '#8b5cf6' }"
        ></span>
      </div>

      <div class="flex items-center space-x-2 mb-4">
        <template v-for="member in (team.member_metrics || []).slice(0, 5)" :key="member.login">
          <Avatar :name="member.name || member.login" :src="member.avatar_url" size="sm" />
        </template>
        <span
          v-if="(team.member_metrics?.length || 0) > 5"
          class="w-8 h-8 rounded-full bg-gray-700 flex items-center justify-center text-gray-300 text-xs font-bold"
        >
          +{{ team.member_metrics.length - 5 }}
        </span>
      </div>

      <div class="grid grid-cols-2 gap-4 text-center">
        <div>
          <div class="text-lg font-semibold bg-gradient-to-r from-primary-400 to-accent-400 bg-clip-text text-transparent">
            {{ formatNumber(team.total_score) }}
          </div>
          <div class="text-xs text-gray-400">Total Score</div>
        </div>
        <div>
          <div class="text-lg font-semibold text-white">
            {{ team.members?.length || 0 }}
          </div>
          <div class="text-xs text-gray-400">Members</div>
        </div>
      </div>
    </Card>
  </RouterLink>
</template>
