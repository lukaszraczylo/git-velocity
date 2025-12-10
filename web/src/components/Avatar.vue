<script setup>
import { computed } from 'vue'

const props = defineProps({
  src: { type: String, default: '' },
  name: { type: String, required: true },
  size: { type: String, default: 'md' }
})

const sizeClasses = {
  sm: 'w-8 h-8 text-xs',
  md: 'w-10 h-10 text-sm',
  lg: 'w-14 h-14 text-xl',
  xl: 'w-16 h-16 text-2xl',
  '2xl': 'w-32 h-32 text-4xl'
}

const initial = computed(() => props.name.charAt(0).toUpperCase())
const classes = computed(() => sizeClasses[props.size] || sizeClasses.md)
</script>

<template>
  <img
    v-if="src"
    :src="src"
    :alt="name"
    :class="classes"
    class="rounded-full"
  />
  <div
    v-else
    :class="classes"
    class="rounded-full bg-gradient-to-br from-primary-500 to-accent-500 flex items-center justify-center text-white font-bold"
  >
    {{ initial }}
  </div>
</template>
