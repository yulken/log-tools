package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

var osExit = os.Exit

type Entry struct {
	Text  string
	Count int
	First string
	Last  string
}

type patternStat struct {
	count int
	first string
	last  string
}

type Input struct {
	minOccurrences int
	groupPatterns  bool
	source         io.Reader
	cutoffDate     string
}

var (
	reTS     = regexp.MustCompile(`\d{12}`)
	reHashes = regexp.MustCompile(`[a-f0-9]{32,64}`)
	rePIDs   = regexp.MustCompile(`pid=\d+|uid=\d+|\d{4,}`)
	reUUID   = regexp.MustCompile(`[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}`)
)

func extractAndMaskTS(line string) (string, string) {
	ts := reTS.FindString(line)
	masked := reTS.ReplaceAllString(line, "<DATE>")
	return masked, ts
}

func normalize(line string) string {
	line = reUUID.ReplaceAllString(line, "<REP>")
	line = reHashes.ReplaceAllString(line, "<REP>")
	line = rePIDs.ReplaceAllString(line, "<REP>")
	return line
}

func validateFlags() Input {
	inputPtr := flag.String("in", "", "Input file path (Optional, defaults to Stdin/Pipe)")
	minPtr := flag.Int("min", 2, "Minimum number of occurrences")
	maskPtr := flag.Bool("mask", false, "Group patterns together (e.g., IDs, Dates)")
	cutoffPtr := flag.String("cutoff", "", "Ignore patterns that stopped occurring before this date (Format: YYYY-MM-DD)")

	flag.Parse()

	if *minPtr < 1 {
		fmt.Fprintf(os.Stderr, "Error: invalid -min value (%d). Minimum occurrences must be at least 1\n", *minPtr)
		flag.CommandLine.Usage()
		osExit(1)
	}

	const dateLayout = "2006-01-02"
	var cutoffTS string

	if *cutoffPtr != "" {
		t, err := time.Parse(dateLayout, *cutoffPtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid -cutoff format '%s'. Expected: YYYY-MM-DD\n", *cutoffPtr)
			osExit(1)
		}
		cutoffTS = t.Format("060102") + "000000"
	}

	var inputSource io.Reader = os.Stdin
	if *inputPtr != "" {
		file, err := os.Open(*inputPtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open input file '%s': %v\n", *inputPtr, err)
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
		minOccurrences: *minPtr,
		groupPatterns:  *maskPtr,
		source:         inputSource,
		cutoffDate:     cutoffTS,
	}
}

func main() {
	input := validateFlags()
	processLogs(input, os.Stdout)
}

func processLogs(input Input, out io.Writer) {
	counts := make(map[string]patternStat)
	totalProcessed := 0
	scanner := bufio.NewScanner(input.source)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasSuffix(line, ":") {
			continue
		}

		pattern, currentTS := extractAndMaskTS(line)
		if input.groupPatterns {
			pattern = normalize(pattern)
		}

		stat := counts[pattern]
		stat.count++

		if stat.first == "" && currentTS != "" {
			stat.first = currentTS
		}
		if currentTS != "" {
			stat.last = currentTS
		}

		counts[pattern] = stat
		totalProcessed++
	}

	var summary []Entry
	for pat, stat := range counts {
		if stat.count >= input.minOccurrences {

			if input.cutoffDate != "" {
				if stat.last != "" {
					if stat.last < input.cutoffDate {
						continue
					}
				} else {
					continue
				}
			}

			summary = append(summary, Entry{
				Text:  pat,
				Count: stat.count,
				First: stat.first,
				Last:  stat.last,
			})
		}
	}

	sort.Slice(summary, func(i, j int) bool {
		return summary[i].Count > summary[j].Count
	})

	writer := bufio.NewWriter(out)

	for _, entry := range summary {
		timeRange := "No Date"
		if entry.First != "" || entry.Last != "" {
			timeRange = fmt.Sprintf("%s -> %s", entry.First, entry.Last)
		}
		fmt.Fprintf(writer, "[%d] [%s]: %s\n", entry.Count, timeRange, entry.Text)
	}

	fmt.Fprintln(writer, "\n"+strings.Repeat("-", 40))
	fmt.Fprintf(writer, "%d summarized lines\n", totalProcessed)
	fmt.Fprintf(writer, "%d unique patterns\n", len(counts))
	fmt.Fprintln(writer, strings.Repeat("-", 40))
	writer.Flush()
}
