package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	oldOsExit := osExit
	osExit = func(code int) {
		panic(fmt.Sprintf("os.Exit called with code %d", code))
	}
	defer func() { osExit = oldOsExit }()
	os.Exit(m.Run())
}

func TestProcessLogs(t *testing.T) {
	t.Run("Basic tree construction", func(t *testing.T) {
		input := `a/b/c.log:log1
a/b/c.log:log2
a/d.txt:log3
x/y.json:log4`

		r := strings.NewReader(input)
		var out bytes.Buffer
		processLogs(r, &out)

		expected := `a:
  b:
    c.log:
      log1
      log2
  d.txt:
    log3
x:
  y.json:
    log4
`
		if out.String() != expected {
			t.Errorf("Unexpected output.\nGot:\n%s\nWant:\n%s", out.String(), expected)
		}
	})

	t.Run("Path cleaning", func(t *testing.T) {
		input := `app.log-2023.10.01:log entry
server.conf-123:config entry`

		r := strings.NewReader(input)
		var out bytes.Buffer
		processLogs(r, &out)

		// app.log-2023.10.01 -> app.log
		// server.conf-123 -> server.conf
		expected := `app.log:
  log entry
server.conf:
  config entry
`
		if out.String() != expected {
			t.Errorf("Unexpected output.\nGot:\n%s\nWant:\n%s", out.String(), expected)
		}
	})

	t.Run("Nested folders and duplicates", func(t *testing.T) {
		input := `root/sub1/file.log:entry1
root/sub1/file.log:entry2
root/sub2/other.log:entry3`

		r := strings.NewReader(input)
		var out bytes.Buffer
		processLogs(r, &out)

		expected := `root:
  sub1:
    file.log:
      entry1
      entry2
  sub2:
    other.log:
      entry3
`
		if out.String() != expected {
			t.Errorf("Unexpected output.\nGot:\n%s\nWant:\n%s", out.String(), expected)
		}
	})

	t.Run("Empty lines and malformed input", func(t *testing.T) {
		input := "\nmalformed line\nvalid/path.log:content\n"
		r := strings.NewReader(input)
		var out bytes.Buffer
		processLogs(r, &out)

		expected := `valid:
  path.log:
    content
`
		if out.String() != expected {
			t.Errorf("Unexpected output.\nGot:\n%s\nWant:\n%s", out.String(), expected)
		}
	})
}
