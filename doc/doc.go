package doc

import (
	"io"
	"os"

	bldrgfx "github.com/gsoultan/gpdf/doc/builder/graphics"
	bldrimg "github.com/gsoultan/gpdf/doc/builder/imgdraw"
	bldrtext "github.com/gsoultan/gpdf/doc/builder/text"
	"github.com/gsoultan/gpdf/reader"
	"github.com/gsoultan/gpdf/reader/file"
)

// Open opens an existing PDF from path and returns a Document.
func Open(path string) (Document, error) {
	return OpenWithPassword(path, "")
}

// OpenWithPassword opens an existing PDF from path and decrypts it with the user password if encrypted.
func OpenWithPassword(path string, userPassword string) (Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	size := info.Size()
	return OpenReaderWithPassword(f, size, userPassword)
}

// OpenReader opens an existing PDF from a random-access reader and returns a Document.
func OpenReader(r io.ReaderAt, size int64) (Document, error) {
	return OpenReaderWithPassword(r, size, "")
}

// OpenReaderWithPassword opens an existing PDF from a random-access reader and decrypts it with the user password if encrypted.
func OpenReaderWithPassword(r io.ReaderAt, size int64, userPassword string) (Document, error) {
	rd := reader.NewPDFReader()
	var doc reader.Document
	var err error
	if userPassword != "" {
		doc, err = rd.ReadDocumentWithPassword(r, size, userPassword)
	} else {
		doc, err = rd.ReadDocument(r, size)
	}
	if err != nil {
		if closer, ok := r.(io.Closer); ok {
			_ = closer.Close()
		}
		return nil, err
	}
	closer, _ := r.(io.Closer)
	return file.NewDocument(closer, doc), nil
}

// New returns a new DocumentBuilder for constructing a PDF from scratch.
// Sub-builders are initialized for text, graphics, and image drawing.
func New() *DocumentBuilder {
	b := &DocumentBuilder{}
	b.fc.onWarning = b.logWarning
	b.textDrawer = bldrtext.NewDrawer(b)
	b.graphicsDrawer = bldrgfx.NewDrawer()
	b.imageDrawer = bldrimg.NewDrawer()
	return b
}
