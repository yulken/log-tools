package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
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

func setupTestFlags(args []string) (func(), *bytes.Buffer) {
	oldArgs, oldStderr, oldStdin := os.Args, os.Stderr, os.Stdin
	os.Args = append([]string{"summarizer"}, args...)
	rIn, wIn, _ := os.Pipe()
	os.Stdin = rIn
	r, w, _ := os.Pipe()
	os.Stderr = w
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.CommandLine.SetOutput(w)
	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		for {
			b := make([]byte, 1024)
			n, err := r.Read(b)
			if n > 0 {
				buf.Write(b[:n])
			}
			if err != nil {
				break
			}
		}
		close(done)
	}()
	var once sync.Once
	teardown := func() {
		once.Do(func() {
			w.Close()
			wIn.Close()
			<-done
			os.Args, os.Stderr, os.Stdin = oldArgs, oldStderr, oldStdin
		})
	}
	return teardown, &buf
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user-550e8400-e29b-41d4-a716-446655440000 logged in", "user-<?> logged in"},
		{"task 1234567890abcdef1234567890abcdef failed", "task <?> failed"},
		{"process pid=1234 is running", "process <?> is running"},
		{"replica pvc-550e8400-e29b-41d4-a716-446655440000-r-a1b2c3", "replica <LH-REP>"},
		{"engine pvc-550e8400-e29b-41d4-a716-446655440000-e-d4e5f6", "engine <LH-E>"},
		{"volume pvc-550e8400-e29b-41d4-a716-446655440000 status", "volume <PVC> status"},
		{"connection from 192.168.1.100:5432", "connection from <IP:PORT>"},
		{"host 8.8.4.4 is up", "host <IP> is up"},
	}
	for _, tt := range tests {
		if got := normalize(tt.input); got != tt.expected {
			t.Errorf("normalize(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestExtractAndMaskTS(t *testing.T) {
	line := "240520100000 some message"
	masked, ts := extractAndMaskTS(line)
	if ts != "240520100000" {
		t.Errorf("expected ts 240520100000, got %s", ts)
	}
	if masked != "<DATE> some message" {
		t.Errorf("expected masked <DATE> some message, got %s", masked)
	}
}

func TestValidateFlagsErrors(t *testing.T) {
	t.Run("Invalid min value", func(t *testing.T) {
		teardown, stderr := setupTestFlags([]string{"-min", "0"})
		defer teardown()
		defer func() {
			recover()
			teardown()
			if !strings.Contains(stderr.String(), "Error: invalid -min value") {
				t.Errorf("expected min value error, got: %s", stderr.String())
			}
		}()
		validateFlags()
	})

	t.Run("Invalid cutoff format", func(t *testing.T) {
		teardown, stderr := setupTestFlags([]string{"-cutoff", "20-05-2024"})
		defer teardown()
		defer func() {
			recover()
			teardown()
			if !strings.Contains(stderr.String(), "Error: invalid -cutoff format") {
				t.Errorf("expected cutoff format error, got: %s", stderr.String())
			}
		}()
		validateFlags()
	})
}

func TestProcessLogs(t *testing.T) {
	t.Run("Basic summary", func(t *testing.T) {
		input := Input{
			minOccurrences: 1,
			source:         strings.NewReader("240520100000 error A\n240520100001 error A\n"),
		}
		var out bytes.Buffer
		processLogs(input, &out)
		if !strings.Contains(out.String(), "[2] [240520100000 -> 240520100001]: <DATE> error A") {
			t.Errorf("output missing expected summary, got: %s", out.String())
		}
	})

	t.Run("With cutoff filter", func(t *testing.T) {
		input := Input{
			minOccurrences: 1,
			cutoffDate:     "240501000000",
			source:         strings.NewReader("240101100000 old error\n240601100000 new error\nno date error\n"),
		}
		var out bytes.Buffer
		processLogs(input, &out)
		res := out.String()
		if !strings.Contains(res, "new error") {
			t.Errorf("expected 'new error' to be kept, but it was not found")
		}
		if strings.Contains(res, "old error") {
			t.Errorf("expected 'old error' to be filtered out, but it was found")
		}
		if strings.Contains(res, "no date error") {
			t.Errorf("expected 'no date error' to be filtered out (no timestamp), but it was found")
		}
	})

	t.Run("Sorting order", func(t *testing.T) {
		lines := []string{
			"240520100002 error C",
			"240520100002 error C",
			"240520100001 error B",
			"240520100001 error B",
			"240520100001 error B",
			"240520100000 error A",
			"240520100000 error A",
			"240520100000 error A",
			"240520100000 error A",
			"240520100000 error A",
		}
		input := Input{
			minOccurrences: 1,
			source:         strings.NewReader(strings.Join(lines, "\n") + "\n"),
		}
		var out bytes.Buffer
		processLogs(input, &out)
		res := out.String()

		posA := strings.Index(res, "error A")
		posB := strings.Index(res, "error B")
		posC := strings.Index(res, "error C")
		if !(posA < posB && posB < posC) {
			t.Errorf("Incorrect sorting order. Expected A (5), B (3), C (2). Indices: A=%d, B=%d, C=%d", posA, posB, posC)
		}
	})
}
