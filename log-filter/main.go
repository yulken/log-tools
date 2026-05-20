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

type LogPattern struct {
	Regex  *regexp.Regexp
	Layout string
}

var patterns = []LogPattern{
	{regexp.MustCompile(`([A-Z][a-z]{2}\s+\d+\s+\d{2}:\d{2}:\d{2}\.\d+)`), "Jan _2 15:04:05.000000"},
	{regexp.MustCompile(`(\d{4}[-/]\d{2}[-/]\d{2} \d{2}:\d{2}:\d{2})`), "2006/01/02 15:04:05"},
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

	inputPtr := flag.String("in", "", "input file")
	startPtr := flag.String("start", now.AddDate(0, 0, -7).Format(flagLayout), "start")
	endPtr := flag.String("end", now.Format(flagLayout), "end")
	flag.Parse()

	var inputSource io.Reader = os.Stdin
	if *inputPtr != "" {
		file, err := os.Open(*inputPtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		inputSource = file
	}

	startTime, _ := time.ParseInLocation(flagLayout, *startPtr, location)
	endTime, _ := time.ParseInLocation(flagLayout, *endPtr, location)

	scanner := bufio.NewScanner(inputSource)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}

		fullPath, content := parts[0], parts[1]

		rawTimeMatch, foundTime, foundTS := findPattern(content, now)

		if foundTS {
			if foundTime.Before(startTime) || foundTime.After(endTime) {
				continue
			}
			shortTime := foundTime.Format("060102150405")
			content = strings.Replace(content, rawTimeMatch, shortTime, 1)
			rawTimeMatch, _, foundTS = findPattern(content, now)
			if foundTS {
				content = strings.Replace(content, `time="`+rawTimeMatch+`" `, "", 1)
			}
		}

		fmt.Fprintf(writer, "%s:%s\n", fullPath, content)
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
			t, err := time.ParseInLocation(p.Layout, rawTimeMatch, time.Local)
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
