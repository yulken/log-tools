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

var osExit = os.Exit

type Node struct {
	Children map[string]*Node
	Logs     []string
}

func NewNode() *Node {
	return &Node{Children: make(map[string]*Node), Logs: []string{}}
}

var reCleanFile = regexp.MustCompile(`(\.(txt|log|yaml|json|conf|cfg|2|1|txt))(-[\s\d\.]+)?$`)

func main() {
	inputPtr := flag.String("in", "", "input")
	flag.Parse()

	var inputSource io.Reader = os.Stdin
	if *inputPtr != "" {
		file, err := os.Open(*inputPtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
		}
		defer file.Close()
		inputSource = file
	}

	processLogs(inputSource, os.Stdout)
}

func processLogs(r io.Reader, w io.Writer) {
	root := NewNode()
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 20*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx == -1 {
			continue
		}

		rawPath := line[:idx]
		content := strings.TrimSpace(line[idx+1:])

		cleanPath := reCleanFile.ReplaceAllString(rawPath, "$1")

		pathParts := strings.Split(cleanPath, "/")
		current := root
		for _, part := range pathParts {
			if part == "" {
				continue
			}
			if _, exists := current.Children[part]; !exists {
				current.Children[part] = NewNode()
			}
			current = current.Children[part]
		}

		current.Logs = append(current.Logs, content)
	}

	writer := bufio.NewWriter(w)
	renderTree(writer, root, "", "")
	writer.Flush()
}

func renderTree(w *bufio.Writer, node *Node, name string, indent string) {
	if name != "" {
		w.WriteString(indent + name + ":\n")
	}

	newIndent := indent
	if name != "" {
		newIndent += "  "
	}

	children := make([]string, 0, len(node.Children))
	for k := range node.Children {
		children = append(children, k)
	}
	sort.Strings(children)

	for _, childName := range children {
		renderTree(w, node.Children[childName], childName, newIndent)
	}

	for _, logLine := range node.Logs {
		w.WriteString(newIndent + logLine + "\n")
	}
}
