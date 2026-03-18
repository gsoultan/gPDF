package doc

// MetadataBuilder sets document-level metadata fields.
type MetadataBuilder interface {
	Title(s string) *DocumentBuilder
	Author(s string) *DocumentBuilder
	Subject(s string) *DocumentBuilder
	Keywords(s string) *DocumentBuilder
	Creator(s string) *DocumentBuilder
	Producer(s string) *DocumentBuilder
	Metadata(xmp []byte) *DocumentBuilder
	SetLanguage(lang string) *DocumentBuilder
}
