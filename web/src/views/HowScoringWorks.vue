<script setup>
import Card from '../components/Card.vue'
import SectionHeader from '../components/SectionHeader.vue'
</script>

<template>
  <div>
    <!-- Hero Section -->
    <header class="py-10 sm:py-16 px-4">
      <div class="container mx-auto text-center animate-[fadeInUp_0.6s_ease-out]">
        <h1 class="text-3xl sm:text-4xl md:text-5xl font-bold mb-3 sm:mb-4 text-gray-900 dark:text-white">
          How <span class="bg-gradient-to-r from-primary-600 to-accent-600 dark:from-primary-400 dark:to-accent-400 bg-clip-text text-transparent">Scoring</span> Works
        </h1>
        <p class="text-base sm:text-lg md:text-xl text-gray-600 dark:text-gray-300 max-w-2xl mx-auto px-2">
          Understanding the point system, leaderboard rankings, and achievement criteria that power Git Velocity.
        </p>
      </div>
    </header>

    <!-- Overview Section -->
    <section class="py-8 px-4">
      <div class="container mx-auto">
        <Card class="shadow-lg mb-8">
          <h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center">
            <i class="fas fa-info-circle mr-3 text-blue-500"></i>
            Overview
          </h2>
          <p class="text-gray-600 dark:text-gray-400 mb-4">
            Git Velocity calculates developer contributions by analyzing GitHub activity across configured repositories.
            The scoring system is designed to encourage well-rounded contributions including code commits, pull requests,
            code reviews, and collaboration.
          </p>
          <div class="grid sm:grid-cols-3 gap-4 mt-6">
            <div class="text-center p-4 bg-primary-50 dark:bg-primary-900/20 rounded-lg">
              <i class="fas fa-calculator text-primary-500 text-2xl mb-2"></i>
              <h3 class="font-medium text-gray-900 dark:text-gray-100">Point-Based</h3>
              <p class="text-sm text-gray-600 dark:text-gray-400">Activities earn configurable points</p>
            </div>
            <div class="text-center p-4 bg-accent-50 dark:bg-accent-900/20 rounded-lg">
              <i class="fas fa-layer-group text-accent-500 text-2xl mb-2"></i>
              <h3 class="font-medium text-gray-900 dark:text-gray-100">Aggregated</h3>
              <p class="text-sm text-gray-600 dark:text-gray-400">Combined across all repositories</p>
            </div>
            <div class="text-center p-4 bg-indigo-50 dark:bg-indigo-900/20 rounded-lg">
              <i class="fas fa-trophy text-indigo-500 text-2xl mb-2"></i>
              <h3 class="font-medium text-gray-900 dark:text-gray-100">Achievement-Driven</h3>
              <p class="text-sm text-gray-600 dark:text-gray-400">Unlock badges for milestones</p>
            </div>
          </div>
        </Card>
      </div>
    </section>

    <!-- Scoring Section -->
    <section id="scoring" class="py-8 px-4">
      <div class="container mx-auto">
        <SectionHeader title="Point Calculations" icon="fas fa-coins" icon-color="text-yellow-500" />

        <div class="space-y-6">
          <!-- Score Formula -->
          <Card class="shadow-lg">
            <h3 class="font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center text-lg sm:text-xl">
              <i class="fas fa-function mr-2 text-primary-500"></i>
              Score Formula
            </h3>
            <div class="bg-gray-900 text-gray-100 p-3 sm:p-4 rounded-lg overflow-x-auto mb-4 -mx-2 sm:mx-0">
              <pre class="text-xs sm:text-sm font-mono whitespace-pre-wrap sm:whitespace-pre"><code>Total Score = Commits + Lines + PRs + Reviews + Comments + Issues + Bonuses

