package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

var osExit = os.Exit

type LogPattern struct {
	Regex  *regexp.Regexp
	Layout string
}

type Input struct {
	Start  time.Time
	End    time.Time
	Source io.Reader
}

var patterns = []LogPattern{
	{regexp.MustCompile(`([A-Z][a-z]{2}\s+\d+\s+\d{2}:\d{2}:\d{2}\.\d+)`), "Jan _2 15:04:05.000000"},
	{regexp.MustCompile(`(\d{4}[-/]\d{2}[-/]\d{2} \d{2}:\d{2}:\d{2})`), "2006-01-02 15:04:05"},
	{regexp.MustCompile(`(\d{4}[-/]\d{2}[-/]\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)`), "2006-01-02T15:04:05.999999999Z"},
	{regexp.MustCompile(`(\d{4}[-/]\d{2}[-/]\d{2}T\d{2}:\d{2}:\d{2}\.\d+)`), "2006-01-02T15:04:05.999999999"},
	{regexp.MustCompile(`(\d{4}[-/]\d{2}[-/]\d{2}T\d{2}:\d{2}:\d{2}Z)`), "2006-01-02T15:04:05Z"},
	{regexp.MustCompile(`(\d{4}[-/]\d{2}[-/]\d{2}T\d{2}:\d{2}:\d{2})`), "2006-01-02T15:04:05"},
	{regexp.MustCompile(`(\d{4}[-/]\d{2}[-/]\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{4})`), "2006-01-02T15:04:05-0700"},
	{regexp.MustCompile(`^[A-Z](\d{4}\s+\d{2}:\d{2}:\d{2}\.\d+)`), "0102 15:04:05.000000"},
}

func main() {
	location := time.Local
	now := time.Now().In(location)

	const flagLayout = "2006-01-02"

	input := flagValidations()

	scanner := bufio.NewScanner(input.Source)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for scanner.Scan() {
		if processed, ok := processLogLine(scanner.Text(), input.Start, input.End, now); ok {
			fmt.Fprintln(writer, processed)
		}
	}
}

func processLogLine(line string, start, end time.Time, now time.Time) (string, bool) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return "", false
	}

	fullPath, content := parts[0], parts[1]

	rawTimeMatch, foundTime, foundTS := findPattern(content, now)

	if foundTS {
		if foundTime.Before(start) || foundTime.After(end) {
			return "", false
		}
		shortTime := foundTime.Format("060102150405")
		content = strings.Replace(content, rawTimeMatch, shortTime, 1)
		rawTimeMatch, _, foundTS = findPattern(content, now)
		if foundTS {
			content = strings.Replace(content, `time="`+rawTimeMatch+`" `, "", 1)
		}
	}

	return fmt.Sprintf("%s:%s", fullPath, content), true
}

func flagValidations() Input {
	const flagLayout = "2006-01-02"
	location := time.Local
	now := time.Now().In(location)

	inputPtr := flag.String("in", "", "input file path")
	startPtr := flag.String("start", now.AddDate(0, 0, -7).Format(flagLayout), "start")
	endPtr := flag.String("end", now.AddDate(0, 0, 1).Format(flagLayout), "end")
	flag.Parse()

	startTime, err := time.ParseInLocation(flagLayout, *startPtr, location)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid start date '%s'. Expected format: YYYY-MM-DD\n", *startPtr)
		flag.CommandLine.Usage()
		osExit(1)
	}

	endTime, err := time.ParseInLocation(flagLayout, *endPtr, location)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid end date '%s'. Expected format: YYYY-MM-DD\n", *endPtr)
		flag.CommandLine.Usage()
		osExit(1)
	}

	if startTime.After(endTime) {
		fmt.Fprintf(os.Stderr, "Error: start date cannot be after end date (-start: %s, -end: %s)\n", *startPtr, *endPtr)
		flag.CommandLine.Usage()
		osExit(1)
	}

	var inputSource io.Reader = os.Stdin
	if *inputPtr != "" {
		file, err := os.Open(*inputPtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			osExit(1)
		}
		inputSource = file
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			fmt.Fprintf(os.Stderr, "Error: no input provided. Please specify an input file using -in or pipe data into Stdin\n")
			flag.CommandLine.Usage()
			osExit(1)
		}
	}

	return Input{
		Start:  startTime,
		End:    endTime,
		Source: inputSource,
	}
}

func findPattern(content string, now time.Time) (string, time.Time, bool) {
	var foundTime time.Time
	var rawTimeMatch string
	currentYear := now.Year()

	for _, p := range patterns {
		match := p.Regex.FindStringSubmatch(content)
		if len(match) > 0 {
			rawTimeMatch = match[0]
			if len(match) > 1 {
				rawTimeMatch = match[1]
			}
			t, err := time.ParseInLocation(p.Layout, strings.Replace(rawTimeMatch, "/", "-", -1), time.Local)
			if err == nil {
				if t.Year() == 0 {
					t = t.AddDate(currentYear, 0, 0)
				}
				foundTime = t
				return rawTimeMatch, t, true
			}
		}
	}

	return rawTimeMatch, foundTime, false
}
