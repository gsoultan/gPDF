package doc

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"gpdf/model"
	"gpdf/writer"
)

func TestNewAndSave(t *testing.T) {
	buf := new(bytes.Buffer)
	doc, err := New().Title("Test").Author("gPDF").PageSize(595, 842).AddPage().AddPage().Build()
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Save(buf); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty PDF output")
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-2.0")) {
		t.Error("expected PDF 2.0 header")
	}
	_ = doc.Close()
}

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.pdf")

	// Build and save (1 catalog, 2 pages, 3 page dicts, 4 info = objects 1..6)
	built, err := New().Title("RoundTrip").Author("gPDF").PageSize(595, 842).AddPage().Build()
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := built.Save(f); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	_ = built.Close()

	// Open and read back
	opened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()

	catalog, err := opened.Catalog()
	if err != nil {
		t.Fatal(err)
	}
	if catalog == nil {
		t.Fatal("expected non-nil catalog")
	}

	pages, err := opened.Pages()
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 1 {
		t.Errorf("expected 1 page, got %d", len(pages))
	}
}

func TestOpenWithPassword_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.pdf")
	built, _ := New().Title("T").AddPage().Build()
	f, _ := os.Create(path)
	_ = built.Save(f)
	f.Close()
	built.Close()
	// OpenWithPassword with empty password should behave like Open
	opened, err := OpenWithPassword(path, "")
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	pages, err := opened.Pages()
	if err != nil || len(pages) != 1 {
		t.Errorf("pages: err=%v len=%d", err, len(pages))
	}
}

func TestDrawText(t *testing.T) {
	buf := new(bytes.Buffer)
	doc, err := New().
		Title("DrawText").
		Author("gPDF").
		PageSize(595, 842).
		AddPage().
		DrawText("Hello, PDF!", 100, 700, "Helvetica", 12).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Save(buf); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty PDF output")
	}
	// Page should have Contents and Resources with Font
	pages, _ := doc.Pages()
	if len(pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(pages))
	}
	contents := pages[0].Contents()
	if contents == nil {
		t.Error("page should have Contents")
	}
	res, ok := pages[0].Resources()
	if !ok || len(res) == 0 {
		t.Error("page should have Resources with Font")
	}
	_ = doc.Close()
}

func TestInfo(t *testing.T) {
	doc, err := New().
		Title("Info Test").
		Author("Alice").
		Subject("Testing").
		Keywords("pdf,gpdf").
		Creator("gPDF").
		Producer("gPDF").
		AddPage().
		Build()
	if err != nil {
		t.Fatal(err)
	}
	info, err := doc.Info()
	if err != nil {
		t.Fatal(err)
	}
	if info == nil {
		t.Fatal("expected non-nil Info")
	}
	if s, ok := info[model.Name("Title")].(model.String); !ok || s != "Info Test" {
		t.Errorf("Title: got %v", info[model.Name("Title")])
	}
	if s, ok := info[model.Name("Author")].(model.String); !ok || s != "Alice" {
		t.Errorf("Author: got %v", info[model.Name("Author")])
	}
	if s, ok := info[model.Name("Subject")].(model.String); !ok || s != "Testing" {
		t.Errorf("Subject: got %v", info[model.Name("Subject")])
	}
	if s, ok := info[model.Name("Keywords")].(model.String); !ok || s != "pdf,gpdf" {
		t.Errorf("Keywords: got %v", info[model.Name("Keywords")])
	}
	if s, ok := info[model.Name("Creator")].(model.String); !ok || s != "gPDF" {
		t.Errorf("Creator: got %v", info[model.Name("Creator")])
	}
	if s, ok := info[model.Name("Producer")].(model.String); !ok || s != "gPDF" {
		t.Errorf("Producer: got %v", info[model.Name("Producer")])
	}
	_ = doc.Close()
}

