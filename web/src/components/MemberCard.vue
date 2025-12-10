<script setup>
import { RouterLink } from 'vue-router'
import Avatar from './Avatar.vue'
import AchievementBadge from './AchievementBadge.vue'
import { formatNumber } from '../composables/formatters'

defineProps({
  member: {
    type: Object,
    required: true
  },
  linkToProfile: {
    type: Boolean,
    default: true
  }
})
</script>

<template>
  <component
    :is="linkToProfile ? RouterLink : 'div'"
    :to="linkToProfile ? { name: 'contributor', params: { login: member.login } } : undefined"
    class="card block"
    :class="{ 'hover:shadow-lg transition cursor-pointer': linkToProfile }"
  >
    <div class="flex items-center space-x-4 mb-4">
      <Avatar :src="member.avatar_url" :name="member.login" size="lg" />
      <div>
        <h3 class="font-semibold text-gray-800 dark:text-white">
          {{ member.name || member.login }}
        </h3>
        <p class="text-sm text-gray-500 dark:text-gray-400">@{{ member.login }}</p>
      </div>
    </div>

    <div class="grid grid-cols-3 gap-4 text-center mb-4">
      <div>
        <div class="text-lg font-semibold text-gray-800 dark:text-white">
          {{ formatNumber(member.commit_count) }}
        </div>
        <div class="text-xs text-gray-500 dark:text-gray-400">Commits</div>
      </div>
      <div>
        <div class="text-lg font-semibold text-gray-800 dark:text-white">
          {{ formatNumber(member.prs_opened) }}
        </div>
        <div class="text-xs text-gray-500 dark:text-gray-400">PRs</div>
      </div>
      <div>
        <div class="text-lg font-semibold text-gray-800 dark:text-white">
          {{ formatNumber(member.reviews_given) }}
        </div>
        <div class="text-xs text-gray-500 dark:text-gray-400">Reviews</div>
      </div>
    </div>

    <div class="flex items-center justify-between pt-4 border-t border-gray-200 dark:border-gray-700">
      <span class="text-sm text-gray-500 dark:text-gray-400">Score</span>
      <span class="text-xl font-bold gradient-text">
        {{ formatNumber(member.score?.total || 0) }}
      </span>
    </div>

    <div v-if="member.achievements?.length" class="mt-4 flex flex-wrap gap-2">
      <AchievementBadge
        v-for="achievement in member.achievements.slice(0, 4)"
        :key="achievement"
        :achievement-id="achievement"
        size="sm"
      />
      <span
        v-if="member.achievements.length > 4"
        class="inline-flex items-center justify-center w-8 h-8 rounded-lg bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-300 text-xs font-bold"
      >
        +{{ member.achievements.length - 4 }}
      </span>
    </div>
  </component>
</template>
