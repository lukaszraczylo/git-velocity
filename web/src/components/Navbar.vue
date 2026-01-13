<script setup>
import { ref, inject, computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

const route = useRoute()
const globalData = inject('globalData')
const mobileMenuOpen = ref(false)

const repositories = computed(() => globalData.value?.Repositories || [])
</script>

<template>
  <nav class="sticky top-0 z-50 bg-gray-900/80 backdrop-blur-md border-b border-gray-700 shadow-lg">
    <div class="container mx-auto px-4">
      <div class="flex items-center justify-between h-16">
        <!-- Logo -->
        <RouterLink to="/" class="flex items-center space-x-2">
          <i class="fas fa-rocket text-2xl bg-gradient-to-r from-primary-400 to-accent-400 bg-clip-text text-transparent"></i>
          <span class="text-xl font-bold bg-gradient-to-r from-primary-400 to-accent-400 bg-clip-text text-transparent">Git Velocity</span>
        </RouterLink>

        <!-- Desktop Navigation -->
        <div class="hidden md:flex items-center space-x-6">
          <RouterLink
            to="/"
            :class="route.path === '/' ? 'text-primary-500 font-medium' : 'text-gray-200 font-medium hover:text-primary-400 transition-colors'"
          >
            Dashboard
          </RouterLink>
          <RouterLink
            to="/leaderboard"
            :class="route.path === '/leaderboard' ? 'text-primary-500 font-medium' : 'text-gray-200 font-medium hover:text-primary-400 transition-colors'"
          >
            Leaderboard
          </RouterLink>
          <RouterLink
            to="/how-scoring-works"
            :class="route.path === '/how-scoring-works' ? 'text-primary-500 font-medium' : 'text-gray-200 font-medium hover:text-primary-400 transition-colors'"
          >
            How Scoring Works
          </RouterLink>
          <RouterLink
            v-for="repo in repositories"
            :key="`${repo.Owner}/${repo.Name}`"
            :to="`/repos/${repo.Owner}/${repo.Name}`"
            :class="route.path.includes(`/repos/${repo.Owner}/${repo.Name}`) ? 'text-primary-500 font-medium' : 'text-gray-200 font-medium hover:text-primary-400 transition-colors'"
          >
            {{ repo.Name }}
          </RouterLink>
        </div>

        <!-- Mobile Menu Button -->
        <button
          class="md:hidden p-2 rounded-lg hover:bg-gray-700 transition"
          @click="mobileMenuOpen = !mobileMenuOpen"
        >
          <i class="fas fa-bars text-gray-200"></i>
        </button>
      </div>

      <!-- Mobile Menu -->
      <div v-if="mobileMenuOpen" class="md:hidden py-2 border-t border-gray-700">
        <div class="flex flex-col space-y-1">
          <RouterLink
            to="/"
            :class="[
              'block px-4 py-3 rounded-lg text-base font-medium transition-colors',
              route.path === '/'
                ? 'bg-primary-900/20 text-primary-400'
                : 'text-gray-200 hover:bg-gray-800'
            ]"
            @click="mobileMenuOpen = false"
          >
            <i class="fas fa-home mr-3 w-5 text-center"></i>Dashboard
          </RouterLink>
          <RouterLink
            to="/leaderboard"
            :class="[
              'block px-4 py-3 rounded-lg text-base font-medium transition-colors',
              route.path === '/leaderboard'
                ? 'bg-primary-900/20 text-primary-400'
                : 'text-gray-200 hover:bg-gray-800'
            ]"
            @click="mobileMenuOpen = false"
          >
            <i class="fas fa-trophy mr-3 w-5 text-center"></i>Leaderboard
          </RouterLink>
          <RouterLink
            to="/how-scoring-works"
            :class="[
              'block px-4 py-3 rounded-lg text-base font-medium transition-colors',
              route.path === '/how-scoring-works'
                ? 'bg-primary-900/20 text-primary-400'
                : 'text-gray-200 hover:bg-gray-800'
            ]"
            @click="mobileMenuOpen = false"
          >
            <i class="fas fa-calculator mr-3 w-5 text-center"></i>How Scoring Works
          </RouterLink>
          <RouterLink
            v-for="repo in repositories"
            :key="`${repo.Owner}/${repo.Name}`"
            :to="`/repos/${repo.Owner}/${repo.Name}`"
            :class="[
              'block px-4 py-3 rounded-lg text-base font-medium transition-colors',
              route.path.includes(`/repos/${repo.Owner}/${repo.Name}`)
                ? 'bg-primary-900/20 text-primary-400'
                : 'text-gray-200 hover:bg-gray-800'
            ]"
            @click="mobileMenuOpen = false"
          >
            <i class="fas fa-code-branch mr-3 w-5 text-center"></i>{{ repo.Name }}
          </RouterLink>
        </div>
      </div>
    </div>
  </nav>
</template>
