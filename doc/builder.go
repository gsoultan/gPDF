package doc

import (
	"fmt"

	"gpdf/content"
	"gpdf/model"
)

// textRun describes one text draw on a page (simple PDF text; no Unicode/CMap).
type textRun struct {
	Text     string
	X, Y     float64
	FontName string
	FontSize float64
}

// imageRun describes one image draw on a page (Image XObject placed via Do).
type imageRun struct {
	X, Y                float64 // position in points (lower-left)
	WidthPt, HeightPt   float64 // display size in points
	Raw                  []byte // raw image stream bytes (decoded)
	WidthPx, HeightPx   int    // dimensions in samples
	BitsPerComponent     int    // 1, 2, 4, 8, 12, 16
	ColorSpace           string // e.g. DeviceGray, DeviceRGB, DeviceCMYK
}

// pageSpec holds a page dict and optional content (text and image runs).
type pageSpec struct {
	dict      model.Dict
	textRuns  []textRun
	imageRuns []imageRun
}

// outlineEntry describes one document outline (bookmark): title and either page, URL, or named dest.
type outlineEntry struct {
	Title     string
	PageIndex int    // 0-based page for /Dest; -1 if using URL or DestName
	URL       string // if set, use /A with URI action
	DestName  string // if set, use /A with GoTo and named destination
}

// DocumentBuilder builds a new PDF via a fluent API. Call Build() to produce a Document.
type DocumentBuilder struct {
	title       string
	author      string
	subject     string
	keywords    string
	creator     string
	producer    string
	metadataXMP []byte   // optional XMP stream for Catalog /Metadata
	pageSize    [2]float64
	pages       []pageSpec
	outlines    []outlineEntry
	namedDests  map[string]int // name -> 0-based page index for Catalog /Dests
}

// Title sets the document title (Info /Title).
func (b *DocumentBuilder) Title(s string) *DocumentBuilder {
	b.title = s
	return b
}

// Author sets the document author (Info /Author).
func (b *DocumentBuilder) Author(s string) *DocumentBuilder {
	b.author = s
	return b
}

// Subject sets the document subject (Info /Subject).
func (b *DocumentBuilder) Subject(s string) *DocumentBuilder {
	b.subject = s
	return b
}

// Keywords sets the document keywords (Info /Keywords).
func (b *DocumentBuilder) Keywords(s string) *DocumentBuilder {
	b.keywords = s
	return b
}

// Creator sets the application that created the document (Info /Creator).
func (b *DocumentBuilder) Creator(s string) *DocumentBuilder {
	b.creator = s
	return b
}

// Producer sets the application that produced the PDF (Info /Producer).
func (b *DocumentBuilder) Producer(s string) *DocumentBuilder {
	b.producer = s
	return b
}

