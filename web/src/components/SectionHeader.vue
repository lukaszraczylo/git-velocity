<script setup>
defineProps({
  title: {
    type: String,
    required: true
  },
  icon: {
    type: String,
    default: ''
  },
  iconColor: {
    type: String,
    default: 'text-primary-500'
  }
})
</script>

<template>
  <h2 class="text-2xl font-bold text-gray-800 dark:text-white mb-6">
    <i v-if="icon" :class="[icon, iconColor]" class="mr-2"></i>{{ title }}
    <slot name="suffix"></slot>
  </h2>
</template>
