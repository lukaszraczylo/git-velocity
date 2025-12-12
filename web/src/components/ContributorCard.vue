<script setup>
import { RouterLink } from 'vue-router'
import Card from './Card.vue'
import Avatar from './Avatar.vue'
import RankBadge from './RankBadge.vue'
import AchievementBadge from './AchievementBadge.vue'
import { formatNumber } from '../composables/formatters'

defineProps({
  contributor: { type: Object, required: true },
  rank: { type: Number, default: 0 },
  showRank: { type: Boolean, default: true },
  featured: { type: Boolean, default: false }
})
</script>

<template>
  <RouterLink
    :to="{ name: 'contributor', params: { login: contributor.login } }"
    class="block group"
  >
    <Card
      hover
      :class="[
        'animate-[fadeInUp_0.6s_ease-out] h-full',
        featured && rank === 1 ? 'ring-2 ring-yellow-400' : '',
        featured && rank === 2 ? 'ring-2 ring-gray-300' : '',
        featured && rank === 3 ? 'ring-2 ring-amber-600' : ''
      ]"
    >
      <div class="flex flex-col h-full">
        <!-- Header with avatar and rank -->
        <div class="flex items-start justify-between mb-4">
          <div class="flex items-center gap-4">
            <div class="relative">
              <Avatar
                :src="contributor.avatar_url"
                :name="contributor.login"
                :size="featured ? 'xl' : 'lg'"
                class="ring-2 ring-gray-700"
              />
              <RankBadge
                v-if="showRank && rank > 0"
                :rank="rank"
                size="sm"
                class="absolute -bottom-1 -right-1"
              />
            </div>
            <div>
              <h3 class="font-bold text-lg text-white group-hover:text-primary-500 transition-colors">
                {{ contributor.name || contributor.login }}
              </h3>
              <p class="text-sm text-gray-400">
                @{{ contributor.login }}
              </p>
              <p v-if="contributor.team" class="text-xs text-accent-500 mt-0.5">{{ contributor.team }}</p>
            </div>
          </div>
        </div>

        <!-- Score display -->
        <div class="flex items-center justify-between py-3 px-4 -mx-2 rounded-lg bg-gradient-to-r from-primary-900/20 to-accent-900/20 mb-4">
          <span class="text-sm font-medium text-gray-300">Score</span>
          <span class="text-2xl font-bold bg-gradient-to-r from-primary-400 to-accent-400 bg-clip-text text-transparent">
            {{ formatNumber(contributor.score?.total || contributor.score || 0) }}
          </span>
        </div>

        <!-- Achievements -->
        <div v-if="contributor.achievements?.length" class="mt-auto">
          <div class="text-xs font-medium text-gray-400 mb-2">Achievements</div>
          <div class="flex flex-wrap gap-1.5">
            <AchievementBadge
              v-for="achievement in contributor.achievements.slice(0, 8)"
              :key="achievement"
              :achievement-id="achievement"
              size="sm"
            />
            <span
              v-if="contributor.achievements.length > 8"
              class="inline-flex items-center justify-center px-2 h-7 rounded-lg bg-gray-700 text-gray-300 text-xs font-bold"
            >
              +{{ contributor.achievements.length - 8 }}
            </span>
          </div>
        </div>
      </div>
    </Card>
  </RouterLink>
</template>
