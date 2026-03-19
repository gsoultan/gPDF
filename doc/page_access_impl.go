package doc

import (
	"gpdf/doc/builder"
	"gpdf/font"
)

var _ builder.PageAccess = (*DocumentBuilder)(nil)

func (b *DocumentBuilder) ValidPageIndex(idx int) bool {
	return b.pc.validPageIndex(idx)
}

func (b *DocumentBuilder) PageCount() int {
	return len(b.pc.pages)
}

func (b *DocumentBuilder) PageHeight(pageIndex int) float64 {
	return b.pc.height(pageIndex)
}

func (b *DocumentBuilder) PageWidth(pageIndex int) float64 {
	return b.pc.width(pageIndex)
}

func (b *DocumentBuilder) PageAt(pageIndex int) *builder.Page {
	if !b.pc.validPageIndex(pageIndex) {
		return nil
	}
	return &b.pc.pages[pageIndex]
}

func (b *DocumentBuilder) NextMCID(pageIndex int) int {
	if !b.pc.validPageIndex(pageIndex) {
		return 0
	}
	ps := &b.pc.pages[pageIndex]
	mcid := ps.NextMCID
	ps.NextMCID++
	return mcid
}

func (b *DocumentBuilder) AppendPage() {
	w, h := b.pc.pageSize[0], b.pc.pageSize[1]
	if w == 0 {
		w, h = 595, 842
	}
	b.pc.addPage(w, h)
}

func (b *DocumentBuilder) FontByName(name string) font.Font {
	if b.fc.fonts == nil {
		return nil
	}
	return b.fc.fonts[name]
}
