<script setup>
import Card from './Card.vue'
import { formatNumber } from '../composables/formatters'

defineProps({
  value: { type: [Number, String], required: true },
  label: { type: String, required: true },
  icon: { type: String, default: '' },
  iconColor: { type: String, default: 'text-gray-500' },
  valueClass: { type: String, default: 'bg-gradient-to-r from-primary-600 to-accent-600 dark:from-primary-400 dark:to-accent-400 bg-clip-text text-transparent' },
  delay: { type: String, default: '0s' }
})
</script>

<template>
  <Card class="animate-[fadeInUp_0.6s_ease-out]" :style="{ animationDelay: delay }">
    <div class="flex items-center justify-between">
      <div class="min-w-0 flex-1">
        <div class="text-xl sm:text-2xl md:text-3xl font-bold truncate" :class="valueClass">
          {{ typeof value === 'number' ? formatNumber(value) : value }}
        </div>
        <div class="text-xs sm:text-sm text-gray-700 dark:text-gray-400 mt-1 truncate">{{ label }}</div>
      </div>
      <div v-if="icon" class="text-2xl sm:text-3xl opacity-50 ml-2 flex-shrink-0" :class="iconColor">
        <i :class="icon"></i>
      </div>
    </div>
  </Card>
</template>
