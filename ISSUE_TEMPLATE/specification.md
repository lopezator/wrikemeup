# Wrikemeup Bot - Complete Specification

## Overview

A GitHub Action bot that logs time to Wrike tasks directly from GitHub issue comments, with intelligent parent/child issue aggregation and minimal storage requirements.

User-facing documentation: https://github.com/wrikemeup/wrikemeup

---

## Commands

### Log Time
```
@wrikemeup log <entries>
```

**Entry Format:** `<date> <duration> ["description"]`

**Examples:**
```
@wrikemeup log today 3h "feature work"
@wrikemeup log today 3h, yesterday 2h "code review"
@wrikemeup log last monday 4h, feb 15 5h
```

### Date Formats
- **Relative:** `today`, `yesterday`
- **This/Next Week:** `monday`, `tuesday`, etc.
- **Last Week:** `last monday`, `last tuesday`, etc.
- **Specific:** `2026-02-15`, `feb 15`, `15 feb`

### Duration Formats
- **Hours:** `3h`, `4.5h`
- **Minutes:** `30m`
- **Combined:** `2h30m`

### Delete with 0h
```
@wrikemeup log today 0h
```
Removes this child's entry for that date, recalculates parent total.

### Invalid Commands
Bot replies: `📖 See documentation: https://github.com/wrikemeup/wrikemeup`

---

## Issue Types & Wrike Mapping

### Standalone Issue (no parent, no children)
- **First log:** Creates Wrike task for this issue
- **Storage:** Wrike ID stored in GitHub Projects custom field "Wrike ID"
- **Logging:** Hours go directly to this Wrike task

### Child Issue (has parent)
- **First log on ANY child:** Creates Wrike task for parent (if doesn't exist)
- **Storage:** Wrike ID stored in PARENT'S "Wrike ID" field (not child's)
- **Logging:** All children's logs aggregate to parent's Wrike task
- **Child has no Wrike task** of its own

### Parent Issue (has children)
- **Cannot log directly** on parent issue
- **Only log via children**
- **Wrike task exists** for parent, contains aggregated child logs

### Standalone → Parent Transition
```
Example:
1. Issue #50 standalone: log today 3h, yesterday 2h
   → Wrike #50: Feb 15: 3h, Feb 14: 2h

2. Add child #51 to #50
   → Previous logs REMAIN in Wrike
   → Can no longer log directly on #50

3. Log on #51: today 5h
   → Wrike #50: Feb 15: 5h (replaced), Feb 14: 2h (untouched)
```

**Key:** Only dates mentioned in new logs are updated; old dates remain.

---

## Logging Algorithm: Full Scan with Date Filter

**On every `@wrikemeup log` command:**

1. **Parse Command**
   - Extract dates: `[today, yesterday]` → `[2026-02-15, 2026-02-14]`
   - Extract hours per date

2. **Determine Context**
   - Is this issue standalone? (no parent, no children)
   - Is this issue a child? (has parent)
   - Is this issue a parent? (has children)

3. **Find Target Wrike Task**
   - If standalone: This issue
   - If child or parent: Parent issue

4. **Aggregate for Each Date** (Full Scan)
   - Get all children of target issue
   - For each date in command:
     - Scan ALL children's comment history
     - Find latest `@wrikemeup log` mentioning this date
     - Extract hours for this date from each child
     - Sum hours across all children

5. **Get/Create Wrike Task**
   - Check target issue's "Wrike ID" field
   - If empty: Create Wrike task, store ID
   - If exists: Use existing task

6. **Update Wrike** (Date-Specific)
   - Update ONLY the dates mentioned in command
   - Other dates remain untouched
   - Set each date to calculated total

7. **Reply**
   ```
   ✅ Logged to #42 (Auth Module)
   Feb 15: 8h, Feb 14: 5h
   View in Wrike: [link]
   ```

---

## Aggregation Examples

### Single Child Updates
```
Parent #42, Child #45

#45: log today 2h
→ Wrike #42: Feb 15: 2h

#45: log today 5h
→ Wrike #42: Feb 15: 5h (overwrites 2h)

#45: log yesterday 3h
→ Wrike #42: Feb 15: 5h, Feb 14: 3h (new date added)
```

### Multiple Children
```
Parent #42
├─ Child #45
└─ Child #46

#45: log today 3h
→ Wrike #42: Feb 15: 3h

#46: log today 2h
→ Wrike #42: Feb 15: 5h (3h + 2h aggregated)

#45: log today 1h (update from 3h to 1h)
→ Wrike #42: Feb 15: 3h (1h + 2h recalculated)

#46: log yesterday 4h
→ Wrike #42: Feb 15: 3h, Feb 14: 4h
```

### Per-Child, Per-Day Latest Wins
```
#45 has comments:
1. "@wrikemeup log today 2h"
2. "some discussion"
3. "@wrikemeup log today 5h"

For "today": Use 5h (latest)
```

### Delete with 0h
```
Current:
- #45: today 3h
- #46: today 2h
→ Wrike: Feb 15: 5h

#45: log today 0h
→ #45's entry removed
→ Wrike: Feb 15: 2h (only #46 remains)

#46: log today 0h
→ Both removed
→ Wrike: Feb 15: 0h (or remove entry entirely)
```

