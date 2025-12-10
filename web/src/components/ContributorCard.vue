<script setup>
import { RouterLink } from 'vue-router'
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
    :class="[
      'card animate-fade-in-up block cursor-pointer hover:shadow-lg transition-shadow',
      featured && rank === 1 ? 'ring-2 ring-yellow-400' : ''
    ]"
  >
    <div class="flex items-center space-x-4">
      <div class="relative">
        <Avatar
          :src="contributor.avatar_url"
          :name="contributor.login"
          :size="featured ? 'xl' : 'lg'"
        />
        <RankBadge
          v-if="showRank && rank > 0"
          :rank="rank"
          size="sm"
          class="absolute -top-1 -right-1"
        />
      </div>

      <div class="flex-1">
        <h3 class="font-semibold text-gray-800 dark:text-white group-hover:text-primary-500 transition-colors">
          {{ contributor.name || contributor.login }}
        </h3>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          <span
            class="hover:text-primary-500 transition-colors"
            @click.stop.prevent="window.open(`https://github.com/${contributor.login}`, '_blank')"
          >
            @{{ contributor.login }}
            <i class="fas fa-external-link-alt text-xs ml-0.5 opacity-50"></i>
          </span>
        </p>
        <p v-if="contributor.team" class="text-xs text-accent-500">{{ contributor.team }}</p>
      </div>

      <div class="text-right">
        <div class="text-2xl font-bold gradient-text">
          {{ formatNumber(contributor.score?.total || contributor.score || 0) }}
        </div>
        <div class="text-xs text-gray-500 dark:text-gray-400">points</div>
      </div>
    </div>

    <div v-if="contributor.achievements?.length" class="mt-4 flex flex-wrap gap-1.5">
      <AchievementBadge
        v-for="achievement in contributor.achievements.slice(0, 6)"
        :key="achievement"
        :achievement-id="achievement"
        size="sm"
      />
      <span
        v-if="contributor.achievements.length > 6"
        class="inline-flex items-center justify-center w-8 h-8 rounded-lg bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-300 text-xs font-bold"
      >
        +{{ contributor.achievements.length - 6 }}
      </span>
    </div>
  </RouterLink>
</template>