func TestMetadataStream(t *testing.T) {
	xmp := []byte(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?><x:xmpmeta xmlns:x="adobe:ns:meta/"/>`)
	doc, err := New().
		Title("XMP Test").
		AddPage().
		Metadata(xmp).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	got, err := doc.MetadataStream()
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil MetadataStream")
	}
	if string(got) != string(xmp) {
		t.Errorf("MetadataStream: got %q", got)
	}
	_ = doc.Close()
}

func TestDrawImage(t *testing.T) {
	// Minimal 2x2 DeviceGray 8bpc image (4 bytes)
	raw := []byte{0x00, 0x40, 0x80, 0xff}
	buf := new(bytes.Buffer)
	doc, err := New().
		Title("DrawImage").
		Author("gPDF").
		PageSize(595, 842).
		AddPage().
		DrawImage(100, 600, 72, 72, raw, 2, 2, 8, "DeviceGray").
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Save(buf); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty PDF output")
	}
	pdf := buf.Bytes()
	if !bytes.Contains(pdf, []byte("Image")) {
		t.Error("expected Image XObject in output")
	}
	if !bytes.Contains(pdf, []byte("Do")) {
		t.Error("expected Do operator in content stream")
	}
	pages, _ := doc.Pages()
	if len(pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(pages))
	}
	res, ok := pages[0].Resources()
	if !ok {
		t.Fatal("page should have Resources")
	}
	xobj, _ := res[model.Name("XObject")].(model.Dict)
	if xobj == nil || len(xobj) == 0 {
		t.Error("page Resources should have XObject with Im1")
	}
	_ = doc.Close()
}

// patchDoc implements writer.Document for incremental update tests.
type patchDoc struct {
	objects map[int]model.Object
	trailer model.Trailer
}

func (p *patchDoc) Trailer() model.Trailer { return p.trailer }
func (p *patchDoc) Resolve(ref model.Ref) (model.Object, error) {
	obj, ok := p.objects[ref.ObjectNumber]
	if !ok {
		return nil, nil
	}
	return obj, nil
}
func (p *patchDoc) ObjectNumbers() []int {
	var nums []int
	for n := range p.objects {
		nums = append(nums, n)
	}
	sort.Ints(nums)
	return nums
}

func TestIncrementalSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inc.pdf")

	// 1. Create initial PDF and save
	built, err := New().Title("Original").Author("Alice").PageSize(595, 842).AddPage().Build()
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := built.Save(f); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	built.Close()

	// 2. Open and get startXRef and root for patch
	opened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	startXRef := opened.StartXRefOffset()
	if startXRef <= 0 {
		t.Fatal("expected positive StartXRefOffset for opened file")
	}
	root := opened.Trailer().Root()
	if root == nil {
		opened.Close()
		t.Fatal("no trailer Root")
	}
	// Base has objects 1..5 (catalog, pages, page, minimal stream, info). Use 6 for new Info.
	newInfoNum := 6
	newInfoDict := model.Dict{
		model.Name("Title"):  model.String("Updated"),
		model.Name("Author"): model.String("Bob"),
	}
	patch := &patchDoc{
		objects: map[int]model.Object{newInfoNum: newInfoDict},
		trailer: model.Trailer{
			Dict: model.Dict{
				model.Name("Root"): model.Ref{ObjectNumber: root.ObjectNumber, Generation: root.Generation},
				model.Name("Info"): model.Ref{ObjectNumber: newInfoNum, Generation: 0},
			},
		},
	}
	opened.Close()

	// 3. Append incremental update
	f, err = os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		t.Fatal(err)
	}
	appendOffset, err := f.Seek(0, os.SEEK_END)
	if err != nil {
		f.Close()
		t.Fatal(err)
	}
	pw := writer.NewPDFWriter()
	if err := pw.WriteIncremental(f, appendOffset, startXRef, patch); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	// 4. Re-open and verify updated Info (reader uses last xref)
	opened2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer opened2.Close()
	info, err := opened2.Info()
	if err != nil {
		t.Fatal(err)
	}
	if info == nil {
		t.Fatal("expected non-nil Info after incremental update")
	}
	if s, ok := info[model.Name("Title")].(model.String); !ok || s != "Updated" {
		t.Errorf("Title after update: got %v", info[model.Name("Title")])
	}
	if s, ok := info[model.Name("Author")].(model.String); !ok || s != "Bob" {
		t.Errorf("Author after update: got %v", info[model.Name("Author")])
	}
}
