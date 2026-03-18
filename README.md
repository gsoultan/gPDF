# gPDF

## Code generation for large PDFs

For large input files, prefer streaming code generation to reduce peak memory usage.

```go
assets, err := pdf.GenerateCodeTo(out, doc.CodeGenOptions{
    PackageName:           "main",
    FunctionName:          "BuildPDF",
    EmbedImages:           true,
    PreservePageSize:      true,
    PreserveTextStyles:    true,
    PreservePositions:     true,
    PreserveTables:        true,
    InlineImageLimit:      128 * 1024,
    MaxDecodedStreamBytes: 256 * 1024 * 1024,
    MaxImageBytes:         32 * 1024 * 1024,
    MaxOpsPerPage:         1_000_000,
})
if err != nil {
    // handle error
}

// When an image exceeds InlineImageLimit, bytes are emitted in assets.
_ = assets
```

### Notes

- `GenerateCode` is still available and returns the full generated source as `GeneratedCode.GoSource`.
- `GenerateCodeTo` streams output to `io.Writer` and returns optional external assets for large images.
- Extraction guardrails (`MaxDecodedStreamBytes`, `MaxImageBytes`, `MaxOpsPerPage`) protect against oversized content during code generation.