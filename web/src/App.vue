<script setup>
import { ref, onMounted, provide } from 'vue'
import Navbar from './components/Navbar.vue'
import Footer from './components/Footer.vue'

const globalData = ref(null)
const loading = ref(true)
const error = ref(null)

provide('globalData', globalData)

onMounted(async () => {
  try {
    const response = await fetch('./data/global.json')
    if (!response.ok) throw new Error('Failed to load data')
    globalData.value = await response.json()
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <Navbar />

    <main class="flex-1">
      <div v-if="loading" class="flex items-center justify-center min-h-[60vh]">
        <div class="text-center">
          <i class="fas fa-spinner fa-spin text-4xl text-primary-500 mb-4"></i>
          <p class="text-gray-600 dark:text-gray-400">Loading dashboard...</p>
        </div>
      </div>

      <div v-else-if="error" class="flex items-center justify-center min-h-[60vh]">
        <div class="text-center">
          <i class="fas fa-exclamation-triangle text-4xl text-red-500 mb-4"></i>
          <p class="text-gray-600 dark:text-gray-400">{{ error }}</p>
        </div>
      </div>

      <router-view v-else />
    </main>

    <Footer />
  </div>
</template>
