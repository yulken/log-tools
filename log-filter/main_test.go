package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestProcessLogLine(t *testing.T) {
	now := time.Date(2024, 5, 20, 0, 0, 0, 0, time.Local)
	start := now.Add(-24 * time.Hour)
	end := now.Add(24 * time.Hour)

	tests := []struct {
		name     string
		line     string
		expected string
		keep     bool
	}{
		{"Valid and in range", "app.log:2024-05-20 10:00:00 message", "app.log:240520100000 message", true},
		{"Out of range (before)", "app.log:2024-05-18 10:00:00 message", "", false},
		{"Double timestamp removal", `app.log:2024-05-20 10:00:00 time="2024-05-20 10:00:00" msg=hi`, "app.log:240520100000 msg=hi", true},
		{"Malformed line (no colon)", "some random log line", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, keep := processLogLine(tt.line, start, end, now)
			if keep != tt.keep || (keep && got != tt.expected) {
				t.Errorf("got (%q, %v), want (%q, %v)", got, keep, tt.expected, tt.keep)
			}
		})
	}
}

func TestFindPattern(t *testing.T) {
	now := time.Date(2024, 5, 20, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name     string
		content  string
		expected string
		found    bool
	}{
		{
			name:     "Syslog format (Pattern 0)",
			content:  "May 20 22:24:00.123456 process[123]: message",
			expected: "2024-05-20 22:24:00.123456",
			found:    true,
		},
		{
			name:     "ISO-like with slash (Pattern 1)",
			content:  "2024/05/20 22:24:00 log content",
			expected: "2024-05-20 22:24:00",
			found:    true,
		},
		{
			name:     "RFC3339 with Z (Pattern 2)",
			content:  "2024-05-20T22:24:00.123Z message",
			expected: "2024-05-20 22:24:00.123",
			found:    true,
		},
		{
			name:     "Android Logcat (Pattern 7)",
			content:  "I0520 22:24:00.123456 1234 5678 Tags: message",
			expected: "2024-05-20 22:24:00.123456",
			found:    true,
		},
		{
			name:     "No timestamp",
			content:  "just a plain text line",
			expected: "",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ts, found := findPattern(tt.content, now)
			if found != tt.found {
				t.Fatalf("found = %v, want %v", found, tt.found)
			}
			if found {
				actual := ts.Format("2006-01-02 15:04:05.999999")
				if actual[:len(tt.expected)] != tt.expected {
					t.Errorf("got %s, want %s", actual, tt.expected)
				}
			}
		})
	}
}

func TestMain(m *testing.M) {
	oldOsExit := osExit
	osExit = func(code int) {
		panic(fmt.Sprintf("os.Exit called with code %d", code))
	}
	defer func() { osExit = oldOsExit }()

	code := m.Run()
	os.Exit(code)
}

func setupTestFlags(args []string) (func(), *bytes.Buffer) {
	oldArgs := os.Args
	oldStderr := os.Stderr
	oldStdin := os.Stdin

	os.Args = append([]string{"log-filter"}, args...)

	rIn, wIn, _ := os.Pipe()
	os.Stdin = rIn

	r, w, _ := os.Pipe()
	os.Stderr = w

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.CommandLine.SetOutput(w)

	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		io.Copy(&buf, r)
		close(done)
	}()

	var once sync.Once
	teardown := func() {
		once.Do(func() {
			w.Close()
			wIn.Close()
			<-done
			os.Args = oldArgs
			os.Stderr = oldStderr
			os.Stdin = oldStdin
		})
	}

	return teardown, &buf
}

