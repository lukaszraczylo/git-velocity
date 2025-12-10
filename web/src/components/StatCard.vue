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
      <div>
        <div class="text-3xl font-bold" :class="valueClass">
          {{ typeof value === 'number' ? formatNumber(value) : value }}
        </div>
        <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">{{ label }}</div>
      </div>
      <div v-if="icon" class="text-3xl opacity-50" :class="iconColor">
        <i :class="icon"></i>
      </div>
    </div>
  </div>
</template>
