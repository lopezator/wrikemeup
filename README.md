# wrikemeup

Never gonna give you up, never gonna let you down... but it will log your hours into Wrike!

## 🚀 Quick Start

**No server needed!** Runs automatically on GitHub Actions.

**Simple workflow:**
1. Add `wrike-parent` label to issue → Bot creates Wrike task
2. Log hours via **comments** (preserves history!):
   ```markdown
   Hours:
   - 16 = 3h
   - 17 = 4.5h
   - 18 = 2h
   ```
3. Edit issue → Hours **automatically** sync to Wrike!
4. Done! ✅

**Smart date formats:**
- `16 = 3h` - Day only (current month/year)
- `03-16 = 4h` - Month-day (current year)
- `2023-12-25 = 5h` - Full date when needed

**📖 [Complete Setup Guide →](SETUP_GUIDE.md)**

---

## Features

WrikeMeUp is a GitHub automation bot that seamlessly integrates GitHub issues with Wrike tasks for hour tracking:

- **💬 Comment-Based Logging**: Log hours via comments (full traceability!)
- **📊 Auto Summary Tables**: Bot responds with formatted hour summaries
- **✏️ Edit Support**: Update hours (e.g., `16 = 2h` → `16 = 5h`)
- **🗑️ Delete Support**: Remove hour entries by deleting comment lines
- **🔄 Smart Sync**: Automatically detects adds/updates/deletes
- **🛡️ Wrike-Safe**: Handles manual Wrike edits gracefully
- **🗓️ Smart Date Format**: Only specify what's needed (day/month-day/full date)
- **📊 Multiple Entries**: Log multiple days in one comment
- **🔗 Auto-link Issues to Wrike Tasks**: Automatically create and link Wrike tasks from GitHub issues
- **⏱️ Hour Tracking**: Track hours in comments or issue body
- **🔄 Automatic Sync**: Hours sync when you add/edit comments or issues
- **📊 Subtask Aggregation**: Automatically sum hours from child issues into parent Wrike tasks
- **🤖 Multiple Workflows**: Support for both label-based and GitHub Projects custom fields
- **☁️ Serverless**: Runs on GitHub Actions - no infrastructure needed!

## How It Works

**Comment-Based Hour Logging (Recommended)**
```markdown
# Add a comment to your issue:

Hours:
- 16 = 3h
- 17 = 4.5h
- 18 = 2h

Bot automatically logs and responds:
```

| Date | Hours | Status |
|------|-------|--------|
| 2024-02-16 | 3.00h | Added |
| 2024-02-17 | 4.50h | Added |
| 2024-02-18 | 2.00h | Added |

**Total: 9.50h**

**Editing Hours:**
```markdown
# Update your comment (change 3h to 5h):

Hours:
- 16 = 5h   ← Changed from 3h
- 17 = 4.5h
- 18 = 2h

Bot detects the change:
```

| Date | Hours | Status |
|------|-------|--------|
| 2024-02-16 | 5.00h | Updated: 3.00h → 5.00h |
| 2024-02-17 | 4.50h | Unchanged: 4.50h |
| 2024-02-18 | 2.00h | Unchanged: 2.00h |

**Total: 11.50h**

**Deleting Hours:**
```markdown
# Remove a line from your comment:

Hours:
- 16 = 5h
- 18 = 2h   ← Deleted line 17

Bot detects deletion:
```

| Date | Hours | Status |
|------|-------|--------|
| 2024-02-16 | 5.00h | Unchanged: 5.00h |
| 2024-02-17 | - | Deleted: 4.50h |
| 2024-02-18 | 2.00h | Unchanged: 2.00h |

**Total: 7.00h**

**Smart Date Formats**
```markdown
Hours:
- 16 = 3h           # Day only → 2024-02-16
- 03-16 = 4h        # Month-day → 2024-03-16
- 2023-12-25 = 5h   # Full date → 2023-12-25

Multiple entries in one comment! 🎉
```

