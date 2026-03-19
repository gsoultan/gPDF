package doc

import (
	bldrgfx "gpdf/doc/builder/graphics"
	bldrimg "gpdf/doc/builder/imgdraw"
	bldrtext "gpdf/doc/builder/text"
	"gpdf/doc/form"
	"gpdf/doc/layer"
	"gpdf/doc/metadata"
	"gpdf/doc/outline"
	"gpdf/doc/pagesize"
	taggedpkg "gpdf/doc/tagged"
	"gpdf/font"
	"gpdf/model"
)

// DocumentBuilder builds a new PDF via a fluent API. Call Build() to produce a Document.
// It is a thin facade that delegates drawing operations to sub-builders.
type DocumentBuilder struct {
	metadata metadata.Support
	outlines outline.Support
	pc       pageConfig
	fc       fontConfig

	useTagged bool
	tagging   taggedpkg.Support
	layers    layer.Support
	forms     form.Support

	noCompressContent bool
	iccProfile        *ICCProfile

	err error

	textDrawer     bldrtext.Drawer
	graphicsDrawer bldrgfx.Drawer
	imageDrawer    bldrimg.Drawer

	defaultStyle TextStyle
	warnings     []string
}

// Err returns the first error recorded during building. Check after Build() or before further fluent calls.
func (b *DocumentBuilder) Err() error { return b.err }

// Warnings returns all font-related warnings collected during the document building.
func (b *DocumentBuilder) Warnings() []string {
	return b.warnings
}

func (b *DocumentBuilder) logWarning(msg string) {
	b.warnings = append(b.warnings, msg)
}

func (b *DocumentBuilder) setErr(err error) {
	if b.err == nil {
		b.err = err
	}
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
func (b *DocumentBuilder) NoCompressContent() *DocumentBuilder {
	b.noCompressContent = true
	return b
}

// RegisterFont registers a parsed font for accurate text measurement and optional embedding.
func (b *DocumentBuilder) RegisterFont(f font.Font) *DocumentBuilder {
	if f == nil {
		return b
	}
	b.fc.registerFont(f)
	return b
}

// RegisterFallbackFont registers a font to be used when the primary font lacks a glyph.
func (b *DocumentBuilder) RegisterFallbackFont(f font.Font) *DocumentBuilder {
	if f == nil {
		return b
	}
	b.RegisterFont(f)
	b.fc.addFallback(f.PostScriptName())
	return b
}

// SetDefaultICCProfile sets an ICC color profile for the document.
func (b *DocumentBuilder) SetDefaultICCProfile(profile ICCProfile) *DocumentBuilder {
	b.iccProfile = &profile
	return b
}

// PageSize sets the default page size in points (width, height). Default is 595 x 842 (A4).
func (b *DocumentBuilder) PageSize(width, height float64) *DocumentBuilder {
	b.pc.pageSize = [2]float64{width, height}
	return b
}

// ApplyPageSize sets the default page size from a predefined Size preset.
func (b *DocumentBuilder) ApplyPageSize(sz pagesize.Size) *DocumentBuilder {
	return b.PageSize(sz.Width, sz.Height)
}

// SetDefaultStyle sets the default text style for the document.
func (b *DocumentBuilder) SetDefaultStyle(s TextStyle) *DocumentBuilder {
	b.defaultStyle = s
	return b
}

func (b *DocumentBuilder) getEffectiveStyle() TextStyle {
	s := b.defaultStyle
	if s.FontName == "" {
		s.FontName = "Helvetica"
	}
	if s.FontSize == 0 {
		s.FontSize = 12
	}
	// Color is already black by default (0,0,0) if not set.
	return s
}

// CurrentPage returns a Page object for the last added page.
func (b *DocumentBuilder) CurrentPage() *Page {
	if len(b.pc.pages) == 0 {
		return nil
	}
	return &Page{builder: b, pageIndex: len(b.pc.pages) - 1}
}

// AddPage adds a blank page with the current default page size.
func (b *DocumentBuilder) AddPage() *DocumentBuilder {
	w, h := b.pc.pageSize[0], b.pc.pageSize[1]
	if w == 0 {
		w, h = 595, 842
	}
	b.pc.addPage(w, h)
	return b
}

// Table starts a new table with the given number of columns.
func (b *DocumentBuilder) Table(cols int) ITableBuilder {
	// Defaults to standard margins/width if not overridden by .At() or .Width()
	return b.BeginTable(len(b.pc.pages)-1, 72, 0, 451, 0, cols)
}

// pageWidth returns the page width in points for the given index.
func (b *DocumentBuilder) pageWidth(pageIndex int) float64 {
	return b.pc.width(pageIndex)
}

// pageHeight returns the page height in points for the given index.
func (b *DocumentBuilder) pageHeight(pageIndex int) float64 {
	return b.pc.height(pageIndex)
}

// GetDefaultPageSize returns the current default page size for new pages.
func (b *DocumentBuilder) GetDefaultPageSize() (float64, float64) {
	return b.pc.pageSize[0], b.pc.pageSize[1]
}

// AddOutline adds a document outline (bookmark) linking to the given page.
func (b *DocumentBuilder) AddOutline(title string, pageIndex int) *DocumentBuilder {
	if title == "" || pageIndex < 0 {
		return b
	}
	b.outlines.Entries = append(b.outlines.Entries, outline.Entry{Title: title, PageIndex: pageIndex})
	return b
}

// AddOutlineURL adds an outline entry that opens the given URL.
func (b *DocumentBuilder) AddOutlineURL(title string, url string) *DocumentBuilder {
	if title == "" || url == "" {
		return b
	}
	b.outlines.Entries = append(b.outlines.Entries, outline.Entry{Title: title, PageIndex: -1, URL: url})
	return b
}

// AddOutlineToDest adds an outline entry that goes to the named destination.
func (b *DocumentBuilder) AddOutlineToDest(title string, destName string) *DocumentBuilder {
	if title == "" || destName == "" {
		return b
	}
	b.outlines.Entries = append(b.outlines.Entries, outline.Entry{Title: title, PageIndex: -1, DestName: destName})
	return b
}

// AddLinkToPage adds a link annotation on a page linking to another page.
func (b *DocumentBuilder) AddLinkToPage(pageIndex int, llx, lly, urx, ury float64, destPageIndex int) *DocumentBuilder {
	if pageIndex < 0 || destPageIndex < 0 {
		return b
	}
	b.outlines.LinkAnnots = append(b.outlines.LinkAnnots, outline.LinkAnnotation{
		PageIndex: pageIndex, Rect: [4]float64{llx, lly, urx, ury}, DestPage: destPageIndex,
	})
	return b
}

// AddLinkToDest adds a link annotation on a page linking to a named destination.
func (b *DocumentBuilder) AddLinkToDest(pageIndex int, llx, lly, urx, ury float64, destName string) *DocumentBuilder {
	if pageIndex < 0 || destName == "" {
		return b
	}
	b.outlines.LinkAnnots = append(b.outlines.LinkAnnots, outline.LinkAnnotation{
		PageIndex: pageIndex, Rect: [4]float64{llx, lly, urx, ury}, DestPage: -1, DestName: destName,
	})
	return b
}

// AddLinkToURL adds a link annotation on a page opening a URL.
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

// BeginSection starts a logical section for tagged content.
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
