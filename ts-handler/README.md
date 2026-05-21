# Log Filter

A utility to filter log lines based on a specific time range. It automatically detects common timestamp formats (RFC3339, Syslog, ISO8601, Logcat, etc.) and normalizes them into a standard `YYMMDDHHMMSS` format for easier analysis.

## Usage

```bash
./ts-handler [flags]
```

### Flags

*   `-in`: Path to the input file. If omitted, it reads from `stdin` (pipe).
*   `-start`: Start date in `YYYY-MM-DD` format (Default: 7 days ago).
*   `-end`: End date in `YYYY-MM-DD` format (Default: tomorrow).

## Examples

**Filter a file for a specific range:**
```bash
./ts-handler -in system.log -start 2024-05-01 -end 2024-05-15
```

**Filtering piped input:**
```bash
cat app.log | ./ts-handler -start 2024-05-20
```

## Features
*   Standardizes multiple timestamp formats found in the same stream.
*   Cleans up redundant structured logging fields like `time="..."`.
