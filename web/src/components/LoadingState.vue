<script setup>
defineProps({
  message: {
    type: String,
    default: 'Loading...'
  }
})
</script>

<template>
  <div class="flex items-center justify-center min-h-[60vh]">
    <div class="text-center">
      <i class="fas fa-spinner fa-spin text-4xl text-primary-500 mb-4"></i>
      <p class="text-gray-400">{{ message }}</p>
    </div>
  </div>
</template>
