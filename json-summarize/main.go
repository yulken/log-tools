package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	// Flags
	inputPtr := flag.String("in", "", "JSON de entrada (Opcional, aceita Stdin)")
	maxPtr := flag.Int("max", 0, "Limite de itens para sumarizar (0 = sumariza tudo que for array)")
	flag.Parse()

	// 1. Define a fonte de entrada (Arquivo ou Pipe)
	var reader io.Reader
	if *inputPtr != "" {
		file, err := os.Open(*inputPtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao abrir arquivo: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		reader = file
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			fmt.Fprintln(os.Stderr, "Aguardando JSON via Pipe ou use -in.")
			return
		}
		reader = os.Stdin
	}

	// 2. Decodifica o JSON para uma estrutura genérica
	var data interface{}
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&data); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao decodificar JSON: %v\n", err)
		os.Exit(1)
	}

	// 3. Processa e sumariza
	summary := summarize(data, *maxPtr)

	// 4. Output formatado (Pretty Print)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	encoder.Encode(summary)
}

// summarize percorre o JSON recursivamente procurando por arrays
func summarize(v interface{}, max int) interface{} {
	switch t := v.(type) {

	case map[string]interface{}:
		// Se for um objeto, processa cada chave recursivamente
		newMap := make(map[string]interface{})
		for k, val := range t {
			newMap[k] = summarize(val, max)
		}
		return newMap

	case []interface{}:
		// Se for um array, verifica o tamanho
		if len(t) > max {
			return []string{fmt.Sprintf("[%d] entries", len(t))}
		}
		// Se for menor que o max, ainda processa os itens internos (caso sejam objetos)
		newSlice := make([]interface{}, len(t))
		for i, val := range t {
			newSlice[i] = summarize(val, max)
		}
		return newSlice

	default:
		// Tipos básicos (string, float, bool) retornam como estão
		return v
	}
}
