package doc

import (
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
		f.Close()
		return nil, err
	}
	size := info.Size()
	r := reader.NewPDFReader()
	var doc reader.Document
	if userPassword != "" {
		doc, err = r.ReadDocumentWithPassword(f, size, userPassword)
	} else {
		doc, err = r.ReadDocument(f, size)
	}
	if err != nil {
		f.Close()
		return nil, err
	}
	return file.NewDocument(f, doc), nil
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
