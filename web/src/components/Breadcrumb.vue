<script setup>
import { RouterLink } from 'vue-router'

defineProps({
  items: {
    type: Array,
    required: true
    // Each item: { label: string, to?: string }
  }
})
</script>

<template>
  <div class="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400 mb-6">
    <template v-for="(item, index) in items" :key="index">
      <RouterLink
        v-if="item.to"
        :to="item.to"
        class="hover:text-primary-500"
      >
        {{ item.label }}
      </RouterLink>
      <span
        v-else
        :class="index === items.length - 1 ? 'text-gray-900 dark:text-white' : ''"
      >
        {{ item.label }}
      </span>
      <i
        v-if="index < items.length - 1"
        class="fas fa-chevron-right text-xs"
      ></i>
    </template>
  </div>
</template>
