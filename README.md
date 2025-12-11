<p align="center">
  <img src="docs/git-velocity-logo.png" alt="Git Velocity Logo" width="400"/>
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
- **Meaningful Lines**: Filter out comments, whitespace, and documentation changes from line counts

### 🎮 Gamification Engine
- **Scoring System**: Earn points for every contribution
- **95 Achievements**: Tiered progression from "First Steps" to "Code Warrior"
- **Leaderboards**: Compete with your team
- **Tier Progression**: Multiple tiers per achievement category
- **Activity Patterns**: Track early bird, night owl, weekend, and out-of-hours commits
- **Streak Tracking**: Daily streaks and work-week streaks (weekends don't break it!)
- **General velocity chart**: Visualize your velocity over time

### 👥 Team Analytics
- Configure teams and see aggregated metrics
- Team leaderboards and comparisons
- Member contribution breakdowns

### ⚡ Performance Optimized
- **Local Git Analysis**: Clone repos locally for 10x faster commit analysis
- **Smart Caching**: File-based caching with configurable TTL
- **Concurrent Requests**: Parallel API calls for faster data fetching
- **Bot Filtering**: Hardcoded patterns automatically exclude common bots (Dependabot, Renovate, GitHub Actions, etc.) with optional custom patterns

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

permissions:
  contents: read
  pages: write
  id-token: write

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Git Velocity Analysis
        uses: lukaszraczylo/git-velocity@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          config_file: '.git-velocity.yaml'
          output_dir: './velocity-report'

      # Fix permissions - Docker container runs as root
      - name: Fix permissions
        run: sudo chown -R $USER:$USER ./velocity-report

      - name: Setup Pages
        uses: actions/configure-pages@v4

      - name: Upload artifact
        uses: actions/upload-pages-artifact@v3
        with:
          path: ./velocity-report

  deploy:
    runs-on: ubuntu-latest
    needs: analyze
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - name: Deploy to GitHub Pages
        id: deployment
        uses: actions/deploy-pages@v4
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

> **Important**: The action runs as a Docker container. Note the following:
> - Your config file **must** include the `auth` section with `github_token: "${GITHUB_TOKEN}"` - the action input does not automatically populate the config
> - You must set the `GITHUB_TOKEN` environment variable on the action step (in addition to the `github_token` input)
> - The "Fix permissions" step is required because the Docker container runs as root, which causes permission errors when uploading artifacts

## 🏆 Achievements

Git Velocity includes **95 hardcoded achievements** across 20 categories with multiple progression tiers. Achievements cannot be modified via configuration to prevent manipulation.

### Achievement Categories

| Category | Tiers | Description |
|----------|-------|-------------|
| **Commits** | 1, 10, 50, 100, 500, 1000 | Track total commits made |
| **PRs Opened** | 1, 10, 25, 50, 100, 250 | Track pull requests created |
| **Reviews** | 1, 10, 25, 50, 100, 250 | Track code reviews performed |
| **Comments** | 10, 50, 100, 250, 500 | Track PR review comments |
| **Lines Added** | 100, 1K, 5K, 10K, 50K | Track code additions |
| **Lines Deleted** | 100, 500, 1K, 5K, 10K | Track code cleanup |
| **Review Time** | 24h, 4h, 1h | Fast review response times |
| **Multi-Repo** | 2, 5, 10 | Contribution across repositories |
| **Unique Reviewees** | 3, 10, 25 | Reviewing different contributors |
| **Large PRs** | 500, 1K, 5K lines | Big changes merged |
| **Small PRs** | 5, 10, 25, 50 | Atomic commits under 100 lines |
| **Perfect PRs** | 1, 5, 10, 25 | Merged without changes requested |
| **Active Days** | 7, 30, 60, 100 | Unique days with activity |
| **Streaks** | 3, 7, 14, 30 days | Consecutive day contributions |
| **Work Week Streak** | 3, 5, 10, 20 days | Weekday streaks (weekends don't break it!) |
| **Early Bird** | 10, 25, 50, 100 | Commits before 9am |
| **Night Owl** | 10, 25, 50, 100 | Commits after 9pm |
| **Midnight** | 5, 10, 25, 50 | Commits between midnight-4am |
| **Weekend** | 5, 10, 25, 50 | Weekend commits |
| **Out of Hours** | 10, 25, 50, 100 | Commits outside 9am-5pm |
| **Documentation** | 100, 500, 1K, 2.5K, 5K | Comment/doc lines added |
| **Comment Cleanup** | 50, 200, 500, 1K, 2.5K | Outdated comments removed |

### Example Achievements

| Achievement | Description |
|-------------|-------------|
| 🍼 First Steps | Made your first commit |
| 👑 Code Warrior | Made 1000 commits |
| ⚡ Speed Demon | Average review response under 1 hour |
| 💎 Flawless | 25 PRs merged without changes requested |
| 🏢 Full Work Week | 5 consecutive weekday streak |
| 🌙 Night Owl | 50 commits after 9pm |
| ♾️ Time Bender | 100 commits outside 9am-5pm |
| 📚 Documentation Hero | Added 1000 lines of comments/docs |
| 🏛️ Code Historian | Added 5000 lines of comments/docs |
| ✂️ Comment Trimmer | Removed 50 outdated comment lines |
| 💀 Dead Code Hunter | Removed 500 outdated comment lines |

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
    use_meaningful_lines: true  # Exclude comments/whitespace from line scoring
    pr_opened: 25
    pr_merged: 50
    pr_reviewed: 30
    review_comment: 5
    issue_opened: 15
    issue_closed: 20
    fast_review_1h: 50
    fast_review_4h: 25
    fast_review_24h: 10
    out_of_hours: 2  # Bonus per commit outside 9am-5pm

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
  # Add custom bot patterns (hardcoded defaults always apply)
  additional_bot_patterns:
    - "my-org-bot"
    - "jenkins*"
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

### Bot Filtering

Bot filtering uses **hardcoded default patterns** that always apply when `include_bots: false`. These cannot be disabled to ensure consistent filtering:

**Default Bot Patterns (always applied):**
- `*[bot]` - GitHub App bots (dependabot[bot], renovate[bot], etc.)
- `dependabot*` - Dependabot variants
- `renovate*` - Renovate bot variants
- `github-actions*` - GitHub Actions
- `codecov*` - Codecov bot
- `snyk*` - Snyk security bot
- `greenkeeper*` - Greenkeeper (legacy)
- `imgbot*` - Image optimization bot
- `allcontributors*` - All Contributors bot
- `semantic-release*` - Semantic release bot

**Add custom patterns** for your organization's bots:

```yaml
options:
  include_bots: false  # When false, hardcoded + additional patterns apply
  additional_bot_patterns:
    - "my-org-bot"     # Exact match
    - "jenkins*"       # Prefix match
    - "*-ci"           # Suffix match
```

### Meaningful Lines Filtering

By default, Git Velocity filters out non-meaningful code changes when scoring line additions and deletions. This provides a more accurate measure of actual code contributions.

**What's filtered out:**
- **Comments**: Single-line (`//`, `#`, `--`), block (`/* */`, `<!-- -->`), docstrings (`"""`, `'''`)
- **Whitespace**: Empty lines, whitespace-only lines
- **Documentation files**: `.md`, `.rst`, `.txt`, `README`, `CHANGELOG`, `LICENSE`, files in `docs/` directories

**Supported comment styles:**
- C-style: `//`, `/* */`, `*` (block continuation)
- Python/Shell: `#`, `"""`, `'''`
- SQL/Lua/Haskell: `--`
- Assembly/Lisp/INI: `;`
- VB: `'`
- HTML/XML: `<!-- -->`

To disable this filtering and score raw line counts:

```yaml
scoring:
  points:
    use_meaningful_lines: false  # Score all lines including comments/whitespace
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
