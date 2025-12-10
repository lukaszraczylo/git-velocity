<script setup>
defineProps({
  url: {
    type: String,
    required: true
  },
  label: {
    type: String,
    default: ''
  },
  showIcon: {
    type: Boolean,
    default: true
  }
})
</script>

<template>
  <a
    :href="url"
    target="_blank"
    rel="noopener noreferrer"
    class="hover:text-primary-500 transition-colors"
    @click.stop
  >
    <slot>{{ label }}</slot>
    <i v-if="showIcon" class="fas fa-external-link-alt text-xs ml-1 opacity-50"></i>
  </a>
</template>
