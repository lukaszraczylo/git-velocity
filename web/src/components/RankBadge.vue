<script setup>
import { computed } from 'vue'

const props = defineProps({
  rank: { type: Number, required: true },
  size: { type: String, default: 'md' }
})

const sizeClasses = {
  sm: 'w-6 h-6 text-xs',
  md: 'w-8 h-8 text-sm',
  lg: 'w-10 h-10 text-base'
}

const rankClass = computed(() => {
  if (props.rank === 1) return 'bg-gradient-to-r from-yellow-400 to-amber-500'
  if (props.rank === 2) return 'bg-gradient-to-r from-slate-400 to-slate-500'
  if (props.rank === 3) return 'bg-gradient-to-r from-amber-600 to-amber-700'
  return 'bg-gray-700 text-gray-300'
})

const classes = computed(() => sizeClasses[props.size] || sizeClasses.md)

const isTopThree = computed(() => props.rank >= 1 && props.rank <= 3)
</script>

<template>
  <span
    :class="[classes, rankClass, { 'text-white': rank <= 3 }]"
    class="inline-flex items-center justify-center rounded-full font-bold"
  >
    <i v-if="isTopThree" class="fas fa-trophy"></i>
    <template v-else>{{ rank }}</template>
  </span>
</template>
