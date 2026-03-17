// readpdf demonstrates gPDF's reader capabilities:
//
//   - Open a PDF file and parse its structure
//   - Read document info (Title, Author, Subject, etc.)
//   - List pages with their MediaBox dimensions
//   - Resolve indirect objects and inspect dictionaries
//   - Parse content stream operators (text extraction)
//   - Read XMP metadata
//
// Usage:
//
//	go run ./cmd/readpdf <input.pdf>
//	go run ./cmd/readpdf -password secret <encrypted.pdf>
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	contentimpl "gpdf/content/impl"
	"gpdf/model"
	"gpdf/reader"
)

func main() {
	password := flag.String("password", "", "User password for encrypted PDFs (empty for unencrypted)")
	verbose := flag.Bool("v", false, "Verbose: print content stream operators per page")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: go run ./cmd/readpdf [-password pw] [-v] <input.pdf>")
		fmt.Println("  Reads a PDF and prints document info, page list, and optionally content operators.")
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
	printPages(doc, *verbose)
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
		return
	}
	for _, key := range []string{"Title", "Author", "Subject", "Keywords", "Creator", "Producer"} {
		if val, ok := infoDict[model.Name(key)]; ok {
			fmt.Printf("  %-10s %v\n", key+":", val)
		}
	}

	trailer := doc.Trailer()
	fmt.Printf("  %-10s %d objects\n", "Size:", dictInt(trailer.Dict, "Size"))
	fmt.Println()
}

func printPages(doc reader.Document, verbose bool) {
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

		if verbose {
			printPageContent(doc, page, i+1)
		}
	}
	fmt.Println()
}

func printPageContent(doc reader.Document, page model.Page, pageNum int) {
	contentsObj := page.Contents()
	if contentsObj == nil {
		return
	}

	var streamData []byte
	switch v := contentsObj.(type) {
	case model.Ref:
		obj, err := doc.Resolve(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "    (error resolving contents: %v)\n", err)
			return
		}
		s, ok := obj.(*model.Stream)
		if !ok || s == nil {
			return
		}
		streamData = s.Content
	case model.Array:
		for _, item := range v {
			ref, ok := item.(model.Ref)
			if !ok {
				continue
			}
			obj, err := doc.Resolve(ref)
			if err != nil {
				continue
			}
			s, ok := obj.(*model.Stream)
			if !ok || s == nil {
				continue
			}
			streamData = append(streamData, s.Content...)
			streamData = append(streamData, '\n')
		}
	}

	if len(streamData) == 0 {
		return
	}

	parser := contentimpl.NewStreamParser()
	ops, err := parser.Parse(streamData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "    (error parsing content stream: %v)\n", err)
		return
	}

	var textParts []string
	inText := false
	for _, op := range ops {
		switch op.Name {
		case "BT":
			inText = true
		case "ET":
			inText = false
		case "Tj":
			if inText && len(op.Args) > 0 {
				if s, ok := op.Args[0].(model.String); ok {
					textParts = append(textParts, string(s))
				}
			}
		case "TJ":
			if inText && len(op.Args) > 0 {
				if arr, ok := op.Args[0].(model.Array); ok {
					for _, elem := range arr {
						if s, ok := elem.(model.String); ok {
							textParts = append(textParts, string(s))
						}
					}
				}
			}
		}
	}
	if len(textParts) > 0 {
		fmt.Printf("    Text: %s\n", strings.Join(textParts, ""))
	}
	fmt.Printf("    Operators: %d\n", len(ops))
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
