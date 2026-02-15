# wrikemeup

Never gonna give you up, never gonna let you down... but it will log your hours into Wrike!

## 🚀 Quick Start

**No server needed!** Runs automatically on GitHub Actions.

**Option 1: Using GitHub Projects Custom Field (Recommended)**
1. Set "Wrike Parent" custom field to "Yes" → Bot creates Wrike task
2. Log hours in child issues: `Hours: 4.5h` (or `Hours: 2024-02-15: 4h, 2024-02-16: 3h`)
3. Edit issue → Hours **automatically** sync to Wrike!
4. Done! ✅

**Option 2: Using Label**
1. Add `wrike-parent` label to issue → Bot creates Wrike task
2. Log hours in child issues: `Hours: 4.5h`
3. Edit issue → Hours **automatically** sync to Wrike!
4. Done! ✅

**📖 [Complete Setup Guide →](SETUP_GUIDE.md)**

---

## Features

WrikeMeUp is a GitHub automation bot that seamlessly integrates GitHub issues with Wrike tasks for hour tracking:

- **🔗 Auto-link Issues to Wrike Tasks**: Automatically create and link Wrike tasks from GitHub issues
- **⏱️ Hour Tracking**: Track hours in issue body or GitHub Projects custom fields
- **📅 Daily Hour Breakdown**: Specify hours per date (e.g., `Hours: 2024-02-15: 4h, 2024-02-16: 3h`)
- **🔄 Incremental Sync**: Only logs new hours since last sync - no duplicates!
- **📊 Subtask Aggregation**: Automatically sum hours from child issues into parent Wrike tasks
- **🤖 Multiple Workflows**: Support for both label-based and GitHub Projects custom fields
- **⚡ Automatic Sync**: Hours sync automatically when you edit issues - no manual commands needed!
- **☁️ Serverless**: Runs on GitHub Actions - no infrastructure needed!

## How It Works

**Automatic Hour Sync - No Commands Needed!**
```
Day 1: Add Hours: 4h → Edit issue → Bot logs 4h to Wrike
Day 2: Update to Hours: 8h → Edit issue → Bot logs 4h more (incremental!)
Day 3: Update to Hours: 12h → Edit issue → Bot logs 4h more

Result: Total 12h in Wrike ✅ No duplicates!
```

**Daily Breakdown Support**
```
Hours: 2024-02-15: 4h, 2024-02-16: 3h, 2024-02-17: 5h

Edit issue → Bot logs:
- 4h on Feb 15
- 3h on Feb 16  
- 5h on Feb 17

Perfect for tracking which hours were on which day! ✅
```

**Child Aggregation**
```
├── Destinations Feature (#100) [wrike-parent]
    Wrike Task ID: IEABC123
    Last Synced: 0h
    │
    ├── Task A (#101) - Hours: 1h
    ├── Task B (#102) - Hours: 3.5h
    └── Task C (#103) - Hours: 5h

Edit parent → Bot aggregates 9.5h to Wrike ✅
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

### Simple Workflow (Recommended)

**1. Create parent issue with label `wrike-parent`**
```markdown
# Destinations Feature

Implement new destinations page
```

**2. Bot auto-creates Wrike task**
```markdown
# Destinations Feature

Wrike Task ID: IEABC123 ← Added by bot
Last Synced: 0h ← Tracks synced hours

Implement new destinations page
```

**3. Add hours to issue (choose format):**

**Option A: Simple total hours**
```markdown
Hours: 12h
```

**Option B: Daily breakdown (specific dates)**
```markdown
Hours: 2024-02-16: 3h, 2024-02-18: 5h
```

This logs:
- 3 hours to Feb 16
- 5 hours to Feb 18

**4. Edit and save** → Bot automatically syncs to Wrike! ✅

### GitHub Projects Custom Field Workflow

**Q: Can I use Projects custom fields for daily hours?**
**A:** Projects V2 fields (Number/Text) can only store single values, so daily breakdown must go in the **issue body**.

**How it works:**
1. Create custom fields:
   - "Wrike Parent" (Single Select: Yes/No) 
   - "Wrike Task ID" (Text - auto-filled)
   
2. Add issue to project, set "Wrike Parent" = "Yes"

3. **Add hours in issue body** (not custom fields):
   ```markdown
   Hours: 2024-02-16: 3h, 2024-02-18: 5h
   ```

4. Edit issue → Bot syncs to correct dates! ✅

### Bot Commands (Optional)

```markdown
@wrikemeup link IEABC123         # Link to existing Wrike task
@wrikemeup sync                  # Manual sync (usually not needed)
@wrikemeup loghours IEABC123 4h  # Log hours manually
@wrikemeup log IEABC123          # Get time logs
```

---

## Examples

### Example 1: Weekly Sprint with Daily Breakdown

**Parent Issue (#100):**
```markdown
# Sprint 23 - Authentication Feature

Wrike Task ID: IEABC123
Last Synced: 0h

Hours: 2024-02-16: 3h, 2024-02-18: 5h, 2024-02-19: 4h

## Subtasks
- #101 Login API
- #102 UI Components
- #103 Testing
```

**Result in Wrike:**
- Feb 16: 3h logged
- Feb 18: 5h logged
- Feb 19: 4h logged

**Update hours later:**
```markdown
Hours: 2024-02-16: 3h, 2024-02-18: 5h, 2024-02-19: 4h, 2024-02-20: 2h
```

Edit & save → Bot logs 2h to Feb 20 (incremental!) ✅

### Example 2: Child Issues Aggregation

**Parent (#200):**
```markdown
# Payment Integration

Wrike Task ID: IEWXYZ789
Last Synced: 0h
```

**Child Issues:**
- #201: `Hours: 2024-02-16: 2h`
- #202: `Hours: 2024-02-17: 3h`
- #203: `Hours: 2024-02-18: 1.5h`

**Edit parent issue** → Bot aggregates:
- Feb 16: 2h
- Feb 17: 3h
- Feb 18: 1.5h
- Total: 6.5h to Wrike ✅

### Example 3: Simple Total Hours

**Issue (#300):**
```markdown
# Bug Fix - Database Migration

Wrike Task ID: IEQUICK456
Last Synced: 0h

Hours: 8h
```

Edit & save → Bot logs 8h to today's date ✅

**Later update to 12h:**
```markdown
Hours: 12h
```

Edit & save → Bot logs 4h more (incremental!) ✅

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