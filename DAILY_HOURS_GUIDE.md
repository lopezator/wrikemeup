# Quick Answer: How to Log Hours to Specific Dates

## Your Question
> "If I want to log 3h to 16/02 and 5h to 18/02 how would that work? Can it work with fields?"

## Answer

### ✅ Yes! Use Daily Breakdown Format

Add this to your **issue body**:
```markdown
Hours: 2024-02-16: 3h, 2024-02-18: 5h
```

Then **edit and save** the issue → Bot automatically logs:
- 3 hours to February 16, 2024
- 5 hours to February 18, 2024

### Why Issue Body (Not Custom Fields)?

GitHub Projects V2 custom fields can only store **single values**:
- Number field: One number (e.g., `8`)
- Text field: One text string (e.g., `"Some text"`)

Daily breakdown needs **multiple date-hour pairs**, so it must go in the issue body.

### Complete Workflow with Projects V2

**1. Setup GitHub Project with custom fields:**
- "Wrike Parent" (Single Select: Yes/No) ← Marks parent issues
- "Wrike Task ID" (Text) ← Auto-filled by bot

**2. Create issue and add to project:**
```markdown
# Sprint 23 - Authentication

Implement user authentication system
```

**3. Set custom field "Wrike Parent" = "Yes":**
- Bot creates Wrike task
- Bot fills "Wrike Task ID" automatically

**4. Add hours to issue body:**
```markdown
# Sprint 23 - Authentication

Wrike Task ID: IEABC123 ← Auto-added by bot
Last Synced: 0h ← Tracks synced hours

Hours: 2024-02-16: 3h, 2024-02-18: 5h ← Add this!

Implement user authentication system
```

**5. Edit and save:**
- Bot automatically syncs
- 3h logged to Feb 16 in Wrike
- 5h logged to Feb 18 in Wrike
- ✅ Done!

### Adding More Hours Later

**Update the issue body:**
```markdown
Hours: 2024-02-16: 3h, 2024-02-18: 5h, 2024-02-20: 2h
```

**Edit and save:**
- Bot logs 2h to Feb 20 (incremental!)
- Previous days already logged, not duplicated
- ✅ Smart!

### Alternative: Simple Total Hours

If you don't need daily breakdown:
```markdown
Hours: 12h
```
- Logs 12h to today's date
- Simpler but less detail

## No Bot Commands Needed!

**Old way (manual):**
```
@wrikemeup sync
```

**New way (automatic):**
```
Just edit the issue!
```

Bot automatically syncs whenever you save changes. No commands needed!

## Summary

✅ **Daily breakdown:** `Hours: 2024-02-16: 3h, 2024-02-18: 5h` in issue body
✅ **Projects V2 fields:** Use for "Wrike Parent" metadata
✅ **Automatic sync:** Just edit and save, no commands
✅ **Incremental logging:** Only logs new hours, no duplicates
✅ **Works everywhere:** Issues, Projects V2, with or without labels

**Time to log 3h to Feb 16 and 5h to Feb 18:**
1. Add `Hours: 2024-02-16: 3h, 2024-02-18: 5h` to issue body
2. Click "Update comment"
3. ✅ Done! (2 seconds)
