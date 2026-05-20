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
)

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

	flag.Parse()

	if *minPtr < 1 {
		fmt.Fprintf(os.Stderr, "Error: invalid -min value (%d). Minimum occurrences must be at least 1\n", *minPtr)
		flag.Usage()
		os.Exit(1)
	}

	var inputSource io.Reader = os.Stdin
	if *inputPtr != "" {
		file, err := os.Open(*inputPtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open input file '%s': %v\n", *inputPtr, err)
			os.Exit(1)
		}
		defer file.Close()
		inputSource = file
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			fmt.Fprintf(os.Stderr, "Error: no input provided. Please specify an input file using -in or pipe data into Stdin\n")
			flag.Usage()
			os.Exit(1)
		}
	}

	return Input{
		minOccurrences: *minPtr,
		groupPatterns:  *maskPtr,
		source:         inputSource,
	}
}

func main() {
	input := validateFlags()
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

	writer := bufio.NewWriter(os.Stdout)

	for _, entry := range summary {
		// Se não encontrou nenhuma data/timestamp na linha, exibe de forma limpa
		timeRange := "No Date"
		if entry.First != "" || entry.Last != "" {
			timeRange = fmt.Sprintf("%s -> %s", entry.First, entry.Last)
		}

		writer.WriteString(fmt.Sprintf("[%d] [%s]: %s\n", entry.Count, timeRange, entry.Text))
	}

	writer.WriteString("\n" + strings.Repeat("-", 40) + "\n")
	writer.WriteString(fmt.Sprintf("%d summarized lines\n", totalProcessed))
	writer.WriteString(fmt.Sprintf("%d unique patterns\n", len(counts)))
	writer.WriteString(strings.Repeat("-", 40) + "\n")
	writer.Flush()
}
