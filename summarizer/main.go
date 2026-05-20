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
}

var (
	reTS     = regexp.MustCompile(`\d{12}`)
	reHashes = regexp.MustCompile(`[a-f0-9]{32,64}`)
	rePIDs   = regexp.MustCompile(`pid=\d+|uid=\d+|\d{4,}`)
	reUUID   = regexp.MustCompile(`[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}`)
)

func removeTS(line string) string {
	line = reTS.ReplaceAllString(line, "<DATE>")
	return line
}

func normalize(line string) string {
	line = reUUID.ReplaceAllString(line, "<REP>")
	line = reHashes.ReplaceAllString(line, "<REP>")
	line = rePIDs.ReplaceAllString(line, "<REP>")
	return line
}

func main() {
	inputPtr := flag.String("in", "", "Input file (Opcional, aceita Stdin/Pipe)")
	minPtr := flag.Int("min", 2, "Mínimo de ocorrências")
	maskPtr := flag.Bool("mask", true, "Agrupar padrões (Ids, Datas, etc)")
	flag.Parse()

	var inputSource io.Reader

	if *inputPtr != "" {
		file, err := os.Open(*inputPtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao abrir entrada: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		inputSource = file
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			fmt.Println("Aguardando entrada via Pipe ou use -in para especificar um arquivo.")
			flag.PrintDefaults()
			return
		}
		inputSource = os.Stdin
	}

	counts := make(map[string]int)
	totalProcessed := 0
	scanner := bufio.NewScanner(inputSource)
	var pattern string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasSuffix(line, ":") {
			continue
		}
		pattern = removeTS(line)
		if *maskPtr {
			pattern = normalize(pattern)
		}
		counts[pattern]++
		totalProcessed++
	}

	var summary []Entry
	for pat, count := range counts {
		if count >= *minPtr {
			summary = append(summary, Entry{Text: pat, Count: count})
		}
	}

	sort.Slice(summary, func(i, j int) bool {
		return summary[i].Count > summary[j].Count
	})

	writer := bufio.NewWriter(os.Stdout)

	for _, entry := range summary {
		writer.WriteString(fmt.Sprintf("[%d]: %s\n", entry.Count, entry.Text))
	}

	writer.WriteString("\n" + strings.Repeat("-", 40) + "\n")
	writer.WriteString(fmt.Sprintf("%d summarized lines\n", totalProcessed))
	writer.WriteString(fmt.Sprintf("%d unique patterns\n", len(counts)))
	writer.WriteString(strings.Repeat("-", 40) + "\n")
	writer.Flush()
}
