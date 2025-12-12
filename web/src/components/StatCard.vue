<script setup>
import { formatNumber } from '../composables/formatters'

defineProps({
  value: { type: [Number, String], required: true },
  label: { type: String, required: true },
  icon: { type: String, default: '' },
  iconColor: { type: String, default: 'text-gray-500' },
  valueClass: { type: String, default: 'gradient-text' },
  delay: { type: String, default: '0s' }
})
</script>

<template>
  <div class="card animate-fade-in-up" :style="{ animationDelay: delay }">
    <div class="flex items-center justify-between">
      <div class="min-w-0 flex-1">
        <div class="text-xl sm:text-2xl md:text-3xl font-bold truncate" :class="valueClass">
          {{ typeof value === 'number' ? formatNumber(value) : value }}
        </div>
        <div class="text-xs sm:text-sm text-gray-500 dark:text-gray-400 mt-1 truncate">{{ label }}</div>
      </div>
      <div v-if="icon" class="text-2xl sm:text-3xl opacity-50 ml-2 flex-shrink-0" :class="iconColor">
        <i :class="icon"></i>
      </div>
    </div>
  </div>
</template>
