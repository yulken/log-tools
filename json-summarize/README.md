# JSON Summarize

A utility to simplify and explore massive JSON structures. It collapses large arrays into a summary string (e.g., `"[150] entries"`) while keeping the rest of the object hierarchy intact.

## Usage

```bash
./json-summarize [flags]
```

### Flags

*   `-in`: JSON input file (Optional, accepts Stdin).
*   `-max`: Limit of items to keep in an array. If an array exceeds this size, it is summarized. (Default: 0 - summarizes all arrays).

## Example

**Get an overview of a huge API response:**
```bash
curl -s https://api.example.com/data | ./json-summarize -max 5
```

**Input:** `{"users": [user1, user2, ..., user100], "status": "ok"}`
**Output:**
```json
{
    "users": ["[100] entries"],
    "status": "ok"
}
```