// Metadata sets the XMP metadata stream for Catalog /Metadata (raw XML bytes).
func (b *DocumentBuilder) Metadata(xmp []byte) *DocumentBuilder {
	b.metadataXMP = xmp
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

// DrawText queues text to be drawn on the last added page at (x, y) using the given font and size.
// FontName should be a standard PDF base font (e.g. Helvetica, Times-Roman). Call after AddPage().
func (b *DocumentBuilder) DrawText(text string, x, y float64, fontName string, fontSize float64) *DocumentBuilder {
	if len(b.pages) == 0 {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	idx := len(b.pages) - 1
	b.pages[idx].textRuns = append(b.pages[idx].textRuns, textRun{
		Text: text, X: x, Y: y, FontName: fontName, FontSize: fontSize,
	})
	return b
}

// AddOutline adds a document outline (bookmark) with the given title linking to the page at 0-based pageIndex.
// Call after AddPage(); pageIndex must be in range [0, len(pages)-1]. Multiple calls add multiple bookmarks in order.
func (b *DocumentBuilder) AddOutline(title string, pageIndex int) *DocumentBuilder {
	if title == "" || pageIndex < 0 {
		return b
	}
	b.outlines = append(b.outlines, outlineEntry{Title: title, PageIndex: pageIndex})
	return b
}

// AddOutlineURL adds an outline entry that opens the given URL when clicked (URI action).
func (b *DocumentBuilder) AddOutlineURL(title string, url string) *DocumentBuilder {
	if title == "" || url == "" {
		return b
	}
	b.outlines = append(b.outlines, outlineEntry{Title: title, PageIndex: -1, URL: url})
	return b
}

// AddOutlineToDest adds an outline entry that goes to the named destination when clicked (GoTo action).
// The name must be added with AddNamedDest before Build().
func (b *DocumentBuilder) AddOutlineToDest(title string, destName string) *DocumentBuilder {
	if title == "" || destName == "" {
		return b
	}
	b.outlines = append(b.outlines, outlineEntry{Title: title, PageIndex: -1, DestName: destName})
	return b
}

// AddNamedDest registers a named destination for the page at 0-based pageIndex.
// Other outlines can link to it via AddOutlineToDest(title, name). Catalog /Dests is written when at least one name is added.
func (b *DocumentBuilder) AddNamedDest(name string, pageIndex int) *DocumentBuilder {
	if name == "" || pageIndex < 0 {
		return b
	}
	if b.namedDests == nil {
		b.namedDests = make(map[string]int)
	}
	b.namedDests[name] = pageIndex
	return b
}

// DrawImage queues an image to be drawn on the last added page at (x, y) with display size (widthPt, heightPt).
// Raw is the decoded image stream; widthPx/heightPx and bitsPerComponent/colorSpace must match.
// colorSpace should be DeviceGray, DeviceRGB, or DeviceCMYK. Call after AddPage().
func (b *DocumentBuilder) DrawImage(x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string) *DocumentBuilder {
	if len(b.pages) == 0 {
		return b
	}
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	if bitsPerComponent <= 0 {
		bitsPerComponent = 8
	}
	idx := len(b.pages) - 1
	b.pages[idx].imageRuns = append(b.pages[idx].imageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: bitsPerComponent, ColorSpace: colorSpace,
	})
	return b
}

// Build returns a Document that can be saved. The document is in-memory.
func (b *DocumentBuilder) Build() (Document, error) {
	// Build object graph: catalog, pages tree, page dicts, info.
	// Object numbers: 1 catalog, 2 pages, 3..N pages, N+1 info (optional)
	// Simplified: 1 = catalog, 2 = pages (with Kids refs), 3,4,... = page dicts, last = info
	var pageRefs model.Array
	objs := make(map[int]model.Object)
	nextNum := 1
	// Catalog
	catalogNum := nextNum
	nextNum++
	// Pages tree
	pagesNum := nextNum
	nextNum++
	pageNums := make([]int, 0, len(b.pages))
	for range b.pages {
		pageNum := nextNum
		nextNum++
		pageNums = append(pageNums, pageNum)
		pageRefs = append(pageRefs, model.Ref{ObjectNumber: pageNum, Generation: 0})
	}
	// Shared minimal content stream for pages with no text
	minimalStreamNum := nextNum
	nextNum++
	minimalContent := []byte("n\n")
	objs[minimalStreamNum] = &model.Stream{
		Dict:    model.Dict{model.Name("Length"): model.Integer(len(minimalContent))},
		Content: minimalContent,
	}
	for idx, pageNum := range pageNums {
		spec := b.pages[idx]
		pageDict := copyPageDict(spec.dict)
		hasContent := len(spec.textRuns) > 0 || len(spec.imageRuns) > 0
		if !hasContent {
			pageDict[model.Name("Contents")] = model.Ref{ObjectNumber: minimalStreamNum, Generation: 0}
			objs[pageNum] = pageDict
			continue
		}
		contentStreamNum := nextNum
		nextNum++
		imageStreamNums := make([]int, len(spec.imageRuns))
		for i := range spec.imageRuns {
			imageStreamNums[i] = nextNum
			nextNum++
		}
		contentBytes, resources, err := b.buildPageContentWithImages(spec.textRuns, spec.imageRuns, imageStreamNums)
		if err != nil {
			return nil, fmt.Errorf("page %d: %w", idx+1, err)
		}
		pageDict[model.Name("Contents")] = model.Ref{ObjectNumber: contentStreamNum, Generation: 0}
		pageDict[model.Name("Resources")] = resources
		objs[contentStreamNum] = &model.Stream{
			Dict:    model.Dict{model.Name("Length"): model.Integer(len(contentBytes))},
			Content: contentBytes,
		}
		for i, im := range spec.imageRuns {
			objs[imageStreamNums[i]] = b.imageXObjectStream(im)
		}
		objs[pageNum] = pageDict
	}
	// Pages tree dict
	pagesDict := model.Dict{
		model.Name("Type"):  model.Name("Pages"),
		model.Name("Kids"):  pageRefs,
		model.Name("Count"): model.Integer(len(b.pages)),
	}
	objs[pagesNum] = pagesDict
	// Catalog dict
	catalogDict := model.Dict{
		model.Name("Type"):  model.Name("Catalog"),
		model.Name("Pages"): model.Ref{ObjectNumber: pagesNum, Generation: 0},
	}
	// Named destinations: Catalog /Dests
	if len(b.namedDests) > 0 && len(pageNums) > 0 {
		destsDict := model.Dict{}
		for name, idx := range b.namedDests {
			if idx >= 0 && idx < len(pageNums) {
				pageRef := model.Ref{ObjectNumber: pageNums[idx], Generation: 0}
				destsDict[model.Name(name)] = model.Array{pageRef, model.Name("Fit")}
			}
		}
		if len(destsDict) > 0 {
			destsNum := nextNum
			nextNum++
			objs[destsNum] = destsDict
			catalogDict[model.Name("Dests")] = model.Ref{ObjectNumber: destsNum, Generation: 0}
		}
	}
	// Outlines (bookmarks): page /Dest, or /A (URI / GoTo named)
	if len(b.outlines) > 0 && len(pageNums) > 0 {
		var validEntries []outlineEntry
		for _, e := range b.outlines {
			hasPage := e.PageIndex >= 0 && e.PageIndex < len(pageNums)
			if hasPage || e.URL != "" || e.DestName != "" {
				validEntries = append(validEntries, e)
			}
		}
		if len(validEntries) > 0 {
			outlineRootNum := nextNum
			nextNum++
			itemNums := make([]int, len(validEntries))
			for i := range validEntries {
				itemNums[i] = nextNum
				nextNum++
			}
			rootDict := model.Dict{
				model.Name("Type"):  model.Name("Outlines"),
				model.Name("First"): model.Ref{ObjectNumber: itemNums[0], Generation: 0},
				model.Name("Last"):  model.Ref{ObjectNumber: itemNums[len(itemNums)-1], Generation: 0},
				model.Name("Count"): model.Integer(int64(len(itemNums))),
			}
			objs[outlineRootNum] = rootDict
			for i, e := range validEntries {
				itemDict := model.Dict{
					model.Name("Title"):  model.String(e.Title),
					model.Name("Parent"): model.Ref{ObjectNumber: outlineRootNum, Generation: 0},
				}
				if e.URL != "" {
					itemDict[model.Name("A")] = model.Dict{
						model.Name("S"):   model.Name("URI"),
						model.Name("URI"): model.String(e.URL),
					}
				} else if e.DestName != "" {
					itemDict[model.Name("A")] = model.Dict{
						model.Name("S"): model.Name("GoTo"),
						model.Name("D"): model.Name(e.DestName),
					}
				} else {
					pageRef := model.Ref{ObjectNumber: pageNums[e.PageIndex], Generation: 0}
					itemDict[model.Name("Dest")] = model.Array{pageRef, model.Name("Fit")}
				}
				if i > 0 {
					itemDict[model.Name("Prev")] = model.Ref{ObjectNumber: itemNums[i-1], Generation: 0}
				}
				if i < len(validEntries)-1 {
					itemDict[model.Name("Next")] = model.Ref{ObjectNumber: itemNums[i+1], Generation: 0}
				}
				objs[itemNums[i]] = itemDict
			}
			catalogDict[model.Name("Outlines")] = model.Ref{ObjectNumber: outlineRootNum, Generation: 0}
		}
	}
	objs[catalogNum] = catalogDict
	infoNum := nextNum
	nextNum++
	infoDict := b.buildInfoDict()
	objs[infoNum] = infoDict
	// Catalog: add /Metadata ref if XMP provided
	if len(b.metadataXMP) > 0 {
		metaNum := nextNum
		nextNum++
		objs[metaNum] = &model.Stream{
			Dict: model.Dict{
				model.Name("Type"):   model.Name("Metadata"),
				model.Name("Subtype"): model.Name("XML"),
				model.Name("Length"): model.Integer(int64(len(b.metadataXMP))),
			},
			Content: b.metadataXMP,
		}
		catalogDict[model.Name("Metadata")] = model.Ref{ObjectNumber: metaNum, Generation: 0}
	}
	trailer := model.Trailer{
		Dict: model.Dict{
			model.Name("Root"): model.Ref{ObjectNumber: catalogNum, Generation: 0},
			model.Name("Size"): model.Integer(nextNum),
			model.Name("Info"): model.Ref{ObjectNumber: infoNum, Generation: 0},
		},
	}
	return &builtDocument{
		trailer: trailer,
		objects: objs,
		size:    nextNum,
	}, nil
}

func (b *DocumentBuilder) buildInfoDict() model.Dict {
	info := model.Dict{}
	if b.title != "" {
		info[model.Name("Title")] = model.String(b.title)
	}
	if b.author != "" {
		info[model.Name("Author")] = model.String(b.author)
	}
	if b.subject != "" {
		info[model.Name("Subject")] = model.String(b.subject)
	}
	if b.keywords != "" {
		info[model.Name("Keywords")] = model.String(b.keywords)
	}
	if b.creator != "" {
		info[model.Name("Creator")] = model.String(b.creator)
	}
	if b.producer != "" {
		info[model.Name("Producer")] = model.String(b.producer)
	}
	return info
}

func copyPageDict(d model.Dict) model.Dict {
	out := make(model.Dict, len(d))
	for k, v := range d {
		out[k] = v
	}
	return out
}

// buildPageContentWithImages returns content stream bytes and /Resources (Font + XObject) for text and image runs.
func (b *DocumentBuilder) buildPageContentWithImages(textRuns []textRun, imageRuns []imageRun, imageStreamNums []int) ([]byte, model.Dict, error) {
	if len(textRuns) == 0 && len(imageRuns) == 0 {
		return nil, nil, fmt.Errorf("no content")
	}
	var ops []content.Op
	var fontName string
	for _, r := range textRuns {
		if fontName == "" {
			fontName = r.FontName
		}
		if fontName == "" {
			fontName = "Helvetica"
		}
		size := r.FontSize
		if size <= 0 {
			size = 12
		}
		ops = append(ops,
			content.Op{Name: "BT", Args: nil},
			content.Op{Name: "Tf", Args: []model.Object{model.Name("F1"), model.Real(size)}},
			content.Op{Name: "Td", Args: []model.Object{model.Real(r.X), model.Real(r.Y)}},
			content.Op{Name: "Tj", Args: []model.Object{model.String(r.Text)}},
			content.Op{Name: "ET", Args: nil},
		)
	}
	for i, im := range imageRuns {
		// q width 0 0 height x y cm /ImN Do Q
		imName := model.Name("Im" + fmt.Sprintf("%d", i+1))
		w, h := im.WidthPt, im.HeightPt
		if w <= 0 {
			w = float64(im.WidthPx)
		}
		if h <= 0 {
			h = float64(im.HeightPx)
		}
		ops = append(ops,
			content.Op{Name: "q", Args: nil},
			content.Op{Name: "cm", Args: []model.Object{
				model.Real(w), model.Real(0), model.Real(0), model.Real(h),
				model.Real(im.X), model.Real(im.Y),
			}},
			content.Op{Name: "Do", Args: []model.Object{imName}},
			content.Op{Name: "Q", Args: nil},
		)
	}
	contentBytes, err := content.EncodeBytes(ops)
	if err != nil {
		return nil, nil, err
	}
	resources := model.Dict{}
	if len(textRuns) > 0 {
		if fontName == "" {
			fontName = "Helvetica"
		}
		resources[model.Name("Font")] = model.Dict{
			model.Name("F1"): model.Dict{
				model.Name("Type"):     model.Name("Font"),
				model.Name("BaseFont"): model.Name(fontName),
			},
		}
	}
	if len(imageRuns) > 0 {
		xobj := make(model.Dict)
		for i, num := range imageStreamNums {
			name := model.Name("Im" + fmt.Sprintf("%d", i+1))
			xobj[name] = model.Ref{ObjectNumber: num, Generation: 0}
		}
		resources[model.Name("XObject")] = xobj
	}
	return contentBytes, resources, nil
}

func (b *DocumentBuilder) imageXObjectStream(im imageRun) *model.Stream {
	dict := model.Dict{
		model.Name("Type"):             model.Name("XObject"),
		model.Name("Subtype"):         model.Name("Image"),
		model.Name("Width"):           model.Integer(int64(im.WidthPx)),
		model.Name("Height"):          model.Integer(int64(im.HeightPx)),
		model.Name("BitsPerComponent"): model.Integer(int64(im.BitsPerComponent)),
		model.Name("ColorSpace"):      model.Name(im.ColorSpace),
		model.Name("Length"):          model.Integer(int64(len(im.Raw))),
	}
	return &model.Stream{Dict: dict, Content: im.Raw}
}
