package doc

import "gpdf/reader"

// ImageReader extracts image metadata and raw bytes from a PDF document.
type ImageReader interface {
	ReadImages() ([]reader.ImageInfo, error)
	ReadImagesPerPage() ([][]reader.ImageInfo, error)
}
