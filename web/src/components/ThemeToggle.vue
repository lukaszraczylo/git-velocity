<script setup>
import { ref, onMounted, watch } from 'vue'

const isDark = ref(false)

onMounted(() => {
  const savedTheme = localStorage.getItem('theme')
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches

  isDark.value = savedTheme === 'dark' || (!savedTheme && prefersDark)
  updateTheme()

  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
    if (!localStorage.getItem('theme')) {
      isDark.value = e.matches
      updateTheme()
    }
  })
})

watch(isDark, () => {
  updateTheme()
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
})

function updateTheme() {
  if (isDark.value) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}

function toggle() {
  isDark.value = !isDark.value
}
</script>

<template>
  <button
    @click="toggle"
    class="p-2 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition"
    :aria-label="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
  >
    <i v-if="isDark" class="fas fa-moon text-purple-400"></i>
    <i v-else class="fas fa-sun text-yellow-500"></i>
  </button>
</template>