**Child Aggregation**
```markdown
Parent Issue (#100) [wrike-parent]
├── Wrike Task ID: IEABC123
│
├── Child #101 comment: "Hours:\n- 16 = 2h\n- 17 = 1h"
├── Child #102 comment: "Hours:\n- 16 = 3h"
└── Child #103 comment: "Hours:\n- 18 = 4h"

Edit parent → Bot aggregates:
- Feb 16: 5h (2h + 3h)
- Feb 17: 1h
- Feb 18: 4h
Total: 10h to Wrike ✅
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

### Comment-Based Logging (Recommended)

**Add a comment to log hours:**
```markdown
Hours:
- 16 = 3h
- 17 = 4.5h
- 18 = 2h
```

**Smart date formats:**
- **Day only**: `16 = 3h` → Uses current month/year
- **Month-day**: `03-16 = 4h` → Uses current year
- **Full date**: `2023-12-25 = 5h` → Specific date

**Benefits:**
- ✅ **Full traceability** - All hour logs visible in comments
- ✅ **No editing** - Just add new comments
- ✅ **Multiple entries** - Log many days at once
- ✅ **Smart dates** - Only specify what's needed

### Label-Based Workflow

1. Create issue + add `wrike-parent` label
2. Bot creates Wrike task and updates issue body
3. Log hours via comments (see format above)
4. Edit issue or add comment → Auto-sync to Wrike ✅

### GitHub Projects Custom Field

1. Create custom field "Wrike Parent" (Single Select: Yes/No)
2. Set to "Yes" → Bot creates Wrike task
3. Log hours via comments on issue
4. Auto-syncs to Wrike!

### Bot Commands (Optional)

```markdown
@wrikemeup link IEABC123         # Link to existing Wrike task
@wrikemeup sync                  # Manual sync (usually not needed)
@wrikemeup loghours IEABC123 4h  # Log hours manually
@wrikemeup log IEABC123          # Get time logs
```

---

## Examples

### Example 1: Weekly Sprint with Smart Dates

**Parent Issue (#100):**
```markdown
# Sprint 23 - Authentication Feature

Wrike Task ID: IEABC123

## Subtasks
- #101 Login API
- #102 UI Components
- #103 Testing
```

**Add comment:**
```markdown
Hours:
- 16 = 3h
- 18 = 5h
- 19 = 4h
```

**Result in Wrike:**
- Feb 16, 2024: 3h
- Feb 18, 2024: 5h
- Feb 19, 2024: 4h

**Add another comment later:**
```markdown
Hours:
- 20 = 2h
- 03-01 = 3h  (March 1st)
```

**New entries logged:**
- Feb 20, 2024: 2h
- Mar 1, 2024: 3h

✅ Full history preserved in comments!

### Example 2: Editing and Correcting Hours

**Initial comment:**
```markdown
Hours:
- 16 = 2h
- 17 = 3h
```

**Bot responds:**
| Date | Hours | Status |
|------|-------|--------|
| 2024-02-16 | 2.00h | Added |
| 2024-02-17 | 3.00h | Added |

**Total: 5.00h**

**Realize you worked more on the 16th, edit comment:**
```markdown
Hours:
- 16 = 5h  ← Changed from 2h
- 17 = 3h
```

**Bot responds:**
| Date | Hours | Status |
|------|-------|--------|
| 2024-02-16 | 5.00h | Updated: 2.00h → 5.00h |
| 2024-02-17 | 3.00h | Unchanged: 3.00h |

**Total: 8.00h**

✅ Wrike updated automatically!

### Example 3: Child Issues Aggregation

**Parent (#200):**
```markdown
# Payment Integration

Wrike Task ID: IEWXYZ789
```

**Child #201 - Comment:**
```markdown
Hours:
- 16 = 2h
- 17 = 1.5h
```

**Child #202 - Comment:**
```markdown
Hours:
- 17 = 3h
- 18 = 2h
```

**Edit parent issue** → Bot aggregates from all children:
- Feb 16: 2h
- Feb 17: 4.5h (1.5h + 3h)
- Feb 18: 2h
- **Total: 8.5h to Wrike** ✅

### Example 4: Cross-Month Logging

**Comment:**
```markdown
Hours:
- 28 = 4h        # Feb 28
- 03-01 = 3h     # March 1 (different month)
- 03-02 = 2h     # March 2
- 2023-12-25 = 1h  # Last year's Christmas (full date)
```

**All 4 entries logged to correct dates!** ✅

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