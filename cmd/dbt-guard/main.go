package main

import (
	"fmt"
	"os"

	"github.com/renatocruz/dbt-guard/internal/parser"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	if err := parser.PrintPIIColumns(root); err != nil {
		fmt.Fprintf(os.Stderr, "erro: %v\n", err)
		os.Exit(1)
	}
}
