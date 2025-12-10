<p align="center">
  <img src="docs/git-velocity-logo.png" alt="Git Velocity Logo" width="200"/>
</p>

<h1 align="center">Git Velocity</h1>

<p align="center">
  <strong>Turn your GitHub activity into a game. Track velocity. Earn achievements. Win at development.</strong>
</p>

<p align="center">
  <a href="https://github.com/lukaszraczylo/git-velocity/releases"><img src="https://img.shields.io/github/v/release/lukaszraczylo/git-velocity?style=flat-square&color=ff69b4" alt="Release"></a>
  <a href="https://github.com/lukaszraczylo/git-velocity/blob/main/LICENSE"><img src="https://img.shields.io/github/license/lukaszraczylo/git-velocity?style=flat-square&color=a855f7" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/lukaszraczylo/git-velocity"><img src="https://goreportcard.com/badge/github.com/lukaszraczylo/git-velocity?style=flat-square" alt="Go Report Card"></a>
  <a href="https://github.com/lukaszraczylo/git-velocity/actions"><img src="https://img.shields.io/github/actions/workflow/status/lukaszraczylo/git-velocity/release.yml?style=flat-square&label=release" alt="Release Status"></a>
</p>

<p align="center">
  <a href="#-features">Features</a> •
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-github-action">GitHub Action</a> •
  <a href="#-configuration">Configuration</a> •
  <a href="#-achievements">Achievements</a>
</p>

---

## What is Git Velocity?

Git Velocity analyzes your GitHub repositories and generates a **beautiful, gamified dashboard** showing developer velocity metrics. It's like Spotify Wrapped, but for your code contributions.

```bash
$ git-velocity analyze --config .git-velocity.yaml
🚀 Fetching data from GitHub...
📊 Processing 3 repositories...
🏆 Calculating scores and achievements...
✨ Generated dashboard at ./dist

$ git-velocity serve --port 8080
🌐 Starting preview server at http://localhost:8080
```

## ✨ Features

### 📊 Comprehensive Metrics
- **Commits**: Count, lines added/deleted, files changed
- **Pull Requests**: Opened, merged, closed, average size, time to merge
- **Code Reviews**: Reviews given, comments, approvals, response time
- **Issues**: Opened, closed, comments

### 🎮 Gamification Engine
- **Scoring System**: Earn points for every contribution
- **34 Achievements**: From "First Steps" to "Code Warrior"
- **Leaderboards**: Compete with your team
- **Tier Progression**: Bronze → Silver → Gold → Diamond

### 👥 Team Analytics
- Configure teams and see aggregated metrics
- Team leaderboards and comparisons
- Member contribution breakdowns

### ⚡ Performance Optimized
- **Local Git Analysis**: Clone repos locally for 10x faster commit analysis
- **Smart Caching**: File-based caching with configurable TTL
- **Concurrent Requests**: Parallel API calls for faster data fetching
- **Bot Filtering**: Automatically excludes Dependabot, Renovate, and other bots

### 🎨 Beautiful Dashboard
- Modern Vue.js SPA with dark/light mode
- Responsive design for desktop and mobile
- Interactive charts and visualizations
- GitHub Pages deployment ready

### 🔐 Flexible Authentication
- Personal Access Token (PAT)
- GitHub App authentication
- Environment variable support

## 🚀 Quick Start

### Installation

```bash
# Go install
go install github.com/lukaszraczylo/git-velocity/cmd/git-velocity@latest

# Or download binary from releases
# https://github.com/lukaszraczylo/git-velocity/releases
```

### Create Configuration

Create `.git-velocity.yaml` in your repository:

```yaml
version: "1.0"

auth:
  github_token: "${GITHUB_TOKEN}"

repositories:
  - owner: "your-org"
    name: "your-repo"
  # Or use patterns:
  - owner: "your-org"
    pattern: "*"  # All repos in org

teams:
  - name: "Backend Team"
    members: ["dev1", "dev2", "dev3"]
    color: "#3B82F6"
  - name: "Frontend Team"
    members: ["dev4", "dev5"]
    color: "#10B981"

scoring:
  enabled: true
  points:
    commit: 10
    commit_with_tests: 15
    pr_opened: 25
    pr_merged: 50
    pr_reviewed: 30
    fast_review_1h: 50
    fast_review_4h: 25

output:
  directory: "./dist"
```

### Run Analysis

```bash
# Set your GitHub token
export GITHUB_TOKEN=ghp_your_token_here

# Run analysis
git-velocity analyze --config .git-velocity.yaml --verbose

# Preview the dashboard
git-velocity serve --port 8080
```

## 🤖 GitHub Action

Automate your velocity reports with our GitHub Action:

```yaml
name: Git Velocity Report

on:
  schedule:
    - cron: '0 0 * * 1'  # Weekly on Monday
  workflow_dispatch:      # Manual trigger

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Git Velocity Analysis
        uses: lukaszraczylo/git-velocity@v1
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          config_file: '.git-velocity.yaml'
          output_dir: './velocity-report'

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: velocity-dashboard
          path: ./velocity-report

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v4
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./velocity-report
```

### Action Inputs

| Input | Description | Default |
|-------|-------------|---------|
| `github_token` | GitHub token for API access | **Required** |
| `config_file` | Path to configuration file | `.git-velocity.yaml` |
| `output_dir` | Output directory for dashboard | `./dist` |
| `verbose` | Enable verbose output | `false` |

### Action Outputs

| Output | Description |
|--------|-------------|
| `output_dir` | Path to the generated dashboard |

> **Note**: The action runs as a Docker container for fast execution. Use separate steps for artifact upload and GitHub Pages deployment as shown in the example above.

