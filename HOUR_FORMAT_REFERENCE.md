# Hour Format Quick Reference

## TL;DR

**One simple line:**
```
Hours: 16: 3h, 17: 4.5h, 18: 2h
```

That's it! Bot handles the rest. ✅

---

## Format Rules

### Basic Format
```
Hours: <date>: <hours>, <date>: <hours>, ...
```

- **Comma-separated** - Clear boundaries
- **YAML-like** - `date: hours` pairs
- **One line** - All entries together
- **Spaces optional** - `16:3h` or `16: 3h` both work

### Date Formats

| Format | Example | Meaning |
|--------|---------|---------|
| Day only | `16: 3h` | Day 16 of current month/year |
| Month-Day | `03-16: 4h` | March 16 of current year |
| Full date | `2024-02-16: 5h` | February 16, 2024 |

**Tip:** Use the shortest format that works!

---

## Examples

### Log Single Day
```
Hours: 16: 3h
```
→ Logs 3h on day 16 of current month

### Log Multiple Days
```
Hours: 16: 3h, 17: 4.5h, 18: 2h
```
→ Logs 3 different days in one comment

### Mix Date Formats
```
Hours: 16: 3h, 03-20: 4h, 2023-12-25: 2h
```
→ Current month, different month, different year

### Incremental Logging
```
# First comment:
Hours: 16: 3h

# Second comment (later):
Hours: 17: 4.5h
```
→ Bot keeps both days (no need to repeat!)

### Update Hours
```
# Original:
Hours: 16: 3h

# Update (new comment):
Hours: 16: 5h
```
→ Latest value wins (day 16 now has 5h)

### Delete Hours
```
Hours: 16: 0h
```
→ Removes day 16 completely

---

## Bot Response

After every comment with hours, bot posts:

```
✅ Hours Synced to Wrike

Current State
| Date | Hours | Status |
|------|-------|--------|
| 2024-02-16 | 3.00h | ✓ |
| 2024-02-17 | 4.50h | Added |
| 2024-02-18 | 2.00h | ✓ |

Total: 9.50h
```

Shows:
- ✅ All current hours (not just changes)
- ✅ Status for each day
- ✅ Total hours logged

---

## Common Mistakes

### ❌ Wrong: Missing colon after date
```
Hours: 16 = 3h
```

### ✅ Correct: Colon separator
```
Hours: 16: 3h
```

---

### ❌ Wrong: No "Hours:" prefix
```
16: 3h, 17: 4h
```

### ✅ Correct: Include "Hours:"
```
Hours: 16: 3h, 17: 4h
```

---

### ❌ Wrong: Multiple "Hours:" lines
```
Hours: 16: 3h
Hours: 17: 4h
```

### ✅ Correct: One line, comma-separated
```
Hours: 16: 3h, 17: 4h
```

---

## What If Format Is Wrong?

Bot helps you! It posts:

```
⚠️ Hour Logging Format Errors

I found some issues:

1. Invalid entry format: '16 = 3h'. Expected: '16: 3h'

Correct Format:
Hours: 16: 3h, 17: 4.5h, 18: 2h

Date formats:
- Day only: 16: 3h (current month/year)
- Month-Day: 03-16: 4h (current year)
- Full date: 2024-02-16: 5h

To delete: Hours: 16: 0h
```

- ✅ Shows what's wrong
- ✅ Provides correct examples
- ✅ Still syncs valid entries

---

## Advanced Usage

### Log Past Dates
```
Hours: 2024-01-15: 8h, 2024-01-16: 7h
```
→ Logs hours to January

### Log Future Dates
```
Hours: 03-20: 4h
```
→ Logs hours to March 20 (useful for planning)

### Fractional Hours
```
Hours: 16: 3.5h, 17: 4.25h, 18: 2.75h
```
→ Supports decimals (e.g., 3.5h = 3 hours 30 minutes)

### Zero Hours (Delete)
```
Hours: 15: 0h, 16: 0h, 17: 0h
```
→ Removes multiple days at once

---

## Tips

1. **Start simple** - `Hours: 16: 3h` is all you need
2. **Add incrementally** - Don't repeat previous days
3. **Use 0h to delete** - Simpler than editing old comments
4. **Check bot response** - Always shows current state
5. **Fix errors quickly** - Bot tells you exactly what's wrong

---

## Need Help?

See [SETUP_GUIDE.md](SETUP_GUIDE.md) for full documentation or [README.md](README.md) for examples.
