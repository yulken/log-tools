# Log Tree

Organizes logs into a visual directory tree structure. This is designed to be used with the output of `grep -r` or similar commands that prefix log lines with file paths (e.g., `path/to/file.log:content`).

## Usage

```bash
./log-tree [flags]
```

### Flags

*   `-in`: Input file path (Defaults to Stdin).

## Example

**Visualize where "ERROR" is occurring across multiple nodes:**
```bash
grep -r "ERROR" ./logs | ./log-tree
```

**Output Example:**
```text
var:
  log:
    syslog:
      Disk space low
      Out of memory
```

## Features
*   **Filename Cleaning:** Automatically strips rotation suffixes (like `.log-2023.10.01` or `.1`) to group logs by their original service name.
*   **Deterministic Output:** Sorted alphabetically for consistent results.