func TestFlagValidations(t *testing.T) {
	location := time.Local
	now := time.Now().In(location)

	t.Run("Default values", func(t *testing.T) {
		teardown, _ := setupTestFlags([]string{})
		defer teardown()

		input := flagValidations()

		expectedStart := now.AddDate(0, 0, -7).Format("2006-01-02")
		expectedEnd := now.AddDate(0, 0, 1).Format("2006-01-02")

		if input.Start.Format("2006-01-02") != expectedStart {
			t.Errorf("Expected default start date %s, got %s", expectedStart, input.Start.Format("2006-01-02"))
		}
		if input.End.Format("2006-01-02") != expectedEnd {
			t.Errorf("Expected default end date %s, got %s", expectedEnd, input.End.Format("2006-01-02"))
		}
		if input.Source != os.Stdin {
			t.Errorf("Expected default source to be os.Stdin")
		}
	})

	t.Run("Custom start/end dates", func(t *testing.T) {
		teardown, _ := setupTestFlags([]string{"-start", "2023-01-01", "-end", "2023-01-02"})
		defer teardown()

		input := flagValidations()

		expectedStart, _ := time.ParseInLocation("2006-01-02", "2023-01-01", location)
		expectedEnd, _ := time.ParseInLocation("2006-01-02", "2023-01-02", location)

		if !input.Start.Equal(expectedStart) {
			t.Errorf("Expected start date %v, got %v", expectedStart, input.Start)
		}
		if !input.End.Equal(expectedEnd) {
			t.Errorf("Expected end date %v, got %v", expectedEnd, input.End)
		}
	})

	t.Run("Invalid start date format", func(t *testing.T) {
		teardown, stderr := setupTestFlags([]string{"-start", "01-01-2023"})
		defer teardown()

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected os.Exit to be called, but it wasn't")
			}
			teardown()
			if !strings.Contains(stderr.String(), "Error: invalid start date") {
				t.Errorf("Expected error message about invalid start date, got: %q", stderr.String())
			}
		}()
		flagValidations()
	})

	t.Run("Invalid end date format", func(t *testing.T) {
		teardown, stderr := setupTestFlags([]string{"-end", "01/01/2023"})
		defer teardown()

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected os.Exit to be called, but it wasn't")
			}
			teardown()
			if !strings.Contains(stderr.String(), "Error: invalid end date") {
				t.Errorf("Expected error message about invalid end date, got: %q", stderr.String())
			}
		}()
		flagValidations()
	})

	t.Run("Start date after end date", func(t *testing.T) {
		teardown, stderr := setupTestFlags([]string{"-start", "2023-01-02", "-end", "2023-01-01"})
		defer teardown()

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected os.Exit to be called, but it wasn't")
			}
			teardown()
			if !strings.Contains(stderr.String(), "Error: start date cannot be after end date") {
				t.Errorf("Expected error message about start date after end date, got: %q", stderr.String())
			}
		}()
		flagValidations()
	})

	t.Run("Input file specified", func(t *testing.T) {
		tmpfile, err := os.CreateTemp("", "testinput*.log")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())
		tmpfile.Close()

		teardown, _ := setupTestFlags([]string{"-in", tmpfile.Name()})
		defer teardown()

		input := flagValidations()

		if input.Source == os.Stdin {
			t.Errorf("Expected source to be a file, got os.Stdin")
		}

		buf := make([]byte, 1)
		n, err := input.Source.Read(buf)
		if n != 0 || err != io.EOF {
			t.Errorf("Expected 0 bytes and io.EOF from empty file, got %d bytes and error %v", n, err)
		}
	})

	t.Run("Invalid input file", func(t *testing.T) {
		teardown, stderr := setupTestFlags([]string{"-in", "/path/to/nonexistent/file.log"})
		defer teardown()

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected os.Exit to be called, but it wasn't")
			}
			teardown()
			if !strings.Contains(stderr.String(), "Error opening input file") {
				t.Errorf("Expected error message about opening input file, got: %q", stderr.String())
			}
		}()
		flagValidations()
	})

	t.Run("No input and not piped (skipped)", func(t *testing.T) {
		t.Skip("Skipping 'No input and not piped' test due to mocking complexity of os.Stdin.Stat()")
	})
}
