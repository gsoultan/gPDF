// Example merge remote demonstrates how to merge PDFs from URLs and simulated S3 sources.
//
// Usage:
//
//	go run ./cmd/example_merge_remote output.pdf
package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gsoultan/gpdf/doc"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/example_merge_remote output.pdf")
		os.Exit(1)
	}
	outputPath := os.Args[1]

	// Examples of different sources
	url1 := "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf"
	url2 := "https://www.africau.edu/images/default/sample.pdf"
	// Simulated S3/Memory source
	s3Content := createDummyPDF()

	fmt.Println("Fetching PDFs from URLs and memory...")

	// 1. Fetch from URL 1
	doc1, err := openFromURL(url1)
	if err != nil {
		fmt.Printf("Failed to open %s: %v\n", url1, err)
		os.Exit(1)
	}
	defer doc1.Close()

	// 2. Fetch from URL 2
	doc2, err := openFromURL(url2)
	if err != nil {
		fmt.Printf("Failed to open %s: %v\n", url2, err)
		os.Exit(1)
	}
	defer doc2.Close()

	// 3. Open from Memory/S3 (simulated)
	// Any io.ReaderAt (like those provided by S3 SDKs) can be used.
	reader := bytes.NewReader(s3Content)
	doc3, err := doc.OpenReader(reader, int64(len(s3Content)))
	if err != nil {
		fmt.Printf("Failed to open from memory: %v\n", err)
		os.Exit(1)
	}
	defer doc3.Close()

	fmt.Printf("Merging %d documents...\n", 3)

	// 4. Merge more than 2 PDFs
	merged, err := doc.Merge(doc1, doc2, doc3)
	if err != nil {
		fmt.Printf("Merge failed: %v\n", err)
		os.Exit(1)
	}
	defer merged.Close()

	// 5. Save the result
	f, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := merged.Save(f); err != nil {
		fmt.Printf("Failed to save merged PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully merged 3 PDFs into %s\n", outputPath)
}

func openFromURL(url string) (doc.Document, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	// For simplicity, we read everything into memory.
	// For very large PDFs, consider using os.CreateTemp.
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(data)
	return doc.OpenReader(reader, int64(len(data)))
}

// createDummyPDF creates a simple PDF in memory to simulate an S3 source.
func createDummyPDF() []byte {
	b := doc.New().
		Title("Memory PDF").
		AddPage().
		Flow(doc.FlowOptions{Margin: 72}).
		Paragraph("This PDF was loaded from memory (simulating S3).").
		End()

	d, err := b.Build()
	if err != nil {
		return nil
	}
	defer d.Close()

	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		return nil
	}
	return buf.Bytes()
}
