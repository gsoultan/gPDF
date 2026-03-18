package doc

import (
	"io"

	"gpdf/writer"
)

// Saver writes a PDF document in various formats and encryption modes.
type Saver interface {
	Save(w io.Writer) error
	SaveWithPassword(w io.Writer, userPassword, ownerPassword string) error
	SaveWithAES256Password(w io.Writer, userPassword, ownerPassword string) error
	SaveLinearized(ws writer.WriteSeeker) error
}
