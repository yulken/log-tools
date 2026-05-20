# Log Tools

A collection of high-performance Go utilities designed to filter, summarize, and visualize log data. These tools are built to handle large datasets efficiently and work seamlessly with piped input.

## Tools Overview

| Tool | Description |
| :--- | :--- |
| [**log-filter**](./log-filter) | Filters logs by date/time range and normalizes timestamps. |
| [**summarizer**](./summarizer) | Groups logs into patterns, masking IDs/UUIDs to find frequent events. |
| [**log-tree**](./log-tree) | Visualizes file paths in a tree structure. |

## Important Notes

- **Input Format**: Most tools expect input lines to be prefixed with the file path (standard `grep -H` format). Always use `grep` with the `-H` (or `--with-filename`) flag when piping data.
- **Pipeline Logic**:
  - **log-filter**: Crucial for normalizing timestamps to a standard 12-digit format (`YYMMDDHHMMSS`).
  - **log-tree**: Can be used directly from `grep` or after filtering to visualize the structure.
  - **summarizer**: Depends on the normalization provided by `log-filter` to correctly identify and mask dates.

**Possible Combinations:**
1. `log-filter` (Standard filtering)
2. `log-filter -> log-tree` (Filtered structure)
3. `log-tree` (Direct visualization from grep)
4. `log-tree -> summarizer` (Summarize patterns within a structure)
5. `log-filter -> log-tree -> summarizer` (Full pipeline: filter, organize, and summarize)

## Installation

Ensure you have Go 1.24+ installed. You can build all tools using:

```bash
make <toolname>
```

## Typical Workflow

A common use case involves combining these tools to drill down into logs:

```bash
# Search logs, filter by date, and summarize patterns
grep -rH "ERROR" /path/to/logs | \
  ./log-filter -start 2024-05-10 -end 2024-05-20 | \
  ./summarizer -mask -min 5
```

## License
MIT