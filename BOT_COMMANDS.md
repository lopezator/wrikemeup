# Bot Commands Quick Reference

## TL;DR

**Log hours with natural dates:**
```
@wrikemeup log 3h, mon:2h, yesterday:5h
```

**Delete hours:**
```
@wrikemeup delete mon, yesterday
```

**Show current state:**
```
@wrikemeup show
```

That's it! Simple and intuitive. ✅

---

## Log Hours Command

### Basic Format
```
@wrikemeup log <entries>
```

Where `<entries>` is a comma-separated list of `[date:]hours`.

### Examples

**Log to today:**
```
@wrikemeup log 3h
```
→ Logs 3h to today

**Log multiple days:**
```
@wrikemeup log 3h, mon:2h, tue:4h
```
→ Logs 3h today, 2h Monday, 4h Tuesday

**Log to yesterday:**
```
@wrikemeup log yesterday:5h
```
→ Logs 5h to yesterday

**Log whole week:**
```
@wrikemeup log mon:8h, tue:7h, wed:8h, thu:8h, fri:6h
```
→ Logs hours for each weekday

---

## Relative Date Formats

### Day of Week
Use short or full day names (most recent occurrence):

| Format | Examples |
|--------|----------|
| Monday | `mon:3h`, `monday:3h` |
| Tuesday | `tue:4h`, `tuesday:4h` |
| Wednesday | `wed:5h`, `wednesday:5h` |
| Thursday | `thu:6h`, `thursday:6h` |
| Friday | `fri:7h`, `friday:7h` |
| Saturday | `sat:2h`, `saturday:2h` |
| Sunday | `sun:1h`, `sunday:1h` |

**How it works:**
- If today is Wednesday and you say `mon:3h`, it means **last Monday**
- If today is Monday and you say `mon:3h`, it means **today**

### Relative Days

| Format | Meaning | Example |
|--------|---------|---------|
| No date | Today | `3h` → today |
| `yesterday` | Yesterday | `yesterday:5h` |
| `-1` | 1 day ago | `-1:5h` → yesterday |
| `-2` | 2 days ago | `-2:4h` |
| `-7` | 7 days ago | `-7:8h` → last week |

### Absolute Dates

| Format | Example | Meaning |
|--------|---------|---------|
| Day only | `15:3h` | Day 15 of current month |
| Month-Day | `03-15:4h` | March 15 of current year |
| Full date | `2024-03-15:5h` | March 15, 2024 |

---

## Delete Hours Command

### Format
```
@wrikemeup delete <dates>
```

Where `<dates>` is a comma-separated list of date specifications.

### Examples

**Delete Monday's hours:**
```
@wrikemeup delete mon
```

**Delete multiple days:**
```
@wrikemeup delete mon, tue, wed
```

**Delete yesterday:**
```
@wrikemeup delete yesterday
```

**Delete specific date:**
```
@wrikemeup delete 2024-03-15
```

**Delete mix of dates:**
```
@wrikemeup delete mon, yesterday, 15
```

---

## Show Hours Command

### Format
```
@wrikemeup show
```

### What It Does
- Displays current logged hours in a table
- Shows all days with hours
- Shows total hours
- No changes made, just displays current state

### Example Output

```
✅ Hours Synced to Wrike

Current State
| Date | Hours | Status |
|------|-------|--------|
| 2024-02-12 | 8.00h | ✓ |
| 2024-02-13 | 7.00h | ✓ |
| 2024-02-14 | 8.00h | ✓ |
| 2024-02-15 | 6.00h | ✓ |
| 2024-02-16 | 8.00h | ✓ |

Total: 37.00h
```

---

## Real-World Examples

### Scenario 1: Daily Standup
Log yesterday's work:
```
@wrikemeup log yesterday:7h
```

### Scenario 2: End of Week
Log whole week at once:
```
@wrikemeup log mon:8h, tue:7.5h, wed:8h, thu:8h, fri:6h
```

### Scenario 3: Forgot to Log
Log past few days:
```
@wrikemeup log -1:5h, -2:6h, -3:7h
```
→ Logs last 3 days

### Scenario 4: Partial Day
Log specific hours to different days:
```
@wrikemeup log 3h, mon:2h, yesterday:4h
```
→ 3h today, 2h Monday, 4h yesterday

### Scenario 5: Correction
Made a mistake? Delete and re-log:
```
@wrikemeup delete mon
@wrikemeup log mon:8h
```

### Scenario 6: Check Status
Before submitting timesheet:
```
@wrikemeup show
```
→ See all logged hours

---

## Advanced Usage

### Mixed Date Formats
```
@wrikemeup log 3h, mon:8h, 15:4h, 2023-12-25:2h
```
→ Today, last Monday, day 15, and Christmas 2023

### Update Existing
Bot uses latest value for each date:
```
# First command:
@wrikemeup log mon:5h

# Later (update):
@wrikemeup log mon:8h
```
→ Monday now has 8h (not 13h)

### Delete Multiple
```
@wrikemeup delete mon, tue, wed, thu, fri
```
→ Clear whole week

### Fractional Hours
```
@wrikemeup log 3.5h, mon:4.25h, tue:7.75h
```
→ Supports decimals (3.5h = 3 hours 30 minutes)

---

## Common Workflows

### Daily Logging
```
# Every day:
@wrikemeup log 8h
```

### Weekly Batch
```
# Friday afternoon:
@wrikemeup log mon:8h, tue:8h, wed:7h, thu:8h, fri:8h
```

### Sprint Review
```
# Check total:
@wrikemeup show
```

### Timesheet Correction
```
# Fix Tuesday:
@wrikemeup delete tue
@wrikemeup log tue:6h
```

---

## Bot Responses

After every command, bot responds with current state:

**After log:**
```
✅ Hours Synced to Wrike

Current State
| Date | Hours | Status |
|------|-------|--------|
| 2024-02-16 | 3.00h | Added |

Total: 3.00h
```

**After delete:**
```
✅ Hours Synced to Wrike

Current State
| Date | Hours | Status |
|------|-------|--------|
| 2024-02-16 | 3.00h | ✓ |

Deleted Entries
| 2024-02-15 | - | Deleted: 5.00h |

Total: 3.00h
```

---

## Error Handling

### Invalid Format
```
@wrikemeup log xyz:3h
```
Bot responds:
```
❌ Error parsing command:

invalid date 'xyz': unknown date format: xyz

Commands:
- @wrikemeup log 3h (log 3h today)
- @wrikemeup log 3h, mon:2h
...
```

### Not Linked
```
@wrikemeup log 3h
```
Bot responds:
```
❌ This issue is not linked to a Wrike task.

Please add the `wrike-parent` label or use `@wrikemeup link <task-id>` first.
```

---

## Tips

1. **Use short forms** - `mon` instead of `monday`
2. **Mix formats** - Combine relative and absolute dates
3. **Check often** - Use `@wrikemeup show` to verify
4. **Batch logging** - Log multiple days in one command
5. **Update freely** - Latest value wins, no duplicates

---

## Other Commands

### Link to Wrike Task
```
@wrikemeup link IEABC123
```
→ Links issue to existing Wrike task

### Sync Now
```
@wrikemeup sync
```
→ Manually trigger hour sync

---

## Need Help?

See:
- [README.md](README.md) for overview
- [SETUP_GUIDE.md](SETUP_GUIDE.md) for setup instructions
- [FEATURE_SUMMARY.md](FEATURE_SUMMARY.md) for technical details
