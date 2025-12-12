import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import App from './App.vue'
import './style.css'

// Views
import Dashboard from './views/Dashboard.vue'
import Leaderboard from './views/Leaderboard.vue'
import Repository from './views/Repository.vue'
import Team from './views/Team.vue'
import Contributor from './views/Contributor.vue'
import HowScoringWorks from './views/HowScoringWorks.vue'

const routes = [
  { path: '/', name: 'dashboard', component: Dashboard },
  { path: '/leaderboard', name: 'leaderboard', component: Leaderboard },
  { path: '/how-scoring-works', name: 'how-scoring-works', component: HowScoringWorks },
  { path: '/repos/:owner/:name', name: 'repository', component: Repository },
  { path: '/teams/:slug', name: 'team', component: Team },
  { path: '/contributors/:login', name: 'contributor', component: Contributor },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  }
})

const app = createApp(App)
app.use(router)
app.mount('#app')
