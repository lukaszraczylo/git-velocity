<script setup>
import { ref, inject, computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import ThemeToggle from './ThemeToggle.vue'

const route = useRoute()
const globalData = inject('globalData')
const mobileMenuOpen = ref(false)

const repositories = computed(() => globalData.value?.Repositories || [])
</script>

<template>
  <nav class="sticky top-0 z-50 glass shadow-modern">
    <div class="container mx-auto px-4">
      <div class="flex items-center justify-between h-16">
        <!-- Logo -->
        <RouterLink to="/" class="flex items-center space-x-2">
          <i class="fas fa-rocket text-2xl gradient-text"></i>
          <span class="text-xl font-bold gradient-text">Git Velocity</span>
        </RouterLink>

        <!-- Desktop Navigation -->
        <div class="hidden md:flex items-center space-x-6">
          <RouterLink
            to="/"
            :class="route.path === '/' ? 'nav-link-active' : 'nav-link'"
          >
            Dashboard
          </RouterLink>
          <RouterLink
            to="/leaderboard"
            :class="route.path === '/leaderboard' ? 'nav-link-active' : 'nav-link'"
          >
            Leaderboard
          </RouterLink>
          <RouterLink
            to="/how-scoring-works"
            :class="route.path === '/how-scoring-works' ? 'nav-link-active' : 'nav-link'"
          >
            How Scoring Works
          </RouterLink>
          <RouterLink
            v-for="repo in repositories"
            :key="`${repo.Owner}/${repo.Name}`"
            :to="`/repos/${repo.Owner}/${repo.Name}`"
            :class="route.path.includes(`/repos/${repo.Owner}/${repo.Name}`) ? 'nav-link-active' : 'nav-link'"
          >
            {{ repo.Name }}
          </RouterLink>
        </div>

        <!-- Actions -->
        <div class="flex items-center space-x-4">
          <ThemeToggle />

          <button
            @click="mobileMenuOpen = !mobileMenuOpen"
            class="md:hidden p-2 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition"
          >
            <i class="fas fa-bars text-gray-700 dark:text-gray-200"></i>
          </button>
        </div>
      </div>

      <!-- Mobile Menu -->
      <div v-if="mobileMenuOpen" class="md:hidden py-2 border-t border-gray-200 dark:border-gray-700">
        <div class="flex flex-col space-y-1">
          <RouterLink
            to="/"
            @click="mobileMenuOpen = false"
            :class="[
              'block px-4 py-3 rounded-lg text-base font-medium transition-colors',
              route.path === '/'
                ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-600 dark:text-primary-400'
                : 'text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800'
            ]"
          >
            <i class="fas fa-home mr-3 w-5 text-center"></i>Dashboard
          </RouterLink>
          <RouterLink
            to="/leaderboard"
            @click="mobileMenuOpen = false"
            :class="[
              'block px-4 py-3 rounded-lg text-base font-medium transition-colors',
              route.path === '/leaderboard'
                ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-600 dark:text-primary-400'
                : 'text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800'
            ]"
          >
            <i class="fas fa-trophy mr-3 w-5 text-center"></i>Leaderboard
          </RouterLink>
          <RouterLink
            to="/how-scoring-works"
            @click="mobileMenuOpen = false"
            :class="[
              'block px-4 py-3 rounded-lg text-base font-medium transition-colors',
              route.path === '/how-scoring-works'
                ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-600 dark:text-primary-400'
                : 'text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800'
            ]"
          >
            <i class="fas fa-calculator mr-3 w-5 text-center"></i>How Scoring Works
          </RouterLink>
          <RouterLink
            v-for="repo in repositories"
            :key="`${repo.Owner}/${repo.Name}`"
            :to="`/repos/${repo.Owner}/${repo.Name}`"
            @click="mobileMenuOpen = false"
            :class="[
              'block px-4 py-3 rounded-lg text-base font-medium transition-colors',
              route.path.includes(`/repos/${repo.Owner}/${repo.Name}`)
                ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-600 dark:text-primary-400'
                : 'text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800'
            ]"
          >
            <i class="fas fa-code-branch mr-3 w-5 text-center"></i>{{ repo.Name }}
          </RouterLink>
        </div>
      </div>
    </div>
  </nav>
</template>
