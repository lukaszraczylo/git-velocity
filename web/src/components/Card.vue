<script setup>
defineProps({
  padding: { type: Boolean, default: true },
  hover: { type: Boolean, default: false }
})
</script>

<template>
  <div
    :class="[
      'rounded-xl bg-white dark:bg-gray-800 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-700 shadow',
      padding ? 'p-6' : '',
      hover ? 'hover:shadow-lg transition-shadow' : ''
    ]"
  >
    <slot />
  </div>
</template>
