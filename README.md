# gPDF

A library for reading, searching, and modifying PDF documents.

## Features

- **Read Content**: Extract text, images, and layout information from PDF files.
- **Search**: Find keywords across pages.
- **Modify**: Replace text content and save back to PDF.
- **Table Detection**: Identify table-like structures in PDF layouts.
- **Image Extraction**: Extract images from PDF pages with color space conversion support.

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

err = pdf.SaveToFile("modified.pdf")
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