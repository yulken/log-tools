# Summarizer

This tool analyzes log streams to group similar lines into patterns. It is extremely useful for identifying "noisy" logs or recurring errors in large datasets.

## Usage

```bash
./summarizer [flags]
```

### Flags

*   `-in`: Input log file (Defaults to Stdin).
*   `-min`: Minimum number of occurrences for a pattern to be displayed (Default: 2).
*   `-mask`: Enable grouping of patterns by masking UUIDs, Hashes, PIDs, and UIDs with `<REP>`.
*   `-cutoff`: Ignore patterns that stopped occurring before this date (`YYYY-MM-DD`).

## Examples

**Summarize frequent errors from a piped source:**
```bash
grep "Exception" logs.txt | ./summarizer -min 10 -mask
```

**Output format:**
```text
[Occurrences] [First Seen -> Last Seen]: Pattern text
[42] [240520100000 -> 240520113000]: User <REP> failed to login
```

## Key Features
*   **Masking:** Automatically detects and masks unique identifiers to aggregate similar log entries.
*   **Timeline:** Shows the first and last occurrence timestamp for every pattern found.
*   **Sorting:** Patterns are sorted by frequency (most frequent first).
