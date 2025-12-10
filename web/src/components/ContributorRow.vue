<script setup>
import { RouterLink } from 'vue-router'
import Avatar from './Avatar.vue'
import { formatNumber } from '../composables/formatters'

defineProps({
  contributor: {
    type: Object,
    required: true
  },
  showGithubLink: {
    type: Boolean,
    default: false
  },
  columns: {
    type: Array,
    default: () => ['commits', 'prs', 'reviews', 'lines', 'score']
  }
})
</script>

<template>
  <RouterLink
    :to="{ name: 'contributor', params: { login: contributor.login } }"
    class="flex items-center space-x-3"
  >
    <Avatar
      :src="contributor.avatar_url"
      :name="contributor.login"
      class="ring-2 ring-transparent group-hover:ring-primary-500 transition-all"
    />
    <div>
      <div class="font-medium text-gray-800 dark:text-white group-hover:text-primary-500 transition-colors">
        {{ contributor.name || contributor.login }}
      </div>
      <div class="text-sm">
        <a
          v-if="showGithubLink"
          :href="`https://github.com/${contributor.login}`"
          target="_blank"
          rel="noopener noreferrer"
          class="text-gray-500 dark:text-gray-400 hover:text-primary-500 transition-colors"
          @click.stop
        >
          @{{ contributor.login }}
          <i class="fas fa-external-link-alt text-xs ml-1 opacity-50"></i>
        </a>
        <span v-else class="text-gray-500 dark:text-gray-400">
          @{{ contributor.login }}
        </span>
      </div>
    </div>
  </RouterLink>
</template>
