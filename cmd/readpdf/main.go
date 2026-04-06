// readpdf demonstrates gPDF's reader capabilities:
//
//   - Open a PDF file and parse its structure
//   - Read document info (Title, Author, Subject, etc.)
//   - List pages with their MediaBox dimensions
//   - Extract text per page (using ToUnicode CMap decoding)
//   - Search for keywords across pages
//   - Read XMP metadata
//
// Usage:
//
//	go run ./cmd/readpdf <input.pdf>
//	go run ./cmd/readpdf -password secret <encrypted.pdf>
//	go run ./cmd/readpdf -text <input.pdf>
//	go run ./cmd/readpdf -search hello -search world <input.pdf>
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/reader"
)

// multiFlag allows -search to be specified multiple times.
type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ", ") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }

func main() {
	password := flag.String("password", "", "User password for encrypted PDFs (empty for unencrypted)")
	showText := flag.Bool("text", false, "Print extracted text for each page")
	verbose := flag.Bool("v", false, "Alias for -text (backwards compatible)")
	var searchTerms multiFlag
	flag.Var(&searchTerms, "search", "Keyword to search for (may be repeated)")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: go run ./cmd/readpdf [-password pw] [-text] [-search kw] <input.pdf>")
		fmt.Println("  Reads a PDF and prints document info, page list, per-page text, and search results.")
		os.Exit(1)
	}
	path := flag.Arg(0)

	doc, cleanup, err := openPDF(path, *password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	printDocumentInfo(doc)
	printPages(doc)

	if *showText || *verbose {
		printTextPerPage(doc)
	}

	if len(searchTerms) > 0 {
		printSearch(doc, searchTerms)
	}

	printMetadata(doc)
}

func openPDF(path, password string) (reader.Document, func(), error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}

	r := reader.NewPDFReader()
	var doc reader.Document
	if password != "" {
		doc, err = r.ReadDocumentWithPassword(f, info.Size(), password)
	} else {
		doc, err = r.ReadDocument(f, info.Size())
	}
	if err != nil {
		f.Close()
		return nil, nil, err
	}

	cleanup := func() { f.Close() }
	return doc, cleanup, nil
}

func printDocumentInfo(doc reader.Document) {
	fmt.Println("=== Document Info ===")

	infoDict, err := doc.Info()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  (error reading Info: %v)\n", err)
		return
	}
	if infoDict == nil {
		fmt.Println("  (no Info dictionary)")
	} else {
		for _, key := range []string{"Title", "Author", "Subject", "Keywords", "Creator", "Producer"} {
			if val, ok := infoDict[model.Name(key)]; ok {
				fmt.Printf("  %-10s %v\n", key+":", val)
			}
		}
	}

	trailer := doc.Trailer()
	fmt.Printf("  %-10s %d objects\n", "Size:", dictInt(trailer.Dict, "Size"))
	fmt.Println()
}

func printPages(doc reader.Document) {
	fmt.Println("=== Pages ===")

	pages, err := doc.Pages()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  (error reading pages: %v)\n", err)
		return
	}
	fmt.Printf("  Total: %d page(s)\n\n", len(pages))

	for i, page := range pages {
		mediaBox, ok := page.MediaBox()
		if ok && len(mediaBox) == 4 {
			w := floatFromObj(mediaBox[2]) - floatFromObj(mediaBox[0])
			h := floatFromObj(mediaBox[3]) - floatFromObj(mediaBox[1])
			fmt.Printf("  Page %d: %.0f x %.0f pt\n", i+1, w, h)
		} else {
			fmt.Printf("  Page %d: (no MediaBox)\n", i+1)
		}
	}
	fmt.Println()
}

func printTextPerPage(doc reader.Document) {
	fmt.Println("=== Extracted Text (per page) ===")

	perPage, err := doc.ContentPerPage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  (error extracting text: %v)\n", err)
		return
	}
	for i, text := range perPage {
		fmt.Printf("  -- Page %d --\n", i+1)
		if text == "" {
			fmt.Println("  (no text)")
		} else {
			fmt.Println(" ", text)
		}
	}
	fmt.Println()
}

func printSearch(doc reader.Document, keywords []string) {
	fmt.Println("=== Search Results ===")

	results, err := doc.Search(keywords...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  (error searching: %v)\n", err)
		return
	}
	for _, r := range results {
		if len(r.Pages) == 0 {
			fmt.Printf("  %q: not found\n", r.Keyword)
			continue
		}
		fmt.Printf("  %q: pages %v", r.Keyword, r.Pages)
		if len(r.Indices) > 0 {
			fmt.Printf(" indices %v", r.Indices)
		}
		fmt.Println()
	}
	fmt.Println()
}

func printMetadata(doc reader.Document) {
	meta, err := doc.MetadataStream()
	if err != nil || meta == nil {
		return
	}
	fmt.Println("=== XMP Metadata ===")
	preview := string(meta)
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}
	fmt.Println(preview)
	fmt.Println()
}

func dictInt(d model.Dict, key string) int64 {
	if v, ok := d[model.Name(key)].(model.Integer); ok {
		return int64(v)
	}
	return 0
}

func floatFromObj(obj model.Object) float64 {
	switch v := obj.(type) {
	case model.Integer:
		return float64(v)
	case model.Real:
		return float64(v)
	}
	return 0
}
