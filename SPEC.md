# WrikeMeUp Specification

## Overview

WrikeMeUp is a GitHub bot that logs hours from GitHub issues to Wrike tasks using simple commands.

## Command Format

### Log Hours

```
@wrikemeup log <entries>
```

Where `<entries>` is comma-separated with format: `<date> <duration> ["description"]`

**Examples:**
```
@wrikemeup log today 3h
@wrikemeup log today 3h "fixed authentication bug"
@wrikemeup log today 3h, yesterday 2h
@wrikemeup log monday 4h, tuesday 5h, wednesday 3h
@wrikemeup log feb 15 8h "sprint work"
```

### Delete Hours

```
@wrikemeup delete <dates>
```

**Examples:**
```
@wrikemeup delete monday
@wrikemeup delete monday, tuesday
@wrikemeup delete yesterday
```

### Show Hours

```
@wrikemeup show
```

Shows all logged hours in a table.

## Date Formats

- `today` - Current day
- `yesterday` - Previous day
- `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday`, `sunday` - Most recent occurrence
- `last monday`, `last tuesday`, etc. - Last week's occurrence
- `feb 15`, `15 feb`, `march 20`, `20 march` - Text month + day (current year)
- `2024-02-15` - Full ISO date
- `15` - Day of current month
- `02-15` - Month-day of current year

## Duration Formats

- `3h` - Hours
- `30m` - Minutes
- `2h30m` - Combined hours and minutes

## Bot Behavior

1. **First log**: Creates Wrike task automatically
2. **Subsequent logs**: Updates existing Wrike task
3. **Always responds**: Posts summary table after each command
4. **Incremental**: Only updates dates mentioned in command

## Wrike Integration

- Wrike task ID stored in GitHub Projects custom field "Wrike ID"
- Each date logs to Wrike with specific date
- Bot user's Wrike token used for API calls
- On issue close: Marks Wrike task as complete

## Implementation Requirements

1. **Language**: Go (Golang)
2. **Hosting**: GitHub Actions (serverless)
3. **Triggers**: Issue comments, issue close
4. **Dependencies**: Minimal - standard library + GitHub/Wrike API clients
5. **Documentation**: README.md + SETUP.md only