---

## Close Behavior

### GitHub Action Trigger
```yaml
on:
  issues:
    types: [closed]
```

### Logic
1. **Check Wrike ID field**
   - No Wrike ID → Exit (no Wrike task exists)
   - Has Wrike ID → Continue

2. **Check Issue Type**
   - Child issue → Exit (parent owns the Wrike task)
   - Parent/Standalone → Continue

3. **Mark Complete in Wrike**
   - Set Wrike task status to complete/closed
   - Keep task (don't delete)
   - Preserve all logged hours

### Examples
```
Close #42 (parent with 20h logged)
→ Wrike task marked complete, 20h preserved

Close #45 (child of #42)
→ No action in Wrike

Close #50 (standalone, never logged)
→ No Wrike ID, no action
```

---

## Storage & Data Model

### GitHub Projects Custom Field
- **Field Name:** `Wrike ID`
- **Type:** Text
- **Stored On:** Parent and standalone issues only
- **Value:** Wrike task/folder ID
- **Children:** No Wrike ID (they log to parent)

### No Other Storage
- All aggregation calculated on-the-fly
- Scan child comments each time
- Self-healing (if Wrike gets out of sync)

---

## GitHub Action Triggers

```yaml
on:
  issue_comment:
    types: [created]
  issues:
    types: [closed]
```

---

## Edge Cases

### Multiple Logs Same Day, Same Child
```
#45: log today 2h
#45: log today 3h

Result: 3h (latest wins, not 5h)
```

### Standalone Becomes Parent (Date Preservation)
```
#50: log today 3h, yesterday 2h, last monday 5h
→ Wrike: Feb 15: 3h, Feb 14: 2h, Feb 10: 5h

Add child #51
#51: log today 4h
→ Wrike: Feb 15: 4h, Feb 14: 2h, Feb 10: 5h
(Only Feb 15 updated, others preserved)
```

### All Children 0h for a Date
```
#45: log today 0h
#46: log today 0h

→ Wrike: Feb 15: 0h (or remove the entry)
```

### Race Conditions - Avoided
Full scan approach means simultaneous logs are safe:
- Both trigger scans
- Both recalculate from scratch
- Last write wins, but both writes are correct

---

## Wrike API Requirements

### Endpoints Needed
1. **Create Task**
   - POST `/tasks`
   - Returns task ID to store in GitHub Projects

2. **Update Time Entries**
   - POST `/tasks/{taskId}/timelogs`
   - Set hours per date
   - Update existing entries if date exists

3. **Mark Complete**
   - PUT `/tasks/{taskId}`
   - Set status to complete/closed

### API Considerations
- Need Wrike API token (stored in GitHub Secrets)
- Handle rate limiting
- Error handling for invalid task IDs

---

## Implementation Checklist

### Phase 1: Comment Parsing
- [ ] Parse `@wrikemeup log` commands
- [ ] Extract dates (relative, specific, day names)
- [ ] Extract hours (3h, 30m, 2h30m)
- [ ] Extract optional descriptions
- [ ] Handle multiple entries (comma-separated)
- [ ] Validate syntax

### Phase 2: Parent-Child Detection
- [ ] Detect if issue is standalone (no parent, no children)
- [ ] Detect if issue is child (has parent)
- [ ] Detect if issue is parent (has children)
- [ ] Get all children of a parent issue
- [ ] Use GitHub Issues API or sub-issues feature

### Phase 3: Aggregation Logic
- [ ] Scan all children's comments
- [ ] Filter comments for `@wrikemeup log`
- [ ] Parse each log entry
- [ ] Find latest log per child, per date
- [ ] Sum hours across children for each date
- [ ] Handle 0h deletion

### Phase 4: GitHub Projects Integration
- [ ] Read "Wrike ID" custom field
- [ ] Write "Wrike ID" custom field
- [ ] Use GitHub GraphQL API for Projects

### Phase 5: Wrike API Integration
- [ ] Create Wrike task
- [ ] Update time entries (per date)
- [ ] Mark task complete
- [ ] Error handling

### Phase 6: Close Behavior
- [ ] Trigger on issue close
- [ ] Check if parent/standalone
- [ ] Mark Wrike task complete

### Phase 7: Bot Replies
- [ ] Success confirmation with link
- [ ] Error messages
- [ ] Invalid command → docs link

### Phase 8: Testing
- [ ] Unit tests for parsing
- [ ] Integration tests with mock Wrike API
- [ ] E2E tests with test repository

---

## Technical Stack (Suggested)

- **Language:** TypeScript/Node.js (standard for GitHub Actions)
- **GitHub APIs:**
  - REST API (issues, comments)
  - GraphQL API (Projects custom fields)
- **Wrike API:** REST API
- **Date Parsing:** `chrono-node` or similar
- **Testing:** Jest

---

## User Documentation Location

End-user facing documentation (how to use the bot) will live at:
**https://github.com/wrikemeup/wrikemeup/README.md**

This spec is for implementation only.