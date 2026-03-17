package doc

import (
	"gpdf/doc/form"
	"gpdf/doc/layer"
	"gpdf/doc/metadata"
	"gpdf/doc/outline"
	taggedpkg "gpdf/doc/tagged"
	"gpdf/font"
	"gpdf/model"
)

// DocumentBuilder builds a new PDF via a fluent API. Call Build() to produce a Document.
type DocumentBuilder struct {
	metadata metadata.Support
	outlines outline.Support

	pageSize [2]float64
	pages    []pageSpec

	useTagged bool // if true, create Catalog /MarkInfo and /StructTreeRoot (tagged PDF)

	// tagging owns tagged-structure-related state (tables, blocks, lists, figures, sections).
	tagging taggedpkg.Support

	// layers owns optional content groups (OCGs) used for simple on/off layers.
	layers layer.Support

	// forms owns AcroForm-related state and field definitions.
	forms form.Support

	// compressContent controls FlateDecode compression of page content streams.
	// Default (false value) enables compression; set to true to disable.
	noCompressContent bool

	// fonts maps PostScript name to Font for accurate text measurement.
	fonts map[string]font.Font

	// embeddedFonts tracks rune usage for registered EmbeddableFont instances.
	embeddedFonts map[string]*embeddedFontUsage

	// iccProfile is an optional ICC color profile for the document's default output intent.
	iccProfile *ICCProfile
}

// Title sets the document title (Info /Title).
func (b *DocumentBuilder) Title(s string) *DocumentBuilder {
	b.metadata.Title = s
	return b
}

// Author sets the document author (Info /Author).
func (b *DocumentBuilder) Author(s string) *DocumentBuilder {
	b.metadata.Author = s
	return b
}

// Subject sets the document subject (Info /Subject).
func (b *DocumentBuilder) Subject(s string) *DocumentBuilder {
	b.metadata.Subject = s
	return b
}

// Keywords sets the document keywords (Info /Keywords).
func (b *DocumentBuilder) Keywords(s string) *DocumentBuilder {
	b.metadata.Keywords = s
	return b
}

// Creator sets the application that created the document (Info /Creator).
func (b *DocumentBuilder) Creator(s string) *DocumentBuilder {
	b.metadata.Creator = s
	return b
}

// Producer sets the application that produced the PDF (Info /Producer).
func (b *DocumentBuilder) Producer(s string) *DocumentBuilder {
	b.metadata.Producer = s
	return b
}

// Metadata sets the XMP metadata stream for Catalog /Metadata (raw XML bytes).
func (b *DocumentBuilder) Metadata(xmp []byte) *DocumentBuilder {
	b.metadata.MetadataXMP = xmp
	return b
}

// SetLanguage sets the default document language (Catalog /Lang), e.g. "en-US".
func (b *DocumentBuilder) SetLanguage(lang string) *DocumentBuilder {
	b.metadata.Lang = lang
	return b
}

// NoCompressContent disables FlateDecode compression of page content streams.
// Useful for debugging or when raw content must be inspectable.
func (b *DocumentBuilder) NoCompressContent() *DocumentBuilder {
	b.noCompressContent = true
	return b
}

// RegisterFont registers a parsed font for accurate text measurement and optional embedding.
// The font is keyed by its PostScript name (e.g. "Helvetica-Bold", "ArialMT").
// If the font implements font.EmbeddableFont, it will be embedded in the PDF with
// proper CID encoding, subsetting, and ToUnicode CMap for full Unicode support.
func (b *DocumentBuilder) RegisterFont(f font.Font) *DocumentBuilder {
	if f == nil {
		return b
	}
	if b.fonts == nil {
		b.fonts = make(map[string]font.Font)
	}
	b.fonts[f.PostScriptName()] = f
	if ef, ok := f.(font.EmbeddableFont); ok {
		if b.embeddedFonts == nil {
			b.embeddedFonts = make(map[string]*embeddedFontUsage)
		}
		b.embeddedFonts[ef.PostScriptName()] = newEmbeddedFontUsage(ef)
	}
	return b
}

// SetDefaultICCProfile sets an ICC color profile for the document.
// The profile is embedded as an ICCBased color space and referenced via an output intent
// in the catalog, which is required for PDF/A conformance.
func (b *DocumentBuilder) SetDefaultICCProfile(profile ICCProfile) *DocumentBuilder {
	b.iccProfile = &profile
	return b
}

// PageSize sets the default page size in points (width, height). Default is 595 x 842 (A4).
func (b *DocumentBuilder) PageSize(width, height float64) *DocumentBuilder {
	b.pageSize = [2]float64{width, height}
	return b
}

// AddPage adds a blank page with the current default page size.
func (b *DocumentBuilder) AddPage() *DocumentBuilder {
	w, h := b.pageSize[0], b.pageSize[1]
	if w == 0 {
		w, h = 595, 842
	}
	dict := model.Dict{
		model.Name("Type"):     model.Name("Page"),
		model.Name("MediaBox"): model.Array{model.Integer(0), model.Integer(0), model.Real(w), model.Real(h)},
	}
	b.pages = append(b.pages, pageSpec{dict: dict})
	return b
}

