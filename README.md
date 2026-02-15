# wrikemeup

Never gonna give you up, never gonna let you down... but it will log your hours into Wrike!

## Features

WrikeMeUp is a GitHub automation bot that seamlessly integrates GitHub issues with Wrike tasks for hour tracking:

- **🔗 Auto-link Issues to Wrike Tasks**: Automatically create and link Wrike tasks from GitHub issues
- **⏱️ Hour Tracking**: Track hours in issue body or GitHub Projects custom fields
- **📊 Subtask Aggregation**: Automatically sum hours from child issues into parent Wrike tasks
- **🤖 Multiple Workflows**: Support for both issue-based and GitHub Projects custom fields (with GitHub Premium)

## Setup

### 1. Fork this repository

### 2. Setup the users

Wrike me up! expects two secrets as environment variables. 

- `USERS` in a base64 encoded JSON format. 

    The JSON must be an array of objects, each containing the following keys:
    
    ```json
    [
        {
            "github_username": "rick",
            "wrike_email": "rick@wrikemeup.com",
            "wrike_token": "someLongWrikeToken"
        },
        {
            "github_username": "roll",
            "wrike_email": "roll@wrikemeup.com",
            "wrike_token": "otherLongWrikeToken"
        }
    ]
    ```

### 3. `BOT_TOKEN` 

Wrike me up! bot will answer to your commands sending a comment to the issue.

For this to work, you need to set the `BOT_TOKEN` environment variable with a GitHub token with `repo` permissions.

### 4. `WRIKE_FOLDER_ID` (Optional)

If you want the bot to automatically create Wrike tasks when issues are labeled with `wrike-parent`, set the `WRIKE_FOLDER_ID` secret with your Wrike folder ID where tasks should be created.

### 5. `GITHUB_PROJECT_NUMBER` (Optional - GitHub Premium)

If you're using GitHub Projects V2 with custom fields, set the `GITHUB_PROJECT_NUMBER` secret to enable project-based workflows.

## Usage

### Option 1: Issue Body (Simple)

Add hours and link Wrike tasks directly in issue bodies:

1. **Mark as parent issue**: Add label `wrike-parent` to auto-create a Wrike task
2. **Add hours**: Include `Hours: 4.5h` in the issue body
3. **Reference subtasks**: Use `#123` to reference child issues
4. **Sync**: Hours automatically sync when issue is edited or closed

**Example Issue Body:**
```markdown
## Task Description
Implement user authentication feature

Wrike Task ID: IEABC123
Hours: 8h

## Subtasks
- #101
- #102
```

### Option 2: GitHub Projects Custom Fields (GitHub Premium)

Use structured custom fields in GitHub Projects V2:

1. **Setup Project**: Create custom fields:
   - "Wrike Parent" (Single Select: Yes/No or Checkbox)
   - "Hours" (Number field)
   - "Wrike Task ID" (Text field)

2. **Add issues to project**: Issues are automatically processed when added

3. **Hours aggregate**: Parent issues automatically sum child issue hours

### Option 3: Bot Commands (Legacy)

Use `@wrikemeup` commands in issue comments:

- `@wrikemeup link <task-id>` - Link issue to existing Wrike task
- `@wrikemeup loghours <task-id> 4.5h` - Log hours to Wrike task
- `@wrikemeup log <task-id>` - Retrieve time logs from Wrike

## Workflow Examples

### Example 1: Auto-create and Track Hours

1. Create parent issue with label `wrike-parent`
2. Bot automatically creates Wrike task and links it
3. Add `Hours: 4h` to issue body
4. Create subtasks #101, #102 with `Hours: 2h` each
5. Reference subtasks in parent with `#101` and `#102`
6. Close parent issue → Bot logs 8h total (4h + 2h + 2h) to Wrike

### Example 2: Manual Link and Sync

1. Create Wrike task manually
2. Comment on GitHub issue: `@wrikemeup link IEABC123`
3. Add `Hours: 5.5h` to issue body
4. Edit issue → Hours auto-sync to Wrike

## Configuration Reference

| Secret/Env | Required | Description |
|------------|----------|-------------|
| `USERS` | ✅ | Base64-encoded JSON array of user mappings |
| `BOT_TOKEN` | ✅ | GitHub token with repo permissions |
| `WRIKE_FOLDER_ID` | ❌ | Wrike folder ID for auto-creating tasks |
| `GITHUB_PROJECT_NUMBER` | ❌ | Project number for GitHub Projects V2 integration |

## How It Works

1. **Triggering**: GitHub Actions workflow triggers on:
   - Issue opened/edited/closed
   - Issue labeled (for auto-link)
   - Issue comments (for bot commands)
   - Project item updates (GitHub Projects V2)

2. **Processing**:
   - Parse hours from issue body or custom fields
   - Aggregate hours from referenced subtasks
   - Create Wrike task (if needed)
   - Log hours to Wrike via API

3. **Feedback**: Bot comments on issues to confirm actions