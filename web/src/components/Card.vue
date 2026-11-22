<script setup>
defineProps({
  padding: { type: Boolean, default: true },
  hover: { type: Boolean, default: false }
})
</script>

<template>
  <div
    :class="[
      'rounded-xl bg-gray-800 text-white border border-gray-700 shadow',
      padding ? 'p-6' : '',
      hover ? 'hover:shadow-lg transition-shadow' : ''
    ]"
  >
    <slot />
  </div>
</template>
