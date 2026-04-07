# gPDF

A library for reading, searching, and modifying PDF documents.

## Features

- **Read Content**: Extract text, images, and layout information from PDF files.
- **Search**: Find keywords across pages.
- **Modify**: Replace text content and save back to PDF.
- **Table Detection**: Identify table-like structures in PDF layouts.
- **Image Extraction**: Extract images from PDF pages with color space conversion support.
- **Merge & Split**: Combine multiple PDFs into one or split a PDF into several documents.

## Usage

### Reading a PDF

```go
pdf, err := doc.Open("example.pdf")
if err != nil {
    log.Fatal(err)
}
defer pdf.Close()

text, err := pdf.ReadContent()
if err == nil {
    fmt.Println(text)
}
```

### Searching and Replacing

```go
err := pdf.Replace("Old Text", "New Text")
if err != nil {
    log.Fatal(err)
}

f, _ := os.Create("modified.pdf")
err = pdf.Save(f)
f.Close()
```

### Table Detection

```go
tables, err := pdf.ReadTables()
if err != nil {
    log.Fatal(err)
}

for _, pageTables := range tables {
    for _, table := range pageTables {
        fmt.Printf("Table at (%f, %f)\n", table.X, table.Y)
    }
}
```
 
### Merging PDFs

```go
doc1, _ := doc.Open("file1.pdf")
doc2, _ := doc.Open("file2.pdf")
defer doc1.Close()
defer doc2.Close()

merged, err := doc.Merge(doc1, doc2)
if err != nil {
    log.Fatal(err)
}
defer merged.Close()

f, _ := os.Create("merged.pdf")
merged.Save(f)
f.Close()
```

### Splitting PDFs

```go
src, _ := doc.Open("input.pdf")
defer src.Close()

// Split into multiple documents, each containing at most 1 page
splitDocs, err := doc.SplitEvery(src, 1)
if err != nil {
    log.Fatal(err)
}

for i, splitDoc := range splitDocs {
    defer splitDoc.Close()
    f, _ := os.Create(fmt.Sprintf("split_%d.pdf", i))
    splitDoc.Save(f)
    f.Close()
}
```

### Extracting Pages

```go
src, _ := doc.Open("input.pdf")
defer src.Close()

// Extract pages 1-3 (inclusive-exclusive range: [1, 3))
extracted, err := doc.Extract(src, 1, 3)
if err != nil {
    log.Fatal(err)
}
defer extracted.Close()

f, _ := os.Create("extracted.pdf")
extracted.Save(f)
f.Close()
```

### Merging from Remote Sources (URL/S3)

You can open PDFs from any `io.ReaderAt` (like a memory buffer, URL, or S3 object):

```go
resp, _ := http.Get("https://example.com/file.pdf")
data, _ := io.ReadAll(resp.Body)
resp.Body.Close()

// Open from memory
reader := bytes.NewReader(data)
pdf, _ := doc.OpenReader(reader, int64(len(data)))
defer pdf.Close()

// Merge with other documents
merged, _ := doc.Merge(doc1, doc2, pdf)
```