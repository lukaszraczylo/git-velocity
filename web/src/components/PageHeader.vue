<script setup>
import Breadcrumb from './Breadcrumb.vue'

defineProps({
  title: {
    type: String,
    required: true
  },
  subtitle: {
    type: String,
    default: ''
  },
  icon: {
    type: String,
    default: ''
  },
  iconColor: {
    type: String,
    default: 'text-primary-500'
  },
  breadcrumbs: {
    type: Array,
    default: () => []
  },
  centered: {
    type: Boolean,
    default: false
  }
})
</script>

<template>
  <header class="py-12 px-4">
    <div class="container mx-auto" :class="{ 'text-center': centered }">
      <Breadcrumb v-if="breadcrumbs.length" :items="breadcrumbs" />

      <div class="flex items-center" :class="centered ? 'justify-center' : ''">
        <slot name="prefix"></slot>
        <h1 class="text-4xl font-bold mb-4">
          <i v-if="icon" :class="[icon, iconColor]" class="mr-3"></i>
          <span class="gradient-text">{{ title }}</span>
        </h1>
      </div>

      <p v-if="subtitle || $slots.subtitle" class="text-gray-600 dark:text-gray-300">
        <slot name="subtitle">{{ subtitle }}</slot>
      </p>

      <slot name="extra"></slot>
    </div>
  </header>
</template>