## 🏆 Achievements

Git Velocity includes 34 unlockable achievements:

### Commit Achievements
| Achievement | Description | Threshold |
|-------------|-------------|-----------|
| 🍼 First Steps | Made your first commit | 1 commit |
| 🌱 Getting Started | Made 10 commits | 10 commits |
| 🔥 Committed | Made 100 commits | 100 commits |
| 🤖 Code Machine | Made 500 commits | 500 commits |
| 👑 Code Warrior | Made 1000 commits | 1000 commits |

### Pull Request Achievements
| Achievement | Description | Threshold |
|-------------|-------------|-----------|
| 🔀 PR Pioneer | Opened your first PR | 1 PR |
| 🌿 Pull Request Pro | Opened 10 PRs | 10 PRs |
| 🔀 Merge Master | Opened 50 PRs | 50 PRs |

### Review Achievements
| Achievement | Description | Threshold |
|-------------|-------------|-----------|
| 🔍 Code Reviewer | Reviewed your first PR | 1 review |
| 👁️ Review Regular | Reviewed 25 PRs | 25 reviews |
| 🎓 Review Guru | Reviewed 100 PRs | 100 reviews |

### Speed Achievements
| Achievement | Description | Threshold |
|-------------|-------------|-----------|
| ⚡ Speed Demon | Avg review response < 1 hour | < 1h |
| ⏰ Quick Responder | Avg review response < 4 hours | < 4h |

### Activity Pattern Achievements
| Achievement | Description | Threshold |
|-------------|-------------|-----------|
| 📅 Week Warrior | 7 day contribution streak | 7 days |
| 📆 Month Master | 30 day contribution streak | 30 days |
| 🌅 Early Bird | 50 commits before 9am | 50 commits |
| 🌙 Night Owl | 50 commits after 9pm | 50 commits |
| 💀 Nosferatu | 25 commits between midnight-4am | 25 commits |
| 🛋️ Weekend Warrior | 25 weekend commits | 25 commits |

### Code Quality Achievements
| Achievement | Description | Threshold |
|-------------|-------------|-----------|
| 🗜️ Small PR Advocate | 10 PRs under 100 lines | 10 PRs |
| ⚛️ Atomic Commits Hero | 50 PRs under 100 lines | 50 PRs |
| ✅ Clean Code | 5 PRs merged without changes requested | 5 PRs |
| 💎 Flawless | 25 PRs merged without changes requested | 25 PRs |

## ⚙️ Configuration

### Full Configuration Reference

```yaml
version: "1.0"

auth:
  # Option 1: Personal Access Token
  github_token: "${GITHUB_TOKEN}"

  # Option 2: GitHub App
  github_app:
    app_id: 123456
    installation_id: 789012
    private_key_path: "./private-key.pem"

repositories:
  # Single repository
  - owner: "your-org"
    name: "repo-name"
  # All repos in organization
  - owner: "your-org"
    pattern: "*"
  # Pattern matching
  - owner: "your-org"
    pattern: "frontend-*"

date_range:
  start: "2024-01-01"
  end: "2024-12-31"

teams:
  - name: "Backend Team"
    members: ["user1", "user2"]
    color: "#3B82F6"

scoring:
  enabled: true
  points:
    commit: 10
    commit_with_tests: 15
    lines_added: 0.1
    lines_deleted: 0.05
    pr_opened: 25
    pr_merged: 50
    pr_reviewed: 30
    review_comment: 5
    issue_opened: 15
    issue_closed: 20
    fast_review_1h: 50
    fast_review_4h: 25
    fast_review_24h: 10

output:
  directory: "./dist"
  format: ["html", "json"]
  deploy:
    gh_pages: true
    artifact: true

cache:
  enabled: true
  directory: "./.cache"
  ttl: "24h"

options:
  concurrent_requests: 5
  include_bots: false
  bot_patterns:
    - "*[bot]"
    - "dependabot*"
    - "renovate*"
  use_local_git: true
  clone_directory: "./.repos"
  user_aliases:
    - github_login: "username"
      emails: ["work@example.com", "personal@example.com"]
      names: ["Full Name", "nickname"]
```

### User Aliases

Map multiple git emails/names to a single GitHub login:

```yaml
options:
  user_aliases:
    - github_login: "johndoe"
      emails:
        - "john.doe@company.com"
        - "johnd@personal.com"
      names:
        - "John Doe"
        - "JD"
```

### Environment Variables

All configuration values support environment variable expansion:

```yaml
auth:
  github_token: "${GITHUB_TOKEN}"
  github_app:
    private_key: "${GITHUB_APP_PRIVATE_KEY}"
```

## 📖 CLI Commands

### `analyze`

Analyze repositories and generate the dashboard.

```bash
git-velocity analyze [flags]

Flags:
  -c, --config string   Path to configuration file (default "config.yaml")
  -o, --output string   Output directory for generated site (default "./dist")
  -v, --verbose         Enable verbose output
```

### `serve`

Start a local preview server.

```bash
git-velocity serve [flags]

Flags:
  -d, --directory string   Directory to serve (default "./dist")
  -p, --port string        Port to listen on (default "8080")
```

### `version`

Print version information.

```bash
git-velocity version
```

## 🛠️ Development

```bash
# Clone repository
git clone https://github.com/lukaszraczylo/git-velocity.git
cd git-velocity

# Install dependencies
go mod download
cd web && npm install && cd ..

# Build
make build

# Run tests
make test

# Build SPA
make build-spa
```

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/lukaszraczylo">Lukasz Raczylo</a>
</p>

<p align="center">
  <a href="https://github.com/lukaszraczylo/git-velocity">⭐ Star this repo</a> if you find it useful!
</p>
