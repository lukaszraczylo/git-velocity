<script setup>
import { computed } from 'vue'

const props = defineProps({
  rank: { type: Number, required: true },
  size: { type: String, default: 'md' }
})

const sizeClasses = {
  sm: 'w-6 h-6 text-xs',
  md: 'w-8 h-8 text-sm'
}

const rankClass = computed(() => {
  if (props.rank === 1) return 'rank-1'
  if (props.rank === 2) return 'rank-2'
  if (props.rank === 3) return 'rank-3'
  return 'bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-300'
})

const classes = computed(() => sizeClasses[props.size] || sizeClasses.md)
</script>

<template>
  <span
    :class="[classes, rankClass, { 'text-white': rank <= 3 }]"
    class="inline-flex items-center justify-center rounded-full font-bold"
  >
    {{ rank }}
  </span>
</template>