Where:
  Commits    = commit_count x 10 pts
  Lines      = (added x 0.1) + (deleted x 0.05) pts
  PRs        = (opened x 25) + (merged x 50) pts
  Reviews    = reviews_given x 30 pts
  Comments   = review_comments x 5 pts
  Issues     = (opened x 10) + (closed x 20) + (comments x 5) + (refs x 5) pts
  Response   = fast review bonus (0-50 pts)
  Out of Hrs = commits outside 9-5 x 2 pts</code></pre>
            </div>
            <p class="text-xs sm:text-sm text-gray-600 dark:text-gray-400">
              <i class="fas fa-info-circle mr-1"></i>
              All point values are configurable in your <code class="text-primary-600 dark:text-primary-400">.git-velocity.yaml</code> file.
            </p>
          </Card>

          <!-- Default Point Values -->
          <Card class="shadow-lg">
            <h3 class="font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center text-xl">
              <i class="fas fa-coins mr-2 text-yellow-500"></i>
              Default Point Values
            </h3>
            <!-- Mobile: Card Layout -->
            <div class="grid grid-cols-1 gap-3 sm:hidden">
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-code-commit text-primary-500"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">Commit</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">10 pts</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-flask text-green-500"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">Commit + Tests</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">15 pts</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-plus text-blue-500"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">Lines Added</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">0.1 pts</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-minus text-red-500"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">Lines Deleted</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">0.05 pts</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-code-pull-request text-accent-500"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">PR Opened</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">25 pts</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-code-merge text-indigo-500"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">PR Merged</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">50 pts</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-eye text-cyan-500"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">PR Reviewed</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">30 pts</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-comment text-orange-500"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">Review Comment</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">5 pts</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-bolt text-yellow-500"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">Fast Review &lt;1h</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">50 pts</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-stopwatch text-yellow-500"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">Fast Review &lt;4h</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">25 pts</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-clock text-yellow-500"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">Fast Review &lt;24h</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">10 pts</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-moon text-gray-600 dark:text-gray-400"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">Out of Hours</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">2 pts</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-circle-exclamation text-teal-500"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">Issue Opened</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">10 pts</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-circle-check text-green-500"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">Issue Closed</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">20 pts</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-comment-dots text-blue-500"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">Issue Comment</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">5 pts</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2">
                  <i class="fas fa-link text-accent-500"></i>
                  <span class="text-sm font-medium text-gray-900 dark:text-gray-100">Issue Reference</span>
                </div>
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400">5 pts</span>
              </div>
            </div>
            <!-- Desktop: Table Layout -->
            <div class="hidden sm:block overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="border-b border-gray-200 dark:border-gray-700">
                    <th class="text-left py-3 text-gray-700 dark:text-gray-400">Activity</th>
                    <th class="text-left py-3 text-gray-700 dark:text-gray-400">Points</th>
                    <th class="text-left py-3 text-gray-700 dark:text-gray-400">Description</th>
                  </tr>
                </thead>
                <tbody class="text-gray-700 dark:text-gray-300">
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-3"><i class="fas fa-code-commit text-primary-500 mr-2"></i>Commit</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">10</td>
                    <td class="py-3">Per commit pushed</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-3"><i class="fas fa-flask text-green-500 mr-2"></i>Commit with Tests</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">15</td>
                    <td class="py-3">Commit that includes test files</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-3"><i class="fas fa-plus text-blue-500 mr-2"></i>Lines Added</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">0.1</td>
                    <td class="py-3">Per meaningful line added</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-3"><i class="fas fa-minus text-red-500 mr-2"></i>Lines Deleted</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">0.05</td>
                    <td class="py-3">Per meaningful line removed</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-3"><i class="fas fa-code-pull-request text-accent-500 mr-2"></i>PR Opened</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">25</td>
                    <td class="py-3">Per pull request created</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-3"><i class="fas fa-code-merge text-indigo-500 mr-2"></i>PR Merged</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">50</td>
                    <td class="py-3">Per pull request merged</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-3"><i class="fas fa-eye text-cyan-500 mr-2"></i>PR Reviewed</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">30</td>
                    <td class="py-3">Per PR review submitted</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-3"><i class="fas fa-comment text-orange-500 mr-2"></i>Review Comment</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">5</td>
                    <td class="py-3">Per comment on PR reviews</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-3"><i class="fas fa-bolt text-yellow-500 mr-2"></i>Fast Review (&lt;1h)</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">50</td>
                    <td class="py-3">Bonus for average response under 1 hour</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-3"><i class="fas fa-stopwatch text-yellow-500 mr-2"></i>Fast Review (&lt;4h)</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">25</td>
                    <td class="py-3">Bonus for average response under 4 hours</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-3"><i class="fas fa-clock text-yellow-500 mr-2"></i>Fast Review (&lt;24h)</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">10</td>
                    <td class="py-3">Bonus for average response under 24 hours</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-3"><i class="fas fa-moon text-gray-500 mr-2"></i>Out of Hours</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">2</td>
                    <td class="py-3">Per commit outside 9am-5pm</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-3"><i class="fas fa-circle-exclamation text-teal-500 mr-2"></i>Issue Opened</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">10</td>
                    <td class="py-3">Per issue created</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-3"><i class="fas fa-circle-check text-green-500 mr-2"></i>Issue Closed</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">20</td>
                    <td class="py-3">Per issue resolved/closed</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-3"><i class="fas fa-comment-dots text-blue-500 mr-2"></i>Issue Comment</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">5</td>
                    <td class="py-3">Per comment on issues</td>
                  </tr>
                  <tr>
                    <td class="py-3"><i class="fas fa-link text-accent-500 mr-2"></i>Issue Reference</td>
                    <td class="py-3 font-mono text-primary-600 dark:text-primary-400">5</td>
                    <td class="py-3">Per commit referencing an issue (#123)</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </Card>

          <!-- Meaningful Lines -->
          <Card class="shadow-lg">
            <h3 class="font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center text-xl">
              <i class="fas fa-filter mr-2 text-green-500"></i>
              Meaningful Lines
            </h3>
            <p class="text-gray-600 dark:text-gray-400 mb-4">
              By default, Git Velocity uses <strong>meaningful lines</strong> instead of raw line counts.
              This filters out noise and rewards actual code contributions:
            </p>
            <div class="grid sm:grid-cols-2 gap-4">
              <div class="p-4 bg-green-50 dark:bg-green-900/20 rounded-lg">
                <h4 class="font-medium text-green-700 dark:text-green-400 mb-2">
                  <i class="fas fa-check mr-2"></i>Counted as Meaningful
                </h4>
                <ul class="text-sm text-gray-600 dark:text-gray-400 space-y-1">
                  <li>Actual code logic</li>
                  <li>Function definitions</li>
                  <li>Variable declarations</li>
                  <li>Import statements</li>
                </ul>
              </div>
              <div class="p-4 bg-red-50 dark:bg-red-900/20 rounded-lg">
                <h4 class="font-medium text-red-700 dark:text-red-400 mb-2">
                  <i class="fas fa-times mr-2"></i>Filtered Out
                </h4>
                <ul class="text-sm text-gray-600 dark:text-gray-400 space-y-1">
                  <li>Empty lines / whitespace</li>
                  <li>Single-line comments</li>
                  <li>Multi-line comment blocks</li>
                  <li>Documentation strings</li>
                </ul>
              </div>
            </div>
            <p class="text-sm text-gray-600 dark:text-gray-400 mt-4">
              <i class="fas fa-info-circle mr-1"></i>
              Meaningful lines filtering is always enabled to accurately reflect code contributions.
            </p>
          </Card>
        </div>
      </div>
    </section>

    <!-- Leaderboard Section -->
    <section id="leaderboard-info" class="py-8 px-4">
      <div class="container mx-auto">
        <SectionHeader title="Leaderboard Rankings" icon="fas fa-list-ol" icon-color="text-accent-500" />

        <div class="space-y-6">
          <!-- Ranking Process -->
          <Card class="shadow-lg">
            <h3 class="font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center text-xl">
              <i class="fas fa-list-ol mr-2 text-accent-500"></i>
              Ranking Process
            </h3>
            <ol class="space-y-4">
              <li class="flex items-start gap-3">
                <span class="flex-shrink-0 w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center text-primary-600 dark:text-primary-400 font-bold">1</span>
                <div>
                  <h4 class="font-medium text-gray-900 dark:text-gray-100">Aggregate Across Repos</h4>
                  <p class="text-sm text-gray-600 dark:text-gray-400">Metrics from all configured repositories are combined per contributor</p>
                </div>
              </li>
              <li class="flex items-start gap-3">
                <span class="flex-shrink-0 w-8 h-8 rounded-full bg-accent-100 dark:bg-accent-900/30 flex items-center justify-center text-accent-600 dark:text-accent-400 font-bold">2</span>
                <div>
                  <h4 class="font-medium text-gray-900 dark:text-gray-100">Calculate Total Score</h4>
                  <p class="text-sm text-gray-600 dark:text-gray-400">Apply point values to each activity type and sum the breakdown</p>
                </div>
              </li>
              <li class="flex items-start gap-3">
                <span class="flex-shrink-0 w-8 h-8 rounded-full bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-600 dark:text-indigo-400 font-bold">3</span>
                <div>
                  <h4 class="font-medium text-gray-900 dark:text-gray-100">Sort by Score</h4>
                  <p class="text-sm text-gray-600 dark:text-gray-400">Contributors are sorted in descending order by total score</p>
                </div>
              </li>
              <li class="flex items-start gap-3">
                <span class="flex-shrink-0 w-8 h-8 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center text-blue-600 dark:text-blue-400 font-bold">4</span>
                <div>
                  <h4 class="font-medium text-gray-900 dark:text-gray-100">Assign Ranks & Percentiles</h4>
                  <p class="text-sm text-gray-600 dark:text-gray-400">Each contributor receives a rank (1st, 2nd...) and percentile position</p>
                </div>
              </li>
            </ol>
          </Card>

          <!-- Top Categories -->
          <Card class="shadow-lg">
            <h3 class="font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center text-xl">
              <i class="fas fa-medal mr-2 text-yellow-500"></i>
              Top Achievers
            </h3>
            <p class="text-gray-600 dark:text-gray-400 mb-4">
              Git Velocity tracks top performers in each category:
            </p>
            <div class="grid sm:grid-cols-2 gap-4">
              <div class="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2 mb-2">
                  <i class="fas fa-trophy text-yellow-500"></i>
                  <span class="font-medium text-gray-900 dark:text-gray-100">Overall Leader</span>
                </div>
                <p class="text-sm text-gray-600 dark:text-gray-400">Highest total score</p>
              </div>
              <div class="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2 mb-2">
                  <i class="fas fa-code-commit text-primary-500"></i>
                  <span class="font-medium text-gray-900 dark:text-gray-100">Top Committer</span>
                </div>
                <p class="text-sm text-gray-600 dark:text-gray-400">Most commits</p>
              </div>
              <div class="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2 mb-2">
                  <i class="fas fa-eye text-accent-500"></i>
                  <span class="font-medium text-gray-900 dark:text-gray-100">Top Reviewer</span>
                </div>
                <p class="text-sm text-gray-600 dark:text-gray-400">Most reviews given</p>
              </div>
              <div class="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex items-center gap-2 mb-2">
                  <i class="fas fa-code-pull-request text-indigo-500"></i>
                  <span class="font-medium text-gray-900 dark:text-gray-100">Top PR Author</span>
                </div>
                <p class="text-sm text-gray-600 dark:text-gray-400">Most PRs opened</p>
              </div>
            </div>
          </Card>

          <!-- Team Scoring -->
          <Card class="shadow-lg">
            <h3 class="font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center text-xl">
              <i class="fas fa-users mr-2 text-blue-500"></i>
              Team Scoring
            </h3>
            <p class="text-gray-600 dark:text-gray-400 mb-4">
              When teams are configured, Git Velocity calculates team metrics:
            </p>
            <ul class="space-y-2 text-gray-600 dark:text-gray-400">
              <li><i class="fas fa-check text-green-500 mr-2"></i><strong>Total Team Score:</strong> Sum of all member scores</li>
              <li><i class="fas fa-check text-green-500 mr-2"></i><strong>Average Score:</strong> Total score / number of members</li>
              <li><i class="fas fa-check text-green-500 mr-2"></i><strong>Member Breakdown:</strong> Individual scores and achievements per team member</li>
            </ul>
          </Card>
        </div>
      </div>
    </section>

    <!-- Achievements Section -->
    <section id="achievements" class="py-8 px-4">
      <div class="container mx-auto">
        <SectionHeader title="Achievement System" icon="fas fa-trophy" icon-color="text-yellow-500" />
        <p class="text-gray-600 dark:text-gray-300 mb-8 text-center">115 achievements across 26 categories with tiered progression</p>

        <div class="space-y-6">
          <!-- Achievement Categories -->
          <Card class="shadow-lg">
            <h3 class="font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center text-xl">
              <i class="fas fa-trophy mr-2 text-yellow-500"></i>
              Achievement Categories
            </h3>
            <div class="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
              <!-- Commits -->
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <h4 class="font-medium text-gray-900 dark:text-gray-100 mb-2">
                  <i class="fas fa-code-commit text-primary-500 mr-2"></i>Commits
                </h4>
                <p class="text-xs text-gray-600 dark:text-gray-400 mb-2">Tiers: 1, 10, 50, 100, 500, 1000</p>
                <div class="text-xs text-gray-600 dark:text-gray-400">
                  First Steps, Getting Started, Contributor, Committed, Code Machine, Code Warrior
                </div>
              </div>
              <!-- PRs Opened -->
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <h4 class="font-medium text-gray-900 dark:text-gray-100 mb-2">
                  <i class="fas fa-code-pull-request text-accent-500 mr-2"></i>PRs Opened
                </h4>
                <p class="text-xs text-gray-600 dark:text-gray-400 mb-2">Tiers: 1, 10, 25, 50, 100, 250</p>
                <div class="text-xs text-gray-600 dark:text-gray-400">
                  PR Pioneer, PR Regular, PR Pro, Merge Master, PR Champion, PR Legend
                </div>
              </div>
              <!-- Reviews -->
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <h4 class="font-medium text-gray-900 dark:text-gray-100 mb-2">
                  <i class="fas fa-eye text-indigo-500 mr-2"></i>Reviews Given
                </h4>
                <p class="text-xs text-gray-600 dark:text-gray-400 mb-2">Tiers: 1, 10, 25, 50, 100, 250</p>
                <div class="text-xs text-gray-600 dark:text-gray-400">
                  First Review, Reviewer, Review Regular, Review Expert, Review Guru, Review Master
                </div>
              </div>
              <!-- Review Comments -->
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <h4 class="font-medium text-gray-900 dark:text-gray-100 mb-2">
                  <i class="fas fa-comment text-blue-500 mr-2"></i>Review Comments
                </h4>
                <p class="text-xs text-gray-600 dark:text-gray-400 mb-2">Tiers: 10, 50, 100, 250, 500</p>
                <div class="text-xs text-gray-600 dark:text-gray-400">
                  Commentator, Feedback Giver, Code Critic, Feedback Expert, Comment Champion
                </div>
              </div>
              <!-- Lines Added -->
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <h4 class="font-medium text-gray-900 dark:text-gray-100 mb-2">
                  <i class="fas fa-plus text-green-500 mr-2"></i>Lines Added
                </h4>
                <p class="text-xs text-gray-600 dark:text-gray-400 mb-2">Tiers: 100, 1K, 5K, 10K, 50K</p>
                <div class="text-xs text-gray-600 dark:text-gray-400">
                  First Hundred, Thousand Lines, Five Thousand, Ten Thousand, Code Mountain
                </div>
              </div>
              <!-- Lines Deleted -->
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <h4 class="font-medium text-gray-900 dark:text-gray-100 mb-2">
                  <i class="fas fa-minus text-red-500 mr-2"></i>Lines Deleted
                </h4>
                <p class="text-xs text-gray-600 dark:text-gray-400 mb-2">Tiers: 100, 500, 1K, 5K, 10K</p>
                <div class="text-xs text-gray-600 dark:text-gray-400">
                  Tidying Up, Spring Cleaning, Code Cleaner, Refactoring Hero, Deletion Master
                </div>
              </div>
              <!-- Response Time -->
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <h4 class="font-medium text-gray-900 dark:text-gray-100 mb-2">
                  <i class="fas fa-bolt text-yellow-500 mr-2"></i>Response Time
                </h4>
                <p class="text-xs text-gray-600 dark:text-gray-400 mb-2">Tiers: &lt;24h, &lt;4h, &lt;1h</p>
                <div class="text-xs text-gray-600 dark:text-gray-400">
                  Same Day Reviewer, Quick Responder, Speed Demon
                </div>
              </div>
              <!-- Streaks -->
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <h4 class="font-medium text-gray-900 dark:text-gray-100 mb-2">
                  <i class="fas fa-fire text-orange-500 mr-2"></i>Contribution Streaks
                </h4>
                <p class="text-xs text-gray-600 dark:text-gray-400 mb-2">Tiers: 3, 7, 14, 30 days</p>
                <div class="text-xs text-gray-600 dark:text-gray-400">
                  Getting Rolling, Week Warrior, Two Week Streak, Month Master
                </div>
              </div>
              <!-- Activity Patterns -->
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <h4 class="font-medium text-gray-900 dark:text-gray-100 mb-2">
                  <i class="fas fa-clock text-cyan-500 mr-2"></i>Activity Patterns
                </h4>
                <p class="text-xs text-gray-600 dark:text-gray-400 mb-2">Early Bird, Night Owl, Weekend Warrior</p>
                <div class="text-xs text-gray-600 dark:text-gray-400">
                  Commits at different times of day unlock special badges
                </div>
              </div>
              <!-- Issues Opened -->
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <h4 class="font-medium text-gray-900 dark:text-gray-100 mb-2">
                  <i class="fas fa-circle-exclamation text-teal-500 mr-2"></i>Issues Opened
                </h4>
                <p class="text-xs text-gray-600 dark:text-gray-400 mb-2">Tiers: 1, 5, 10, 25, 50</p>
                <div class="text-xs text-gray-600 dark:text-gray-400">
                  Issue Opener, Reporter, Bug Hunter, Issue Tracker, Issue Master
                </div>
              </div>
              <!-- Issues Closed -->
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <h4 class="font-medium text-gray-900 dark:text-gray-100 mb-2">
                  <i class="fas fa-circle-check text-green-500 mr-2"></i>Issues Closed
                </h4>
                <p class="text-xs text-gray-600 dark:text-gray-400 mb-2">Tiers: 1, 5, 10, 25, 50</p>
                <div class="text-xs text-gray-600 dark:text-gray-400">
                  Issue Closer, Problem Solver, Resolver, Issue Crusher, Closure King
                </div>
              </div>
              <!-- Issue Comments -->
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <h4 class="font-medium text-gray-900 dark:text-gray-100 mb-2">
                  <i class="fas fa-comment-dots text-blue-500 mr-2"></i>Issue Comments
                </h4>
                <p class="text-xs text-gray-600 dark:text-gray-400 mb-2">Tiers: 5, 10, 25, 50, 100</p>
                <div class="text-xs text-gray-600 dark:text-gray-400">
                  Issue Commenter, Discussion Starter, Feedback Provider, Issue Conversationalist, Discussion Champion
                </div>
              </div>
            </div>
          </Card>

          <!-- Achievement Conditions -->
          <Card class="shadow-lg">
            <h3 class="font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center text-xl">
              <i class="fas fa-unlock mr-2 text-green-500"></i>
              How Achievements Are Earned
            </h3>
            <p class="text-gray-600 dark:text-gray-400 mb-4">
              Each achievement has a <strong>condition type</strong> and <strong>threshold</strong>.
              When your metrics meet or exceed the threshold, the achievement is unlocked.
            </p>
            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="border-b border-gray-200 dark:border-gray-700">
                    <th class="text-left py-2 text-gray-700 dark:text-gray-400">Condition Type</th>
                    <th class="text-left py-2 text-gray-700 dark:text-gray-400">Metric Checked</th>
                    <th class="text-left py-2 text-gray-700 dark:text-gray-400">Comparison</th>
                  </tr>
                </thead>
                <tbody class="text-gray-700 dark:text-gray-300">
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-2 font-mono text-xs">commit_count</td>
                    <td class="py-2">Total commits</td>
                    <td class="py-2">&ge; threshold</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-2 font-mono text-xs">pr_opened_count</td>
                    <td class="py-2">PRs opened</td>
                    <td class="py-2">&ge; threshold</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-2 font-mono text-xs">review_count</td>
                    <td class="py-2">Reviews given</td>
                    <td class="py-2">&ge; threshold</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-2 font-mono text-xs">avg_review_time_hours</td>
                    <td class="py-2">Average review response</td>
                    <td class="py-2">&le; threshold (lower is better)</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-2 font-mono text-xs">longest_streak</td>
                    <td class="py-2">Consecutive active days</td>
                    <td class="py-2">&ge; threshold</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-2 font-mono text-xs">perfect_prs</td>
                    <td class="py-2">PRs with no changes requested</td>
                    <td class="py-2">&ge; threshold</td>
                  </tr>
                  <tr class="border-b border-gray-100 dark:border-gray-800">
                    <td class="py-2 font-mono text-xs">issues_opened</td>
                    <td class="py-2">Issues created</td>
                    <td class="py-2">&ge; threshold</td>
                  </tr>
                  <tr>
                    <td class="py-2 font-mono text-xs">issues_closed</td>
                    <td class="py-2">Issues resolved/closed</td>
                    <td class="py-2">&ge; threshold</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p class="text-sm text-gray-600 dark:text-gray-400 mt-4">
              <i class="fas fa-shield-halved mr-1"></i>
              Achievement definitions are hardcoded and cannot be customized to prevent manipulation.
            </p>
          </Card>

          <!-- Tiered Progression -->
          <Card class="shadow-lg">
            <h3 class="font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center text-xl">
              <i class="fas fa-layer-group mr-2 text-accent-500"></i>
              Tiered Progression
            </h3>
            <p class="text-gray-600 dark:text-gray-400 mb-4">
              Most achievements have multiple tiers. As you progress, you unlock higher tiers:
            </p>
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-2 sm:gap-3 mb-4">
              <div class="flex items-center gap-2 p-2 sm:p-3 bg-gray-100 dark:bg-gray-800 rounded-lg">
                <span class="w-6 h-6 sm:w-8 sm:h-8 flex items-center justify-center rounded-full bg-gray-400 text-white text-xs sm:text-sm font-bold">1</span>
                <div class="text-xs sm:text-sm"><span class="font-medium text-gray-700 dark:text-gray-300">Tier 1</span></div>
              </div>
              <div class="flex items-center gap-2 p-2 sm:p-3 bg-gray-100 dark:bg-gray-800 rounded-lg">
                <span class="w-6 h-6 sm:w-8 sm:h-8 flex items-center justify-center rounded-full bg-gray-500 text-white text-xs sm:text-sm font-bold">10</span>
                <div class="text-xs sm:text-sm"><span class="font-medium text-gray-700 dark:text-gray-300">Tier 2</span></div>
              </div>
              <div class="flex items-center gap-2 p-2 sm:p-3 bg-green-50 dark:bg-green-900/20 rounded-lg">
                <span class="w-6 h-6 sm:w-8 sm:h-8 flex items-center justify-center rounded-full bg-green-500 text-white text-xs sm:text-sm font-bold">25</span>
                <div class="text-xs sm:text-sm"><span class="font-medium text-green-700 dark:text-green-400">Tier 3</span></div>
              </div>
              <div class="flex items-center gap-2 p-2 sm:p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
                <span class="w-6 h-6 sm:w-8 sm:h-8 flex items-center justify-center rounded-full bg-blue-500 text-white text-xs sm:text-sm font-bold">50</span>
                <div class="text-xs sm:text-sm"><span class="font-medium text-blue-700 dark:text-blue-400">Tier 4</span></div>
              </div>
              <div class="flex items-center gap-2 p-2 sm:p-3 bg-purple-50 dark:bg-purple-900/20 rounded-lg">
                <span class="w-6 h-6 sm:w-8 sm:h-8 flex items-center justify-center rounded-full bg-purple-500 text-white text-xs sm:text-sm font-bold">100</span>
                <div class="text-xs sm:text-sm"><span class="font-medium text-purple-700 dark:text-purple-400">Tier 5</span></div>
              </div>
              <div class="flex items-center gap-2 p-2 sm:p-3 bg-primary-50 dark:bg-primary-900/20 rounded-lg">
                <span class="w-6 h-6 sm:w-8 sm:h-8 flex items-center justify-center rounded-full bg-primary-500 text-white text-xs sm:text-sm font-bold">250</span>
                <div class="text-xs sm:text-sm"><span class="font-medium text-primary-700 dark:text-primary-400">Tier 6</span></div>
              </div>
              <div class="flex items-center gap-2 p-2 sm:p-3 bg-orange-50 dark:bg-orange-900/20 rounded-lg">
                <span class="w-6 h-6 sm:w-8 sm:h-8 flex items-center justify-center rounded-full bg-orange-500 text-white text-xs sm:text-sm font-bold">500</span>
                <div class="text-xs sm:text-sm"><span class="font-medium text-orange-700 dark:text-orange-400">Tier 7</span></div>
              </div>
              <div class="flex items-center gap-2 p-2 sm:p-3 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg">
                <span class="w-6 h-6 sm:w-8 sm:h-8 flex items-center justify-center rounded-full bg-gradient-to-r from-yellow-500 to-amber-500 text-white text-xs sm:text-sm font-bold">1k+</span>
                <div class="text-xs sm:text-sm"><span class="font-medium text-yellow-700 dark:text-yellow-400">Tier 8+</span></div>
              </div>
            </div>
            <p class="text-sm text-gray-600 dark:text-gray-400">
              The leaderboard shows only the highest tier achieved per category for each contributor.
            </p>
          </Card>
        </div>
      </div>
    </section>

    <!-- Data Sources Section -->
    <section id="data-sources" class="py-8 px-4">
      <div class="container mx-auto">
        <SectionHeader title="Data Sources" icon="fab fa-github" icon-color="text-gray-700 dark:text-gray-300" />

        <div class="space-y-6">
          <Card class="shadow-lg">
            <h3 class="font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center text-xl">
              <i class="fab fa-github mr-2 text-gray-700 dark:text-gray-300"></i>
              GitHub API Data
            </h3>
            <div class="grid sm:grid-cols-2 gap-6">
              <div>
                <h4 class="font-medium text-gray-800 dark:text-gray-200 mb-3">Commits</h4>
                <ul class="text-sm text-gray-600 dark:text-gray-400 space-y-1">
                  <li><i class="fas fa-check text-green-500 mr-2"></i>SHA, message, timestamp</li>
                  <li><i class="fas fa-check text-green-500 mr-2"></i>Author (login, name, email)</li>
                  <li><i class="fas fa-check text-green-500 mr-2"></i>Additions, deletions, files changed</li>
                  <li><i class="fas fa-check text-green-500 mr-2"></i>Patch/diff for line analysis</li>
                </ul>
              </div>
              <div>
                <h4 class="font-medium text-gray-800 dark:text-gray-200 mb-3">Pull Requests</h4>
                <ul class="text-sm text-gray-600 dark:text-gray-400 space-y-1">
                  <li><i class="fas fa-check text-green-500 mr-2"></i>State (open, merged, closed)</li>
                  <li><i class="fas fa-check text-green-500 mr-2"></i>Author and timestamps</li>
                  <li><i class="fas fa-check text-green-500 mr-2"></i>Size (additions, deletions)</li>
                  <li><i class="fas fa-check text-green-500 mr-2"></i>Comments count</li>
                </ul>
              </div>
              <div>
                <h4 class="font-medium text-gray-800 dark:text-gray-200 mb-3">Reviews</h4>
                <ul class="text-sm text-gray-600 dark:text-gray-400 space-y-1">
                  <li><i class="fas fa-check text-green-500 mr-2"></i>Review state (approved, changes requested)</li>
                  <li><i class="fas fa-check text-green-500 mr-2"></i>Reviewer login</li>
                  <li><i class="fas fa-check text-green-500 mr-2"></i>Submission timestamp</li>
                  <li><i class="fas fa-check text-green-500 mr-2"></i>Comment count</li>
                </ul>
              </div>
              <div>
                <h4 class="font-medium text-gray-800 dark:text-gray-200 mb-3">User Profiles</h4>
                <ul class="text-sm text-gray-600 dark:text-gray-400 space-y-1">
                  <li><i class="fas fa-check text-green-500 mr-2"></i>GitHub login (username)</li>
                  <li><i class="fas fa-check text-green-500 mr-2"></i>Display name</li>
                  <li><i class="fas fa-check text-green-500 mr-2"></i>Avatar URL</li>
                  <li><i class="fas fa-check text-green-500 mr-2"></i>Public email (for deduplication)</li>
                </ul>
              </div>
            </div>
          </Card>

          <!-- Calculated Metrics -->
          <Card class="shadow-lg">
            <h3 class="font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center text-xl">
              <i class="fas fa-calculator mr-2 text-blue-500"></i>
              Derived Metrics
            </h3>
            <p class="text-gray-600 dark:text-gray-400 mb-4">
              These metrics are calculated from raw data:
            </p>
            <div class="grid sm:grid-cols-2 gap-4 text-sm">
              <div class="p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <strong class="text-gray-900 dark:text-gray-100">Meaningful Lines</strong>
                <p class="text-gray-600 dark:text-gray-400">Parsed from commit diffs, filtering comments/whitespace</p>
              </div>
              <div class="p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <strong class="text-gray-900 dark:text-gray-100">Average Review Time</strong>
                <p class="text-gray-600 dark:text-gray-400">Time between PR creation and first review</p>
              </div>
              <div class="p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <strong class="text-gray-900 dark:text-gray-100">Contribution Streaks</strong>
                <p class="text-gray-600 dark:text-gray-400">Consecutive days with activity</p>
              </div>
              <div class="p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <strong class="text-gray-900 dark:text-gray-100">Perfect PRs</strong>
                <p class="text-gray-600 dark:text-gray-400">PRs merged without "changes requested" reviews</p>
              </div>
              <div class="p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <strong class="text-gray-900 dark:text-gray-100">Out of Hours</strong>
                <p class="text-gray-600 dark:text-gray-400">Commits outside 9am-5pm based on commit timestamp</p>
              </div>
              <div class="p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <strong class="text-gray-900 dark:text-gray-100">Issue References</strong>
                <p class="text-gray-600 dark:text-gray-400">Commits containing #123 patterns (fixes, closes, resolves, refs)</p>
              </div>
            </div>
          </Card>

          <!-- Bot Filtering -->
          <Card class="shadow-lg">
            <h3 class="font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center text-xl">
              <i class="fas fa-robot mr-2 text-red-500"></i>
              Bot Filtering
            </h3>
            <p class="text-gray-600 dark:text-gray-400 mb-4">
              By default, bot activity is excluded from metrics. The following patterns are automatically filtered:
            </p>
            <div class="flex flex-wrap gap-2">
              <code class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-sm">*[bot]</code>
              <code class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-sm">dependabot*</code>
              <code class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-sm">renovate*</code>
              <code class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-sm">github-actions*</code>
              <code class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-sm">codecov*</code>
              <code class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-sm">snyk*</code>
              <code class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-sm">greenkeeper*</code>
              <code class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-sm">imgbot*</code>
              <code class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-sm">allcontributors*</code>
              <code class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-sm">semantic-release*</code>
            </div>
            <p class="text-sm text-gray-600 dark:text-gray-400 mt-4">
              <i class="fas fa-cog mr-1"></i>
              Enable with <code class="text-primary-600 dark:text-primary-400">include_bots: true</code> or add custom patterns with <code class="text-primary-600 dark:text-primary-400">additional_bot_patterns</code>.
            </p>
          </Card>
        </div>
      </div>
    </section>
  </div>
</template>
