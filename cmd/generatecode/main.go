// generatecode demonstrates Open -> GenerateCode -> Close.
//
// Usage (run from repo root so "CV CONTOH.pdf" is found):
//
//	go run ./cmd/generatecode
//	go run ./cmd/generatecode "CV CONTOH.pdf"
package main

import (
	"fmt"
	"os"

	"gpdf/doc"
)

func main() {
	path := "CV CONTOH.pdf"
	if len(os.Args) >= 2 {
		path = os.Args[1]
	}

	pdf, err := doc.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Open %q: %v\n", path, err)
		os.Exit(1)
	}
	defer pdf.Close()

	assets, err := pdf.GenerateCodeTo(os.Stdout, doc.CodeGenOptions{
		PackageName:        "main",
		FunctionName:       "BuildPDF",
		EmbedImages:        true,
		PreservePageSize:   true,
		PreserveTextStyles: true,
		PreservePositions:  true,
		PreserveTables:     true,
		InlineImageLimit:   128 * 1024,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "GenerateCodeTo: %v\n", err)
		os.Exit(1)
	}
	if len(assets) > 0 {
		fmt.Fprintf(os.Stderr, "\n// %d external assets were generated\n", len(assets))
	}
}
