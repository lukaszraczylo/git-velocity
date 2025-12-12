<script setup>
defineProps({
  message: {
    type: String,
    default: 'An error occurred'
  },
  icon: {
    type: String,
    default: 'fas fa-exclamation-triangle'
  }
})
</script>

<template>
  <div class="flex items-center justify-center min-h-[60vh]">
    <div class="text-center">
      <i :class="icon" class="text-4xl text-red-500 mb-4"></i>
      <p class="text-gray-400">{{ message }}</p>
      <slot name="actions"></slot>
    </div>
  </div>
</template>
