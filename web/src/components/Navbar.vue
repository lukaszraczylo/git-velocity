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
      <div v-if="mobileMenuOpen" class="md:hidden py-4 border-t border-gray-200 dark:border-gray-700">
        <div class="flex flex-col space-y-3">
          <RouterLink
            to="/"
            @click="mobileMenuOpen = false"
            :class="route.path === '/' ? 'nav-link-active' : 'nav-link'"
          >
            Dashboard
          </RouterLink>
          <RouterLink
            to="/leaderboard"
            @click="mobileMenuOpen = false"
            :class="route.path === '/leaderboard' ? 'nav-link-active' : 'nav-link'"
          >
            Leaderboard
          </RouterLink>
          <RouterLink
            v-for="repo in repositories"
            :key="`${repo.Owner}/${repo.Name}`"
            :to="`/repos/${repo.Owner}/${repo.Name}`"
            @click="mobileMenuOpen = false"
            :class="route.path.includes(`/repos/${repo.Owner}/${repo.Name}`) ? 'nav-link-active' : 'nav-link'"
          >
            {{ repo.Name }}
          </RouterLink>
        </div>
      </div>
    </div>
  </nav>
</template>
