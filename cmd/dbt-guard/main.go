package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/renatocruz/dbt-guard/internal/parser"
	"github.com/renatocruz/dbt-guard/internal/validator"
)

func main() {
	if len(os.Args) >= 3 && os.Args[1] == "manifest" {
		if err := parser.PrintManifestPII(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "erro: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) >= 3 && os.Args[1] == "sensitive" {
		if err := parser.PrintSensitiveNodes(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "erro: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) >= 3 && os.Args[1] == "validate" {
		violations, err := validator.RunValidate(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "erro: %v\n", err)
			os.Exit(1)
		}
		if len(violations) > 0 {
			for _, v := range violations {
				fmt.Fprintf(os.Stderr, "[dbt-guard] modelo em analysis descende de PII sem mascaramento: %s\n", v.ModelID)
				fmt.Fprintf(os.Stderr, "  linhagem: %s\n", strings.Join(v.LineagePath, " → "))
			}
			os.Exit(1)
		}
		return
	}

	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	if err := parser.PrintPIIColumns(root); err != nil {
		fmt.Fprintf(os.Stderr, "erro: %v\n", err)
		os.Exit(1)
	}
}
