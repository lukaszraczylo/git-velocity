<script setup>
import { inject, computed, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Card from '../components/Card.vue'
import StatCard from '../components/StatCard.vue'
import ContributorCard from '../components/ContributorCard.vue'
import RepoCard from '../components/RepoCard.vue'
import TeamCard from '../components/TeamCard.vue'
import SectionHeader from '../components/SectionHeader.vue'
import VelocityChart from '../components/VelocityChart.vue'
import { formatNumber, formatDate } from '../composables/formatters'

const globalData = inject('globalData')

const metrics = computed(() => globalData.value || {})
const leaderboard = computed(() => metrics.value.leaderboard?.slice(0, 3) || [])
const repositories = computed(() => metrics.value.repositories || [])
const teams = computed(() => metrics.value.teams || [])
const velocityTimeline = computed(() => metrics.value.velocity_timeline)

const showScoreInChart = ref(false)
</script>

<template>
  <div>
    <!-- Hero Section -->
    <header class="py-10 sm:py-16 px-4">
      <div class="container mx-auto text-center animate-[fadeInUp_0.6s_ease-out]">
        <h1 class="text-3xl sm:text-4xl md:text-6xl font-bold mb-3 sm:mb-4">
          <span class="bg-gradient-to-r from-primary-400 to-accent-400 bg-clip-text text-transparent">Git Velocity</span>
        </h1>
        <p class="text-base sm:text-xl text-gray-300 max-w-2xl mx-auto px-2">
          Celebrate your team's achievements and contributions with beautiful insights.
        </p>
        <!-- Period and Generation Info -->
        <div class="flex flex-col items-center space-y-2 mt-4 text-sm text-gray-400">
          <p v-if="metrics.period?.start || metrics.period?.end">
            <i class="fas fa-calendar-alt mr-1 text-primary-500"></i>
            <span class="font-medium">Period:</span>
            <span v-if="metrics.period.start">{{ formatDate(metrics.period.start) }}</span>
            <span v-if="metrics.period.start && metrics.period.end"> &mdash; </span>
            <span v-if="metrics.period.end">{{ formatDate(metrics.period.end) }}</span>
          </p>
          <p v-if="metrics.generated_at">
            <i class="fas fa-clock mr-1"></i>
            Generated on {{ formatDate(metrics.generated_at) }}
          </p>
        </div>
      </div>
    </header>

    <!-- Velocity Timeline Chart -->
    <section v-if="velocityTimeline" class="py-6 sm:py-8 px-4">
      <div class="container mx-auto">
        <Card>
          <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-4 sm:mb-6">
            <SectionHeader title="Velocity Timeline" icon="fas fa-chart-line" icon-color="text-primary-500" />
            <label class="flex items-center space-x-2 text-sm text-gray-400 cursor-pointer">
              <input
                type="checkbox"
                v-model="showScoreInChart"
                class="rounded border-gray-600 text-primary-500 focus:ring-primary-500"
              />
              <span>Show Score</span>
            </label>
          </div>
          <div class="h-[200px] sm:h-[280px] md:h-[320px]">
            <VelocityChart :timeline="velocityTimeline" :show-score="showScoreInChart" height="100%" />
          </div>
        </Card>
      </div>
    </section>

    <!-- Stats Overview -->
    <section class="py-8 px-4">
      <div class="container mx-auto">
        <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
          <StatCard
            :value="metrics.total_contributors || 0"
            label="Contributors"
            delay="0s"
          />
          <StatCard
            :value="metrics.total_commits || 0"
            label="Commits"
            delay="0.1s"
          />
          <StatCard
            :value="metrics.total_prs || 0"
            label="Pull Requests"
            delay="0.2s"
          />
          <StatCard
            :value="metrics.total_reviews || 0"
            label="Reviews"
            delay="0.3s"
          />
          <StatCard
            :value="'+' + formatNumber(metrics.total_lines_added || 0)"
            label="Lines Added"
            value-class="text-green-500"
            delay="0.4s"
          />
          <StatCard
            :value="'-' + formatNumber(metrics.total_lines_deleted || 0)"
            label="Lines Deleted"
            value-class="text-red-500"
            delay="0.5s"
          />
        </div>
      </div>
    </section>

    <!-- Top Contributors -->
    <section class="py-8 px-4">
      <div class="container mx-auto">
        <SectionHeader title="Top Contributors" icon="fas fa-trophy" icon-color="text-yellow-500" />

        <div class="grid md:grid-cols-3 gap-6">
          <ContributorCard
            v-for="(entry, index) in leaderboard"
            :key="entry.login"
            :contributor="entry"
            :rank="index + 1"
            featured
          />
        </div>

        <div class="mt-6 text-center">
          <RouterLink to="/leaderboard" class="inline-flex items-center px-6 py-3 bg-gradient-to-r from-primary-500 to-accent-500 text-white font-medium rounded-lg shadow-lg hover:from-primary-600 hover:to-accent-600 transition-all">
            View Full Leaderboard
            <i class="fas fa-arrow-right ml-2"></i>
          </RouterLink>
        </div>
      </div>
    </section>

    <!-- Repositories -->
    <section class="py-8 px-4">
      <div class="container mx-auto">
        <SectionHeader title="Repositories" icon="fas fa-code-branch" icon-color="text-accent-500" />

        <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          <RepoCard v-for="repo in repositories" :key="`${repo.owner}/${repo.name}`" :repo="repo" />
        </div>
      </div>
    </section>

    <!-- Teams -->
    <section v-if="teams.length" class="py-8 px-4">
      <div class="container mx-auto">
        <SectionHeader title="Teams" icon="fas fa-users" icon-color="text-blue-500" />

        <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          <TeamCard v-for="team in teams" :key="team.name" :team="team" />
        </div>
      </div>
    </section>
  </div>
</template>