// pageHeight returns the page height in points for the given index.
func (b *DocumentBuilder) pageHeight(pageIndex int) float64 {
	if pageIndex < 0 || pageIndex >= len(b.pages) {
		return 0
	}
	spec := b.pages[pageIndex]
	if mb, ok := spec.dict[model.Name("MediaBox")].(model.Array); ok && len(mb) == 4 {
		if h, ok := mb[3].(model.Real); ok {
			return float64(h)
		}
		if h, ok := mb[3].(model.Integer); ok {
			return float64(h)
		}
	}
	if b.pageSize[1] > 0 {
		return b.pageSize[1]
	}
	return 842
}

// AddOutline adds a document outline (bookmark) with the given title linking to the page at 0-based pageIndex.
func (b *DocumentBuilder) AddOutline(title string, pageIndex int) *DocumentBuilder {
	if title == "" || pageIndex < 0 {
		return b
	}
	b.outlines.Entries = append(b.outlines.Entries, outline.Entry{Title: title, PageIndex: pageIndex})
	return b
}

// AddOutlineURL adds an outline entry that opens the given URL when clicked (URI action).
func (b *DocumentBuilder) AddOutlineURL(title string, url string) *DocumentBuilder {
	if title == "" || url == "" {
		return b
	}
	b.outlines.Entries = append(b.outlines.Entries, outline.Entry{Title: title, PageIndex: -1, URL: url})
	return b
}

// AddOutlineToDest adds an outline entry that goes to the named destination when clicked (GoTo action).
func (b *DocumentBuilder) AddOutlineToDest(title string, destName string) *DocumentBuilder {
	if title == "" || destName == "" {
		return b
	}
	b.outlines.Entries = append(b.outlines.Entries, outline.Entry{Title: title, PageIndex: -1, DestName: destName})
	return b
}

// AddLinkToPage adds a link annotation on the page at pageIndex (0-based) with the given rectangle, linking to destPageIndex.
func (b *DocumentBuilder) AddLinkToPage(pageIndex int, llx, lly, urx, ury float64, destPageIndex int) *DocumentBuilder {
	if pageIndex < 0 || destPageIndex < 0 {
		return b
	}
	b.outlines.LinkAnnots = append(b.outlines.LinkAnnots, outline.LinkAnnotation{
		PageIndex: pageIndex, Rect: [4]float64{llx, lly, urx, ury}, DestPage: destPageIndex,
	})
	return b
}

// AddLinkToDest adds a link annotation on the page at pageIndex with the given rect, linking to the named destination.
func (b *DocumentBuilder) AddLinkToDest(pageIndex int, llx, lly, urx, ury float64, destName string) *DocumentBuilder {
	if pageIndex < 0 || destName == "" {
		return b
	}
	b.outlines.LinkAnnots = append(b.outlines.LinkAnnots, outline.LinkAnnotation{
		PageIndex: pageIndex, Rect: [4]float64{llx, lly, urx, ury}, DestPage: -1, DestName: destName,
	})
	return b
}

// AddLinkToURL adds a link annotation on the page at pageIndex with the given rect, opening the URL when clicked.
func (b *DocumentBuilder) AddLinkToURL(pageIndex int, llx, lly, urx, ury float64, url string) *DocumentBuilder {
	if pageIndex < 0 || url == "" {
		return b
	}
	b.outlines.LinkAnnots = append(b.outlines.LinkAnnots, outline.LinkAnnotation{
		PageIndex: pageIndex, Rect: [4]float64{llx, lly, urx, ury}, DestPage: -1, URL: url,
	})
	return b
}

// SetTagged marks the document as tagged PDF.
func (b *DocumentBuilder) SetTagged() *DocumentBuilder {
	b.useTagged = true
	return b
}

// SetAcroForm ensures the document has an AcroForm dictionary.
func (b *DocumentBuilder) SetAcroForm() *DocumentBuilder {
	b.forms.UseAcroForm = true
	return b
}

// SetAcroFormSigFlags sets /SigFlags on the AcroForm dictionary.
// Implies SetAcroForm(). Call before Build().
func (b *DocumentBuilder) SetAcroFormSigFlags(flags int) *DocumentBuilder {
	b.forms.UseAcroForm = true
	b.forms.SigFlags = flags
	return b
}

// AddNamedDest registers a named destination for the page at 0-based pageIndex.
func (b *DocumentBuilder) AddNamedDest(name string, pageIndex int) *DocumentBuilder {
	if name == "" || pageIndex < 0 {
		return b
	}
	if b.outlines.NamedDests == nil {
		b.outlines.NamedDests = make(map[string]int)
	}
	b.outlines.NamedDests[name] = pageIndex
	return b
}

// BeginSection starts a logical section; all following tagged content
// until EndSection is called will be grouped under one Sect structure element.
func (b *DocumentBuilder) BeginSection() *DocumentBuilder {
	b.useTagged = true
	b.tagging.Sections = append(b.tagging.Sections, taggedpkg.Section{})
	b.tagging.CurrentSection = len(b.tagging.Sections) - 1
	return b
}

// EndSection ends the current section started by BeginSection.
func (b *DocumentBuilder) EndSection() *DocumentBuilder {
	b.tagging.CurrentSection = -1
	return b
}

func copyPageDict(d model.Dict) model.Dict {
	out := make(model.Dict, len(d))
	for k, v := range d {
		out[k] = v
	}
	return out
}
