# wrikemeup

Never gonna give you up, never gonna let you down... but it will log your hours into Wrike!

## 🚀 Quick Start

**No server needed!** Runs automatically on GitHub Actions.

1. Add `wrike-parent` label to issue → Bot creates Wrike task
2. Log hours in child issues: `Hours: 4.5h`
3. Close parent → Bot aggregates all child hours to Wrike
4. Done! ✅

**📖 [Complete Setup Guide →](SETUP_GUIDE.md)**

---

## Features

WrikeMeUp is a GitHub automation bot that seamlessly integrates GitHub issues with Wrike tasks for hour tracking:

- **🔗 Auto-link Issues to Wrike Tasks**: Automatically create and link Wrike tasks from GitHub issues
- **⏱️ Hour Tracking**: Track hours in issue body or GitHub Projects custom fields
- **📊 Subtask Aggregation**: Automatically sum hours from child issues into parent Wrike tasks
- **🤖 Multiple Workflows**: Support for both issue-based and GitHub Projects custom fields (with GitHub Premium)
- **☁️ Serverless**: Runs on GitHub Actions - no infrastructure needed!

## How It Works

```
GitHub Issue Hierarchy:
├── Destinations Feature (#100) [wrike-parent] ← Bot creates Wrike task
    Wrike Task ID: IEABC123 (auto-added by bot)
    │
    ├── Task A (#101) - Hours: 1h
    ├── Task B (#102) - Hours: 3.5h
    └── Task C (#103) - Hours: 5h

When parent closes → Bot logs 9.5h to Wrike task ✅
```

## Quick Example

**1. Create parent issue:**
```markdown
# Destinations Feature

Implement new destinations page

Label: wrike-parent
```

**2. Bot auto-creates Wrike task and updates issue:**
```markdown
# Destinations Feature

Wrike Task ID: IEABC123 ← Added automatically!

Implement new destinations page
```

**3. Create child tasks:**
```markdown
# Task A

Parent: #100
Hours: 1h
```

**4. Close parent → Bot aggregates 9.5h to Wrike!**

---

## Setup (5 minutes)

### Requirements
- Wrike account with API token
- GitHub repository
- 5 minutes to configure secrets

### Quick Setup

1. **Get Wrike API Token** ([guide](SETUP_GUIDE.md#step-1-get-wrike-api-token))
   
2. **Add GitHub Secrets:**
   - `USERS` - Base64 encoded user mappings
   - `BOT_TOKEN` - GitHub token with repo permissions
   - `WRIKE_FOLDER_ID` - Wrike folder for tasks (optional)

3. **Done!** The GitHub Actions workflow runs automatically.

**📖 [Detailed Setup Instructions →](SETUP_GUIDE.md)**

---

## Usage

### Option 1: Simple Workflow (Recommended)

1. Create issue + add `wrike-parent` label
2. Add hours to child issues: `Hours: 4.5h`
3. Close parent issue
4. Hours auto-sync to Wrike ✅

### Option 2: Bot Commands

```markdown
@wrikemeup link IEABC123         # Link to existing Wrike task
@wrikemeup loghours IEABC123 4h  # Log hours manually
@wrikemeup log IEABC123           # Get time logs
```

### Option 3: GitHub Projects (Premium)

Use custom fields: "Wrike Parent", "Hours", "Wrike Task ID"

---

## Configuration

| Secret | Required | Description |
|--------|----------|-------------|
| `USERS` | ✅ | Base64-encoded JSON of GitHub→Wrike user mappings |
| `BOT_TOKEN` | ✅ | GitHub token with `repo` permissions |
| `WRIKE_FOLDER_ID` | ❌ | Wrike folder ID for auto-creating tasks |
| `GITHUB_PROJECT_NUMBER` | ❌ | GitHub Projects V2 number (Premium feature) |

### USERS Format

```json
[
  {
    "github_username": "yourname",
    "wrike_email": "you@company.com",
    "wrike_token": "yourWrikeAPIToken"
  }
]
```

Base64 encode before adding as secret.

---

## Architecture

```
┌─────────────────┐
│ GitHub Issue    │
│ (labeled/edited)│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ GitHub Actions  │ ← Workflow runs automatically
│ (no server!)    │    (2000 free minutes/month)
└────────┬────────┘
         │
         ├──► Parse hours from issues
         ├──► Find child issues
         ├──► Aggregate hours
         │
         ▼
┌─────────────────┐
│ Wrike API       │ ← Log hours to task
└─────────────────┘
```

---

## Troubleshooting

### Bot not responding?
1. Check **Actions** tab for failed runs
2. Verify secrets are set correctly
3. Check [Troubleshooting Guide](SETUP_GUIDE.md#troubleshooting)

### Common Issues
- "missing USERS" → Check `USERS` secret is base64 encoded
- "missing BOT_TOKEN" → Create GitHub token with `repo` scope
- No child issues found → Ensure child has `Parent: #123` in body

**📖 [Full Troubleshooting Guide →](SETUP_GUIDE.md#troubleshooting)**

---

## Examples

### Sprint Workflow
```
Epic: User Authentication (#100) [wrike-parent]
├── Login UI (#101) - Hours: 4h
├── API Integration (#102) - Hours: 3h
└── Testing (#103) - Hours: 2h

Close #100 → 9h logged to Wrike ✅
```

### Feature Development
```
Feature: Payment System (#200) [wrike-parent]
├── Stripe Integration (#201) - Hours: 8h
├── PayPal Integration (#202) - Hours: 6h
├── UI Components (#203) - Hours: 4h
└── E2E Tests (#204) - Hours: 5h

Close #200 → 23h logged to Wrike ✅
```

---

## FAQ

**Q: Do I need a server?**  
A: No! Runs on GitHub Actions (serverless).

**Q: How much does it cost?**  
A: Free! GitHub Actions includes 2,000 minutes/month.

**Q: Can I use it on existing issues?**  
A: Yes! Just add the `wrike-parent` label.

**Q: What if I don't want auto-creation?**  
A: Don't set `WRIKE_FOLDER_ID`. Use `@wrikemeup link <task-id>`.

**📖 [More FAQs →](SETUP_GUIDE.md#faq)**

---

## Contributing

Contributions welcome! Please read our contributing guidelines.

## License

MIT License - see LICENSE file

---

## Support

- **Setup Guide**: [SETUP_GUIDE.md](SETUP_GUIDE.md)
- **Issues**: https://github.com/lopezator/wrikemeup/issues
- **Discussions**: GitHub Discussions

---

**⭐ Star this repo if it saves you time!**