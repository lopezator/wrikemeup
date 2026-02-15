# WrikeMeUp Setup Guide

## Quick Overview

**WrikeMeUp** automatically syncs hours between GitHub issues and Wrike tasks. No server needed - it runs on GitHub Actions!

### How It Works

1. You mark a GitHub issue as a "Wrike Parent" (using label `wrike-parent`)
2. Bot auto-creates a Wrike task and links it (adds task ID to issue body)
3. Developers log hours in child issues: `Hours: 4.5h`
4. When parent issue closes/updates → Bot aggregates ALL child hours and logs to Wrike
5. Done! ✅

---

## Prerequisites

- [ ] **GitHub Account** (Projects V2 is free for everyone!)
- [ ] **Wrike Account** with API access
- [ ] **GitHub Repository** with Issues enabled
- [ ] Admin access to repository settings

---

## Step 1: Get Wrike API Token

### 1.1 Login to Wrike
- Go to https://www.wrike.com/
- Login to your account

### 1.2 Generate API Token
1. Click your profile picture (top right)
2. Go to **Settings** → **Apps & Integrations**
3. Click **API** tab
4. Click **+ Create new token**
5. Give it a name: `WrikeMeUp Bot`
6. Click **Generate**
7. **Copy the token** (you won't see it again!)

### 1.3 Get Wrike Folder ID (for auto-creating tasks)
1. Open Wrike and navigate to the folder where you want tasks created
2. Look at the URL: `https://app-eu.wrike.com/workspace/#/folder/IEABC123DEFG`
3. The folder ID is `IEABC123DEFG` (everything after `/folder/`)
4. **Save this ID** for later

---

## Step 2: Setup GitHub Repository

### 2.1 Fork/Clone This Repository
1. Fork https://github.com/lopezator/wrikemeup
2. Or use it directly in your project repository

### 2.2 Configure GitHub Secrets

Go to your repository → **Settings** → **Secrets and variables** → **Actions**

#### Create Secret: `USERS`
This maps GitHub users to Wrike accounts.

1. Create a JSON array:
```json
[
  {
    "github_username": "yourGitHubUsername",
    "wrike_email": "your@email.com",
    "wrike_token": "YOUR_WRIKE_TOKEN_FROM_STEP_1.2"
  }
]
```

2. Base64 encode it:
```bash
# On macOS/Linux:
echo '[{"github_username":"yourGitHubUsername","wrike_email":"your@email.com","wrike_token":"YOUR_TOKEN"}]' | base64

# On Windows PowerShell:
[Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes('[{"github_username":"yourGitHubUsername","wrike_email":"your@email.com","wrike_token":"YOUR_TOKEN"}]'))
```

3. Create secret named `USERS` with the base64 value

#### Create Secret: `BOT_TOKEN`
This allows the bot to comment on issues.

1. Go to https://github.com/settings/tokens
2. Click **Generate new token** → **Generate new token (classic)**
3. Name: `WrikeMeUp Bot`
4. Select scopes:
   - ✅ `repo` (Full control of private repositories)
   - ✅ `workflow` (Update GitHub Actions workflows)
5. Click **Generate token**
6. **Copy the token**
7. Create secret named `BOT_TOKEN` with this value

#### Create Secret: `WRIKE_FOLDER_ID` (Optional)
Use the folder ID from Step 1.3

1. Create secret named `WRIKE_FOLDER_ID`
2. Value: `IEABC123DEFG` (your folder ID)

#### Create Secret: `GITHUB_PROJECT_NUMBER` (Optional)
Only needed if using GitHub Projects V2.

1. Go to your Projects tab
2. Open your project
3. Look at URL: `https://github.com/users/USERNAME/projects/5`
4. The project number is `5`
5. Create secret named `GITHUB_PROJECT_NUMBER` with value `5`

---

## Step 3: GitHub Workflow Setup

The workflow file `.github/workflows/wrikemeup.yaml` is already configured! It will:

- ✅ Trigger when issue gets `wrike-parent` label
- ✅ Trigger when issue is edited/closed
- ✅ Trigger on bot commands in comments
- ✅ Run on GitHub Actions (no server needed!)

**No action needed** - it's ready to go!

---

## Step 4: Using WrikeMeUp

### Method 1: Simple Issue-Based Workflow (Recommended)

#### For Parent Issues (Epics/Features):

1. **Create parent issue** in GitHub
   ```markdown
   # Feature: User Authentication
   
   Implement user login and registration system.
   
   ## Child Tasks
   Will be created as separate issues
   ```

2. **Add label** `wrike-parent` to the issue
   - Bot automatically creates Wrike task
   - Bot updates issue body with: `Wrike Task ID: IEABC123`
   - ✅ No copy-paste needed!

3. **Create child issues** (sub-tasks)
   ```markdown
   # Task: Implement Login Form
   
   Parent: #123 (reference parent issue)
   Hours: 4h
   
   Create login form with email/password fields.
   ```

4. **Work on tasks**
   - Developers log hours in child issues: `Hours: 2.5h`
   - Update hours anytime by editing issue body

5. **Sync hours to Wrike (choose one):**
   - **Option A:** Comment `@wrikemeup sync` to sync anytime
   - **Option B:** Edit the parent issue → Auto-syncs
   - **Option C:** Close parent issue → Auto-syncs
   - Bot automatically:
     - Finds ALL child issues (that reference parent)
     - Aggregates hours: 4h + 2.5h + 3h = 9.5h
     - Logs 9.5h to Wrike task
   - ✅ Done!

**💡 Tip:** Use `@wrikemeup sync` to log partial work without closing the parent issue!

#### Example Hierarchy:

```
📋 Destinations Feature (#100) [wrike-parent]
   Wrike Task ID: IEABC123
   Hours: 0h (parent has no direct hours)
   │
   ├── 📝 Task A: API Integration (#101)
   │   Parent: #100
   │   Hours: 1h
   │
   ├── 📝 Task B: UI Components (#102)
   │   Parent: #100
   │   Hours: 3.5h
   │
   └── 📝 Task C: Testing (#103)
       Parent: #100
       Hours: 5h

Result in Wrike:
✅ "Destinations Feature" task = 9.5h logged
```

### Method 2: Bot Commands (Alternative)

Use `@wrikemeup` commands in issue comments:

#### Sync hours (without closing issue):
```
@wrikemeup sync
```
**💡 Use this for partial work logging!**

#### Link existing Wrike task:
```
@wrikemeup link IEABC123
```

#### Log hours manually:
```
@wrikemeup loghours IEABC123 4.5h
```

#### Get time logs:
```
@wrikemeup log IEABC123
```

### Method 3: GitHub Projects V2 (Recommended for Teams)

If you want to use Projects V2 (free for everyone!), use custom fields:

1. **Create Project** with custom fields:
   - `Wrike Parent` (Single Select: Yes/No)
   - `Hours` (Number)
   - `Wrike Task ID` (Text - auto-filled by bot)

2. **Add issue to project**

3. **Set "Wrike Parent" = Yes**
   - Bot creates Wrike task
   - Bot auto-fills "Wrike Task ID" field

4. **Log hours** in child issue bodies: `Hours: 4.5h`

5. **Sync hours anytime:**
   - Comment `@wrikemeup sync` on parent issue
   - Or edit parent issue → Auto-syncs
   
**💡 Projects V2 gives you a visual board to track all tasks and hours!**

---

## Step 5: Testing Your Setup

### 5.1 Test Auto-Creation

1. Create a test issue:
   ```markdown
   # Test Parent Issue
   
   Testing WrikeMeUp integration
   ```

2. Add label `wrike-parent`

3. Wait 30 seconds

4. Check:
   - ✅ Issue body updated with `Wrike Task ID: xxx`
   - ✅ Bot comment confirming creation
   - ✅ Wrike task created in your folder

### 5.2 Test Hour Aggregation

1. Create child issue:
   ```markdown
   # Test Child Task
   
   Parent: #1 (your test parent issue number)
   Hours: 2.5h
   ```

2. Create another child:
   ```markdown
   # Test Child Task 2
   
   Parent: #1
   Hours: 3h
   ```

3. Close the parent issue (#1)

4. Check:
   - ✅ Bot comment showing aggregated hours
   - ✅ Wrike task shows 5.5h logged

### 5.3 Check GitHub Actions Logs

1. Go to **Actions** tab in your repository
2. Look for workflow runs
3. Click to see logs
4. Debug any issues

---

## Troubleshooting

### Bot Not Responding?

**Check GitHub Actions:**
1. Go to **Actions** tab
2. Look for failed runs (red ❌)
3. Click to see error logs

**Common Issues:**

| Problem | Solution |
|---------|----------|
| "missing USERS environment variable" | Check `USERS` secret is set correctly |
| "missing BOT_TOKEN" | Create GitHub token and add as secret |
| "WRIKE_FOLDER_ID not configured" | Add `WRIKE_FOLDER_ID` secret |
| Wrike API error 401 | Check Wrike token is valid |
| No child issues found | Ensure child issues have `Parent: #123` in body |

### How to Debug

1. **Enable verbose logging:**
   - Check Actions workflow logs
   - Look for "Found X child issues" messages

2. **Test manually:**
   ```bash
   # Set environment variables
   export USERS='base64encodedusers'
   export BOT_TOKEN='ghp_xxxx'
   export GITHUB_REPO='owner/repo'
   export GITHUB_ISSUE_NUMBER='123'
   export GITHUB_ACTION_TYPE='sync-hours'
   export GITHUB_USERNAME='yourname'
   
   # Run locally
   go run cmd/wrikemeup/main.go
   ```

3. **Check issue body format:**
   - Hours must be: `Hours: 4.5h` or `Hours: 4h`
   - Parent reference: `Parent: #123`
   - Wrike task ID: `Wrike Task ID: IEABC123`

---

## Advanced Configuration

### Custom Hour Formats

The bot accepts:
- `Hours: 4h`
- `Hours: 4.5h`
- `Hours: 4.25h`
- Case-insensitive: `hours: 4h`

### Child Issue Detection

Bot finds child issues that have ANY of:
- `Parent: #123`
- `Related to #123`
- `Part of #123`
- Mentioned in tasklist: `- [ ] #123`

### Multiple Users

Add more users to the `USERS` JSON:
```json
[
  {
    "github_username": "alice",
    "wrike_email": "alice@company.com",
    "wrike_token": "token1"
  },
  {
    "github_username": "bob",
    "wrike_email": "bob@company.com",
    "wrike_token": "token2"
  }
]
```

---

## FAQ

### Q: Do I need a server?
**A: No!** GitHub Actions runs the bot automatically. It's serverless.

### Q: How much does it cost?
**A: Free!** GitHub Actions gives 2,000 minutes/month for free.

### Q: Can I use it on existing issues?
**A: Yes!** Just add the `wrike-parent` label to any issue.

### Q: What if I don't want auto-creation?
**A: Don't set `WRIKE_FOLDER_ID`**. Use `@wrikemeup link <task-id>` to link manually.

### Q: Can child issues have their own Wrike tasks?
**A: Yes!** Mark them as `wrike-parent` too. Hours roll up to the ultimate parent.

### Q: Does it work with private repos?
**A: Yes!** Just ensure the `BOT_TOKEN` has `repo` scope.

### Q: Can I customize the Wrike task description?
**A: Currently no**, but you can edit the Wrike task manually after creation.

### Q: What happens if I close then reopen an issue?
**A: It logs hours again**. Be careful not to double-log!

---

## Security Best Practices

1. **Never commit tokens** to git
2. **Use repository secrets** for all sensitive data
3. **Rotate tokens** periodically
4. **Limit token scopes** to minimum required
5. **Audit workflow runs** regularly

---

## Support

- **Issues**: https://github.com/lopezator/wrikemeup/issues
- **Discussions**: Create GitHub Discussion in the repo

---

## Summary Checklist

- [ ] Get Wrike API token
- [ ] Get Wrike Folder ID
- [ ] Create GitHub `USERS` secret (base64)
- [ ] Create GitHub `BOT_TOKEN` secret
- [ ] Create GitHub `WRIKE_FOLDER_ID` secret
- [ ] Create test issue with `wrike-parent` label
- [ ] Verify Wrike task created
- [ ] Create child issues with hours
- [ ] Close parent and verify aggregation
- [ ] 🎉 You're done!

**Estimated setup time: 15-20 minutes**
